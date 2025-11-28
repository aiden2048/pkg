package redisDeal

import (
	"context"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/aiden2048/pkg/utils"
	jsoniter "github.com/json-iterator/go"

	"github.com/redis/go-redis/v9"
)

// 直接操作key 的
// del
func RedisSendDel(key *redisKeys.RedisKeys) error {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.Del(context.TODO(), key.Key)
	_, err := cmd.Result()

	logRedisCommand(key.Name, err, "DEL", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "DEL", err != nil, duration)

	return err
}

func RedisDoDel(key *redisKeys.RedisKeys) error {
	return RedisSendDel(key)
}

// SETEX
func RedisSendSetEx(key *redisKeys.RedisKeys, ttl int64, value interface{}) error {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.SetEx(context.TODO(), key.Key, value, time.Duration(ttl)*time.Second)
	err := cmd.Err()

	logRedisCommand(key.Name, err, "SETEX", key.Key, ttl, value)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SETEX", err != nil, duration)

	return err
}

// get int64
func RedisDoGetInt(key *redisKeys.RedisKeys) int64 {
	ret, _ := RedisDoGetIntV2(key)
	return ret
}

func RedisDoGetIntV2(key *redisKeys.RedisKeys) (int64, error) {
	retStr, err := RedisDoGetStrV2(key)
	if err != nil {
		return 0, err
	}
	ret := utils.StrToInt64(retStr)
	return ret, nil
}

// get string
func RedisDoGetStr(key *redisKeys.RedisKeys) string {
	ret, _ := RedisDoGetStrV2(key)
	return ret
}

// get string
func RedisDoGetStrV2(key *redisKeys.RedisKeys) (string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.Get(context.TODO(), key.Key)
	retStr, err := cmd.Result()

	logRedisCommand(key.Name, err, "GET", key.Key)
	reportRedisStats(key.Name, "GET", err != nil, time.Since(startTime))

	return retStr, nil
}

// set
func RedisSendSet(key *redisKeys.RedisKeys, data interface{}, ttl ...int64) error {
	str, ok := data.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = datastr
	}
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.Set(context.TODO(), key.Key, str, 0) // 默认没有 TTL
	err := cmd.Err()

	logRedisCommand(key.Name, err, "SET", key.Key, data)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SET", err != nil, duration)
	// 设置 TTL（如果提供了）
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}

	return err
}

// set
func RedisDoSet(key *redisKeys.RedisKeys, data interface{}, ttl ...int64) error {

	return RedisSendSet(key, data, ttl...)
}

// INCRBY
func RedisSendIncrby(key *redisKeys.RedisKeys, inc int64, ttl ...int64) error {
	_, err := RedisDoIncrbyV2(key, inc, ttl...)
	return err
}

// INCRBY
func RedisDoIncrbyV2(key *redisKeys.RedisKeys, inc int64, ttl ...int64) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.IncrBy(context.TODO(), key.Key, inc)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "INCRBY", key.Key, inc)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "INCRBY", err != nil, duration)

	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}

	// 设置 TTL（如果传入了 ttl 参数）
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}

	return ret, nil
}

// INCRBY
func RedisDoIncrby(key *redisKeys.RedisKeys, inc int64, ttl ...int64) int64 {
	ret, _ := RedisDoIncrbyV2(key, inc, ttl...)
	return ret
}

// SETNXInt //互斥写一个 数字 1 返回0 说明不成功
func RedisDoSetnx(key *redisKeys.RedisKeys, data interface{}, ttl int64) int {
	// 如果 TTL 小于等于 0，设置默认 TTL
	if ttl <= 0 {
		logs.Errorf("RedisDoSetnx Set Key %s With No TTL, Set Default TTL = 60", key.Key)
		ttl = 60
	}

	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.SetNX(context.TODO(), key.Key, data, 0) // 默认没有 TTL，后续手动设置
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "SETNX", key.Key, data)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SETNX", err != nil, duration)

	if err != nil {
		logs.Errorf("RedisDoSetnx key:%s, err:%+v", key.Key, err)
		return 0
	}
	// 如果设置成功，设置 TTL
	if ret {
		RedisSetTtl(key, ttl)
		return 1
	}

	return 0
}

// SETNX
func RedisSendSetnx(key *redisKeys.RedisKeys, data interface{}, ttl int64) error {
	if ttl <= 0 {
		logs.Errorf("RedisDoSetnx Set Key %s With No TTL, Set Default TTL = 60", key.Key)
		ttl = 60
	}

	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.SetNX(context.TODO(), key.Key, data, 0) // 默认没有 TTL，后续手动设置
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "SETNX", key.Key, data)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SETNX", err != nil, duration)

	// 如果设置成功，设置 TTL
	if ret {
		_ = RedisSetTtl(key, ttl) // 错误处理通过 _ 忽略
	}

	return err
}

// 判断key是否存在
func RedisKeyExist(key *redisKeys.RedisKeys) bool {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.Exists(context.TODO(), key.Key)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "EXISTS", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "EXISTS", err != nil, duration)

	if err != nil {
		logs.Errorf("RedisKeyExist key:%s, err:%+v", key.Key, err)
		return false
	}

	if ret <= 0 {
		return false
	}

	return true
}

func RedisDoGetFloat(key *redisKeys.RedisKeys) float64 {
	retStr := RedisDoGetStr(key)
	ret := utils.StrToFloat(retStr)
	return ret
}
