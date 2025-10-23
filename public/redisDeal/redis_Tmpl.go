package redisDeal

import (
	"strings"

	"github.com/aiden2048/pkg/public/redisKeys"
)

func InMoneyKey(key string) bool {
	if strings.Index(key, "money.") != 0 {
		return false
	}
	return true
}

// 检查是否属于我的key
func check(key *redisKeys.RedisKeys) bool {
	if key == nil || key.Name == "" {
		return false
	}
	switch key.Name {
	case REDIS_INDEX_COMMON: //默认的
		return true //不检查
	case REDIS_INDEX_SESSION: //session
		return redisKeys.InSessKey(key.Key)
	case REDIS_INDEX_USER: //user
		return redisKeys.InUserKey(key.Key)
	case REDIS_INDEX_Money: //money
		return InMoneyKey(key.Key)
	case REDIS_INDEX_PAY: //pay
		return redisKeys.InPayKey(key.Key)
	case REDIS_INDEX_RISK: //risk
		return true //不检查
	case REDIS_INDEX_ACTIVITY: //activity
		return true
	case REDIS_INDEX_TOPURLMAIN:
		return true
	case REDIS_INDEX_RECORD:
		return true
	}
	return false
}
