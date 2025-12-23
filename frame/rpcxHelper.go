package frame

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiden2048/pkg/public/errorMsg"

	"github.com/aiden2048/pkg/frame/logs"

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
