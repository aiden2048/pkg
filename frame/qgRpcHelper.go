package frame

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aiden2048/pkg/public/errorMsg"

	"github.com/aiden2048/pkg/frame/logs"

	"time"

	jsoniter "github.com/json-iterator/go"
)

//框架rpc服务, 不是rpcx

// 推送消息至客户端
func SendMsgToClient(sess *Session, errno int32, errstr string, f string, msgParam interface{}, isEncrypt bool) error {
	if sess == nil {
		logs.LogError("sess is nil")
		return errors.New("sess is nil")
	}
	if sess.GetServerFE() == "" || sess.GetServerID() == 0 {
		logs.LogError("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
		return fmt.Errorf("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
	}
	rspmsg := NatsMsg{Sess: *sess}
	rspmsg.Encrypt = isEncrypt
	//rspmsg.MsgBody.Mod = GetServerName()
	rspmsg.MsgBody.Func = f
	// rspmsg.MsgBody.Check = sess.MsgBody.Check
	//rspmsg.MsgBody.Ver = "1.0"
	//rspmsg.MsgBody.Type = "nti"
	//rspmsg.MsgBody.Src = fmt.Sprintf("%s.%d", server_config.ServerName, server_config.ServerID)
	var str []byte
	var err error
	if msgParam != nil {
		str, err = jsoniter.Marshal(msgParam)
		if err != nil {
			logs.LogError("Failed to jsoniter.Marshal: %+v", msgParam)
			return errors.New("jsoniter.Marshal rsp failed")
		}
	}
	rspmsg.MsgBody.ErrNo = errno
	rspmsg.MsgBody.ErrStr = errstr
	rspmsg.MsgBody.Seq = sess.Seq
	if sess.Check != "" {
		rspmsg.MsgBody.Check = sess.Check
	}
	rspmsg.MsgBody.Param = str
	traceId := ""
	if sess.Trace != nil {
		traceId = sess.Trace.TraceId
	}
	if traceIdPos := strings.Index(traceId, ":"); traceIdPos > 0 {
		traceId = traceId[:traceIdPos]
	}
	rspmsg.MsgBody.Trace = traceId
	//reply := fmt.Sprintf("rpc.p2p.%s.%d", m.GetSession().SvrFE, m.GetSession().SvrID)
	//if m.GetNatsMsg() != nil {
	//	reply = m.GetNatsMsg().Reply
	//}
	// if sess.GetUid() > 0 {
	// 	//	logs.LogUser(sess.GetUid(), "\n============================\nNatsSendMsgToClient to %s:%d Param: %s\n============================\n", sess.GetServerFE(), sess.GetServerID(), string(str))
	// }
	toNats := 0
	if sess.RpcType == "nats" {
		toNats = 1
	}
	return NatsSend(sess.GetServerFE(), sess.GetServerID(), "p2p", &rspmsg, sess.GetPlatID(), int32(toNats))
}

// 只发送不等回包
func NatsSend(mod string, svrid int32, cmd string, req *NatsMsg, pids ...int32) (err *errorMsg.ErrRsp) {
	req.Sess.PlatId = GetPlatformId()
	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()
	if req.Sess.Cmd == "" {
		req.Sess.Cmd = cmd
	}
	platId := int32(GetPlatformId())
	if len(pids) > 0 {
		platId = pids[0]
	}
	// toNats := 0
	// if len(pids) > 1 {
	// 	toNats = int(pids[1])
	// }
	//检查通过rpcx请求
	if /*toNats == 0 &&*/ CheckRpcxService(platId, mod, svrid, cmd) {
		req.Sess.RpcType = "rpcx"
		e := SendRpcx(req.GetUid(), platId, mod, cmd, svrid, req)
		if etcdConfig.IsRpcxOnly() || e.ErrorNo() != errorMsg.NoService.ErrorNo() {
			return e
		}
	}
	//正常nats请求

	//	req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())
	subj := GenReqSubject(mod, cmd, svrid)
	req.Sess.RpcType = "nats"

	return NatsPublish(getMixNatsConn(platId), subj, req, false, GetCheckFunc(platId))
}

// 推送消息至客户端
func SendBinaryMsgToClient(sess *Session, msgParam []byte) error {
	if sess == nil {
		logs.LogError("sess is nil")
		return errors.New("sess is nil")
	}
	if sess.GetServerFE() == "" || sess.GetServerID() == 0 {
		logs.LogError("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
		return fmt.Errorf("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
	}
	// if sess.GetUid() > 0 {
	// 	//	logs.LogUser(sess.GetUid(), "\n============================\nNatsSendMsgToClient to %s:%d Param: %s\n============================\n", sess.GetServerFE(), sess.GetServerID(), string(str))
	// }
	rspmsg := &NatsTransMsg{Sess: *sess}
	rspmsg.MsgData = msgParam

	return SendTrans(sess.GetServerFE(), sess.GetServerID(), "p2p", rspmsg, sess.GetPlatID())
}

// 只发送不等回包
func SendTrans(mod string, svrid int32, cmd string, req *NatsTransMsg, platId int32) (err *errorMsg.ErrRsp) {
	if platId <= 0 {
		platId = GetPlatformId()
	}
	req.Sess.PlatId = GetPlatformId()
	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()
	if req.Sess.Cmd == "" {
		req.Sess.Cmd = cmd
	}
	//检查通过rpcx请求
	if CheckRpcxService(platId, mod, svrid, cmd) {
		req.Sess.RpcType = "rpcx"
		e := SendRpcx(req.GetUid(), platId, mod, cmd, svrid, req)
		if etcdConfig.IsRpcxOnly() || e.ErrorNo() != errorMsg.NoService.ErrorNo() {
			return e
		}
	}
	//正常nats请求
	req.Sess.RpcType = "nats"
	//	req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())
	subj := GenReqSubject(mod, cmd, svrid)
	return NatsPublish(getMixNatsConn(platId), subj, req, false, GetCheckFunc(platId))
}
