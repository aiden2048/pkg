// request.go 提供Service接口，包括请求定义（Msg、Cookie、Get/Post参数）、SendResponse/SendHttpResponse
package frame

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/session_def"
	"github.com/aiden2048/pkg/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

const (
	ESMR_SUCCEED int32 = 0
	ESMR_FAILED  int32 = 1
	ESMR_TIMEOUT int32 = 2

	LangDefault = 0  // 默认
	LangEN      = 1  // 英语
	LangVND     = 2  // 越南语
	LangTHA     = 3  // 泰语
	LangTC      = 4  // 繁体
	LangIDR     = 5  // 印尼语
	LangBR      = 11 // 葡萄牙
	LangES      = 12 // 西班牙语
	LangKR      = 60 // 韩语
	LangPHL     = 61 // 菲律宾
	LangJP      = 62 // 日本
)

// 从登录态获取的信息
type LoginInfo = session_def.LoginInfo

/*
	struct {
		SessionID  int64  `json:"SessionID,omitempty"`
		PriKey     string `json:"PriKey,omitempty"`
		AppID      int32  `json:"AppID,omitempty"`
		ChannelGetChannelNo     int32  `json:"ChannelGetChannelNo,omitempty"` // 用户渠道号
		LoginType  int32  `json:"utype"`
		LoginAddr  string `json:"LoginAddr"`            //登录的ip
		RemoteAddr string `json:"RemoteAddr,omitempty"` //当前访问ip
		Mac        string `json:"Mac,omitempty"`
		Src        string `json:"src,omitempty"` // checkLogin 赋值
		Os         int32  `json:"os"`            // 系统类型
		Vip        int32  `json:"vip"`           //vip等级, 登录时候是什么就是什么
		Tag        int32  `json:"tag"`           //玩家tag
		Lang       int32  `json:"lang"`          //语言
		IsFree     bool   `json:"is_free"`       // 是否试玩
		RealUid    uint64 `json:"real_uid"`      // 试玩玩家的真实UID
	}
*/
type Session = session_def.Session

/*struct {
	Uid          uint64     `json:"Uid,omitempty"`
	AppId        int32      `json:"AppID,omitempty"`
	PlatId       int32      `json:"PlatID,omitempty"`
	SvrFE        string     `json:"SvrFE,omitempty"`
	SvrID        int32      `json:"SvrID,omitempty"`
	Channel      uint64     `json:"Channel,omitempty"` //用户连接的管道ID
	Cmd          string     `json:"Cmd,omitempty"`
	Time         int64      `json:"Time,omitempty"`
	LoginInfo    *LoginInfo `json:"LoginInfo,omitempty"`
	Seq          int64      `json:"Seq"`
	Check        string     `json:"-"`
	X_RemoteAddr string     `json:"-"`
	RpcType      string     `json:"rpc_type"`
}*/
/*
func NewSession( args ...interface{}) *Session {
	sess := &Session{}
	sess.PlatId = GetPlatformId()
	sess.SvrFE = GetServerName()
	sess.SvrID = GetServerID()
	sess.Time = time.Now().Unix()
	if len(args)>0{
		sess.Uid = uint64(utils.AutoToInt(args[0]))
	}
	if len(args)>1{

		sess.AppId  = int32(utils.AutoToInt(args[1]))
	}

	return sess
}*/

func NewSession(uid uint64, appid int32) *Session {
	sess := &Session{}
	trace := logs.GetTraceId(sess)
	// trace绑定了线程起始的session，copy线程起始session
	if trace.Sess != nil && trace.Sess != sess {
		sess = trace.Sess.Copy()
	}

	sess.PlatId = GetPlatformId()
	sess.SvrFE = GetServerName()
	sess.SvrID = GetServerID()
	sess.Times = time.Now()
	sess.Time = sess.Times.Unix()

	sess.Uid = uid
	sess.AppId = appid
	sess.Trace = trace

	if sess.Trace.Uid != uid || sess.Trace.AppId != appid {
		sess.Trace.AppId, sess.Trace.Uid = appid, uid
	}

	return sess
}
func NewSessionOnlyApp(appid int32) *Session {
	return NewSession(1000, appid)
}
func NewSessionOnly() *Session {
	return NewSession(1000, 1000)
}

type MsgBody struct {
	Mod   string      `json:"_mod,omitempty"`
	Func  string      `json:"_func,omitempty"`
	Check interface{} `json:"_check,omitempty"`

	Seq   interface{} `json:"_sn,omitempty"`
	Token string      `json:"_token,omitempty"`
	Lang  int32       `json:"_lang,omitempty"`
	//Ver    string          `json:"_ver,omitempty"`
	Type      string `json:"_type,omitempty"`
	SvrT      string `json:"_svr_t,omitempty"`
	SvrID     int32  `json:"_svr_id,omitempty"`
	ErrNo     int32  `json:"_errno"`            // 错误码
	ErrStr    string `json:"_errstr,omitempty"` // 错误内容
	PopupNo   int32  `json:"_popno,omitempty"`  // 弹窗样式
	PopupJump int32  `json:"_popjp,omitempty"`  // 弹窗跳转
	Trace     string `json:"trace"`
	//Pkgid  int32               `json:"_pkgid,omitempty"`
	//Appid  int32               `json:"_appid,omitempty"`
	Param jsoniter.RawMessage `json:"_param,omitempty"`
	//兼容chess协议的_dst
	Dst jsoniter.RawMessage `json:"_dst,omitempty"`
	//兼容老协议, 目前只有online和chess用到,其他新模块不建议使用
	//Events jsoniter.RawMessage `json:"_events,omitempty"`

	Filters []string `json:"_filters,omitempty"` //回包过滤
}
type NatsMsg struct {
	Sess    Session `json:"Sess,omitempty"`
	Type    string  `json:"type,omitempty"`
	Encrypt bool    `json:"encrypt"`
	MsgBody MsgBody `json:"Body,omitempty"`
	MsgData []byte  `json:"Data,omitempty"`
	Conn    *nats.Conn
	NatsMsg *nats.Msg `json:"-,omitempty"`
	//NatsMsgStr *string   `json:"-"`
	Reply *NatsMsg `json:"-,omitempty"`

	ExtValue int32 `json:"-,omitempty"`

	Cookie      []string `json:"cookie"`
	IsSetCookie bool     `json:"-,omitempty"`
	//Host   string   `json:"host"`
	RpcxConn net.Conn
}

func (m *NatsMsg) AddCookie(k, v string, maxAges ...int) {
	c := Cookie{Name: k, Value: v, SameSite: SameSiteNoneMode, Secure: true, Path: "/"}
	//c.Expires = time.Now().Add(time.Hour * 24 * 30)
	c.MaxAge = 60 * 60 * 24 * 45
	if len(maxAges) > 0 {
		c.MaxAge = maxAges[0]
	}

	m.Cookie = append(m.Cookie, c.String())
	//delete cookie
	if v == "" && isCookieDomainName(m.Sess.Host) {
		c.Domain = m.Sess.GetHost()
		m.Cookie = append(m.Cookie, c.String())

		if strings.HasPrefix(m.Sess.GetHost(), "www.") {
			c.Domain = strings.TrimPrefix(c.Domain, "www.")
			m.Cookie = append(m.Cookie, c.String())
		}

		if strings.HasPrefix(m.Sess.GetHost(), "api.") {
			c.Domain = m.Sess.GetHost()
			c.Domain = strings.TrimPrefix(c.Domain, "api.")
			m.Cookie = append(m.Cookie, c.String())
		}
	}
	m.IsSetCookie = true
}

func (m *NatsMsg) AddCookieKvOnly(k, v string) {
	c := Cookie{Name: k, Value: v}
	m.Cookie = append(m.Cookie, c.String())
	m.IsSetCookie = true
}

func (m *NatsMsg) ClearCookie() {
	m.Cookie = make([]string, 0)
}

func (m *NatsMsg) GetCookie(key string) string {
	for _, cookie := range m.Cookie {
		for _, kv := range strings.Split(cookie, ";") {
			kv = strings.TrimSpace(kv)
			arr := strings.Split(kv, "=")
			if len(arr) == 2 && arr[0] == key && len(arr[1]) > 0 {
				return arr[1]
			}
		}
	}
	return ""
}

func (m *NatsMsg) GetSession( /*uid ...uint64*/ ) *Session {
	if m == nil {
		return nil
	}
	/*	if m.GetUid() == 0 && len(uid) > 0 {
		m.Sess.Uid = uid[0]
	}*/
	return &m.Sess
}
func (m *NatsMsg) IsLogin() bool {
	if m == nil {
		return false
	}
	return m.Sess.GetUid() > 10000
}
func (m *NatsMsg) GetUid() uint64 {
	if m == nil {
		return 0
	}
	return uint64(m.Sess.GetUid())
}
func (m *NatsMsg) GetLang() string {
	if m == nil {
		return "en_US"
	}
	return m.Sess.GetLang()
}
func (m *NatsMsg) GetAppID() int32 {
	if m == nil {
		return 0
	}
	return m.GetSession().GetAppID()
}
func (m *NatsMsg) GetPlatID() int32 {
	if m == nil {
		return 0
	}
	return m.GetSession().GetPlatID()
}
func (m *NatsMsg) GetExtValue() int32 {
	if m == nil {
		return 0
	}
	return m.ExtValue
}
func (m *NatsMsg) GetChannelNo() int32 {
	if m == nil {
		return 0
	}
	return m.GetSession().GetChannelNo()
}

func (m *NatsMsg) GetMod() string {
	if m == nil {
		return ""
	}
	return m.MsgBody.Mod
}
func (m *NatsMsg) GetFunc() string {
	if m == nil {
		return ""
	}
	return m.MsgBody.Func
}
func (m *NatsMsg) GetCheck() string {
	if m == nil || m.MsgBody.Check == nil {
		return ""
	}
	return utils.AutoToString(m.MsgBody.Check)
}

//func (m *NatsMsg) GetType() string {
//	if m == nil {
//		return "req"
//	}
//	if m.MsgBody.Type == "" {
//		return "json"
//	}
//	return m.MsgBody.Type
//}

//	func (m *NatsMsg) GetRequestBody() *string {
//		return m.NatsMsgStr
//	}
func (m *NatsMsg) GetMsgErrNo() int32 {
	if m == nil {
		return -1
	}
	return m.MsgBody.ErrNo
}
func (m *NatsMsg) GetMsgErrStr() string {
	if m == nil {
		return "noMsg"
	}
	return m.MsgBody.ErrStr
}
func (m *NatsMsg) GetMsgBody() *MsgBody {
	if m == nil {
		return nil
	}
	return &m.MsgBody
}

func (m *NatsMsg) GetParam() []byte {
	if m == nil {
		return nil
	}
	return m.MsgBody.Param
}
func (m *NatsMsg) GetNatsMsg() *nats.Msg {
	if m == nil {
		return nil
	}
	return m.NatsMsg
}

func (m *NatsMsg) GetLoginInfo() *LoginInfo {
	if m == nil {
		return nil
	}
	return m.Sess.LoginInfo
}
func (m *NatsMsg) GetLoginAddr() string {
	if m == nil {
		return ""
	}
	return m.Sess.GetLoginAddr()
}

func (m *NatsMsg) GetRemoteAddr() string {
	if m == nil {
		return ""
	}
	return m.Sess.GetRemoteAddr()
}
func (m *NatsMsg) GetLoginType() int32 {
	if m == nil {
		return 0
	}
	return m.Sess.GetLoginType()
}
func (m *NatsMsg) GetRemoteMac() string {
	if m == nil {
		return ""
	}
	return m.Sess.GetRemoteMac()
}

// func (m *NatsMsg) ResponeData(t string, data []byte) (err error) {
// 	if m == nil {
// 		return nil
// 	}
// 	if m.RpcxConn != nil && m.Reply == nil && m.GetSession().Channel == 0 {
// 		//这是rpcx请求不需要回包的
// 		return nil
// 	}
// 	//是rpcx请求
// 	if m.Reply != nil {
// 		rspmsg := m.Reply
// 		rspmsg.Sess = *m.GetSession()
// 		rspmsg.MsgBody.Mod = GetServerName()
// 		rspmsg.MsgBody.Func = m.MsgBody.Func
// 		rspmsg.MsgBody.Check = m.MsgBody.Check
// 		rspmsg.MsgData = data
// 		rspmsg.Type = t
// 		return nil
// 	}

// 	rsp := NatsMsg{Sess: *m.GetSession(), Type: t}
// 	rsp.MsgData = data

// 	//rsp.Sess.SvrFE = GetServerName()
// 	//rsp.Sess.SvrID = GetServerID()
// 	//rsp.Sess.Time = time.Now().Unix()
// 	//req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())

// 	return NatsSendReply(m, &rsp)
// }

//	func (m *NatsMsg) IsNeedRsp() bool {
//		if m == nil {
//			return false
//		}
//		if m.GetNatsMsg() != nil && m.GetNatsMsg().Reply != "" {
//			return true
//		}
//		if m.GetSession().SvrID > 0 && m.GetSession().Channel > 0 {
//			return true
//		}
//		if m.Reply != nil {
//			return true
//		}
//		return false
//	}
func (m *NatsMsg) ResponeSucc(rspparam interface{}) (err error) {
	return m.SendResponse(0, "", rspparam)
}

// 加密方式回包
func (m *NatsMsg) SendResponseEcrypt(errno int32, errstr string, rspparam interface{}) (err error) {
	return m.response(errno, errstr, rspparam, true)
}

func (m *NatsMsg) SendResponse(errno int32, errstr string, rspparam interface{}) (err error) {
	return m.response(errno, errstr, rspparam, false)
}

func (m *NatsMsg) SendResponseNew(errno int32, rspparam interface{}) (err error) {
	return m.response(errno, "", rspparam, false)
}

func (m *NatsMsg) response(errno int32, errstr string, rspparam interface{}, isEncrypt bool) (err error) {
	if m == nil {
		return nil
	}
	if m.RpcxConn != nil && m.Reply == nil && m.GetSession().Channel == 0 {
		//这是rpcx请求不需要回包的
		return nil
	}
	traceId := m.GetSession().Trace.TraceId
	if traceIdPos := strings.Index(traceId, ":"); traceIdPos > 0 {
		traceId = traceId[:traceIdPos]
	}
	if rspparam != nil && len(m.MsgBody.Filters) > 0 {
		str, err := jsoniter.Marshal(rspparam)
		if err == nil {
			mm := make(map[string]interface{})
			err = jsoniter.Unmarshal(str, &mm)
			if err == nil {
				mmm := make(map[string]interface{}, len(mm))
				for k, v := range mm {
					if utils.InArray(m.MsgBody.Filters, k) {
						mmm[k] = v
					}
				}
				rspparam = mmm
			}
		}

	}

	//是rpcx请求
	if m.Reply != nil {
		rspmsg := m.Reply
		rspmsg.Encrypt = isEncrypt
		rspmsg.Sess = *m.GetSession()
		rspmsg.MsgBody.Mod = GetServerName()
		rspmsg.MsgBody.Func = m.MsgBody.Func
		rspmsg.MsgBody.Check = m.MsgBody.Check
		rspmsg.MsgBody.ErrNo = errno
		rspmsg.MsgBody.ErrStr = errstr
		rspmsg.MsgBody.Trace = traceId
		rspmsg.Cookie = m.Cookie

		if rspparam != nil {
			rspmsg.MsgBody.Param, err = jsoniter.Marshal(rspparam)
			if err != nil {
				logs.Errorf("Failed to jsoniter.Marshal: %+v, err:%+v", rspparam, err)
				return errors.New("jsoniter.Marshal rsp failed")
			}
		}
		return nil
	}
	//这个是从rpcx收到websocket过来的请求, conn转发过来的websocket的请求都是onway
	if m.GetSession().Channel > 0 {
		rsp := QgMsp{
			Cmd:    m.MsgBody.Func,
			ErrNo:  errno,
			ErrStr: errstr,
		}
		rsp.BytePara, _ = jsoniter.Marshal(rspparam)
		return SendMsgToClient(m.GetSession(), errno, errstr, m.MsgBody.Func, rsp, isEncrypt)
	}

	return nil
}

type QgMsp struct {
	BytePara jsoniter.RawMessage `json:"_para,omitempty"`
	ErrNo    int32               `json:"_errno,omitempty"`
	ErrStr   string              `json:"_errstr,omitempty"`
	Cmd      string              `json:"_cmd"`
}

type NatsTransMsg struct {
	Sess    Session             `json:"Sess"`
	Type    string              `json:"type,omitempty"`
	Encrypt bool                `json:"encrypt"` //是否需要加密
	MsgBody jsoniter.RawMessage `json:"Body,omitempty"`
	MsgData []byte              `json:"Data,omitempty"`
	NatsMsg *nats.Msg           `json:"-"`
	//NatsMsgStr *string   `json:"-"`

	Cookie []string `json:"cookie"`
}

func (m *NatsTransMsg) GetSession() *Session {
	if m == nil {
		return nil
	}
	return &m.Sess
}
func (m *NatsTransMsg) GetAppUid() uint64 {
	if m == nil {
		return 0
	}
	return GenAppUid(m.Sess.GetAppID(), m.Sess.GetUid())
}
func (m *NatsTransMsg) GetUid() uint64 {
	if m == nil {
		return 0
	}
	return uint64(m.Sess.GetUid())
}
func (m *NatsTransMsg) GetType() string {
	if m == nil || m.Type == "" {
		return "json"
	}
	return m.Type
}
func (m *NatsTransMsg) GetBody() jsoniter.RawMessage {
	if m == nil || m.MsgBody == nil {
		return jsoniter.RawMessage{}
	}
	return m.MsgBody
}

func (m *NatsTransMsg) GetData() jsoniter.RawMessage {
	if m == nil || m.MsgData == nil {
		return jsoniter.RawMessage{}
	}
	return m.MsgData
}

func (m *NatsTransMsg) GetNatsMsg() *nats.Msg {
	if m == nil {
		return nil
	}
	return m.NatsMsg
}

//	func (m *NatsTransMsg) ResponeData(t string, data []byte) (err error) {
//		if m == nil {
//			return nil
//		}
//		rsp := NatsTransMsg{Sess: *m.GetSession(), Type: t}
//		rsp.MsgData = data
//
//		rsp.Sess.SvrFE = GetServerName()
//		rsp.Sess.SvrID = GetServerID()
//		rsp.Sess.Time = time.Now().Unix()
//		//req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())
//
//		subj := ""
//		if m.GetNatsMsg() != nil && m.GetNatsMsg().Reply != "" {
//			subj = m.GetNatsMsg().Reply
//		} else if m.GetSession().SvrID > 0 && m.GetSession().Channel > 0 {
//			//兼容老的svrframe
//			subj = GenRspSubject(m.GetSession().SvrFE, m.GetSession().SvrID, m.GetSession().Channel)
//		} else {
//			return nil
//		}
//		return NatsPublish(subj, rsp, true, false)
//	}
func (m *NatsTransMsg) IsNeedRsp() bool {
	if m == nil {
		return false
	}
	if m.GetNatsMsg() != nil && m.GetNatsMsg().Reply != "" {
		return true
	}
	if m.GetSession().SvrID > 0 && m.GetSession().Channel > 0 {
		return true
	}
	return false
}

//func (m *NatsTransMsg) ResponeSucc(rspparam interface{}) (err error) {
//	return m.SendResponse(0, "", rspparam)
//}
//
//func (m *NatsTransMsg) SendResponse(errno int32, errstr string, rspparam interface{}) (err error) {
//	if m == nil {
//		return nil
//	}
//	rspmsg := NatsMsg{Sess: *m.GetSession()}
//	if rspparam != nil {
//		rspmsg.MsgBody.Param, err = jsoniter.Marshal(rspparam)
//		if err != nil {
//			logs.Errorf("Failed to jsoniter.Marshal: %+v", rspparam)
//			return errors.New("jsoniter.Marshal rsp failed")
//		}
//	}
//	rspmsg.MsgBody.Ret = errno
//	rspmsg.MsgBody.ErrStr = errstr
//
//	subj := ""
//	if m.GetNatsMsg() != nil && m.GetNatsMsg().Reply != "" {
//		subj = m.GetNatsMsg().Reply
//	} else if m.GetSession().SvrID > 0 && m.GetSession().Channel > 0 {
//		//兼容老的svrframe
//		subj = GenRspSubject(m.GetSession().SvrFE, m.GetSession().SvrID, m.GetSession().Channel)
//	} else {
//		return nil
//	}
//	return NatsPublish(subj, rspmsg, true, true)
//}

type BaseHandler struct {
	req *NatsMsg
}

func (handler *BaseHandler) SetReq(req *NatsMsg) {
	handler.req = req
}
func (handler *BaseHandler) GetReq() *NatsMsg {
	return handler.req
}
func (handler *BaseHandler) GetAppID() int32 {
	return handler.req.GetAppID()
}
func (handler *BaseHandler) GetUid() uint64 {
	return handler.req.GetUid()
}
func (handler *BaseHandler) GetChannelNo() int32 {
	return handler.req.GetChannelNo()
}

func (handler *BaseHandler) GetUserLang() string {
	return handler.req.GetLang()
}
