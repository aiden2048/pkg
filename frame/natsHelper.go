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

	//sSub := &serviceSubscription{subj: subj, queue: queue, handler: handler, threadNum: num, oSubs: oSubs}
	g_natsSubjects = append(g_natsSubjects, sSub)
	logs.LogDebug("append subject %+v", sSub)

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

func UnloadAllSubject() {

	nsubjs := g_natsSubjects
	g_natsSubjects = nil

	for _, sSub := range nsubjs {

		for _, oSub := range sSub.oSubs {
			logs.Trace("UnRegist Nats Subject :%s, %s", oSub.Subject, oSub.Queue)
			oSub.Unsubscribe()
			//if sSub.ch != nil {
			//	close(sSub.ch)
			//	sSub.ch = nil
			//	logs.Trace("Close channel for Nats Subject :%s, %s", oSub.Subject, oSub.Queue)
			//}
		}
	}

}
func UnregistNatsSubject(subj, queue string) {
	var natsSubjects []*serviceSubscription
	for _, sSub := range g_natsSubjects {
		if sSub.subj == subj && sSub.queue == queue {
			for _, oSub := range sSub.oSubs {
				logs.Trace("UnRegist Nats Subject :%s, %s", oSub.Subject, oSub.Queue)
				oSub.Unsubscribe()
			}
		} else {
			natsSubjects = append(natsSubjects, sSub)
		}
	}
	g_natsSubjects = natsSubjects
}
func GenReqSubject(mod string, cmd string, svrid int32) string {
	if svrid <= 0 {
		return fmt.Sprintf("%d.msg.%s.%s", GetPlatformId(), mod, cmd)
	} else {
		return fmt.Sprintf("%d.msg.%s.%d.%s", GetPlatformId(), mod, svrid, cmd)
	}

	//if svrid <= 0 {
	//	return fmt.Sprintf("%d.msg.%s.%s", GetPlatformId(), mod, cmd)
	//} else {
	//	return fmt.Sprintf("%d.msg.%s.%d.%s", GetPlatformId(), mod, svrid, cmd)
	//}
}
func GenRegHandlerSubject(mod string, cmd string, svrid int32) string {
	if svrid <= 0 {
		return fmt.Sprintf("msg.%s.%s", mod, cmd)
	} else {
		return fmt.Sprintf("msg.%s.%d.%s", mod, svrid, cmd)
	}

	//if svrid <= 0 {
	//	return fmt.Sprintf("%d.msg.%s.%s", GetPlatformId(), mod, cmd)
	//} else {
	//	return fmt.Sprintf("%d.msg.%s.%d.%s", GetPlatformId(), mod, svrid, cmd)
	//}
}

// 兼容老的svrframe
func GenRspSubject(svrName string, svrID int32, channel uint64) string {
	//return fmt.Sprintf("msg.%s.%d.p2p.%d", svrName, svrID, channel)
	return fmt.Sprintf("%d.msg.%s.%d.p2p", GetPlatformId(), svrName, svrID)
}

// 远程调用并且回包
func NatsCall(mod string, svrid int32, cmd string, req *NatsMsg, args ...int32) (*NatsMsg, *errorMsg.ErrRsp) {
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
		logs.LogError("NatsCall %s.%s Unmarshal rsp %s failed:%s", mod, cmd, string(msg.Data), err.Error())
		//ret = ESMR_FAILED
		return nil, errorMsg.RspError.Copy(err)
	}
	/*	if rsp.GetMsgErrNo() != 0 {
		logs.LogError("NatsCall %s.%s GetMsgBody().Ret = %d, %s", mod, cmd, rsp.GetMsgBody().Ret, rsp.GetMsgBody().ErrStr)
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
	if CheckRpcxService(platId, mod, svrid, cmd) {
		// return CallRpcxForTrans(platId, &req.Sess, mod, cmd, svrid, req, timeout)
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
	msg, err := NatsCallMsg(nats, mod, svrid, cmd, req, timeout, GetCheckFunc(platId))
	if err != nil {
		return nil, err
	}
	rsp = &NatsTransMsg{}
	errs := jsoniter.Unmarshal(msg.Data, rsp)
	if errs != nil {
		logs.LogError("NatsCall %s.%s Unmarshal rsp %s failed:%s", mod, cmd, string(msg.Data[:]), err.Error())
		//ret = ESMR_FAILED
		return nil, errorMsg.RspError.Copy(errs)
	}
	return rsp, nil
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
		logs.LogError("NatsCall %s.%s Unmarshal rsp %s failed:%s", mod, cmd, string(msg.Data[:]), err.Error())
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

func genErrNatsMsg(errno int32, errstr string) *NatsMsg {
	rspmsg := &NatsMsg{Sess: *NewSessionOnly()}
	//rspmsg.MsgBody.Mod = GetServerName()
	rspmsg.MsgBody.ErrNo = errno
	rspmsg.MsgBody.ErrStr = errstr

	return rspmsg
}

// 注册handler
func registNatsHandler(sname, func_name string, svrid int32, handler func(context.Context, *NatsMsg) int32, tNum int) error {
	if IsDev() {
		sname = fmt.Sprintf("%s_%d", sname, GetServerID())
	}
	subj := GenRegHandlerSubject(sname, func_name, svrid)
	_func := func(nConn *nats.Conn, msg *nats.Msg) int32 {
		start := time.Now()
		req := NatsMsg{}
		err := jsoniter.Unmarshal(msg.Data, &req)
		if err != nil {
			if msg.Reply != "" {
				rsp := genErrNatsMsg(-1, "")
				NatsPublish(nConn, msg.Reply, rsp, false, nil)
			}
			logs.LogError("HandleRpc Unmarshal req %s failed:%s", string(msg.Data[:]), err.Error())
			t := time.Since(start)
			ReportDoRpcStat("Unmarshal", subj, ESMR_FAILED, t)
			return ESMR_FAILED
		}
		if req.MsgBody.Func == "" {
			req.MsgBody.Func = func_name
		}
		if req.Sess.RpcType == "" {
			req.Sess.RpcType = "nats"
		}
		var gpid int64
		if req.Sess.Trace == nil {
			req.Sess.Trace, gpid = logs.CreateTraceId(&req.Sess)
		} else {
			gpid = logs.StoreTraceId(req.Sess.Trace, &req.Sess)
		}
		req.NatsMsg = msg
		req.Conn = nConn
		ret := handler(context.Background(), &req)
		t := time.Since(start)
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		ReportDoRpcStat(s, subj, ret, t)
		//用完移出traceid映射
		if gpid != 0 {
			logs.RemoveTraceId(gpid)
		}
		return ret
	}

	err := DoRegistNatsHandler(gNatsconn, subj, sname, _func, tNum, onRegSubject)
	logs.Print("DoRegistNatsHandler", gNatsconn.Opts.Servers, sname, subj, tNum, err)

	return err
}

// 注册handler收到消息后透传, 并且不起协程
func registNatsTransHandler(sname, func_name string, svrid int32, handler func(msg *NatsTransMsg) int32, tNum int) error {
	if IsDev() {
		sname = fmt.Sprintf("%s_%d", sname, GetServerID())
	}
	subj := GenRegHandlerSubject(sname, func_name, svrid)
	_func := func(nConn *nats.Conn, msg *nats.Msg) int32 {
		start := time.Now()
		req := NatsTransMsg{}
		err := jsoniter.Unmarshal(msg.Data, &req)
		if err != nil {
			//if msg.Reply != "" {
			//	rsp := genErrNatsMsg(-1, "")
			//	NatsPublish(nConn, msg.Reply, rsp, false, nil)
			//}
			logs.LogError("HandleRpc Unmarshal req %s failed:%s", string(msg.Data[:]), err.Error())
			t := time.Since(start)
			ReportDoRpcStat("Unmarshal", subj, ESMR_FAILED, t)
			return ESMR_FAILED
		}

		if req.Sess.RpcType == "" {
			req.Sess.RpcType = "nats"
		}
		req.NatsMsg = msg

		ret := handler(&req)
		t := time.Since(start)
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		ReportDoRpcStat(s, subj, ret, t)
		return ret
	}
	err := DoRegistNatsHandler(gNatsconn, subj, GetServerName(), _func, tNum, onRegSubject)
	logs.Print("DoRegistNatsHandler", gNatsconn.Opts.Servers, GetServerName(), subj, tNum, err)

	//err := DoRegistNatsHandler(g_natsConn, subj, GetServerName(), _func, tNum, onRegSubject)
	return err
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

func HandlerNatsAndRpcxCmd(cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	needP2p := false
	if len(p2p) > 0 && p2p[0] {
		needP2p = true
	}
	b := 0
	//if len(tnums) > 0 {
	//	b = tnums[0]
	//}
	if err := registNatsHandler(GetServerName(), cmd, -1, handler, b); err != nil {
		return err
	}
	//if needP2p {
	//	subj = GenRegHandlerSubject(GetServerName(), cmd, GetServerID())
	//	if err := registNatsHandler(cmd, subj, handler, b); err != nil {
	//		return err
	//	}
	//}
	//添加rpcx注册
	if err := RegisterRpcxHandler(cmd, handler, needP2p); err != nil {
		return err
	}
	return nil
}

// 监听来自go-conn的请求
func HandlerHttpCmd(cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	//b := 0
	//if len(tnums) > 0 {
	//	b = tnums[0]
	//}
	return HandlerConnCmd("mini", cmd, handler, p2p...)
}

// 监听来自go-conn.xxx的请求
func HandlerConnCmd(subName string, cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	needP2p := true
	if len(p2p) > 0 {
		needP2p = p2p[0]
	}
	if IsDebug() {
		needP2p = true
	}
	if IsDebug() && !strings.HasPrefix(cmd, "NA.") {
		_ = HandlerConnCmdByServerName(subName, GetServerName(), "NA."+cmd, handler, needP2p)
	}
	return HandlerConnCmdByServerName(subName, GetServerName(), cmd, handler, needP2p)
}

// 监听来自go-conn.xxx的请求
func HandlerConnCmdMustNats(subName string, cmd string, handler func(context.Context, *NatsMsg) int32, p2p ...bool) error {
	needP2p := false
	if len(p2p) > 0 {
		needP2p = p2p[0]
	}
	return HandlerConnCmdByServerNameMustNats(subName, GetServerName(), cmd, handler, needP2p)
}

// 监听来自go-conn.xxx的请求, 指定自己的servername
func HandlerConnCmdByServerName(subName string, serverName string, cmd string, handler func(context.Context, *NatsMsg) int32, needP2p bool /*, tnums ...int*/) error {

	httpMod := fmt.Sprintf("http_%s.%s", subName, serverName)
	if !GetEtcdConfig().IsRpcxOnly() {
		b := 0
		//if len(tnums) > 0 {
		//	b = tnums[0]
		//}

		if err := registNatsHandler(httpMod, cmd, -1, handler, b); err != nil {
			return err
		}
		if needP2p {
			if err := registNatsHandler(httpMod, cmd, GetServerID(), handler, b); err != nil {
				return err
			}
		}
	}

	//添加rpcx注册
	if err := RegisterRpcxHandlerBySName(httpMod, cmd, handler, needP2p); err != nil {
		return err
	}
	return nil
}

// 监听来自go-conn.xxx的请求, 指定自己的servername
func HandlerConnCmdByServerNameMustNats(subName string, serverName string, cmd string, handler func(context.Context, *NatsMsg) int32, needP2p bool /*, tnums ...int*/) error {

	httpMod := fmt.Sprintf("http_%s.%s", subName, serverName)
	b := 0
	if err := registNatsHandler(httpMod, cmd, -1, handler, b); err != nil {
		return err
	}
	if needP2p {
		if err := registNatsHandler(httpMod, cmd, GetServerID(), handler, b); err != nil {
			return err
		}
	}

	//添加rpcx注册
	if err := RegisterRpcxHandlerBySName(httpMod, cmd, handler, needP2p); err != nil {
		return err
	}
	return nil
}

// 处理gconn过来的透传信息, 有可能二进制协议
func HandlerConnTransCmd(subName string, cmd string, handler func(*NatsTransMsg) int32) error {
	b := 1
	//if len(tnums) > 0 {
	//	b = tnums[0]
	//}
	httpMod := GetServerName()
	if subName != "" {
		httpMod = fmt.Sprintf("http_%s.%s", subName, server_config.ServerName)
	}
	/*subj := GenReqSubject(httpMod, cmd, -1)
	if err := registNatsHandler(cmd, subj, handler, b); err != nil {
		return err
	}*/
	if !GetEtcdConfig().IsRpcxOnly() {
		if err := registNatsTransHandler(httpMod, cmd, -1, handler, b); err != nil {
			return err
		}
		//强制注册p2p接口
		if err := registNatsTransHandler(httpMod, cmd, GetServerID(), handler, b); err != nil {
			return err
		}
	}
	//添加rpcx注册
	if err := RegisterRpcxWsHandler(httpMod, cmd, handler); err != nil {
		return err
	}

	return nil
}

// 监听内部事件的广播, 如chess的
func HandlerEvent(server string, cmd string, handler func(ctx context.Context, msg *NatsTransMsg) int32) error {
	// TODO Add options logic
	return HandlerEvent_Ex(server, cmd, handler, "")
}

// 监听内部事件的广播,所有监听者都会收到, 而且都会处理 HandlerEventBroadcast 和 HandlerEventSveID 一样
func HandlerEventBroadcast(server string, cmd string, handler func(ctx context.Context, msg *NatsTransMsg) int32) error {
	// TODO Add options logic
	return HandlerEvent_Ex(server, cmd, handler, GetStrServerID())
}

func HandlerEventSveID(server string, cmd string, handler func(ctx context.Context, msg *NatsTransMsg) int32) error {
	// TODO Add options logic
	return HandlerEvent_Ex(server, cmd, handler, GetStrServerID())
}

// 监听内部事件的广播, 如chess的
func GenEventSubj(server string, cmd string) string {
	return fmt.Sprintf("%d.event.%s.%s", GetPlatformId(), server, cmd)
}
func HandlerEvent_Ex(server string, cmd string, handler func(ctx context.Context, msg *NatsTransMsg) int32, task string, tnums ...int) error {
	subj := fmt.Sprintf("event.%s.%s", server, cmd)
	return HandlerEvent_Ex1(subj, cmd, handler, task, tnums...)
	//subj := fmt.Sprintf("%d.event.%s.%s", GetPlatformId(), server, cmd)
	//HandlerEvent_Ex1(subj, cmd, handler, task, tnums...)
	//return nil
}
func HandlerEvent_Ex1(subj string, cmd string, handler func(ctx context.Context, msg *NatsTransMsg) int32, task string, tnums ...int) error {
	// TODO Add options logic
	b := defFrameOption.EventHandlerNum
	if len(tnums) > 0 {
		b = tnums[0]
	}
	//subj := GenEventSubj(server, cmd)
	if task == "" {
		task = GetServerName()
	} else {
		task = GetServerName() + "." + task
	}
	err := DoRegistNatsHandler(gNatsconn, subj, task, func(nConn *nats.Conn, msg *nats.Msg) int32 {
		start := time.Now()
		req := NatsTransMsg{}
		err := jsoniter.Unmarshal(msg.Data, &req)
		if err != nil {
			logs.LogError("Handle Event Unmarshal req %s failed:%s", string(msg.Data[:]), err.Error())
			t := time.Since(start)
			ReportDoRpcStat("Unmarshal", subj, ESMR_FAILED, t)
			return ESMR_FAILED
		}
		req.NatsMsg = msg
		req.Sess.RpcType = "nats"
		var gpid int64
		ts := fmt.Sprintf("%02d%02d%02d", start.Hour(), start.Minute(), start.Second())
		if req.Sess.Trace == nil {
			req.Sess.Trace, gpid = logs.CreateTraceId(&req.Sess)
			req.Sess.Trace.TraceId += ":" + cmd + ":" + ts
		} else {
			req.GetSession().Trace.TraceId = req.GetSession().Trace.TraceId +
				fmt.Sprintf(":%s%d:%s:%s", req.GetSession().GetServerFE(), req.GetSession().GetServerID(), cmd, ts)
			gpid = logs.StoreTraceId(req.Sess.Trace, &req.Sess)
		}

		ret := handler(context.TODO(), &req)
		t := time.Since(start)
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		ReportDoRpcStat(s, subj, ret, t)

		//用完移出traceid映射
		if gpid != 0 {
			logs.RemoveTraceId(gpid)
		}
		return ret
	}, b, onRegSubject)

	return err
}

//func UnhandlerEvent(server string, cmd string, task string) {
//	subj := fmt.Sprintf("event.%s.%s", server, cmd)
//	if task == "" {
//		task = GetServerName()
//	}
//	UnregistNatsSubject(subj, task)
//}

// ----------------------------------------------------------------------------
func NatsPulishEvent(ename string, eobj interface{}) error {
	return NatsPulishEventByName(GetServerName(), ename, eobj)
}

func NatsPulishGameEvent(ename string, eobj interface{}, pids ...int32) error {
	return NatsPulishEventByName("game", ename, eobj, pids...)
}
func NatsPulishEventByName(server_name string, ename string, eobj interface{}, pids ...int32) error {

	sess := NewSessionOnly() //Session{SvrFE: GetServerName(), SvrID: GetServerID() /*, Cmd: ename*/}
	msg := NatsTransMsg{Sess: *sess}
	switch eobj.(type) {
	case []byte, jsoniter.RawMessage, string:
		msg.MsgBody = eobj.([]byte)
	default:
		bdata, err := jsoniter.Marshal(eobj)
		if err != nil {
			logs.LogError("Failed to jsoniter.Marshal: %+v", eobj)
			return err
		}
		msg.MsgBody = bdata
	}
	//{
	//	subj := fmt.Sprintf("event.%s.%s", server_name, ename)
	//	return NatsPublish(gNatsconn, subj, msg, false, GetCheckFunc(-1))
	//}
	subj := GenEventSubj(server_name, ename)
	return NatsPublish(gNatsconn, subj, msg, false, GetCheckFunc(-1))

}

// NotifyReloadConfig("online", "push")
func NotifyReloadConfig(cfg string, obj interface{}) {
	subj := fmt.Sprintf("%d.config.%s", GetPlatformId(), cfg)
	logs.Trace("NotifyReloadConfig subj:%s,obj:%+v", subj, obj)
	NatsPublish(gNatsconn, subj, obj, false, nil)
}

// NotifyStopConfig
func NotifyStopService(serverName string, serverId int32, obj interface{}) {
	stopKey := fmt.Sprintf("%d.config.stop_server.%s-%d", GetPlatformId(), serverName, serverId)
	//stopKey = "config" + "." + stopKey
	_ = NatsPublish(gNatsconn, stopKey, obj, false, nil)
}

// 注册可以不校验登录态的接口
func HandlerHttpCmdNoAuth(cmd string, handler func(ctx context.Context, msg *NatsMsg) int32, p2p ...bool) error {
	if err := HandlerConnCmdMustNats("http", cmd, handler, p2p...); err != nil {
		return err
	}
	cmd = "NA." + cmd
	return HandlerConnCmd("http", cmd, handler, p2p...)
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
