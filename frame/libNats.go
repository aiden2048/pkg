package frame

import (
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
