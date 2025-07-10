package redisDeal

import (
	"context"
	"strconv"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"

	jsoniter "github.com/json-iterator/go"
)

func RedisDoRPop(key *redisKeys.RedisKeys) (string, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.RPop(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "RPOP", key.Key)
	duration := time.Since(start)
	reportRedisStats(key.Name, "RPOP", err != nil, duration)
	return ret, err
}

// RPOP
func RedisDoRpopStr(key *redisKeys.RedisKeys) string {
	ret, _ := RedisDoRPop(key)
	return ret
}
func RedisDoLLen(key *redisKeys.RedisKeys) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.LLen(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "LLEN", key.Key)
	duration := time.Since(start)
	reportRedisStats(key.Name, "LLEN", err != nil, duration)
	return ret, err
}

func RedisLlen(key *redisKeys.RedisKeys) int64 {
	ret, _ := RedisDoLLen(key)
	return ret
}

// LPUSH
func RedisDoLPush(key *redisKeys.RedisKeys, values ...interface{}) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.LPush(context.TODO(), key.Key, values...)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "LPUSH", key.Key, values)
	duration := time.Since(start)
	reportRedisStats(key.Name, "LPUSH", err != nil, duration)
	return ret, err
}

func RedisSendLpush(key *redisKeys.RedisKeys, data interface{}) error {
	str, ok := data.(string)
	if !ok {
		strtmp, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = strtmp
	}
	_, err := RedisDoLPush(key, str)
	return err
}

// LPUSH
func RedisDoLpush(key *redisKeys.RedisKeys, data interface{}) error {

	str, ok := data.(string)
	if !ok {
		strtmp, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = strtmp
	}
	_, err := RedisDoLPush(key, str)
	return err
}

// LPOP
func RedisDoLPop(key *redisKeys.RedisKeys) (string, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.LPop(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "LPOP", key.Key)
	duration := time.Since(start)
	reportRedisStats(key.Name, "LPOP", err != nil, duration)
	return ret, err
}

func RedisDoLpopStr(key *redisKeys.RedisKeys) string {
	ret, _ := RedisDoLPop(key)
	return ret
}

// RPUSH
func RedisDoRPush(key *redisKeys.RedisKeys, values ...interface{}) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.RPush(context.TODO(), key.Key, values...)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "RPUSH", key.Key, values)
	duration := time.Since(start)
	reportRedisStats(key.Name, "RPUSH", err != nil, duration)
	return ret, err
}

func RedisSendRpush(key *redisKeys.RedisKeys, data interface{}) error {
	str, ok := data.(string)
	if !ok {
		strtmp, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = strtmp
	}
	_, err := RedisDoRPush(key, str)
	return err
}

// lrem
func RedisDoLRem(key *redisKeys.RedisKeys, count int64, value interface{}) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.LRem(context.TODO(), key.Key, count, value)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "LREM", key.Key, count, value)
	duration := time.Since(start)
	reportRedisStats(key.Name, "LREM", err != nil, duration)
	return ret, err
}

func RedisSendLrem(key *redisKeys.RedisKeys, data interface{}) error {
	str, ok := data.(string)
	if !ok {
		strtmp, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = strtmp
	}
	_, err := RedisDoLRem(key, 0, str)
	if err != nil {
		return err
	}
	return nil
}

// lrange strings
func RedisDoLRange(key *redisKeys.RedisKeys, start, stop int64) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.LRange(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "LRANGE", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "LRANGE", err != nil, duration)
	return ret, err
}

func RedisLrangeStrings(key *redisKeys.RedisKeys, start int32, end int32) ([]string, error) {

	return RedisDoLRange(key, int64(start), int64(end))

}
func RedisLrangeInt64(key *redisKeys.RedisKeys, start int32, end int32) ([]int64, error) {
	res, err := RedisDoLRange(key, int64(start), int64(end))
	if err != nil {
		return nil, err
	}
	ret := make([]int64, len(res))
	for i, v := range res {
		ret[i], _ = strconv.ParseInt(v, 10, 64)
	}
	return ret, nil
}

// ltrim  删除范围外数据
func RedisDoLTrim(key *redisKeys.RedisKeys, start, stop int64) error {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.LTrim(context.TODO(), key.Key, start, stop)
	_, err := cmd.Result()
	logRedisCommand(key.Name, err, "LTRIM", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "LTRIM", err != nil, duration)
	return err
}

func RedisLtrim(key *redisKeys.RedisKeys, start int32, end int32) error {
	return RedisDoLTrim(key, int64(start), int64(end))
}

// lrange int64
