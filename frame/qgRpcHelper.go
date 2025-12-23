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

// 推送消息至客户端
func SendMsgToClient(sess *Session, errno int32, errstr string, f string, msgParam interface{}, isEncrypt bool) error {
	if sess == nil {
		logs.Errorf("sess is nil")
		return errors.New("sess is nil")
	}
	if sess.GetServerFE() == "" || sess.GetServerID() == 0 {
		logs.Errorf("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
		return fmt.Errorf("Server %s:%d is invalid", sess.GetServerFE(), sess.GetServerID())
	}
	rspmsg := NatsMsg{Sess: *sess}
	rspmsg.Encrypt = isEncrypt
	rspmsg.MsgBody.Func = f
	var str []byte
	var err error
	if msgParam != nil {
		str, err = jsoniter.Marshal(msgParam)
		if err != nil {
			logs.Errorf("Failed to jsoniter.Marshal: %+v", msgParam)
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

	return RpcxSend(sess.GetServerFE(), sess.GetServerID(), "p2p", &rspmsg, sess.GetPlatID())
}

// 只发送不等回包
func RpcxSend(mod string, svrid int32, cmd string, req *NatsMsg, pids ...int32) (err *errorMsg.ErrRsp) {
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

	if !CheckRpcxService(platId, mod, svrid, cmd) {
		return errorMsg.NoService.Line()
	}
	req.Sess.RpcType = "rpcx"
	return SendRpcx(req.GetUid(), platId, mod, cmd, svrid, req)
}
