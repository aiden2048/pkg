package frame

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/rcrowley/go-metrics"
)

// CustomEtcdV3RegisterPlugin implements etcd registry with Lease+KeepAlive.
// Unlike the official EtcdV3RegisterPlugin which does a PUT every UpdateInterval
// (generating a new Etcd Revision each time), this implementation:
//  1. Grants a lease with a TTL on startup
//  2. Puts the service key once, bound to the lease
//  3. Only calls KeepAlive to renew the lease TTL — no new Revision is created
//  4. On lease expiry/reconnect, re-grants a new lease and re-puts the key
type CustomEtcdV3RegisterPlugin struct {
	ServiceAddress string
	EtcdServers    []string
	BasePath       string
	Metrics        metrics.Registry
	Services       []string
	UpdateInterval time.Duration
	Options        *clientv3.Config

	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	// leaseMu protects leaseID, which is written by keepAliveLoop and read by registerOne
	leaseMu sync.RWMutex
	leaseID clientv3.LeaseID

	metasLock sync.RWMutex
	metas     map[string]string

	// ctx is cancelled by Stop() to cleanly exit keepAliveLoop and all retries
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// Start initializes the etcd client, grants a lease, and starts the KeepAlive loop.
func (p *CustomEtcdV3RegisterPlugin) Start() error {
	p.done = make(chan struct{})
	p.ctx, p.cancel = context.WithCancel(context.Background())

	if p.client == nil {
		cfg := clientv3.Config{
			Endpoints:   p.EtcdServers,
			DialTimeout: 5 * time.Second,
		}
		if p.Options != nil {
			cfg = *p.Options
		}
		cli, err := clientv3.New(cfg)
		if err != nil {
			return err
		}
		p.client = cli
	}

	p.kv = clientv3.NewKV(p.client)
	p.lease = clientv3.NewLease(p.client)

	if err := p.grantLease(); err != nil {
		return err
	}

	go p.keepAliveLoop()
	return nil
}

func (p *CustomEtcdV3RegisterPlugin) leaseTTL() int64 {
	// TTL = UpdateInterval + 5s buffer, minimum 10s
	ttl := int64(p.UpdateInterval.Seconds()) + 5
	if ttl < 10 {
		ttl = 10
	}
	return ttl
}

// grantLease requests a new lease from etcd and stores its ID.
func (p *CustomEtcdV3RegisterPlugin) grantLease() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := p.lease.Grant(ctx, p.leaseTTL())
	if err != nil {
		return err
	}

	// RISK FIX #2: protect leaseID with a mutex since keepAliveLoop and
	// registerOne can access it concurrently.
	p.leaseMu.Lock()
	p.leaseID = resp.ID
	p.leaseMu.Unlock()
	return nil
}

func (p *CustomEtcdV3RegisterPlugin) currentLeaseID() clientv3.LeaseID {
	p.leaseMu.RLock()
	defer p.leaseMu.RUnlock()
	return p.leaseID
}

// keepAliveLoop runs in a goroutine. It calls KeepAlive with the plugin's
// cancellable ctx so that Stop() cleanly terminates the loop.
func (p *CustomEtcdV3RegisterPlugin) keepAliveLoop() {
	defer close(p.done)

	for {
		if p.ctx.Err() != nil {
			return
		}

		// RISK FIX #3: pass p.ctx so that when Stop() cancels it,
		// the etcd client library's internal goroutine also exits cleanly.
		ch, err := p.lease.KeepAlive(p.ctx, p.currentLeaseID())
		if err != nil {
			if p.ctx.Err() != nil {
				return
			}
			log.Printf("[CustomEtcdV3Plugin] KeepAlive error: %v, retrying...", err)
			// RISK FIX #1: wait with ctx so Stop() unblocks immediately
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}

		// Drain responses until the channel closes (lease expired / network lost)
		if stopped := p.drainKeepAlive(ch); stopped {
			return
		}

		// Channel closed → lease expired, re-grant and re-register
		log.Printf("[CustomEtcdV3Plugin] KeepAlive channel closed, re-granting lease...")
		if err := p.reGrantAndRegister(); err != nil {
			// ctx was cancelled by Stop()
			return
		}
	}
}

// drainKeepAlive reads from ch until it closes or ctx is done.
// Returns true if Stop() was called (ctx cancelled), false if channel closed normally.
func (p *CustomEtcdV3RegisterPlugin) drainKeepAlive(ch <-chan *clientv3.LeaseKeepAliveResponse) (stopped bool) {
	for {
		select {
		case <-p.ctx.Done():
			return true
		case _, ok := <-ch:
			if !ok {
				return false
			}
		}
	}
}

// reGrantAndRegister retries granting a lease until success or Stop() is called.
func (p *CustomEtcdV3RegisterPlugin) reGrantAndRegister() error {
	for {
		if p.ctx.Err() != nil {
			return p.ctx.Err()
		}
		if err := p.grantLease(); err != nil {
			log.Printf("[CustomEtcdV3Plugin] grantLease failed: %v", err)
			// RISK FIX #1: all retry waits respect ctx
			select {
			case <-p.ctx.Done():
				return p.ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}
		p.registerAll()
		return nil
	}
}

func (p *CustomEtcdV3RegisterPlugin) registerAll() {
	p.metasLock.RLock()
	defer p.metasLock.RUnlock()
	for name, metadata := range p.metas {
		if err := p.registerOne(name, metadata); err != nil {
			log.Printf("[CustomEtcdV3Plugin] registerOne %s failed: %v", name, err)
		}
	}
}

func (p *CustomEtcdV3RegisterPlugin) registerOne(name, metadata string) error {
	nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// RISK FIX #2: read leaseID through the mutex-protected accessor
	_, err := p.kv.Put(ctx, nodePath, metadata, clientv3.WithLease(p.currentLeaseID()))
	return err
}

// Stop gracefully shuts down the plugin:
//  1. Cancels the ctx → keepAliveLoop exits via ctx.Done() checks
//  2. Waits for keepAliveLoop to finish
//  3. Revokes the lease (key is removed from etcd immediately)
//  4. Closes the etcd client connection
func (p *CustomEtcdV3RegisterPlugin) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.done != nil {
		<-p.done
	}
	if p.lease != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = p.lease.Revoke(ctx, p.currentLeaseID())
	}
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// Register registers a service key in etcd bound to the current lease.
func (p *CustomEtcdV3RegisterPlugin) Register(name string, rcvr interface{}, metadata string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("Register service `name` can't be empty")
	}

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	p.metas[name] = metadata
	// RISK FIX #4: avoid duplicate entries in Services slice
	found := false
	for _, s := range p.Services {
		if s == name {
			found = true
			break
		}
	}
	if !found {
		p.Services = append(p.Services, name)
	}
	p.metasLock.Unlock()

	return p.registerOne(name, metadata)
}

func (p *CustomEtcdV3RegisterPlugin) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	return p.Register(serviceName, fn, metadata)
}

func (p *CustomEtcdV3RegisterPlugin) Unregister(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("Register service `name` can't be empty")
	}

	nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.kv.Delete(ctx, nodePath)

	p.metasLock.Lock()
	if p.metas != nil {
		delete(p.metas, name)
	}
	newServices := p.Services[:0]
	for _, s := range p.Services {
		if s != name {
			newServices = append(newServices, s)
		}
	}
	p.Services = newServices
	p.metasLock.Unlock()

	return err
}

func (p *CustomEtcdV3RegisterPlugin) GetServices() []string {
	p.metasLock.RLock()
	defer p.metasLock.RUnlock()
	return p.Services
}

func (p *CustomEtcdV3RegisterPlugin) PreCall(ctx context.Context, serviceName, methodName string, args interface{}) (interface{}, error) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("calls", p.Metrics).Mark(1)
	}
	return args, nil
}

func (p *CustomEtcdV3RegisterPlugin) PostCall(ctx context.Context, serviceName, methodName string, args interface{}, reply interface{}, err error) (interface{}, error) {
	return reply, nil
}

func (p *CustomEtcdV3RegisterPlugin) RegisterConn(conn net.Conn) (net.Conn, bool) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("connections", p.Metrics).Mark(1)
	}
	return conn, true
}
