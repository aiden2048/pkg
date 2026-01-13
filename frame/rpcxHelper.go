package frame

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aiden2048/pkg/public/errorMsg"
	jsoniter "github.com/json-iterator/go"
	"github.com/smallnest/rpcx/client"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/stat"

	"time"
)

func GenReqSubject(mod string, cmd string, svrid int32) string {
	if svrid <= 0 {
		return fmt.Sprintf("%d.msg.%s.%s", GetPlatformId(), mod, cmd)
	} else {
		return fmt.Sprintf("%d.msg.%s.%d.%s", GetPlatformId(), mod, svrid, cmd)
	}
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
	timeout := GetRpcCallTimeout()

	if len(args) > 0 {
		xtime := args[0]
		timeout = time.Duration(xtime) * time.Second
	}
	platId := int32(0)
	if len(args) > 1 {
		platId = args[1]
	}
	if platId <= 0 {
		platId = GetPlatformId()
	}
	//检查通过rpcx请求
	if !CheckRpcxService(platId, mod, svrid, cmd) {
		return nil, errorMsg.NoService.Line()
	}
	m, e := CallRpcx(platId, mod, cmd, svrid, req, timeout)
	logs.PrintDebug(req.GetSession(), "call ", platId, mod, cmd, svrid, "err", e)
	return m, e
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

func CallRpcx(platId int32, uid uint64, sname, fname string, svrid int32, req interface{}, timeout time.Duration) (*NatsMsg, *errorMsg.ErrRsp) {
	rsp := &NatsMsg{}
	rspBuf, errs := rpcxCall(uid, platId, sname, fname, svrid, req, timeout, true)
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
