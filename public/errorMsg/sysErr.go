package errorMsg

var (
	ReqError            = NewConstError(400, "[Network connection failed]") //.Return("[Network connection failed]")
	TE_Event_NetChanged = NewConstError(401, "[Login invalidity]")          //.Return("[Login invalidity]") //客户端收到后回大厅, 等待大厅心跳触发被踢回到登录界面
	ReqParamError       = NewConstError(402, "[Request data error]")        //.Return("[Request data error]")
	NotAllow            = NewConstError(403, "[Forbidden]")                 //.Return("[Forbidden]")
	NoService           = NewConstError(404, "[Service is invalid]")        //.Return("[Service is invalid]")
	RspTimeout          = NewConstError(405, "[Timeout]")                   //.Return("[Timeout]")
	RspError            = NewConstError(406, "[service internal error]")    //.Return("[service internal error]")
	TimeOut             = NewConstError(407, "[Timeout]")                   //.Return("[Timeout]")
)
var (
	Err_NoRedis     = NewError(450, "[no redis]") //.Return("[Network connection failed]")
	Err_NoMysql     = NewError(451, "[no mysql]")
	Err_NoApp       = NewError(452, "[no app]")
	Err_NoRedisTmpl = NewError(453, "[no redis.tmpl]")
)
