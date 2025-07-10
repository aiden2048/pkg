package redisKeys

import (
	"fmt"
	"strconv"

	"github.com/aiden2048/pkg/utils"
)

func CacheUserInfoKey(appId int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.cache.user.%d.%d", appId, uid)
	return key
}

// 用户信息 string
func GetUserInfoKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.userInfo.%d.%d", appid, uid)
	return key
}

// 玩家总充值 hmset
func GetRiskPayKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.u.pay.all.%d.%d", appid, uid)
	return key
}

// 玩家 充值 每天 hmset
func GetRiskPayDayKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "risk.u.pay.all." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(idate))
	return key
}

// 玩家总提现 hmset
func GetRiskPutKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.u.put.all.%d.%d", appid, uid)
	return key
}

// 玩家 充值提现 hmset
func GetRiskPutDayKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "risk.u.put.day." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(idate))
	return key
}

// 玩家投注 hmset
func GetRiskBetKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.u.bet.all.%d.%d", appid, uid)
	return key
}

// 玩家返奖 hmset
func GetAllRiskHitKey(appid int32, uid uint64) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("risk.u.hit.all.%d.%d", appid, uid)
	return k
}

// 玩家 投注 hmset
func GetRiskBetDayKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.bet.day." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(idate))
	return key
}

// 玩家投注次数 hmset
func GetRiskBetNumKey(appid int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.bet.num.all." + strconv.Itoa(int(appid))
	return key
}

// 玩家 投注 次数 hmset
func GetRiskBetDayNumKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.bet.num.day." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(idate))
	return key
}

// 玩家 返奖 hmset
func GetRiskHitDayKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.hit.day." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(idate))
	return key
}

// 游戏 ip 地址 区分 app
func GetIpKey(appid int32, ip string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.ips." + strconv.Itoa(int(appid)) + "_" + ip
	return key
}

// 游戏 mac 区分 app
func GetMacKey(appid int32, mac string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.macss." + strconv.Itoa(int(appid)) + "_" + mac
	return key
}

// 玩家单次充值后的投注次数
func GetUserPayBetNumKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.pay.bet.num.%d.%d", appid, uid)
	return key
}

// 玩家单次充值后的投注金额
func GetUserPayBetPointKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.pay.bet.p.%d.%d", appid, uid)
	return key
}

// 玩家单次充值后的返奖金额
func GetUserPayHitPointKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.pay.hit.p.%d.%d", appid, uid)
	return key
}

// 玩家充值后第一次投注时间
func GetUserPayFirstBetKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.pay.first.bet." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(uid))
	return key
}

// 玩家最后一次充值金额
func GetUserLastPayPoint(appid int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "r.u.last.pay." + strconv.Itoa(int(appid))
	return key
}

// 智能风控执行次数
func GetSmartSuccKey(appid, smartId, gameType int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("smart.succ.%d.%d.%d", appid, smartId, gameType)
	return key
}

// 最后执行的策略
func GetLastSmartKey(appid int32, now int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "last.exec.smart.id." + strconv.Itoa(int(appid)) + "." + utils.GetStrIDate(now)
	return key
}

// 最后执行的策略次数
func GetLastSmartSumKey(appid int32, now int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "last.exec.smart.sum." + strconv.Itoa(int(appid)) + "." + utils.GetStrIDate(now)
	return key
}

// 税收
func GetRiskPerKey(appid int32, key string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.per.%d.%s", appid, key)
	return k
}

// 游戏投注 1小时颗粒 每天一个key
func GetRiskGameBetKey(appid int32, ymd string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.g.bet.%d.%s", appid, ymd)
	return k
}

// 游戏返奖 1小时颗粒 每天一个key
func GetRiskGameHitKey(appid int32, ymd string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.g.hit.%d.%s", appid, ymd)
	return k
}

// 游戏返奖 1小时颗粒 每天一个key
func GetRiskGamePerKey(appid int32, ymd string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.g.per.%d.%s", appid, ymd)
	return k
}

// 玩家流水
// risk  玩家每天一个KEY
func GetRiskUserBetKey(appid int32, date string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.user.bet.%s", date)
	return k
}
func GetRiskUserHitKey(appid int32, date string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.user.hit.%s", date)
	return k
}

// risk  玩家每天一个KEY
func GetRiskUserBlackBetKey(appid int32, date string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.user.bet.black.%s", date)
	return k
}
func GetRiskUserBlackHitKey(appid int32, date string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf("app.risk.user.hit.black.%s", date)
	return k
}

// 总登录次数
func GetRiskUserLoginKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = fmt.Sprintf("risk.u.l.n.all.%d.%d", appid, uid)
	return key
}

// 玩家登陆天数
func GetUserLoginDayKey(appid int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_RISK
	key.Key = "risk.login.day." + strconv.Itoa(int(appid)) + "." + strconv.Itoa(int(uid))
	return key
}
