package redisKeys

import (
	"strings"
)

// 必须是 login 开头
func InSessKey(key string) bool {
	if strings.Index(key, "session.") != 0 && strings.Index(key, "login.") != 0 {
		return false
	}
	return true
}

// 每天的访问IP的set的key
func GetVisitIpSetKey(appid int32, shour string) *RedisKeys {

	k := GenSessionRedisKey("ipset.%d.%s", appid, shour)
	return k
}
