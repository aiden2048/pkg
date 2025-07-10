package frame

import (
	"fmt"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/nats-io/nats.go"
)

//监听配置变更通知
// func ListenConfig(config_key string, _func func([]byte)){
// 	subj := "config." + config_key
// 	queue := fmt.Sprintf("%s-%d", GetServerName(), GetServerID())
// 	_, err := g_natsConn.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
// 		logs.Infof("Recv %s to reload config", msg.Subject)
// 		_func(msg.Data)
// 	})
// 	logs.Infof("Listen Config key: %s %+v", config_key, err)
// }

// 监听配置变更通知
func ListenConfig(config_key string, _func func([]byte)) {
	ListenConfigEx("config", config_key, _func)
	//ListenConfigEx(fmt.Sprintf("%d.config", GetPlatformId()), config_key, _func)
}
func ListenConfigEx(c string, config_key string, _func func([]byte)) {
	subj := c + "." + config_key
	queue := fmt.Sprintf("%s-%d", GetServerName(), GetServerID())
	//osub, err := g_natsConn.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
	//	logs.Infof("Recv %s to reload config", msg.Subject)
	//	_func(msg.Data)
	//
	//}):q\
	err := DoRegistNatsHandler(gNatsconn, subj, queue, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		logs.Infof("Recv_Config %s to reload config:%s", msg.Subject, string(msg.Data))
		_func(msg.Data)
		return 0
	}, 1, onRegSubject)

	logs.Infof("ListenConfig key: %s @ %s %+v", subj, queue, err)

	subj = c + "." + config_key + "." + GetServerName()
	err = DoRegistNatsHandler(gNatsconn, subj, queue, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		logs.Infof("Recv_Config %s to reload config:%s", msg.Subject, string(msg.Data))
		_func(msg.Data)
		return 0
	}, 1, onRegSubject)
	logs.Infof("ListenConfig key: %s @ %s %+v", subj, queue, err)

	subj = c + "." + config_key + "." + queue
	err = DoRegistNatsHandler(gNatsconn, subj, queue, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		logs.Infof("Recv_Config %s to reload config:%s", msg.Subject, string(msg.Data))
		_func(msg.Data)
		return 0
	}, 1, onRegSubject)
	logs.Infof("ListenConfig key: %s @ %s %+v", subj, queue, err)
}
