package frame

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/aiden2048/pkg/utils"

	"github.com/rpcxio/libkv/store"

	"github.com/aiden2048/pkg/frame/rpcPorts"
	"github.com/aiden2048/pkg/utils/maddr"
	"github.com/aiden2048/pkg/utils/mnet"

	"github.com/aiden2048/pkg/public/errorMsg"

	"github.com/alitto/pond"

	"github.com/aiden2048/pkg/frame/logs"

	jsoniter "github.com/json-iterator/go"
	"github.com/smallnest/rpcx/client"
	rpcxlog "github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"

	"github.com/aiden2048/pkg/frame/stat"
)

var rpcxStarted bool

var listener net.Listener
var rpcxServer *server.Server
var address string

type TRpcxMsg struct {
	Data []byte
}

func init() {
	//_allServiceDiscovery.diss = make(map[string]client.ServiceDiscovery)
	_allRpcxClients.clients = make(map[string]*XClientPool)
	_allRpcxEtcdStore.stores = make(map[string]store.Store)

}

func ReclearRpcxClient() {

	//dss := _allServiceDiscovery.diss

	css := _allRpcxClients.clients
	stores := _allRpcxEtcdStore.stores
	//_allServiceDiscovery.diss = make(map[string]client.ServiceDiscovery)

	_allRpcxClients.clients = make(map[string]*XClientPool)
	_allRpcxEtcdStore.stores = make(map[string]store.Store)

	for _, cs := range css {
		cs.Close()
	}

	//for _, ds := range dss {
	//	ds.Close()
	//}

	for _, kv := range stores {
		kv.Close()
	}
}

const myContextOneway = "_frame_oneway"

type myRpcxPlugin struct {
}

// HandleConnAccept check ip.
func (plugin *myRpcxPlugin) PreHandleRequest(ctx context.Context, r *protocol.Message) error {
	if r.IsOneway() {
		if rpcxContext, ok := ctx.(*share.Context); ok {
			rpcxContext.SetValue(myContextOneway, "true")
		}
	}
	return nil
}

type regFunc struct {
	Sevice   string
	Name     string
	Fn       interface{}
	Metadata string
}

var regs []*regFunc

func registerFunctionName(servicePath string, name string, fn interface{}, metadata string) error {
	if IsDev() {
		servicePath = fmt.Sprintf("%s_%d", servicePath, GetServerID())
		log.Printf("registerFunctionName as %s,%s metadata:%s", servicePath, name, metadata)
	}
	regs = append(regs, &regFunc{servicePath, name, fn, metadata})
	err := rpcxServer.RegisterFunctionName(servicePath, name, fn, metadata)
	logs.Print("registerFunctionName v2", servicePath, name, fn, metadata, "err", err)
	if err != nil {
		return err
	}
	return nil // registerV3FunctionName(servicePath, name, fn, metadata)
}

func OnEtcdReload() {
	if !IsUseRpcx() {
		logs.Importantf("The Server is Not Use Rpcx")
		return
	}
	if !rpcxStarted && rpcxServer == nil {
		return
	}
	logs.Importantf("收到重载通知, 重新加载rpcx发现服务...")
	//err := addEtcdRgistryPlugin(rpcxServer)
	//if err != nil {
	//	logs.PrintError("addEtcdRgistryPlugin", "error", err)
	//}
	err := addEtcdV3RgistryPlugin(rpcxServer)
	if err != nil {
		logs.PrintError("addEtcdV3RgistryPlugin", "error", err)
	}
}

func initRpcxServer() error {
	if !IsUseRpcx() {
		logs.Importantf("The Server is Not Use Rpcx")
		return nil
	}
	rpcxlog.SetLogger(&rpcxLogger{})

	rpcxServer = server.NewServer()
	const defaultPort = 21000
	port := GetServerPort()
	if port == 0 {
		port = rpcPorts.GetRpcPort(GetServerName())
	}

	if port == 0 {
		port = defaultPort
		logs.PrintImportant("本进程没有设置监听端口,使用默认端口", port)
		log.Print("本进程没有设置监听端口,使用默认端口", port)
	} else if port < 0 {
		log.Panic("有进程监听与进程同样端口,无法启动,先检查rpcPorts")
	}
	host, err := maddr.Extract("0.0.0.0", GetEtcdConfig().IpBlocks...)
	if err != nil {
		log.Panicf("获取本机地址失败: %s", err.Error())
		return err
	}
	//port = 20000
	address = mnet.HostPort(host, port)
	listener, err = net.Listen("tcp", address)
	logs.Importantf("Listen RpcxServer at: %s, err: %+v", address, err)
	//if port == defaultPort {
	for i := 1; i < 10; i++ {
		if err == nil {
			break
		}
		port += i * 100
		address = mnet.HostPort(host, port)
		listener, err = net.Listen("tcp", address)
		logs.Importantf("Listen RpcxServer at: %s, err: %+v", address, err)
	}
	//}
	if err != nil {
		logs.Errorf("Start rpcx MakeListener:%s failed:%s", address, err.Error())
		log.Panicf("Start rpcx MakeListener:%s failed:%s", address, err.Error())
		return err
	}

	rpcxServer.Plugins.Add(&myRpcxPlugin{})
	//
	//if err := addEtcdRgistryPlugin(rpcxServer); err != nil {
	//	logs.PrintError("addEtcdRgistryPlugin", "error", err)
	//	return err
	//}

	if err := addEtcdV3RgistryPlugin(rpcxServer); err != nil {
		logs.PrintError("addEtcdV3RgistryPlugin", "error", err)
	}
	rpcxStarted = true
	return nil
}

func startRpcServer() error {
	if !IsUseRpcx() {
		return nil
	}
	logs.Importantf("startRpcServer")
	go func() {
		err := rpcxServer.ServeListener(listener.Addr().Network(), listener)
		if err != nil {
			logs.PrintError("rpcxServer.ServeListener:", listener.Addr().String(), " failed:", err.Error())
		}
	}()

	log.Printf("启动rpcx服务: %s\n", address)
	logs.Importantf("启动rpcx服务: %s", address)
	return nil
}
func stopRegistryPlugin() {

	//plugins := registryPlugin
	//registryPlugin = nil
	//for _, pl := range plugins {
	//	_ = pl.Stop()
	//	logs.Importantf("Stop EtcdPlugin, Address:%s Service:%+v, Server:%+v", pl.ServiceAddress, pl.Services, pl.EtcdServers)
	//	log.Printf("Stop EtcdPlugin, Address:%s Service:%+v, Server:%+v\n", pl.ServiceAddress, pl.Services, pl.EtcdServers)
	//}
	pluginsV3 := registryPluginV3
	registryPluginV3 = nil
	for _, pl := range pluginsV3 {
		_ = pl.Stop()
		logs.Importantf("Stop EtcdPluginV3, Address:%s Service:%+v, Server:%+v", pl.ServiceAddress, pl.Services, pl.EtcdServers)
		log.Printf("Stop EtcdPluginV3, Address:%s Service:%+v, Server:%+v\n", pl.ServiceAddress, pl.Services, pl.EtcdServers)
	}
}
func stopRpcxServer() {
	if !IsUseRpcx() {
		logs.Importantf("The Server is Not Use Rpcx")
		return
	}

	if rpcxStarted {
		//_ = listener.Close()

		//logs.Importantf("Close Rpcx Listener: %s", listener.Addr().String())
		//log.Printf("Close Rpcx Listener: %s\n", listener.Addr().String())
		_ = rpcxServer.UnregisterAll()
		//_ = rpcxServer.Close()
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		_ = rpcxServer.Shutdown(ctx)
		rpcxStarted = false
	}
	stopRegistryPlugin()
	logs.Importantf("Stop RpcxServer at: %s", address)
	log.Printf("Stop RpcxServer at: %s\n", address)
}

func SetRpcxWorkPool(maxWorkers, maxCapacity int) {
	if rpcxServer != nil {
		server.WithPool(maxWorkers, maxCapacity, pond.MinWorkers(1))(rpcxServer)
	}
}

// ===============Register==============================
func RegisterRpcxHandler(fname string, handler func(context.Context, *NatsMsg) int32, needP2p bool) error {
	return RegisterRpcxHandlerBySName(GetServerName(), fname, handler, needP2p)
}

func RegisterRpcxHandlerBySName(sname, fname string, handler func(context.Context, *NatsMsg) int32, needP2p bool) error {
	logs.Print("RegisterRpcxHandlerBySName", "sname", sname, "fname", fname, "needP2p", needP2p)
	return doRegisterRpcxHandler(sname, fname, handler, needP2p)
}

func doRegisterRpcxHandler(sname, fname string, handler func(context.Context, *NatsMsg) int32, needP2p bool) error {
	if !IsUseRpcx() {
		logs.Importantf("The Server is Not Use Rpcx")
		return nil
	}

	if rpcxServer == nil {
		return errors.New("no rpcx server")
	}

	err := registerFunctionName(sname, fname, func(ctx context.Context, reqbuff *TRpcxMsg, replyBuff *TRpcxMsg) error {

		req := &NatsMsg{}
		if reqbuff != nil && reqbuff.Data != nil {
			_ = jsoniter.Unmarshal(reqbuff.Data, req)
		}
		isOneWay := false
		if rpcxContext, ok := ctx.(*share.Context); ok {
			if o, ok1 := rpcxContext.Value(myContextOneway).(string); ok1 && o == "true" {
				isOneWay = true
			}
			if o := rpcxContext.Value(server.RemoteConnContextKey); o != nil {
				req.RpcxConn, _ = o.(net.Conn)
			}
		}

		if !isOneWay {
			reply := &NatsMsg{}
			req.Reply = reply
		}

		start := time.Now()
		req.MsgBody.Func = fname

		var gpid int64
		ts := fmt.Sprintf("%02d%02d%02d", start.Hour(), start.Minute(), start.Second())
		if req.Sess.Trace == nil {
			req.Sess.Trace, gpid = logs.CreateTraceId(&req.Sess)
			req.Sess.Trace.TraceId += ":" + fname + ":" + ts
		} else {
			req.GetSession().Trace.TraceId = req.GetSession().Trace.TraceId +
				fmt.Sprintf(":%s%d:%s:%s", req.GetSession().GetServerFE(), req.GetSession().GetServerID(), fname, ts)
			gpid = logs.StoreTraceId(req.Sess.Trace, &req.Sess)
		}
		if req.Sess.Trace != nil {
			req.Sess.Trace.Request = req
		}
		result := handler(ctx, req)

		if req.Reply != nil && replyBuff != nil {
			replyBuff.Data, _ = jsoniter.Marshal(req.Reply)
		}
		now := time.Now()
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		stat.ReportStat("rp:rpcx.do."+fname+"."+s, int(result), now.Sub(start))
		if now.Unix()-req.GetSession().Time > 1 {
			logs.Bill("rpcx_slow", "RpcxCallFunction.timeout: %s.%s result: %d, platid:%d,Src:%s-%d,Uid:%d,time [req:%d, start:%d, now:%d,deal:%d,req-rsp:%d]",
				sname, fname, result, req.GetSession().GetPlatID(), req.GetSession().GetServerFE(), req.GetSession().GetServerID(), req.GetUid(),
				req.GetSession().Time, start.Unix(), now.Unix(), now.Unix()-start.Unix(), now.Unix()-req.GetSession().Time)
		}

		//用完移出traceid映射
		if gpid != 0 {
			logs.RemoveTraceId(gpid)
		}

		return nil
	}, "")
	if err != nil {
		logs.Warnf("doRegisterRpcxHandler.RegisterFunctionName:'%s','%s' failed:%s", sname, fname, err.Error())
		logs.Bill("RegisterFunctionName", "failed|doRegisterRpcxHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		return err
	}
	logs.Importantf("doRegisterRpcxHandler.RegisterRpcxHandler %s.%s", sname, fname)
	logs.Bill("RegisterFunctionName", "ok|doRegisterRpcxHandler.RegisterRpcxHandler %s.%s", sname, fname)

	if !needP2p {
		return nil
	}

	//必须把名字放在后面
	sname = fmt.Sprintf("%s.%d", sname, GetServerID())
	err = registerFunctionName(sname, fname, func(ctx context.Context, reqbuff *TRpcxMsg, replyBuff *TRpcxMsg) error {

		req := &NatsMsg{}
		if reqbuff != nil && reqbuff.Data != nil {
			_ = jsoniter.Unmarshal(reqbuff.Data, req)
		}
		isOneWay := false
		if rpcxContext, ok := ctx.(*share.Context); ok {
			if o, ok1 := rpcxContext.Value(myContextOneway).(string); ok1 && o == "true" {
				isOneWay = true
			}
		}
		if !isOneWay {
			reply := &NatsMsg{}
			req.Reply = reply
		}

		start := time.Now()
		req.MsgBody.Func = fname
		if req.Sess.RpcType == "" {
			req.Sess.RpcType = "rpcx"
		}
		var gpid int64
		ts := fmt.Sprintf("%02d%02d%02d", start.Hour(), start.Minute(), start.Second())
		if req.Sess.Trace == nil {
			req.Sess.Trace, gpid = logs.CreateTraceId(&req.Sess)
			req.Sess.Trace.TraceId += ":" + fname + ":" + ts
		} else {
			req.GetSession().Trace.TraceId = req.GetSession().Trace.TraceId +
				fmt.Sprintf(":%s%d:%s:%s", req.GetSession().GetServerFE(), req.GetSession().GetServerID(), fname, ts)
			gpid = logs.StoreTraceId(req.Sess.Trace, &req.Sess)
		}
		if req.Sess.Trace != nil {
			req.Sess.Trace.Request = req
		}
		result := handler(ctx, req)
		if req.Reply != nil && replyBuff != nil {
			replyBuff.Data, _ = jsoniter.Marshal(req.Reply)
		}
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		now := time.Now()
		stat.ReportStat("rp:rpcx.do."+fname+"."+s, int(result), now.Sub(start))

		if now.Unix()-req.GetSession().Time > 1 {
			logs.Bill("rpcx_slow", "RpcxCallFunction.timeout: %s.%s result: %d, platid:%d,Src:%s-%d,Uid:%d,time [req:%d, start:%d, now:%d,deal:%d,req-rsp:%d]",
				sname, fname, result, req.GetSession().GetPlatID(), req.GetSession().GetServerFE(), req.GetSession().GetServerID(), req.GetUid(),
				req.GetSession().Time, start.Unix(), now.Unix(), now.Unix()-start.Unix(), now.Unix()-req.GetSession().Time)
		}
		//用完移出traceid映射
		if gpid != 0 {
			logs.RemoveTraceId(gpid)
		}
		return nil
	}, "")
	if err != nil {
		logs.Warnf("doRegisterRpcxHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		logs.Bill("RegisterFunctionName", "failed|doRegisterRpcxHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		return err
	}
	logs.Importantf("doRegisterRpcxHandler.RegisterRpcxHandler %s.%s", sname, fname)
	logs.Bill("RegisterFunctionName", "ok|doRegisterRpcxHandler.RegisterRpcxHandler %s.%s", sname, fname)
	return err
}

// 只处理websocket请求, 不做回包
func RegisterRpcxWsHandler(sname, fname string, handler func(msg *NatsTransMsg) int32) error {
	if !IsUseRpcx() {
		logs.Importantf("The Server is Not Use Rpcx")
		return nil
	}

	if rpcxServer == nil {
		return errors.New("no rpcx server")
	}
	_fn := func(ctx context.Context, reqbuff *TRpcxMsg, replyBuff *TRpcxMsg) error {

		req := &NatsTransMsg{}
		if reqbuff != nil && reqbuff.Data != nil {
			_ = jsoniter.Unmarshal(reqbuff.Data, req)
		}
		start := time.Now()
		if req.Sess.RpcType == "" {
			req.Sess.RpcType = "rpcx"
		}

		result := handler(req)
		s := fmt.Sprintf("%s.%d", req.GetSession().SvrFE, req.GetSession().SvrID)
		//		logs.LogUser(req.Sess.Uid, "do %s.%s result: %d", sname, fname, result)
		stat.ReportStat("rp:rpcx.do."+fname+"-"+s, int(result), time.Now().Sub(start))
		return nil
	}
	err := registerFunctionName(sname, fname, _fn, "")
	if err != nil {
		logs.LogWarn("RegisterRpcxWsHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		logs.WriteBill("RegisterFunctionName", "failed|RegisterRpcxWsHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		return err
	}
	logs.WriteBill("RegisterFunctionName", "ok|rpcxServer.RegisterRpcxWsHandler %s.%s", sname, fname)
	sname = fmt.Sprintf("%s.%d", sname, GetServerID())
	err = registerFunctionName(sname, fname, _fn, "")
	if err != nil {
		logs.LogWarn("RegisterRpcxWsHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		logs.WriteBill("RegisterFunctionName", "failed|RegisterRpcxWsHandler.RegisterFunctionName %s, %s failed:%s", sname, fname, err.Error())
		return err
	}
	logs.Importantf("rpcxServer.RegisterRpcxWsHandler %s.%s", sname, fname)
	logs.WriteBill("RegisterFunctionName", "ok|rpcxServer.RegisterRpcxWsHandler %s.%s", sname, fname)
	return err
}

func rpcxCall(uid uint64, platId int32, sname, fname string, svrid int32, req interface{}, timeout time.Duration, needReply bool) (*TRpcxMsg, *errorMsg.ErrRsp) {
	if platId <= 0 {
		platId = GetPlatformId()
	}
	if svrid > 0 {
		sname = fmt.Sprintf("%s.%d", sname, svrid)
	}

	xclient := getXClient(uid, platId, sname)
	if xclient == nil {
		logs.Debugf("rpcxCall getXClient is nil, platId:%d, sname:%s, uid:%d", platId, sname, uid)
		return nil, errorMsg.NoService
	}

	start := time.Now()
	reqBuf := &TRpcxMsg{}
	var err error
	reqBuf.Data, err = jsoniter.Marshal(req)
	if err != nil {
		logs.Errorf("rpcxClient.Call %s.%s failed:%s,uid:%d", sname, fname, err.Error(), uid)
		return nil, errorMsg.ReqError.Copy(err)
	}
	//timeout = time.Second
	if timeout <= 0 || timeout > 5*time.Minute {
		timeout = GetRpcCallTimeout()
	}
	ctx := context.Background()
	var rspBuf *TRpcxMsg
	if needReply {
		rspBuf = &TRpcxMsg{}
		ctx, _ = context.WithTimeout(context.Background(), timeout)
		err = xclient.Call(ctx, fname, reqBuf, rspBuf)
	} else {
		err = xclient.Call(ctx, fname, reqBuf, nil)
	}
	ret := int(ESMR_SUCCEED)
	cost := time.Since(start)

	if !strings.Contains(fname, "HeartBeat") && IsDebug() {
		logs.Infof("rpc uid:%d plat:%d, %s.%s,%t cost:%+v req:%+v, error:%+v", uid, platId, sname, fname, needReply, cost, req, err)
	}
	//logs.LogDebug("RpcxCall fname:%s, err:%+v", fname, err)
	var errs *errorMsg.ErrRsp
	if err != nil {
		//不打conn的404错误
		errs = errorMsg.ReqError.Copy(err)
		if err == context.DeadlineExceeded {
			errs = errorMsg.TimeOut.Copy(err) //.Return("x-dead")
		} else if err == client.ErrXClientNoServer || err == client.ErrXClientShutdown || err == client.ErrServerUnavailable {
			errs = errorMsg.NoService.Copy(err) //.Return("x-nos")
		} else if strings.Contains(err.Error(), "rpcx: can't find") ||
			strings.Contains(err.Error(), "connect: connection refused") ||
			strings.Contains(err.Error(), "dial tcp") {
			errs = errorMsg.NoService.Copy(err) //.Return("x-dial")
		} else {
			errs = errorMsg.ReqError.Copy(err) //.Return("Unknown")
		}
		ret = int(ESMR_FAILED)
	}

	stat.ReportStat("rp:rpcx.call."+sname+"."+fname+"."+strconv.Itoa(int(platId)), ret, cost)
	return rspBuf, errs
}

func CallRpcxForTrans(platId int32, sess *Session, sname, fname string, svrid int32, req interface{}, timeout time.Duration) (*NatsTransMsg, *errorMsg.ErrRsp) {
	rsp := &NatsTransMsg{}
	rspBuf, errs := rpcxCall(sess.GetUid(), platId, sname, fname, svrid, req, timeout, true)
	if errs != nil {
		// logs.Errorf("CallRpcxForTrans plat:%d, %s.%s.%d, Sess:%+v failed:%s", platId, sname, fname, svrid, sess, err.Error())
		return rsp, errs
	}
	err := jsoniter.Unmarshal(rspBuf.Data, rsp)
	if err != nil {
		errs = errorMsg.RspError.Copy(err)
		logs.Errorf("CallRpcxForTrans  plat:%d,%s.%s.%d Unmarshal rsp (%+v),rsp(%+v) failed:%s", platId, sname, fname, svrid, string(rspBuf.Data), string(rspBuf.Data), err.Error())
	}
	return rsp, errs
}

func CallRpcx(platId int32, sname, fname string, svrid int32, req interface{}, timeout time.Duration) (*NatsMsg, *errorMsg.ErrRsp) {
	rsp := &NatsMsg{}
	rspBuf, errs := rpcxCall(0, platId, sname, fname, svrid, req, timeout, true)
	if errs != nil {
		//logs.Errorf("CallRpcx  plat:%d,%s.%s.%d  failed:%s", platId, sname, fname, svrid, err.Error())
		return rsp, errs
	}
	err := jsoniter.Unmarshal(rspBuf.Data, rsp)
	if err != nil {
		errs = errorMsg.RspError.Copy(err)
		logs.Errorf("CallRpcx  plat:%d,%s.%s.%d Unmarshal rsp (%+v) failed:%s", platId, sname, fname, svrid, string(rspBuf.Data), err.Error())
	}
	return rsp, errs
}

func SendRpcx(uid uint64, platId int32, sname, fname string, svrid int32, req interface{}) *errorMsg.ErrRsp {

	if svrid <= -999 {
		err := BroadcastRpcx(uid, platId, sname, fname, req)
		if err != nil {
			// logs.Errorf("SendRpcx  plat:%d,%s.%s.%d failed:%s", platId, sname, fname, svrid, err.Error())
			return err
		}
	} else {
		_, err := rpcxCall(uid, platId, sname, fname, svrid, req, GetRpcCallTimeout(), false)
		if err != nil {
			// logs.Errorf("SendRpcx  plat:%d,%s.%s.%d failed:%s", platId, sname, fname, svrid, err.Error())
			return err
		}
	}

	return nil
}

func BroadcastRpcx(uid uint64, platId int32, sname, fname string, req interface{}) *errorMsg.ErrRsp {
	if platId <= 0 {
		platId = GetPlatformId()
	}

	xclient := getXClient(uid, platId, sname)
	if xclient == nil {
		logs.Debugf("rpcxCall getXClient is nil, platId:%d, sname:%s, uid:%d", platId, sname, uid)
		return errorMsg.NoService
	}

	reqBuf := &TRpcxMsg{}
	var err error
	reqBuf.Data, err = jsoniter.Marshal(req)
	if err != nil {
		logs.Errorf("rpcxClient.Call %s.%s failed:%s,uid:%d", sname, fname, err.Error(), uid)
		return errorMsg.ReqError.Copy(err)
	}
	ctx := context.Background()
	err = xclient.Broadcast(ctx, fname, reqBuf, nil)
	var errs *errorMsg.ErrRsp
	if err != nil {
		//不打conn的404错误
		errs = errorMsg.ReqError.Copy(err)
		if err == context.DeadlineExceeded {
			errs = errorMsg.TimeOut.Copy(err) //.Return("x-dead")
		} else if err == client.ErrXClientNoServer || err == client.ErrXClientShutdown || err == client.ErrServerUnavailable {
			errs = errorMsg.NoService.Copy(err) //.Return("x-nos")
		} else if strings.Contains(err.Error(), "rpcx: can't find") ||
			strings.Contains(err.Error(), "connect: connection refused") ||
			strings.Contains(err.Error(), "dial tcp") {
			errs = errorMsg.NoService.Copy(err) //.Return("x-dial")
		} else {
			errs = errorMsg.ReqError.Copy(err) //.Return("Unknown")
		}
		//ret = int(ESMR_FAILED)
	}
	return errs
}

type rpcxLogger struct {
}

func wlog(tag, msg string) {
	//测试环境不用关注这些信息
	//if GetGlobalConfig().IsTestServer {
	//	return
	//}
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	_, fileName := path.Split(file)
	f := fmt.Sprintf("rpcx[%s:%d] %s: ", fileName, line, tag)
	msg = f + msg
	trace := logs.GetTraceId()
	if trace != nil {
		msg = msg + " request:" + utils.AutoToString(trace.Request)
	}
	msg2 := ":stack:" + string(debug.Stack())
	if tag == "Debug" || tag == "Debugf" {
		logs.LogDebug(msg)
		logs.LogDebug(msg2)
	} else if tag == "Info" || tag == "Infof" {
		logs.Importantf(msg)
		logs.LogDebug(msg2)
	} else {
		logs.Importantf(msg)
		logs.Importantf(msg2)
		if IsDebug() {
			log.Printf(msg + "\n")
			//log.Printf(msg2 + "\n")
		}
	}
}
func (l *rpcxLogger) Debug(v ...interface{}) {
	wlog("Debug", fmt.Sprint(v...))
}

func (l *rpcxLogger) Debugf(format string, v ...interface{}) {
	wlog("Debugf", fmt.Sprintf(format, v...))
}

func (l *rpcxLogger) Info(v ...interface{}) {
	wlog("Info", fmt.Sprint(v...))
}

func (l *rpcxLogger) Infof(format string, v ...interface{}) {
	wlog("Infof", fmt.Sprintf(format, v...))
}

func (l *rpcxLogger) Warn(v ...interface{}) {
	wlog("Warn", fmt.Sprint(v...))
}

func (l *rpcxLogger) Warnf(format string, v ...interface{}) {
	wlog("Warnf", fmt.Sprintf(format, v...))
}

func (l *rpcxLogger) Error(v ...interface{}) {
	wlog("Error", fmt.Sprint(v...))
}

func (l *rpcxLogger) Errorf(format string, v ...interface{}) {
	wlog("Errorf", fmt.Sprintf(format, v...))
}

func (l *rpcxLogger) Fatal(v ...interface{}) {
	logs.Errorf("rpcx: " + fmt.Sprint(v...))
	os.Exit(1)
}

func (l *rpcxLogger) Fatalf(format string, v ...interface{}) {
	wlog("Fatalf", fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *rpcxLogger) Panic(v ...interface{}) {
	log.Panic(v...)
}

func (l *rpcxLogger) Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}
