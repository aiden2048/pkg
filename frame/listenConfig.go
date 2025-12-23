package frame

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/nats-io/nats.go"
)

func getConfigSubj(cfg string) string {
	return fmt.Sprintf("%d.config.%s", GetPlatformId(), cfg)
}
func getCacheSubj(cfg string) string {
	return fmt.Sprintf("%d.cache.%s", GetPlatformId(), cfg)
}

// NotifyReloadConfig("online", "push")
func NotifyReloadConfig(cfg string, obj interface{}) {
	subj := getConfigSubj(cfg)
	logs.Infof("NotifyReloadConfig subj:%s,obj:%+v", subj, obj)
	NatsPublish(GetNatsConn(GetPlatformId()), subj, obj, false, nil)
}

// 监听配置变更通知
func ListenConfig(config_key string, _func func([]byte)) {
	subj := getConfigSubj(config_key)
	queue := fmt.Sprintf("%s-%d", GetServerName(), GetServerID())
	err := DoRegistNatsHandler(GetNatsConn(GetPlatformId()), subj, queue, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		logs.Infof("Recv_Config %s to reload config:%s", msg.Subject, string(msg.Data))
		_func(msg.Data)
		return 0
	}, 1)

	logs.Infof("ListenConfig key: %s @ %s %+v", subj, queue, err)
}
func NotifyReloadCache(key string, obj interface{}) {
	subj := getCacheSubj(key)
	logs.Infof("NotifyReloadCache subj:%s,obj:%+v", subj, obj)
	NatsPublish(GetNatsConn(GetPlatformId()), subj, obj, false, nil)
}

// 监听配置变更通知
func ListenCache(config_key string, _func func([]byte)) {
	subj := getCacheSubj(config_key)
	queue := fmt.Sprintf("%s-%d", GetServerName(), GetServerID())
	err := DoRegistNatsHandler(GetNatsConn(GetPlatformId()), subj, queue, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		logs.Infof("Recv_Cache %s to reload config:%s", msg.Subject, string(msg.Data))
		_func(msg.Data)
		return 0
	}, 1)
	logs.Infof("ListenCache key: %s @ %s %+v", subj, queue, err)
}

func DoRegistNatsHandler(natsConn *nats.Conn, subj string, queue string, handler func(*nats.Conn, *nats.Msg) int32, threadNum int) error {

	// TODO Add options logic
	if natsConn == nil {
		logs.Errorf("no natsConn, subj:%s, queue:%s", subj, queue)
		return errors.New("no natsConn")
	}

	var oSubjs []*nats.Subscription = nil

	if threadNum <= 0 {
		oSubj, err := natsConn.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
			//起一个协程去做这个事
			if !strings.Contains(subj, "HeartBeat") && IsDebug() {
				logs.Debugf("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
			}

			go handler(natsConn, msg)

		})
		if err != nil {
			logs.Errorf("natsConn %+v.QueueSubscribe %s queue:%s failed:%s", natsConn.Servers(), subj, queue, err.Error())
			return err
		}
		logs.Infof("natsConn %+v.QueueSubscribe %s queue:%s, threadNum: %d succ", natsConn.Servers(), subj, queue, threadNum)

		oSubjs = append(oSubjs, oSubj)
	} else {
		for i := 0; i < threadNum; i++ {
			oSubj, err := natsConn.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
				if !strings.Contains(subj, "HeartBeat") && IsDebug() {
					logs.Debugf("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
				}
				handler(natsConn, msg)
				//go handler(msg)
			})
			if err != nil {
				logs.Errorf("natsConn %+v.QueueSubscribe %s,Index:%d,threadNum:%d queue:%s failed:%s", natsConn.Servers(), subj, i, threadNum, queue, err.Error())
				return err
			}
			logs.Infof("natsConn %+v.QueueSubscribe %s queue:%s,Index:%d,threadNum:%d succ", natsConn.Servers(), subj, queue, i, threadNum)
			oSubjs = append(oSubjs, oSubj)
		}

	}

	natsConn.Flush()

	return nil
}
