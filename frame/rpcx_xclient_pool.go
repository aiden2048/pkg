package frame

import (
	"sync"
	"sync/atomic"

	"github.com/smallnest/rpcx/client"

	"github.com/smallnest/rpcx/protocol"
)

// XClientPool is a xclient pool with fixed size.
// It uses roundrobin algorithm to call its xclients.
// All xclients share the same configurations such as ServiceDiscovery and serverMessageChan.
type XClientPool struct {
	count    uint64
	index    uint64
	xclients []client.XClient
	mu       sync.RWMutex

	servicePath string
	failMode    client.FailMode
	selectMode  client.SelectMode
	discovery   client.ServiceDiscovery
	option      client.Option
	auth        string

	serverMessageChan chan<- *protocol.Message
}

// NewXClientPool creates a fixed size XClient pool.
func NewXClientPool(count int, servicePath string, failMode client.FailMode, selectMode client.SelectMode, discovery client.ServiceDiscovery, option client.Option) *XClientPool {
	pool := &XClientPool{
		count:       uint64(count),
		xclients:    make([]client.XClient, count),
		servicePath: servicePath,
		failMode:    failMode,
		selectMode:  selectMode,
		discovery:   discovery,
		option:      option,
	}

	for i := 0; i < count; i++ {
		xclient := client.NewXClient(servicePath, failMode, selectMode, discovery, option)
		pool.xclients[i] = xclient
	}
	return pool
}

// Auth sets s token for Authentication.
func (c *XClientPool) Auth(auth string) {
	c.auth = auth
	c.mu.RLock()
	for _, v := range c.xclients {
		v.Auth(auth)
	}
	c.mu.RUnlock()
}

// Get returns a xclient.
// It does not remove this xclient from its cache so you don't need to put it back.
// Don't close this xclient because maybe other goroutines are using this xclient.
func (p *XClientPool) Get(u uint64) client.XClient {
	//小于10000以下的不是真实UID, 随机取一个客户端
	if u < 10000 {
		i := atomic.AddUint64(&p.index, 1)
		picked := int(i % p.count)
		return p.xclients[picked]
	} else {
		picked := int(u % p.count)
		return p.xclients[picked]
	}

}

// Close this pool.
// Please make sure it won't be used any more.
func (p *XClientPool) Close() {
	for _, c := range p.xclients {
		c.Close()
	}
	p.xclients = nil
}
