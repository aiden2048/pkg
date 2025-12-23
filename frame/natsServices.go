package frame

import (
	"fmt"
	"os"

	"github.com/aiden2048/pkg/frame/logs"

	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var gNatsconn = &sync.Map{}

var (
	serverNameToNats string
	serverNameOnce   sync.Once
)

func genNameToNats() string {
	serverNameOnce.Do(func() {
		hostName, _ := os.Hostname()
		serverNameToNats = fmt.Sprintf(
			"goserver.%s.%d-%s.%d",
			hostName,
			os.Getpid(),
			server_config.ServerName,
			server_config.ServerID,
		)
	})
	return serverNameToNats
}

func GetNatsConn(plat_id int32) *nats.Conn {
	conn, ok := gNatsconn.Load(plat_id)
	if !ok {
		return nil
	}
	return conn.(*nats.Conn)
}
func GetAllNatsConn() map[int32]*nats.Conn {
	result := make(map[int32]*nats.Conn)
	gNatsconn.Range(func(key, value interface{}) bool {
		result[key.(int32)] = value.(*nats.Conn)
		return true
	})
	return result
}

func StartNatsService(cfg *NatsConfig) error {
	newNatsConn, err := ConnToNats(cfg)
	if err != nil {
		return err
	}

	logs.Infof("Connect to NatsConn: %s", newNatsConn.ConnectedUrl())
	if old, ok := gNatsconn.Load(cfg.PlatId); ok {
		old.(*nats.Conn).Close()
	}
	gNatsconn.Store(cfg.PlatId, newNatsConn)
	time.Sleep(10 * time.Millisecond)

	return nil
}

func ConnToNats(cfg *NatsConfig) (*nats.Conn, error) {

	var NatsOpts = nats.Options{
		User:           cfg.User,
		Password:       cfg.Password,
		Servers:        cfg.Servers,
		Name:           genNameToNats(),
		AllowReconnect: true,
		MaxReconnect:   -1,
		ReconnectWait:  100 * time.Millisecond,
		Timeout:        time.Duration(3) * time.Second,
		Secure:         false,
	}

	var err = error(nil)
	newNatsConn, err := NatsOpts.Connect()
	// 初始化Nats
	if err != nil {
		fmt.Printf("nats.Connect error:%s NatsOpts:%+v", err, NatsOpts)
		logs.Errorf("nats.Connect error:%s", err)
		return nil, err
	}
	logs.Debugf("MaxPayload:%d", newNatsConn.MaxPayload())
	return newNatsConn, nil
}
