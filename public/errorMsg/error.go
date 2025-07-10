package errorMsg

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"
)

func New(format string, v ...interface{}) error {
	msg := fmt.Sprintf(format, v...)
	return errors.New(msg)
}

type ErrRsp struct {
	Ret int32  `json:"ret"`
	Str string `json:"str"`
	Msg string `json:"msg"`

	Params []interface{}
}

var _errMap map[int32]string
var lock sync.Mutex

// 固定的错误码, 不允许重复
func NewConstError(ret int32, msgfmt string, p ...interface{}) *ErrRsp {

	lock.Lock()
	defer lock.Unlock()
	if _errMap == nil {
		_errMap = make(map[int32]string)
	}
	if v, ok := _errMap[ret]; ret > 0 && ok {
		fmt.Printf("发现重复定义错误码:%d, %s --> %s", ret, v, msgfmt)
		log.Panicf("发现重复定义错误码:%d, %s --> %s", ret, v, msgfmt)
	}
	_errMap[ret] = msgfmt

	go func() {
		time.Sleep(30 * time.Second)
		logs.WriteBill("ErrorCode", ",%d,%s,%+v", ret, msgfmt, p)
	}()

	// 有fmt參數時使用
	if len(p) == 0 {
		return &ErrRsp{Ret: ret, Msg: msgfmt, Params: nil}
	}
	return &ErrRsp{Ret: ret, Msg: msgfmt, Params: p}
}

// service 自定义错误, 规定在1000以上
func NewError(ret int32, msgfmt string, p ...interface{}) *ErrRsp { // 有fmt參數時使用
	ret = ret + 10000
	lock.Lock()
	defer lock.Unlock()
	if _errMap == nil {
		_errMap = make(map[int32]string)
	}
	if v, ok := _errMap[ret]; ret > 0 && ok {
		file1 := utils.GetCallFile(1)
		file2 := utils.GetCallFile(2)
		go func(f1, f2 string) {
			time.Sleep(30 * time.Second)
			logs.LogError("发现重复定义错误码:%d, %s --> %s at: %s -> %s", ret, v, msgfmt, file2, file1)
		}(file1, file2)
	}
	_errMap[ret] = msgfmt

	if len(p) == 0 {
		return &ErrRsp{Ret: ret, Msg: msgfmt, Params: nil}
	}
	return &ErrRsp{Ret: ret, Msg: msgfmt, Params: p}
}

func (this *ErrRsp) IsEqual(err *ErrRsp) bool {
	if this == nil {
		if err == nil {
			return true
		}
		return false
	}
	return err != nil && err.Ret == this.Ret
}
func (this *ErrRsp) Return(n interface{}, p ...interface{}) *ErrRsp { // 兼容 error 類型
	if this == nil {
		return NewConstError(-100, "")
	}

	line := utils.GetLine(2)
	e := &ErrRsp{Ret: this.Ret, Msg: this.Msg, Params: this.Params}
	if this.Str != "" {
		e.Str = this.Str + "," + utils.AutoToString(n) + "-" + utils.AutoToString(line)
	} else {
		e.Str = utils.AutoToString(n) + "-" + utils.AutoToString(line)
	}

	if len(p) > 0 {
		e.Params = append(e.Params, p...)
	}
	return e
}
func (this *ErrRsp) Line(prefixes ...string) *ErrRsp { // 兼容 error 類型
	if this == nil {
		return NewConstError(-100, "")
	}
	var prefix string
	if len(prefixes) > 0 {
		prefix = prefixes[0]
	}
	n := utils.GetLine(2)
	e := &ErrRsp{Ret: this.Ret, Msg: this.Msg, Params: this.Params}
	if this.Str != "" {
		e.Str = this.Str + "," + prefix + "-" + utils.AutoToString(n)
	} else {
		e.Str = prefix + "-" + utils.AutoToString(n)
	}
	return e
}

// 设置param, 会传给调用者
func (this *ErrRsp) Copy(p ...interface{}) *ErrRsp { // 兼容 error 類型
	if this == nil {
		return NewConstError(-100, "")
	}

	e := &ErrRsp{Ret: this.Ret, Msg: this.Msg, Str: this.Str}
	if len(p) > 0 {
		e.Params = append(e.Params, p...)
	}
	return e
}

// 返回给客户端的主错误码
func (this *ErrRsp) ErrorNo() int32 { // 兼容 error 類型
	if this == nil {
		return 0
	}
	return this.Ret /**1000 + this.Str%1000*/
}

// 返回详细的错误信息, 用于打日志等
func (this *ErrRsp) Error() string { // 兼容 error 類型
	//if this.Params != nil {
	//	return fmt.Sprintf(this.Msg, *this.Params...)
	//}
	if this == nil {
		return "-"
	}
	ss := fmt.Sprintf("%d-%s-%s", this.Ret, this.Msg, this.Str)
	if len(this.Params) > 0 {

		for _, pp := range this.Params {
			ss += "\n" + utils.AutoToString(pp)
		}
	}
	return ss
}

// 返回给客户端的附加错误信息
func (this *ErrRsp) ErrorStr() string { // 兼容 error 類型
	if this == nil {
		return "-"
	}
	return this.Str /**1000 + this.Str%1000*/
}

// 组合 ret 和 str , 一般给三方
func (this *ErrRsp) ErrorRet() string { // 兼容 error 類型
	if this == nil {
		return "-"
	}
	return fmt.Sprintf("%d-%s", this.Ret, this.Str)
}

// 后台发送中文
func (this *ErrRsp) SendMsg(msg ...string) *ErrRsp {
	if this == nil {
		return NewConstError(-100, "")
	}
	var prefix string
	if len(msg) > 0 {
		prefix = msg[0]
	}
	n := utils.GetLine(2)
	e := &ErrRsp{Ret: this.Ret, Msg: this.Msg, Params: this.Params}
	if this.Str != "" {
		e.Str = this.Str + "," + this.Msg + "," + prefix + utils.AutoToString(n)
	} else {
		e.Str = this.Msg + "," + prefix + utils.AutoToString(n)
	}
	return e
}
