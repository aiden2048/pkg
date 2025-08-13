package natsHandle

import (
	"fmt"
	"strings"
	"time"

	"github.com/aiden2048/pkg/frame/stat"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	jsoniter "github.com/json-iterator/go"
)

func HandleLogic[R any, T any](r *frame.NatsMsg, logic func(req *R) (resp *T, err *errorMsg.ErrRsp), chkSess ...bool) int32 {
	if r == nil {
		logs.Error("req is nil")
		return frame.ESMR_FAILED
	}

	// 解析请求
	var req R
	start := time.Now()

	_err := jsoniter.Unmarshal(r.GetParam(), &req)
	if _err != nil {
		err := errorMsg.ReqParamError.Return("GetParam")
		//errstr := err.ErrorStr()
		//if r.GetSession().SvrFE != "conn" {
		//	errstr = err.Error()
		//}
		r.SendResponse(err.ErrorNo(), err.ErrorStr(), err.Params)
		return frame.ESMR_FAILED
	}

	if len(chkSess) > 0 && chkSess[0] && (r.GetSession().GetAppID() <= 0 || r.GetUid() <= 0) && !strings.HasPrefix(r.GetFunc(), "NA.") {
		logs.PrintErr("请求没有初始化Session信息, 检查调用者代码:", r.GetSession(), "func", r.GetMod(), r.GetFunc(), "check", r.GetCheck())
	}
	needLog := !strings.HasPrefix(r.GetFunc(), "DoCode")
	if needLog {
		remote := ""
		if r.Conn != nil {
			ip, _ := r.Conn.GetClientIP()
			if ip != nil {
				remote = ip.String()
			}
		} else if r.RpcxConn != nil {
			remote = r.RpcxConn.RemoteAddr().String()
		}
		logs.PrintInfo(r.GetFunc(), "rpc-begin", r.GetSession(), "req", &req, "conn", remote)
	}

	resp, err := logic(&req)

	if needLog {
		defer func() {
			//errs := ""
			//if err != nil {
			//	errs = err.Error()
			//}
			logs.PrintInfo(r.GetFunc(), "rpc-end", "cost(ms)", time.Now().Sub(start)/time.Millisecond, "err", err, "resp", resp, "req", req)
		}()
	}
	if err != nil {
		logs.PrintBill("user_app_err_code", r.GetSession().GetAppID(), r.GetSession().GetUid(), "errCode", err.ErrorNo(), "errMsg", err.ErrorStr())
		if r.GetSession().SvrFE == "conn" && err.ErrorNo() >= 10000 && r.GetSession().Channel <= 0 {
			logs.PrintError("rpc-end", r.GetFunc(), "from", r.GetSession(), "返回不应该返回到connd的错误码", err)
		}
		stat.ReportStat(fmt.Sprintf("rp:rpcx_do_err.%s", err.Error()), 1, time.Now().Sub(start))
		r.SendResponse(err.ErrorNo(), err.ErrorStr(), err.Params)
		if err.IsEqual(errorMsg.TimeOut) {
			return frame.ESMR_TIMEOUT
		}
		return frame.ESMR_FAILED
	}

	/*if frame.IsTestUid(r.GetUid()) {
		r.SendResponseEcrypt(0, "", resp)
	} else {
		r.SendResponse(0, "", resp)
	}*/
	if r.GetSession().Channel > 0 {
		r.GetSession().Check = r.GetCheck()
	}
	r.SendResponse(0, "", resp)
	return frame.ESMR_SUCCEED
}

func HandleEvent[R any](r *frame.NatsTransMsg, logic func(req *R) (err *errorMsg.ErrRsp)) int32 {
	if r == nil {
		logs.Errorf("HandleEvent is nil")
		return frame.ESMR_FAILED
	}

	// 解析请求
	var req R
	_err := jsoniter.Unmarshal(r.MsgBody, &req)
	if _err != nil {
		logs.Errorf("HandleEvent is err:%+v,r.MsgBody:%+v", _err, r.MsgBody)
		return frame.ESMR_FAILED
	}

	err := logic(&req)
	if err != nil {
		return frame.ESMR_FAILED
	}
	return frame.ESMR_SUCCEED
}

func HandleConfig[R any](b []byte, logic func(req *R) (err *errorMsg.ErrRsp)) int32 {

	// 解析请求
	var req R
	_err := jsoniter.Unmarshal(b, &req)
	if _err != nil {
		logs.Errorf("HandleEvent is err:%+v,r.MsgBody:%+v", _err, string(b))
		return frame.ESMR_FAILED
	}
	err := logic(&req)
	if err != nil {
		return frame.ESMR_FAILED
	}
	return frame.ESMR_SUCCEED
}

func Request[T any, H any](group, funcName string, req_param *T, needRsp bool, pids []int32) (rsppara *H, erro *errorMsg.ErrRsp) {
	platId := int32(0)
	if len(pids) > 0 {
		platId = pids[0]
	}

	timeout := frame.GetCallTimeout()
	if len(pids) > 1 && pids[1] > 0 {
		timeout = pids[1]
	}
	svrId := int32(-1)
	if len(pids) > 2 {
		svrId = pids[2]
	}
	req := frame.NatsMsg{Sess: *frame.NewSessionOnly()}

	req.MsgBody.Func = funcName
	var err error
	req.MsgBody.Param, err = jsoniter.Marshal(&req_param)
	if err != nil {
		logs.Errorf("%s->%s failed:%v", group, funcName, err)
		return nil, errorMsg.RspError.Copy(err)
	}
	if needRsp { //需要返回
		rsp, errs := frame.RpcxCall(group, svrId, funcName, &req, timeout, platId)
		if errs != nil {
			logs.Errorf("%s->%s failed:%v", group, funcName, errs)
			return nil, errs
		}
		if rsp.GetMsgErrNo() != 0 {
			parmars := []any{}
			jsoniter.Unmarshal(rsp.GetParam(), &parmars)
			return nil, &errorMsg.ErrRsp{Ret: rsp.GetMsgBody().ErrNo, Str: rsp.GetMsgErrStr(), Params: parmars}
		}
		rsppara = new(H)
		err = jsoniter.Unmarshal(rsp.GetParam(), &rsppara)
		if err != nil {
			logs.PrintErr("Request", group, funcName, "respone from", rsp.Sess, "err:", err, "Unmarshal", rsp.GetParam())
			return nil, errorMsg.RspError.Copy(err)
		}
		return rsppara, nil
	} else {
		req.Sess.Channel = 0
		err := frame.RpcxSend(group, svrId, funcName, &req, platId)
		if err != nil {
			logs.Errorf("%s->%s failed:%s", group, funcName, err)
			return nil, errorMsg.RspError.Copy(err)
		}
		return nil, nil
	}
}

func RequestWithSess[T any, H any](sess *frame.Session, group, funcName string, req_param *T, needRsp bool, pids []int32) (rsppara *H, erro *errorMsg.ErrRsp) {
	platId := int32(0)
	if len(pids) > 0 {
		platId = pids[0]
	}
	timeout := frame.GetCallTimeout()
	if len(pids) > 1 && pids[1] > 0 {
		timeout = pids[1]
	}
	svrId := int32(-1)
	if len(pids) > 2 {
		svrId = pids[2]
	}

	req := frame.NatsMsg{}
	if sess != nil {
		req.Sess = *sess
		if req.GetSession().GetUid() == 0 && req.GetSession().GetServerFE() == "conn" && strings.HasPrefix(req.GetSession().Cmd, "NA.") {
			req.GetSession().Uid = 1001
			logs.PrintDebug("从非登录态请求转后端, 自动补齐UID")
		}

	} else {
		req.Sess = *frame.NewSessionOnly()
	}

	req.MsgBody.Func = funcName
	var err error
	req.MsgBody.Param, err = jsoniter.Marshal(&req_param)
	if err != nil {
		logs.Errorf("%s->%s failed:%s", group, funcName, err)
		return nil, errorMsg.RspError.Copy(err)
	}
	if needRsp { //需要返回
		rsp, errs := frame.RpcxCall(group, svrId, funcName, &req, timeout, platId)
		var respBody frame.MsgBody
		if rsp != nil {
			respBody = rsp.MsgBody
		}
		logs.PrintInfo("rpc-call", group, svrId, funcName, platId, "timeout", timeout, "req", &req.MsgBody, "Resp", respBody, "err", errs)
		if errs != nil {
			if errs.IsEqual(errorMsg.TimeOut) {
				logs.PrintError("rpc-call", group, svrId, funcName, platId, "timeout", timeout, "req", &req.MsgBody, "Resp", rsp, "err", errs)
			}
			return nil, errs
		}
		if rsp.GetMsgErrNo() != 0 {
			parmars := []any{}
			jsoniter.Unmarshal(rsp.GetParam(), &parmars)
			return nil, &errorMsg.ErrRsp{Ret: rsp.GetMsgBody().ErrNo, Str: rsp.GetMsgErrStr(), Params: parmars}

		}
		rsppara = new(H)
		err = jsoniter.Unmarshal(rsp.GetParam(), &rsppara)
		if err != nil {
			//logs.Errorf("%s->%s rsp:%+v, param:%s, failed:%s", group, funcName, rsp, string(rsp.GetParam()), err)
			logs.PrintError("RequestWithSess", group, funcName, "respone from", rsp.Sess, "err:", err, "Unmarshal", rsp.GetParam())
			return nil, errorMsg.RspError.Copy(err)
		}
		return rsppara, nil
	} else {
		req.Sess.Channel = 0
		errs := frame.RpcxSend(group, svrId, funcName, &req, platId)
		logs.PrintInfo("rpc-send", group, svrId, funcName, &req, timeout, platId, "err", errs)
		if errs != nil {
			//logs.Errorf("%s->%s failed:%s", group, funcName, err)
			return nil, errs
		}
		return nil, nil
	}
}
