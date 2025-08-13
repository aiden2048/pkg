package frame

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiden2048/pkg/public/errorMsg"

	"github.com/aiden2048/pkg/frame/logs"

	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

type serviceSubscription struct {
	conn      *nats.Conn
	subj      string
	queue     string
	ch        chan *nats.Msg
	handler   func(*nats.Conn, *nats.Msg) int32
	threadNum int
	oSubs     []*nats.Subscription
}

var g_natsSubjects []*serviceSubscription

func onRegSubject(sSub *serviceSubscription) {
	if sSub == nil || sSub.conn != gNatsconn {
		return
	}
	if g_natsSubjects == nil {
		g_natsSubjects = make([]*serviceSubscription, 0, 100)
	}
	g_natsSubjects = append(g_natsSubjects, sSub)
	logs.Debugf("append subject %+v", sSub)

}
func UnloadAllSubject() {

	nsubjs := g_natsSubjects
	g_natsSubjects = nil

	for _, sSub := range nsubjs {

		for _, oSub := range sSub.oSubs {
			logs.Infof("UnRegist Nats Subject :%s, %s", oSub.Subject, oSub.Queue)
			oSub.Unsubscribe()
		}
	}

}
func ReregistSubject() {
	natsSubjects := make([]*serviceSubscription, 0, 100)
	for _, sSub := range g_natsSubjects {
		natsSubjects = append(natsSubjects, sSub)
	}
	UnloadAllSubject()
	for _, sSub := range natsSubjects {
		DoRegistNatsHandler(gNatsconn, sSub.subj, sSub.queue, sSub.handler, sSub.threadNum, onRegSubject)
	}

}
func GenReqSubject(mod string, cmd string, svrid int32) string {
	if svrid <= 0 {
		return fmt.Sprintf("%d.msg.%s.%s", GetPlatformId(), mod, cmd)
	} else {
		return fmt.Sprintf("%d.msg.%s.%d.%s", GetPlatformId(), mod, svrid, cmd)
	}
}
func GenRegHandlerSubject(mod string, cmd string, svrid int32) string {
	if svrid <= 0 {
		return fmt.Sprintf("msg.%s.%s", mod, cmd)
	} else {
		return fmt.Sprintf("msg.%s.%d.%s", mod, svrid, cmd)
	}
}

// 兼容老的svrframe
func GenRspSubject(svrName string, svrID int32, channel uint64) string {
	return fmt.Sprintf("%d.msg.%s.%d.p2p", GetPlatformId(), svrName, svrID)
}

// 远程调用并且回包
func RpcxCall(mod string, svrid int32, cmd string, req *NatsMsg, args ...int32) (*NatsMsg, *errorMsg.ErrRsp) {
	req.Sess.PlatId = GetPlatformId()

	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()
	req.Sess.Channel = 0
	if req.MsgBody.Check == nil {
		req.MsgBody.Check = time.Unix(req.Sess.Time, 0).Format("2006-01-02 15:04:05.999")
	}
	if req.Sess.Trace == nil {
		req.Sess.Trace = logs.GetTraceId(&req.Sess)
	}
	//if req.Sess.Trace != nil {
	//	req.Sess.Trace.TraceId = fmt.Sprintf("%s.%d-", GetServerName(), GetServerID()) + req.Sess.Trace.TraceId
	//}

	timeout := GetRpcCallTimeout()

	if len(args) > 0 {
		xtime := args[0]
		timeout = time.Duration(xtime) * time.Second
	}
	platId := GetPlatformId()
	if len(args) > 1 {
		platId = args[1]
	}
	//检查通过rpcx请求
	if CheckRpcxService(platId, mod, svrid, cmd) {
		m, e := CallRpcx(platId, mod, cmd, svrid, req, timeout)
		if etcdConfig.IsRpcxOnly() || e.ErrorNo() != errorMsg.NoService.ErrorNo() {
			return m, e
		}
		logs.PrintSess(req.GetSession(), "call ", platId, mod, cmd, svrid, "err", e)
	}
	if req.Sess.RpcType == "" {
		req.Sess.RpcType = "nats"
	}
	nats := getMixNatsConn(platId)
	msg, errs := NatsCallMsg(nats, mod, svrid, cmd, req, timeout, GetCheckFunc(platId))
	if errs != nil {
		return nil, errs
	}
	rsp := &NatsMsg{}
	err := jsoniter.Unmarshal(msg.Data, rsp)
	if err != nil {
		logs.Errorf("NatsCall %s.%s Unmarshal rsp %s failed:%s", mod, cmd, string(msg.Data), err.Error())
		//ret = ESMR_FAILED
		return nil, errorMsg.RspError.Copy(err)
	}
	/*	if rsp.GetMsgErrNo() != 0 {
		logs.Errorf("NatsCall %s.%s GetMsgBody().Ret = %d, %s", mod, cmd, rsp.GetMsgBody().Ret, rsp.GetMsgBody().ErrStr)
		//ret = ESMR_FAILED
		return rsp, fmt.Errorf("[%d] %s", rsp.GetMsgBody().Ret, rsp.GetMsgBody().ErrStr)
	}*/
	return rsp, nil
}

func NatsCallForTrans(mod string, svrid int32, cmd string, req *NatsMsg, args ...int32) (rsp *NatsTransMsg, err *errorMsg.ErrRsp) {
	req.Sess.PlatId = GetPlatformId()
	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()
	req.Sess.Channel = 0
	if req.Sess.Trace == nil {
		req.Sess.Trace = logs.GetTraceId(&req.Sess)
	}
	if req.Sess.Cmd == "" {
		req.Sess.Cmd = cmd
	}
	timeout := GetRpcCallTimeout()
	if len(args) > 0 {
		xtime := args[0]
		timeout = time.Duration(xtime) * time.Second
	}
	platId := GetPlatformId()
	if len(args) > 1 {
		platId = args[1]
	}
	//检查通过rpcx请求
	if !CheckRpcxService(platId, mod, svrid, cmd) {
		return nil, errorMsg.NoService.Line()
	}
	return CallRpcxForTrans(platId, &req.Sess, mod, cmd, svrid, req, timeout)
}

func NatsCallTrans(mod string, svrid int32, cmd string, req *NatsTransMsg, args ...int32) (*NatsTransMsg, *errorMsg.ErrRsp) {
	req.Sess.PlatId = GetPlatformId()
	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()
	req.Sess.Channel = 0
	if req.Sess.Trace == nil {
		req.Sess.Trace = logs.GetTraceId(&req.Sess)
	}
	if req.Sess.Cmd == "" {
		req.Sess.Cmd = cmd
	}

	timeout := GetRpcCallTimeout()
	if len(args) > 0 {
		xtime := args[0]
		timeout = time.Duration(xtime) * time.Second
	}
	platId := GetPlatformId()
	if len(args) > 1 {
		platId = args[1]
	}
	//检查通过rpcx请求
	if CheckRpcxService(platId, mod, svrid, cmd) {
		m, e := CallRpcxForTrans(platId, &req.Sess, mod, cmd, svrid, req, timeout)
		if etcdConfig.IsRpcxOnly() || e.ErrorNo() != errorMsg.NoService.ErrorNo() {
			return m, e
		}
	}
	if req.Sess.RpcType == "" {
		req.Sess.RpcType = "nats"
	}
	//正常nats请求
	nats := getMixNatsConn(platId)
	msg, errs := NatsCallMsg(nats, mod, svrid, cmd, req, timeout, GetCheckFunc(platId))
	if errs != nil {
		return nil, errs
	}
	rsp := &NatsTransMsg{}
	err := jsoniter.Unmarshal(msg.Data, rsp)
	if err != nil {
		logs.Errorf("NatsCall %s.%s Unmarshal rsp %s failed:%s", mod, cmd, string(msg.Data[:]), err.Error())
		//ret = ESMR_FAILED
		return nil, errorMsg.RspError.Copy(err)
	}
	return rsp, nil
}

// 回包,由于c++过来的请求不会带有replay,因此这里有两种发送模式
func NatsSendReply(req *NatsMsg, rsp *NatsMsg) (err *errorMsg.ErrRsp) {
	req.Sess.PlatId = GetPlatformId()
	req.Sess.SvrFE = GetServerName()
	req.Sess.SvrID = GetServerID()
	req.Sess.Time = time.Now().Unix()

	//req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())

	subj := ""
	if req.GetNatsMsg() != nil && req.GetNatsMsg().Reply != "" {
		subj = req.GetNatsMsg().Reply
	} else if req.GetSession().SvrID > 0 && req.GetSession().Channel > 0 {
		//兼容老的svrframe
		subj = GenRspSubject(req.GetSession().SvrFE, req.GetSession().SvrID, req.GetSession().Channel)
	} else {
		return nil
	}
	conn := req.Conn
	if conn == nil {
		conn = gNatsconn
	}
	return NatsPublish(conn, subj, rsp, true, nil)
}

func HandlerAutoCmd(cmd string, handler func(context.Context, *NatsMsg) int32) error {

	//测试阶段打开, 正式不能打开
	if IsDebug() || GetGlobalConfig().IsTestServer {
		_ = HandlerHttpCmdNoAuth(cmd, handler, true)
	}
	sname := GetServerName()
	return handlerNatsCmdBySname(sname, cmd, handler, IsDebug())
}

// 监听内部服务的请求, 完整协议解析
func HandlerNatsCmd(cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	return handlerNatsCmdBySname(GetServerName(), cmd, handler, p2p...)
}

func handlerNatsCmdBySname(sname, cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	needP2p := false
	if len(p2p) > 0 && p2p[0] {
		needP2p = true
	}

	//添加rpcx注册
	if err := RegisterRpcxHandlerBySName(sname, cmd, handler, needP2p); err != nil {
		return err
	}
	return nil
}

// 监听来自go-conn的请求
func HandlerHttpCmd(cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	return HandlerConnCmd(cmd, handler, p2p...)
}
func GetConnSubName() string {
	return fmt.Sprintf("http.%s", GetServerName())
}

// 监听来自go-conn.xxx的请求
func HandlerConnCmd(cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	needP2p := true
	if len(p2p) > 0 {
		needP2p = p2p[0]
	}
	if IsDebug() {
		needP2p = true
	}
	if IsDebug() && !strings.HasPrefix(cmd, "NA.") {
		_ = HandlerConnCmdByServerName("NA."+cmd, handler, needP2p)
	}
	return HandlerConnCmdByServerName(cmd, handler, needP2p)
}

// 监听来自go-conn.xxx的请求, 指定自己的servername
func HandlerConnCmdByServerName(cmd string, handler func(context.Context, *NatsMsg) int32, needP2p bool /*, tnums ...int*/) error {

	httpMod := GetConnSubName()
	//添加rpcx注册
	if err := RegisterRpcxHandlerBySName(httpMod, cmd, handler, needP2p); err != nil {
		return err
	}
	return nil
}

// NotifyStopConfig
func NotifyStopService(serverName string, serverId int32, obj interface{}) {
	stopKey := fmt.Sprintf("%d.config.stop_server.%s-%d", GetPlatformId(), serverName, serverId)
	//stopKey = "config" + "." + stopKey
	_ = NatsPublish(gNatsconn, stopKey, obj, false, nil)
}

// 注册可以不校验登录态的接口
func HandlerHttpCmdNoAuth(cmd string, handler func(ctx context.Context, msg *NatsMsg) int32, p2p ...bool) error {
	cmd = "NA." + cmd
	return HandlerConnCmd(cmd, handler, p2p...)
}

// 通服
func HandlerAutoCmdMix(cmd string, handler func(context.Context, *NatsMsg) int32) error {
	//测试阶段打开, 正式不能打开

	if IsDebug() || GetGlobalConfig().IsTestServer {
		_ = HandlerHttpCmdNoAuth(cmd, handler, true)
	}
	sname := GetServerName()

	if GetFrameOption().EnableMixServer {
		sname = sname + GetGlobalConfig().MixSuffix
		logs.Print("HandlerAutoCmdMix 接口", cmd, "启用通服混合接口, 服务名为", sname)
	}

	return handlerNatsCmdBySname(sname, cmd, handler, IsDebug() || GetGlobalConfig().IsTestServer)
}
