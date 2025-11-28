package redisDeal

import (
	"context"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/aiden2048/pkg/public/redisKeys"
)

// ttl
func RedisSendTtl(key *redisKeys.RedisKeys, ttl int64) error {
	if !check(key) {
		logs.Errorf("key =%s 不属于 %s", key.Key, key.Name)
		return errorMsg.Err_NoRedisTmpl
	}
	return RedisSetTtl(key, ttl)
}
func RedisSetTtl(key *redisKeys.RedisKeys, ttl int64) error {
	// 更新ttl缓存
	if autoGetTtl(key, ttl, 10) == 0 {
		// 减少操作频率
		return nil
	}
	conn := GetRedisPool(key.Name)
	return conn.Expire(context.TODO(), key.Key, time.Duration(ttl)*time.Second).Err()
}
func RedisGetTtl(key *redisKeys.RedisKeys) (int64, error) {
	conn := GetRedisPool(key.Name)
	cmd := conn.TTL(context.TODO(), key.Key)
	ttl, err := cmd.Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}
