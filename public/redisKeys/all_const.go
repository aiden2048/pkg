package redisKeys

const (
	RedisExpire_2Days  = 60 * 60 * 24 * 2
	RedisExpire_1Days  = 60 * 60 * 24 * 1
	RedisExpire_1Week  = 60 * 60 * 24 * 7
	RedisExpire_30Days = 60 * 60 * 24 * 30

	RedisExpire_Money_UserCoin = 60 * 60 * 24 * 30 //用户金币
	RedisExpire_Money_Lock     = 60 * 60 * 12      //房间锁
	RedisExpire_Money_Server   = 60 * 3            //Money服务
	RedisExpire_login_session  = 60 * 60 * 72      //登录session
	RedisExpire_Admin_Total    = 60 * 60 * 24 * 45 //统计
	InfoTtlHour                = 3600              //1小时缓存
	InfoTtlDay                 = 3600 * 24         //1天缓存
	InfoTtlWeek                = 7 * 3600 * 24     //info 数据 缓存时间
	DEFULT_REDIS_WEEk_TTL      = 7 * 3600 * 24
	MonthTtl                   = 30 * 3600 * 24 //一个月缓存
	HalfMonthTtl               = 15 * 3600 * 24 //半个月缓存
)

const (
	REDIS_INDEX_COMMON  = "default"
	REDIS_INDEX_Money   = "money"
	REDIS_INDEX_USER    = "user"
	REDIS_INDEX_SESSION = "session"
	REDIS_INDEX_RISK    = "risk"
	REDIS_INDEX_PAY     = "pay"
	REDIS_INDEX_LOTTERY = "lottery" //彩票
)

const (
	InfoForeverKey = "info.forever"
)

// redis 分类
type RedisKeys struct {
	Name string
	Key  string
}
