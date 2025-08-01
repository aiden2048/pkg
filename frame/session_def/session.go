package session_def

import "time"

// 从登录态获取的信息
type LoginInfo struct {
	Idx        string `json:"idx,omitempty"`
	ChannelNo  int32  `json:"ChannelNo,omitempty"` // 用户渠道号
	LoginType  int32  `json:"utype"`
	Os         int32  `json:"os"`          // 系统类型
	AppType    int32  `json:"appType"`     // 应用类型
	ClientType int32  `json:"client_type"` //登录客户端类型: Web=0 App[or PWA]=1 等
	Birth      int64  `json:"birth"`
	Expire     int64  `json:"expire"`

	// 增加地区信息，做区域验证
	Country string `json:"country,omitempty"` // 请求的国家
	City    string `json:"city,omitempty"`    // 请求的城市
	Region  string `json:"region,omitempty"`  // 所在地区

	LoginAddr  string `json:"LoginAddr"`            //登录的ip
	RemoteAddr string `json:"RemoteAddr,omitempty"` //当前访问ip, 临时打开
	Mac        string `json:"Mac,omitempty"`
	Src        string `json:"src,omitempty"` // checkLogin 赋值

	//Vip     int32  `json:"vip"`           //vip等级, 登录时候是什么就是什么
	//Tag int32 `json:"tag"` //玩家tag
	//Lang    int32  `json:"lang"`          //语言
	IsFree  bool   `json:"is_free"`  // 是否试玩
	RealUid uint64 `json:"real_uid"` // 试玩玩家的真实UID

	Sign          int64          `json:"sign"` //随机数校验
	OldTokenInfos []OldTokenInfo `json:"old_token_infos,omitempty"`
}
type OldTokenInfo struct {
	TokenDelTime int64 `json:"token_del_time"` // 老token删除时间
	Expire       int64 `json:"expire"`         // 老token计划过期时间
	Sign         int64 `json:"sign"`           //随机数校验
}
type STrace struct {
	Uid     uint64 `json:"Uid,omitempty"`
	AppId   int32  `json:"AppID,omitempty"`
	TraceId string `json:"trace_id"`
	//Begin   string      `json:"begin"`
	Request interface{} `json:"-"`
	Sess    *Session    `json:"-"`
}
type Session struct {
	Uid     uint64 `json:"Uid,omitempty"`
	AppId   int32  `json:"AppID,omitempty"`
	Lang    int32  `json:"lang,omitempty"`
	Os      int32  `json:"os"`
	PlatId  int32  `json:"PlatID,omitempty"`
	SvrFE   string `json:"SvrFE,omitempty"`
	SvrID   int32  `json:"SvrID,omitempty"`
	Channel uint64 `json:"Channel,omitempty"` //用户连接的管道ID

	Cmd         string `json:"cmd,omitempty"`
	Seq         int64  `json:"Seq,omitempty"`
	Addr        string `json:"addr,omitempty"`
	Country     string `json:"country,omitempty"`      // 请求的国家
	City        string `json:"city,omitempty"`         // 请求的城市
	Region      string `json:"region,omitempty"`       // 所在地区
	RegionCode  string `json:"region_code,omitempty"`  //地区缩写
	IpLatitude  string `json:"ip_latitude,omitempty"`  // 经度
	IpLongitude string `json:"ip_longitude,omitempty"` // 纬度
	Referer     string `json:"referer,omitempty"`      // 请求的域名

	Host  string `json:"host,omitempty"` // 请求的域名
	Check string `json:"-"`

	LoginInfo *LoginInfo `json:"LoginInfo,omitempty"`

	Times   time.Time `json:"Times,omitempty"`
	Time    int64     `json:"Time,omitempty"`
	RpcType string    `json:"rpc_type"`
	Trace   *STrace   `json:"trace"`
	AppType int32     `json:"App_type"`
}

func (s *Session) Copy() *Session {
	return &Session{
		Uid:         s.Uid,
		AppId:       s.AppId,
		Lang:        s.Lang,
		Os:          s.Os,
		PlatId:      s.PlatId,
		SvrFE:       s.SvrFE,
		SvrID:       s.SvrID,
		Channel:     s.Channel,
		Cmd:         s.Cmd,
		Times:       time.Now(),
		Time:        s.Time,
		Seq:         s.Seq,
		Addr:        s.Addr,
		Country:     s.Country,
		City:        s.City,
		Region:      s.Region,
		RegionCode:  s.RegionCode,
		IpLatitude:  s.IpLatitude,
		IpLongitude: s.IpLongitude,
		Check:       s.Check,
		RpcType:     s.RpcType,
		LoginInfo:   s.LoginInfo,
		Trace:       s.Trace,
		Host:        s.Host,
	}
}

func (s *Session) GetServerID() int32 {
	if s == nil {
		return 0
	}
	return s.SvrID
}
func (s *Session) GetServerFE() string {
	if s == nil {
		return ""
	}
	return s.SvrFE
}
func (s *Session) GetUid() uint64 {
	if s == nil {
		return 0
	}
	return s.Uid
}
func (s *Session) GetLang() int32 {
	if s == nil {
		return 0
	}
	return s.Lang
}

func (s *Session) GetOsType() int32 {
	if s == nil {
		return 0
	}
	return s.Os
}
func (s *Session) GetHost() string {
	if s == nil {
		return ""
	}
	return s.Host
}

//	func (s *Session) GetPirKey() string {
//		if s == nil || s.LoginInfo == nil {
//			return ""
//		}
//		return s.LoginInfo.PriKey
//	}
//
//	func (s *Session) GetSessionID() int64 {
//		if s == nil || s.LoginInfo == nil {
//			return 0
//		}
//		return s.LoginInfo.SessionID
//	}
func (s *Session) GetAppID() int32 {

	return s.AppId
}
func (s *Session) GetPlatID() int32 {
	if s == nil {
		return 0
	}
	return s.PlatId
}
func (s *Session) GetChannelNo() int32 {
	if s == nil || s.LoginInfo == nil {
		return 0
	}
	return s.LoginInfo.ChannelNo
}
func (s *Session) GetRemoteAddr() string {
	if s == nil {
		return ""
	}
	return s.Addr
}

// 是否免费用户
func (s *Session) IsFree() bool {
	if s == nil || s.LoginInfo == nil {
		return false
	}
	return s.LoginInfo.IsFree
}

// 是否虚拟用户
func (s *Session) IsVirtualUser() bool {
	if s == nil || s.LoginInfo == nil {
		return false
	}
	return s.LoginInfo.LoginType > 100
}

func (s *Session) GetLoginAddr() string {
	if s == nil || s.LoginInfo == nil {
		return ""
	}
	return s.LoginInfo.LoginAddr
}

func (s *Session) GetRemoteMac() string {
	if s == nil || s.LoginInfo == nil {
		return ""
	}
	return s.LoginInfo.Mac
}

func (s *Session) GetCountry() string {
	if s == nil {
		return ""
	}
	return s.Country
}

func (s *Session) GetReferer() string {
	if s == nil {
		return ""
	}
	return s.Referer
}

func (s *Session) GetCity() string {
	if s == nil {
		return ""
	}
	return s.City
}

/*func (s *Session) GetXRemoteAddr() string {
	if s == nil {
		return ""
	}
	return s.X_RemoteAddr
}*/

func (s *Session) GetLoginType() int32 {
	if s == nil {
		return 0
	}

	return s.LoginInfo.LoginType
}

/*
	func (s *Session) GetVipLevel() int32 {
		if s == nil || s.LoginInfo == nil {
			return 0
		}

		return s.LoginInfo.Vip
	}
*/
func (s *Session) GetLoginInfo() *LoginInfo {
	if s == nil {
		return &LoginInfo{}
	}

	return s.LoginInfo
}
func (s *Session) Islogin() bool {
	return s.GetUid() > 10000
}
func (s *Session) MutableLoginInfo() *LoginInfo {
	if s == nil {
		return nil
	}
	if s.LoginInfo == nil {
		s.LoginInfo = &LoginInfo{}
	}
	return s.LoginInfo
}
func (l *LoginInfo) AddOldTokenInfo(expire, sign, del_time int64) {
	if l == nil {
		return
	}
	if len(l.OldTokenInfos) == 0 {
		l.OldTokenInfos = []OldTokenInfo{}
	}
	old := OldTokenInfo{}
	old.Expire = expire
	old.Sign = sign
	old.TokenDelTime = del_time
	l.OldTokenInfos = append(l.OldTokenInfos, old)
}
func (l *LoginInfo) CheckToken(expire, sign int64) bool {
	if l == nil {
		return false
	}
	if l.Expire == expire && l.Sign == sign {
		// 当前token
		return true
	}
	now := time.Now().Unix()
	for _, old := range l.OldTokenInfos {
		if old.TokenDelTime < now {
			// 老token失效
			continue
		}
		if old.Expire == expire && old.Sign == sign {
			return true
		}
	}
	return false
}
