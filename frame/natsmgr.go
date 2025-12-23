package frame

import (
	"strings"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/nats-io/nats.go"
)

type NatsManager struct {
	mu    sync.RWMutex
	conns map[int32]*nats.Conn // platId -> conn
}

func NewNatsManager() *NatsManager {
	return &NatsManager{
		conns: make(map[int32]*nats.Conn),
	}
}

func (m *NatsManager) StartOrUpdate(cfg NatsConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	old, exists := m.conns[cfg.PlatId]

	if exists && old != nil && !NeedReconnect(old, cfg) {
		return nil
	}

	// 需要重连 → 先关旧连接
	if exists && old != nil {
		old.Close()
		delete(m.conns, cfg.PlatId)
	}

	// 建新连接
	nc, err := Connect(cfg)
	if err != nil {
		return err
	}

	m.conns[cfg.PlatId] = nc
	return nil
}
func (m *NatsManager) ClosePlat(platId int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if nc, ok := m.conns[platId]; ok {
		nc.Close()
		delete(m.conns, platId)
	}
}

func (m *NatsManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, nc := range m.conns {
		nc.Close()
		delete(m.conns, k)
	}
}

func (m *NatsManager) Get(platId int32) *nats.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conns[platId]
}

func (m *NatsManager) GetAll() map[int32]*nats.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cp := make(map[int32]*nats.Conn, len(m.conns))
	for k, v := range m.conns {
		cp[k] = v
	}
	return cp
}
func NeedReconnect(old *nats.Conn, cfg NatsConfig) bool {
	// 你可以按需扩展
	if old.ConnectedUrl() == "" {
		return true
	}
	// server list / auth 变化 → 重连
	return true
}

func Connect(cfg NatsConfig) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name(genNameToNats()),
		nats.UserInfo(cfg.User, cfg.Password),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
		nats.Timeout(5 * time.Second),

		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logs.Errorf("NATS disconnected plat=%d err=%v", cfg.PlatId, err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logs.Infof("NATS reconnected plat=%d url=%s", cfg.PlatId, nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logs.Warnf("NATS closed plat=%d", cfg.PlatId)
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logs.Errorf("NATS async error: %v", err)
		}),
	}

	return nats.Connect(strings.Join(cfg.Servers, ","), opts...)
}
