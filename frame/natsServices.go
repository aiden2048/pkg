package frame

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	t "log"
	"os"

	"github.com/aiden2048/pkg/frame/stat"

	"github.com/aiden2048/pkg/frame/logs"

	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var gNatsconn *nats.Conn // 默认连接的nats

// 混服下链接到其他平台的nats
// var gMixnatsconn map[int32]*nats.Conn // 混服连接的nats
//var mMixNatsConnLocker sync.Mutex // 混服的nats conn锁

const subjectForReportService = "system.service.report"

var subscribeForReportService *nats.Subscription
var subscribeForReportServerAsk *nats.Subscription

const timeSecondToReport = 5

var mServiceLocker sync.Mutex

type NatsConnect struct {
	ServerName string    `json:"server_name,omitempty"`
	ServerId   int32     `json:"server_id,omitempty"`
	NatsName   string    `json:"name,omitempty"`
	SubList    []string  `json:"subscriptions_list,omitempty"`
	UpdateTime time.Time `json:"-"`
	Cmd        string    `json:"cmd"`
}

var serverNameToNats string

func init() {

}
func genNameToNats() string {
	if serverNameToNats != "" {
		return serverNameToNats
	}
	hostName, _ := os.Hostname()
	serverNameToNats = fmt.Sprintf("goserver.%s.%d-%s.%d", hostName, os.Getpid(), server_config.ServerName, server_config.ServerID)
	return serverNameToNats
}

func genAskSubjToNats() string {
	return fmt.Sprintf("%d.system.service.ask.%s.%d", GetPlatformId(), GetServerName(), GetServerID())
}

func StartLoadNatsServices() error {
	go func() {
		//tick := time.Tick(timeSecondToReport * time.Second)
		for {
			select {
			case <-IsStop():
				logs.Infof("G_CloseChan close")
				log.Printf("G_CloseChan close\n")
				return
			}
		}
	}()
	return nil
}

func GetCheckFunc(platid int32) func(string) bool {
	//if platid <= 0 || platid == GetPlatformId() {
	//	return CheckNatsService
	//}
	return nil
}

func GetNatsConn() *nats.Conn {
	return gNatsconn
}

func getMixNatsConn(platId int32) *nats.Conn {

	return gNatsconn
}

func onServerAsk(natsMsg *nats.Msg) {
	_ = gNatsconn.Publish(natsMsg.Reply, []byte(genNameToNats()))
}

func askServer(natsConn *nats.Conn) (string, bool) {
	nadd := int32(1000)
	if GetServerID() > 0 {
		//nadd = 1000
	} else {
		server_config.ServerID = 1
	}
	for i := 1; i < 100; i++ {
		fmt.Printf("Is there is anthor goServer run as %s ........\n", genAskSubjToNats())
		reply, err := natsConn.Request(genAskSubjToNats(), []byte("are you here?"), time.Millisecond*500)
		logs.Infof("Ask Server %s, replay: %+v, err: %+v", genAskSubjToNats(), reply, err)
		if err == nil {
			fmt.Printf("some one reply as %s\n", string(reply.Data))
			server_config.ServerID = GetServerID() + nadd
			continue
		}
		fmt.Printf("nobody answer me as %s\n", genAskSubjToNats())
		return "", true
	}
	return "no Empty ServerId for me", false
}

func StartNatsService(cfg *NatsConfig) error {
	oldNatsConn := gNatsconn
	newNatsConn, err := ConnToNats(cfg)
	if err != nil {
		return err
	}
	if oldNatsConn == nil {
		//第一次连接, 先查询下id
		if ret, b := askServer(newNatsConn); ret != "" {
			return fmt.Errorf("there is other node for goServer:[%s, %d] at: %s", GetServerName(), GetServerID(), ret)
		} else if b {
			newNatsConn.Close()
			//重新连接
			serverNameToNats = ""
			newNatsConn, err = ConnToNats(cfg)
			if err != nil {
				return err
			}
		}
	}

	logs.Infof("Connect to NatsConn: %s", newNatsConn.ConnectedUrl())
	gNatsconn = newNatsConn

	if oldNatsConn != nil {
		if subscribeForReportServerAsk != nil {
			_ = subscribeForReportServerAsk.Unsubscribe()
			subscribeForReportServerAsk = nil
		}
		go func() {
			time.Sleep(2 * time.Second)
			logs.Infof("Close NatsConn: %s", oldNatsConn.ConnectedUrl())
			oldNatsConn.Close()
		}()
	}
	ReregistSubject()

	time.Sleep(10 * time.Millisecond)

	subscribeForReportServerAsk, _ = gNatsconn.Subscribe(genAskSubjToNats(), onServerAsk)
	return nil
}

func ReportBillStat(billName string) {
	stat.ReportStat("bill:"+billName, 0, 0)
}

func ConnToNats(cfg *NatsConfig) (*nats.Conn, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2
	}

	var tlscfg *tls.Config
	if cfg.Secure {
		tlscfg = &tls.Config{}
		tlscfg.InsecureSkipVerify = true
		if cfg.RootCa != "" {
			rootCAs := x509.NewCertPool()
			loaded, err := ioutil.ReadFile(cfg.RootCa)
			if err != nil {
				t.Fatalf("unexpected missing certfile: %v", err)
			}
			rootCAs.AppendCertsFromPEM(loaded)
			tlscfg.RootCAs = rootCAs
		}

	}

	var NatsOpts = nats.Options{
		//	Url:            natsConfig.Url,
		User:     cfg.User,
		Password: cfg.Password,
		//	Token:          natsConfig.Token,
		Servers:        cfg.Servers,
		Name:           genNameToNats(),
		AllowReconnect: true,
		MaxReconnect:   -1,
		ReconnectWait:  100 * time.Millisecond,
		Timeout:        time.Duration(cfg.Timeout) * time.Second,
		Secure:         cfg.Secure,
		TLSConfig:      tlscfg,
	}

	var err = error(nil)
	newNatsConn, err := NatsOpts.Connect()
	// 初始化Nats
	if err != nil {
		fmt.Printf("nats.Connect error:%s NatsOpts:%+v", err, NatsOpts)
		logs.Errorf("nats.Connect error:%s", err)
		return nil, err
	}
	logs.LogDebug("MaxPayload:%d", newNatsConn.MaxPayload())
	return newNatsConn, nil
}
