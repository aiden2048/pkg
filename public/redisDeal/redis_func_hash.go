package redisDeal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
)

func RedisDoHGet(key *redisKeys.RedisKeys, field string) (string, error) {
	conn := GetRedisPool(key.Name)

	start := time.Now()
	cmd := conn.HGet(context.TODO(), key.Key, field)
	val, err := cmd.Result()

	logRedisCommand(key.Name, err, "HGet", key.Key, field)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HGet", err != nil, duration)
	if err == redis.Nil {
		return "", nil
	}

	return val, err
}
func RedisDoHGetSrt(key *redisKeys.RedisKeys, key2 interface{}) string {
	ret, err := RedisDoHGet(key, fmt.Sprintf("%v", key2))
	if err != nil {
		return ""
	}
	return ret
}

func RedisDoHGetStrV2(key *redisKeys.RedisKeys, key2 interface{}) (string, error) {
	return RedisDoHGet(key, fmt.Sprintf("%v", key2))
}

// hget int64
func RedisDoHGetInt(key *redisKeys.RedisKeys, key2 interface{}) int64 {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HGet(context.TODO(), key.Key, fmt.Sprintf("%v", key2))
	ret, err := cmd.Int64()
	logRedisCommand(key.Name, err, "HGet", key.Key, key2)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HGet", err != nil, duration)
	if err != nil {
		return 0
	}
	return ret
}

func RedisDoHGetIntV2(key *redisKeys.RedisKeys, key2 interface{}) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HGet(context.TODO(), key.Key, fmt.Sprintf("%v", key2))
	ret, err := cmd.Int64()
	logRedisCommand(key.Name, err, "HGet", key.Key, key2)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HGet", err != nil, duration)
	return ret, nil
}

func RedisDoHGetFloat(key *redisKeys.RedisKeys, key2 interface{}) float64 {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HGet(context.TODO(), key.Key, fmt.Sprintf("%v", key2))
	ret, err := cmd.Float64()
	logRedisCommand(key.Name, err, "HGet", key.Key, key2)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HGet", err != nil, duration)
	if err != nil {
		return 0
	}
	return ret
}
func RedisDoHGetAll(key *redisKeys.RedisKeys) (map[string]string, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HGetAll(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "HGETALL", key.Key)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HGETALL", err != nil, duration)
	return ret, nil
}

// hget []string
func RedisDoHGetAllSrt(key *redisKeys.RedisKeys) []string {
	retMap, err := RedisDoHGetAll(key)
	if err != nil {
		return nil
	}
	ret := make([]string, 0, len(retMap))
	for k, v := range retMap {
		ret = append(ret, k)
		ret = append(ret, v)
	}
	return ret
}

// hget []string
func RedisDoHGetAllSrtV2(key *redisKeys.RedisKeys) ([]string, error) {
	retMap, err := RedisDoHGetAll(key)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(retMap))
	for k, v := range retMap {
		ret = append(ret, k)
		ret = append(ret, v)
	}
	return ret, nil
}

// HGETALL uint64[int64] => userId[Point]
func RedisDoHGetAllUint64AndInt64(key *redisKeys.RedisKeys) map[uint64]int64 {
	ret := map[uint64]int64{}
	data, err := RedisDoHGetAll(key)
	if err != nil {
		return ret
	}
	for k, v := range data {
		uintk, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		intv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ret[uintk] = intv
	}
	return ret
}
func RedisDoHGetAllUint64AndInt64V2(key *redisKeys.RedisKeys) (map[uint64]int64, error) {
	ret := map[uint64]int64{}
	data, err := RedisDoHGetAll(key)
	if err != nil {
		return ret, err
	}
	for k, v := range data {
		uintk, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		intv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ret[uintk] = intv
	}
	return ret, nil
}

// HGETALL int64
func RedisDoHGetAllInt64(key *redisKeys.RedisKeys) map[int64]int64 {
	ret := map[int64]int64{}
	data, err := RedisDoHGetAll(key)
	if err != nil {
		return ret
	}
	for k, v := range data {
		intk, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			continue
		}
		intv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ret[intk] = intv
	}
	return ret
}

// hsetnx
func RedisSendHSetNx(key *redisKeys.RedisKeys, key2 string, data interface{}, ttl ...int64) error {
	str, ok := data.(string)
	if !ok { //不是字符串
		var err error
		str, err = jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
	}
	conn := GetRedisPool(key.Name)
	start := time.Now()
	err := conn.HSetNX(context.TODO(), key.Key, key2, str).Err()
	logRedisCommand(key.Name, err, "HSetNX", key.Key, key2, str)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HSetNX", err != nil, duration)

	if err != nil {
		logs.Errorf("RedisSendHSet HSETNX key:%s,key2:%+v err:%+v", key.Key, key2, err)
		return err
	}
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}
	return nil
}

// hset

func RedisSendHSet(key *redisKeys.RedisKeys, key2, data interface{}, ttl ...int64) error {
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
	start := time.Now()
	cmd := conn.HSet(context.TODO(), key.Key, key2, str)
	err := cmd.Err()
	logRedisCommand(key.Name, err, "HSET", key.Key, key2, str)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HSET", err != nil, duration)
	if err != nil {
		return err
	}
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}
	return nil
}

// hset
func RedisDoHSet(key *redisKeys.RedisKeys, key2, data interface{}, ttl ...int64) error {
	return RedisSendHSet(key, key2, data, ttl...)
}

func RedisDoHMSet(key *redisKeys.RedisKeys, fields map[string]interface{}, ttl int64) error {
	conn := GetRedisPool(key.Name)
	args := []interface{}{}
	for f, v := range fields {
		args = append(args, f, v)
	}
	start := time.Now()
	cmd := conn.HMSet(context.TODO(), key.Key, args...)
	err := cmd.Err()
	logRedisCommand(key.Name, err, "HMSET", key.Key, args)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HMSET", err != nil, duration)
	if err != nil {
		return err
	}
	if ttl > 0 {
		RedisSetTtl(key, ttl)
	}
	return err
}
func RedisSendHmSetTtl(key *redisKeys.RedisKeys, ttl int64, data ...interface{}) error {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HMSet(context.TODO(), key.Key, data...)
	err := cmd.Err()
	logRedisCommand(key.Name, err, "HMSET", key.Key, data)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HMSET", err != nil, duration)
	if err != nil {
		return err
	}
	if ttl > 0 {
		RedisSetTtl(key, ttl)
	}
	return nil
}

// hmget
func RedisDoHMGet(key *redisKeys.RedisKeys, fields ...string) ([]string, error) {
	conn := GetRedisPool(key.Name)

	start := time.Now()
	cmd := conn.HMGet(context.TODO(), key.Key, fields...)
	res, err := cmd.Result()
	logRedisCommand(key.Name, err, "HMGET", key.Key, fields)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HMGET", err != nil, duration)
	rets := make([]string, 0, len(fields))
	for _, v := range res {
		rets = append(rets, fmt.Sprintf("%v", v))
	}
	if err != nil && err != redis.Nil {
		return rets, err
	}
	return rets, nil
}

func RedisDoHmGet(key *redisKeys.RedisKeys, fields []string) []int64 {
	var ret []int64
	res, err := RedisDoHMGet(key, fields...)
	if err != nil {
		return ret
	}
	for _, v := range res {
		if v == "" {
			continue
		}
		intv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ret = append(ret, intv)
	}
	return ret
}

// hmget
func RedisDoHmGetV2(key *redisKeys.RedisKeys, keys []string) (map[string]string, error) {
	k := []interface{}{key.Key}
	for _, v := range keys {
		k = append(k, v)
	}

	ret := map[string]string{}
	data, err := RedisDoHMGet(key, keys...)
	if err != nil {
		return nil, err
	}
	for i, v := range data {
		if v == "" {
			continue
		}

		key := keys[i]
		ret[key] = v
	}
	return ret, nil
}

// HINCRBY do 返回结果值得
func RedisDoHIncrBy(key *redisKeys.RedisKeys, field any, incr int64, ttl ...int64) (int64, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HIncrBy(context.TODO(), key.Key, fmt.Sprintf("%v", field), incr)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "HINCRBY", key.Key, fmt.Sprintf("%v", field), incr)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HINCRBY", err != nil, duration)

	if err != nil {
		return 0, err
	}
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}
	return ret, nil
}

func RedisDoHincrby(key *redisKeys.RedisKeys, key2 interface{}, inc int64, ttl ...int64) (int64, error) {
	return RedisDoHIncrBy(key, key2, inc, ttl...)
}

// HINCRBY
func RedisSendHincrby(key *redisKeys.RedisKeys, key2 interface{}, inc int64, ttl ...int64) error {
	_, err := RedisDoHIncrBy(key, key2, inc, ttl...)
	return err
}

// hexists
func RedisDoHExists(key *redisKeys.RedisKeys, field string) (bool, error) {
	conn := GetRedisPool(key.Name)
	start := time.Now()
	cmd := conn.HExists(context.TODO(), key.Key, field)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "HEXISTS", key.Key, field)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HEXISTS", err != nil, duration)
	return ret, err
}

func RedisHExist(key *redisKeys.RedisKeys, key2 interface{}) bool {
	ret, _ := RedisDoHExists(key, fmt.Sprintf("%v", key2))
	return ret

}

func RedisDoHDel(key *redisKeys.RedisKeys, fields ...string) error {
	conn := GetRedisPool(key.Name)
	args := []interface{}{key.Key}
	for _, f := range fields {
		args = append(args, f)
	}
	start := time.Now()
	cmd := conn.HDel(context.TODO(), key.Key, fields...)
	_, err := cmd.Result()
	logRedisCommand(key.Name, err, "HDEL", key.Key, fields)
	duration := time.Since(start)
	reportRedisStats(key.Name, "HDEL", err != nil, duration)
	return err
}

// hdel
func RedisSendHDel(key *redisKeys.RedisKeys, key2 interface{}) error {
	return RedisDoHDel(key, fmt.Sprintf("%v", key2))
}
