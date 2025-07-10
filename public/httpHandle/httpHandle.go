package httpHandle

import (
	"time"

	"github.com/aiden2048/pkg/frame/runtime"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/go-playground/form/v4"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var decoder = form.NewDecoder()

func HandleLogic[R any, T any](fastCtx *fasthttp.RequestCtx, msg *frame.NatsMsg, logic func(req *R) (resp *T, err *errorMsg.ErrRsp), chkSess ...bool) {
	if fastCtx == nil {
		logs.Error("req is nil")
		return
	}
	// 解析请求
	var req R
	start := time.Now()

	if fastCtx.IsGet() {
		args := fastCtx.QueryArgs()
		params := make(map[string][]string)
		args.VisitAll(func(key, value []byte) {
			params[string(key)] = []string{string(value)}
		})
		if err := decoder.Decode(&req, params); err != nil {
			err := errorMsg.ReqParamError.Return("GetParam")
			fastCtx.SetStatusCode(fasthttp.StatusBadRequest)
			SendResponse(fastCtx, err.ErrorNo(), err.ErrorStr(), err.Params)
			return
		}
	} else {
		_err := jsoniter.Unmarshal(fastCtx.Request.Body(), &req)
		if _err != nil {
			err := errorMsg.ReqParamError.Return("GetParam")
			fastCtx.SetStatusCode(fasthttp.StatusBadRequest)
			SendResponse(fastCtx, err.ErrorNo(), err.ErrorStr(), err.Params)
			return
		}
	}

	logs.PrintInfo(string(fastCtx.Path()), "func", msg.GetFunc(), "begin", msg, "req", &req)
	resp, err := logic(&req)
	defer func() {
		logs.PrintInfo(string(fastCtx.Path()), "func", msg.GetFunc(), "end", "cost(ms)", time.Since(start).Milliseconds(), "err", err, "resp", resp, "req", req)
	}()

	if err != nil {
		SendResponse(fastCtx, err.ErrorNo(), err.Error(), err.Params)
		return
	}
	SendResponse(fastCtx, 0, "", resp)
	return
}

func SendResponse(fastCtx *fasthttp.RequestCtx, err_no int32, err_str string, rspparam interface{}) {
	rsp := Response{}
	rsp.ErrNo = err_no
	rsp.ErrStr = err_str
	rsp.Trace = runtime.GetTrace().TraceId
	if rspparam != nil {
		rsp.Param, _ = jsoniter.Marshal(rspparam)
	}
	body, _ := jsoniter.Marshal(rsp)
	fastCtx.SetBody(body)
}

type Response struct {
	ErrNo  int32               `json:"code"`    // 错误码
	ErrStr string              `json:"message"` // 错误内容
	Param  jsoniter.RawMessage `json:"data"`
	Trace  string              `json:"trace"`
}
