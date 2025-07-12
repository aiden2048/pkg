package redisDeal

import (
	"time"

	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/patrickmn/go-cache"
)

/*
用于缓存redis设置是ttl, 没必要每次都要更新ttl, 减少对redis的操作
*/
var ttlCache = cache.New(time.Second, time.Minute)

// ttlttl 本身缓存的有效期
func autoGetTtl(keys *redisKeys.RedisKeys, ttl, ttlttl int64) int64 {
	key := keys.Name + keys.Key
	_, ok := ttlCache.Get(key)
	if !ok {
		if ttlttl == 0 {
			ttlttl = ttl
		}
		ttlCache.Set(key, ttl, time.Second*time.Duration(ttlttl))
		return ttl
	}
	return 0
}
