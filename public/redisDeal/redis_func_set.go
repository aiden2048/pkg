package redisDeal

import (
	"context"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/aiden2048/pkg/utils"
	jsoniter "github.com/json-iterator/go"
)

// Sadd
func RedisSendSadd(key *redisKeys.RedisKeys, data interface{}, ttl ...int64) error {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	str, ok := data.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = datastr
	}
	cmd := conn.SAdd(context.TODO(), key.Key, str)
	_, err := cmd.Result()
	logRedisCommand(key.Name, err, "SADD", key.Key, data)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SADD", err != nil, duration)
	return err

}

// 从redis set 里面检查 是否存在 这个值 sismember 1有
func RedisDoSisMember(key *redisKeys.RedisKeys, value interface{}) int64 {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	str, ok := value.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(value)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, value, err)
			return 0
		}
		str = datastr
	}
	cmd := conn.SIsMember(context.TODO(), key.Key, str)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SISMEMBER", key.Key, value)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SISMEMBER", err != nil, duration)
	if err != nil {
		return 0
	}
	if ret {
		return 1
	}
	return 0
}

// 从redis set 里面检查 是否存在 这个值 sismember 1有
func RedisDoSisMemberV2(key *redisKeys.RedisKeys, value interface{}) (bool, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	str, ok := value.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(value)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, value, err)
			return false, err
		}
		str = datastr
	}
	cmd := conn.SIsMember(context.TODO(), key.Key, str)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SISMEMBER", key.Key, value)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SISMEMBER", err != nil, duration)
	return ret, err
}

// SCARD
func RedisDoSCARD(key *redisKeys.RedisKeys) int {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.SCard(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SCARD", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SCARD", err != nil, duration)

	return int(ret)
}
func RedisDoZCARD(key *redisKeys.RedisKeys) int {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZCard(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZCARD", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZCARD", err != nil, duration)
	return int(ret)
}

func RedisDoSPop(key *redisKeys.RedisKeys) (string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.SPop(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SPOP", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SPOP", err != nil, duration)
	return ret, err
}

// SPOP 随机pop一个元素
func RedisDoSPOP(key *redisKeys.RedisKeys) int64 {
	ret, err := RedisDoSPop(key)
	if err != nil {
		return 0
	}
	return utils.StrToInt64(ret)
}

// SPOP 随机pop一个元素
func RedisDoSPopStr(key *redisKeys.RedisKeys) string {
	ret, _ := RedisDoSPop(key)

	return ret
}
func RedisDoSMembers(key *redisKeys.RedisKeys) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.SMembers(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SMEMBERS", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SMEMBERS", err != nil, duration)
	return ret, err
}

// SMEMBERS 返回集合所有成员
func RedisDoSMEMBERS(key *redisKeys.RedisKeys) []int64 {
	res, err := RedisDoSMembers(key)
	if err != nil {
		return nil
	}
	if len(res) == 0 {
		return nil
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets
}

// SMEMBERS 返回集合所有成员
func RedisDoSMEMBERSStr(key *redisKeys.RedisKeys) []string {
	rets, _ := RedisDoSMembers(key)
	return rets
}
func RedisDoSRem(key *redisKeys.RedisKeys, members ...interface{}) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.SRem(context.TODO(), key.Key, members...)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "SREM", key.Key, members)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SREM", err != nil, duration)
	return ret, err
}

// SREM 删除指定成员
func RedisDoSREM(key *redisKeys.RedisKeys, value interface{}) int64 {
	str, ok := value.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(value)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, value, err)
			return 0
		}
		str = datastr
	}
	ret, err := RedisDoSRem(key, str)
	if err != nil {
		return 0
	}
	return ret
}
