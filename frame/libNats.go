package frame

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	t "log"
	"strings"
	"time"

	"github.com/aiden2048/pkg/public/errorMsg"

	"github.com/aiden2048/pkg/frame/logs"

	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

//var RpcTimeout = errors.New(`{"errno":401, "err":"Timeout"}`)
//var errorMsg.NoService = errors.New(`{"errno":404, "err":"No Service"}`)
//var RpcNoRequest = errors.New(`{"errno":405, "err":"No Request"}`)

// 发送消息到nats
func NatsPublish(natsConn *nats.Conn, subj string, req interface{}, isReply bool, CheckNatsService func(string) bool) *errorMsg.ErrRsp {
	if natsConn == nil {
		logs.Errorf("NatsPublish no natsConn")
		return errorMsg.NoService
	}
	if !isReply && CheckNatsService != nil {
		if ok := CheckNatsService(subj); ok == false {
			logs.LogDebug("%+v not find service for %s", natsConn.Servers(), subj)
			return errorMsg.NoService
		}
	}
	data, err := jsoniter.Marshal(req)
	if err != nil {
		logs.Errorf("NatsPublish Marshal req (%+v) failed:%s", req, err.Error())
		return errorMsg.ReqError.Copy(err)
	}
	start := time.Now()
	err = natsConn.Publish(subj, data)
	since := time.Since(start)
	ret := 0
	if err != nil {
		ret = 1
		logs.Errorf("natsConn %+v Publish[%s] (%d)(%s) failed：%s", natsConn.Servers(), subj, len(data), string(data), err.Error())
		ReportStat("Send", subj, ret, since)
		return errorMsg.RspError.Copy(err)
	}
	if strings.Index(subj, "_INBOX.") != 0 {
		ReportStat("Send", subj, int(ret), since)
	}

	if !strings.Contains(subj, "HeartBeat") && IsDebug() {
		//logs.LogDebug("natsConn %+v Publish[%s] (%+v) ", natsConn.Servers(), subj, string(data))
	}
	return nil
}

// 发送消息到natsJetStream
func NatsPublishJs(natsConn *nats.Conn, subj string, req interface{}, isReply bool, CheckNatsService func(string) bool) *errorMsg.ErrRsp {

	if !isReply && CheckNatsService != nil {
		if ok := CheckNatsService(subj); ok == false {
			logs.LogDebug("%+v not find service for %s", natsConn.Servers(), subj)
			return errorMsg.NoService
		}
	}
	data, err := jsoniter.Marshal(req)
	if err != nil {
		logs.Errorf("NatsCall Marshal req (%+v) failed:%s", req, err.Error())
		return errorMsg.ReqError.Copy(err)
	}
	natsJs, err := natsConn.JetStream()
	if err != nil {
		logs.LogDebug("%+v get JetStream failed:%s", natsConn.Servers(), err.Error())
		return errorMsg.NoService
	}
	start := time.Now()
	_, err = natsJs.Publish(subj, data)
	t := time.Since(start)
	ret := 0
	if err != nil {
		ret = 1
		logs.Errorf("natsConn %+v Publish[%s] (%d)(%s) failed：%s", natsConn.Servers(), subj, len(data), string(data), err.Error())
		ReportStat("Send", subj, int(ret), t)
		return errorMsg.RspError.Copy(err)
	}
	if strings.Index(subj, "_INBOX.") != 0 {
		ReportStat("Send", subj, int(ret), t)
	}

	if !strings.Contains(subj, "HeartBeat") && IsDebug() {
		//logs.LogDebug("natsConn %+v Publish[%s] (%+v) ", natsConn.Servers(), subj, string(data))
	}
	return nil
}

func NatsCallMsg(natsConn *nats.Conn, mod string, svrid int32, cmd string, req interface{}, timeout time.Duration, CheckNatsService func(string) bool) (*nats.Msg, *errorMsg.ErrRsp) {
	if natsConn == nil {
		logs.Errorf("no conn")
		return nil, errorMsg.RspError.Copy("NoConn")
	}
	if timeout <= 0 || timeout > 300*time.Second {
		timeout = GetRpcCallTimeout()
	}
	subj := GenReqSubject(mod, cmd, svrid)
	if CheckNatsService != nil {
		if ok := CheckNatsService(subj); ok == false {
			logs.LogDebug("CheckNatsService subj:%s, err:%s", subj, errorMsg.NoService.Copy(subj).Error())
			return nil, errorMsg.NoService
		}
	}

	data, err := jsoniter.Marshal(req)
	if err != nil {
		logs.Errorf("NatsCall %s Marshal req (%+v) failed:%s", subj, req, err.Error())
		return nil, errorMsg.ReqParamError.Copy(err)
	}
	start := time.Now()
	msg, err := natsConn.Request(subj, data, timeout)

	t := time.Since(start)
	ret := ESMR_SUCCEED
	var errs *errorMsg.ErrRsp
	if err != nil {
		if err == nats.ErrTimeout {
			errs = errorMsg.TimeOut.Copy(err)
			ret = ESMR_TIMEOUT
		} else {
			errs = errorMsg.ReqError.Copy(err)
			ret = ESMR_FAILED
		}
		logs.Errorf("natsConn %+v.Request %s , timeout:%d failed:%s", natsConn.Servers(), subj, timeout/time.Millisecond, err.Error())
		return nil, errs
	} else {
		logs.LogDebug("natsConn %+v.Request %s (cost:%v)\n: %s , Response  %s", natsConn.Opts.Servers, subj, t, string(data), string(msg.Data))
	}
	ReportCallRpcStat(mod, cmd, ret, t)
	return msg, nil
}

func DoRegistNatsHandler(natsConn *nats.Conn, subj string, queue string, handler func(*nats.Conn, *nats.Msg) int32, threadNum int,
	onReg func(sSub *serviceSubscription)) error {
	//err := DoRegistNatsHandlerEx(natsConn, subj, queue, handler, threadNum, onReg)
	//if err != nil {
	//	return err
	//}
	subj = fmt.Sprintf("%d.%s", GetPlatformId(), subj)
	return DoRegistNatsHandlerEx(natsConn, subj, queue, handler, threadNum, onReg)
}

func DoRegistNatsHandlerEx(natsConn *nats.Conn, subj string, queue string, handler func(*nats.Conn, *nats.Msg) int32, threadNum int,
	onReg func(sSub *serviceSubscription)) error {

	// TODO Add options logic
	if natsConn == nil {
		logs.Errorf("no natsConn, subj:%s, queue:%s", subj, queue)
		return errors.New("no natsConn")
	}
	//subj := fmt.Sprintf("rpc.%s.%s", server_config.ServiceName, cmd)
	//queue := server_config.ServerName
	var oSubjs []*nats.Subscription = nil
	//var err error = nil

	if threadNum <= 0 {
		oSubj, err := natsConn.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
			//起一个协程去做这个事
			if !strings.Contains(subj, "HeartBeat") && IsDebug() {
				logs.LogDebug("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
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
					logs.LogDebug("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
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
	if onReg != nil {
		sSub := &serviceSubscription{conn: natsConn, subj: subj, queue: queue, handler: handler, threadNum: threadNum, oSubs: oSubjs}
		onReg(sSub)
	}

	return nil
}

/*
func DoRegistNatsJsHandler(natsConn *nats.Conn, subj string, queue string, handler func(*nats.Conn, *nats.Msg) int32, threadNum int,

		onReg func(sSub *serviceSubscription)) error {
		if err := DoRegistNatsJsHandlerEx(natsConn, subj, queue, handler, threadNum, onReg); err != nil {
			return err
		}
		subj = fmt.Sprintf("%d.%s", GetPlatformId(), subj)
		if err := DoRegistNatsJsHandlerEx(natsConn, subj, queue, handler, threadNum, onReg); err != nil {
			return err
		}
		return nil
	}

func DoRegistNatsJsHandlerEx(natsConn *nats.Conn, subj string, queue string, handler func(*nats.Conn, *nats.Msg) int32, threadNum int,

		onReg func(sSub *serviceSubscription)) error {

		// TODO Add options logic
		if natsConn == nil {
			logs.Errorf("no natsConn, subj:%s, queue:%s", subj, queue)
			return errors.New("no natsConn")
		}
		natsJs, err := natsConn.JetStream()
		if err != nil {
			logs.LogDebug("%+v get JetStream failed:%s", natsConn.Servers(), err.Error())
			return errors.New("no natsJetStream")
		}
		//subj := fmt.Sprintf("rpc.%s.%s", server_config.ServiceName, cmd)
		//queue := server_config.ServerName
		var oSubjs []*nats.Subscription = nil
		//var err error = nil

		if threadNum <= 0 {
			oSubj, err := natsJs.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
				//起一个协程去做这个事
				if !strings.Contains(subj, "HeartBeat") && IsDebug() {
					logs.LogDebug("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
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
				oSubj, err := natsJs.QueueSubscribe(subj, queue, func(msg *nats.Msg) {
					if !strings.Contains(subj, "HeartBeat") && IsDebug() {
						logs.LogDebug("natsConn.Recive At: %s - GetMsg:%s[%s]", subj, queue, string(msg.Data))
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
		if onReg != nil {
			sSub := &serviceSubscription{conn: natsConn, subj: subj, queue: queue, handler: handler, threadNum: threadNum, oSubs: oSubjs}
			onReg(sSub)
		}

		return nil
	}
*/
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
