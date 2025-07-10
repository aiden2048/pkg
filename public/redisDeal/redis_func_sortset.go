package redisDeal

import (
	"context"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/aiden2048/pkg/utils"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/errors/fmt"

	jsoniter "github.com/json-iterator/go"
)

func RedisDoSScen(key *redisKeys.RedisKeys, cursor, count int64) (int64, []string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.SScan(context.TODO(), key.Key, uint64(cursor), "COUNT", count)
	ret, cur, err := cmd.Result()
	logRedisCommand(key.Name, err, "SSCAN", key.Key, uint64(cursor), "COUNT", count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "SSCAN", err != nil, duration)
	return int64(cur), ret, err
}

func RedisDoZScan(key *redisKeys.RedisKeys, cursor, count int64) (int64, []string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZScan(context.TODO(), key.Key, uint64(cursor), "COUNT", count)
	ret, cur, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZSCAN", key.Key, uint64(cursor), "COUNT", count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZSCAN", err != nil, duration)

	return int64(cur), ret, err
}
func RedisDoZscan(key *redisKeys.RedisKeys, cursor, count int64) (int64, []string, error) {

	return RedisDoZScan(key, cursor, count)
}
func RedisDoZAdd(key *redisKeys.RedisKeys, score int64, member string, ttl ...int64) error {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	mem := redis.Z{
		Score:  float64(score),
		Member: member,
	}
	cmd := conn.ZAdd(context.TODO(), key.Key, mem)
	_, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZADD", key.Key, mem)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZADD", err != nil, duration)

	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}
	return err
}

func RedisDoZScore(key *redisKeys.RedisKeys, member string) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZScore(context.TODO(), key.Key, member)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZSCORE", key.Key, member)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZSCORE", err != nil, duration)

	return int64(ret), err
}

// ZCARD 获取成员数
func RedisDoZCard(key *redisKeys.RedisKeys) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZCard(context.TODO(), key.Key)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZCARD", key.Key)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZCARD", err != nil, duration)

	return ret, err
}

// ZINCRBY
func RedisDoZIncrBy(key *redisKeys.RedisKeys, increment int64, member string) (float64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZIncrBy(context.TODO(), key.Key, float64(increment), member)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZINCRBY", key.Key, increment, member)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZINCRBY", err != nil, duration)

	return ret, err
}
func RedisDoZincrby(key *redisKeys.RedisKeys, increment int64, member string, ttl ...int64) error {
	_, err := RedisDoZIncrBy(key, increment, member)
	if len(ttl) > 0 && ttl[0] > 0 {
		RedisSetTtl(key, ttl[0])
	}
	return err
}
func RedisDoZCount(key *redisKeys.RedisKeys, min, max string) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.ZCount(context.TODO(), key.Key, min, max)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZCOUNT", key.Key, min, max)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZCOUNT", err != nil, duration)
	return ret, err
}

func RedisDoZRevRangeByScore(key *redisKeys.RedisKeys, max, min string, count int64) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRevRangeByScore(context.TODO(), key.Key, &redis.ZRangeBy{
		Max:    max,
		Min:    min,
		Offset: 0,
		Count:  count,
	})

	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZREVRANGEBYSCORE", key.Key, max, min, "LIMIT", 0, count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREVRANGEBYSCORE", err != nil, duration)

	return ret, err
}
func RedisDoZRevRangeByScoreWithScore(key *redisKeys.RedisKeys, max, min string, count int64) ([]redis.Z, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRevRangeByScoreWithScores(context.TODO(), key.Key, &redis.ZRangeBy{
		Max:    max,
		Min:    min,
		Offset: 0,
		Count:  count,
	})

	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZREVRANGEBYSCORE", key.Key, max, min, "LIMIT", 0, count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREVRANGEBYSCORE", err != nil, duration)

	return ret, err
}

// 获取Key 分数从大到小 排列的 的成员 start (分数)等于 到end等于  返回 int64
func RedisDoZrevrangebyscoreInts(key *redisKeys.RedisKeys, start interface{}, end interface{}) []int64 {
	res, _ := RedisDoZRevRangeByScore(key, fmt.Sprintf("%v", start), fmt.Sprintf("%v", end), 0)

	if len(res) == 0 {
		return nil
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets
}
func RedisDoZrevrangebyscoreIntsV2(key *redisKeys.RedisKeys, max interface{}, min interface{}) ([]int64, error) {
	res, err := RedisDoZRevRangeByScore(key, fmt.Sprintf("%v", max), fmt.Sprintf("%v", min), 0)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("没有数据")
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets, nil

}

// 获取Key 分数从大到小 排列的 的成员 start (分数)等于 到end等于  返回 string
func RedisDoZrevrangebyscoreStrings(key *redisKeys.RedisKeys, start interface{}, end interface{}) []string {
	res, _ := RedisDoZRevRangeByScore(key, fmt.Sprintf("%v", start), fmt.Sprintf("%v", end), 0)
	return res

}
func RedisDoZRangeByScore(key *redisKeys.RedisKeys, min, max string, count int64) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	conn.ZRangeByScoreWithScores(context.TODO(), key.Key, &redis.ZRangeBy{})
	cmd := conn.ZRangeByScore(context.TODO(), key.Key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: 0,
		Count:  count,
	})

	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZRANGEBYSCORE", key.Key, min, max, "LIMIT", 0, count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZRANGEBYSCORE", err != nil, duration)

	return ret, err
}
func RedisDoZRangeByScoreWithScore(key *redisKeys.RedisKeys, min, max string, count int64) ([]redis.Z, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	conn.ZRangeByScoreWithScores(context.TODO(), key.Key, &redis.ZRangeBy{})
	cmd := conn.ZRangeByScoreWithScores(context.TODO(), key.Key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: 0,
		Count:  count,
	})

	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZRANGEBYSCORE", key.Key, min, max, "LIMIT", 0, count)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZRANGEBYSCORE", err != nil, duration)

	return ret, err
}

// 获取Key 分数从小到大 排列的 的成员 start (分数)等于 到end等于  返回 int64数组
func RedisDoZrangerbyscoreInts(key *redisKeys.RedisKeys, start interface{}, end interface{}) []int64 {
	res, _ := RedisDoZRangeByScore(key, fmt.Sprintf("%v", start), fmt.Sprintf("%v", end), 0)
	if len(res) == 0 {
		return nil
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets
}

// 获取Key 分数从小到大 排列的 的成员 start (分数)等于 到end等于  返回 string数组
func RedisDoZrangerbyscoreStr(key *redisKeys.RedisKeys, start interface{}, end interface{}) []string {
	res, _ := RedisDoZRangeByScore(key, fmt.Sprintf("%v", start), fmt.Sprintf("%v", end), 0)
	return res
}

func RedisDoZRemRangeByScore(key *redisKeys.RedisKeys, min, max string) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRemRangeByScore(context.TODO(), key.Key, min, max)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZREMRANGEBYSCORE", key.Key, min, max)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREMRANGEBYSCORE", err != nil, duration)

	return ret, err
}

// ZREMRANGEBYSCORE  移除
func RedisSendZremrangebyscore(key *redisKeys.RedisKeys, start interface{}, end interface{}) error {
	_, err := RedisDoZRemRangeByScore(key, fmt.Sprintf("%v", start), fmt.Sprintf("%v", end))
	return err
}
func RedisDoZRemRangeByRank(key *redisKeys.RedisKeys, start, stop int64) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRemRangeByRank(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZREMRANGEBYRANK", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREMRANGEBYRANK", err != nil, duration)

	return ret, err
}

func RedisDoZRange(key *redisKeys.RedisKeys, start, stop int64) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRange(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZRANGE", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZRANGE", err != nil, duration)

	return ret, err
}
func RedisDoZRangeWithScore(key *redisKeys.RedisKeys, start, stop int64) ([]redis.Z, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRangeWithScores(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZRANGE", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZRANGE", err != nil, duration)

	return ret, err
}

// ZRANGE, 获取Key 指定下标区间内的成员, 成员的位置按分数值递增(从小到大)来排列, 返回 int64数组
func RedisDoZrangeInts(key *redisKeys.RedisKeys, start, stop int64) []int64 {
	res, _ := RedisDoZRange(key, start, stop)
	if len(res) == 0 {
		return nil
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets
}

// ZRANGE, 获取Key 指定下标区间内的成员, 成员的位置按分数值递增(从小到大)来排列, 返回 string数组
func RedisDoZrangeStr(key *redisKeys.RedisKeys, start, stop int64) []string {
	rets, _ := RedisDoZRange(key, start, stop)
	return rets
}
func RedisDoZRevRange(key *redisKeys.RedisKeys, start, stop int64) ([]string, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRevRange(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZREVRANGE", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREVRANGE", err != nil, duration)

	return ret, err
}
func RedisDoZRevRangeWithScore(key *redisKeys.RedisKeys, start, stop int64) ([]redis.Z, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRevRangeWithScores(context.TODO(), key.Key, start, stop)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZREVRANGE", key.Key, start, stop)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREVRANGE", err != nil, duration)

	return ret, err
}

// ZREVRANGE, 获取Key 指定下标区间内的成员, 成员的位置按分数值递减(从大到小)来排列, 返回 int64数组
func RedisDoZrevrangeInts(key *redisKeys.RedisKeys, start, stop int64) []int64 {
	res, _ := RedisDoZRevRange(key, start, stop)
	if len(res) == 0 {
		return nil
	}
	rets := make([]int64, len(res))
	for i, v := range res {
		rets[i] = utils.StrToInt64(v)
	}
	return rets
}

// ZREVRANGE, 获取Key 指定下标区间内的成员, 成员的位置按分数值递减(从大到小)来排列, 返回 string数组
func RedisDoZrevrangeStr(key *redisKeys.RedisKeys, start, stop int64) []string {
	rets, _ := RedisDoZRevRange(key, start, stop)
	return rets
}

func RedisDoZRevRank(key *redisKeys.RedisKeys, member string) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRevRank(context.TODO(), key.Key, member)
	ret, err := cmd.Result()

	logRedisCommand(key.Name, err, "ZREVRANK", key.Key, member)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREVRANK", err != nil, duration)

	return ret, err
}

// ZREVRANK 获取成员的排名, 返回<0表示没有这个成员
func RedisDoZrevrank(key *redisKeys.RedisKeys, data interface{}) int64 {

	str, ok := data.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return -1
		}
		str = datastr
	}
	ret, err := RedisDoZRevRank(key, str)
	if err != nil {
		return -1
	}
	return ret
}

func RedisDoZRem(key *redisKeys.RedisKeys, members ...string) (int64, error) {
	conn := GetRedisPool(key.Name)
	startTime := time.Now()

	cmd := conn.ZRem(context.TODO(), key.Key, members)
	ret, err := cmd.Result()
	logRedisCommand(key.Name, err, "ZREM", key.Key, members)

	duration := time.Since(startTime)
	reportRedisStats(key.Name, "ZREM", err != nil, duration)

	return ret, err
}

// ZREM 删除成员
func RedisDoZrem(key *redisKeys.RedisKeys, data interface{}) error {
	if !check(key) {
		logs.Errorf("key =%s 不属于 %s", key.Key, key.Name)
		return errorMsg.Err_NoRedisTmpl
	}
	str, ok := data.(string)
	if !ok { //不是字符串
		datastr, err := jsoniter.MarshalToString(data)
		if err != nil {
			logs.Errorf("解析失败 key:%s,value:%+v error:%+v", key.Key, data, err)
			return err
		}
		str = datastr
	}
	_, err := RedisDoZRem(key, str)
	return err
}

func RedisDoZExist(key *redisKeys.RedisKeys, member string) (bool, error) {
	_, err := RedisDoZScore(key, member)
	if err != nil {
		return false, err
	}
	return true, nil
}
