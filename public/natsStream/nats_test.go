package natsStream

import (
	"fmt"
	"testing"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/nats-io/nats.go"
)

var test = false

func TestNats(t *testing.T) {
	logs.InitServer("test", 123, "../logs/", true, func(s1, s2, s3 string) {}, func(billName string) {})

	push_test()
	publish_test()
	time.Sleep(1000 * time.Second)
}

func push_test() {
	err := PushjsEvent("Pay", "PaySuccess", "test1", push_test_handler)
	if err != nil {
		return
	}
	err = PushjsEvent("Pay", "PaySuccess1", "test1", push_test_handler)
	if err != nil {
		return
	}
}
func publish_test() {
	t1 := TestEvent{Uid: 1}
	PublishSub("Pay", "PaySuccess", 1012, t1, nil)
	time.Sleep(5 * time.Second)
	t2 := TestEvent{Uid: 2}
	PublishSub("Pay", "PaySuccess1", 1013, t2, nil)
	time.Sleep(5 * time.Second)
	t3 := TestEvent{Uid: 3}
	PublishSub("Pay", "PaySuccess", 1014, t3, nil)
}
func push_test_handler(req *TestEvent) *errorMsg.ErrRsp {
	fmt.Println(req)
	return nil
}

type TestEvent struct {
	Uid uint64
}

func TGetStreamContext() nats.JetStreamContext {
	conn, _ := nats.Connect(nats.DefaultURL)
	if conn == nil {
		logs.Errorf("GetNatsConn Failed")
		return nil
	}
	stream, err := conn.JetStream()
	if err != nil {
		logs.Errorf("GetJetStream Failed")
		return nil
	}
	return stream
}
