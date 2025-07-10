package redisKeys

import "golang.org/x/exp/errors/fmt"

// 缓存最后请求时间
func GetTaskRegPointKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "task.reg.point.last.time.key1"
	return key
}

// 缓存最后请求时间
func GetTaskRegPointSameIPKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "task.reg.point.same.ip.key"
	return key
}

// 缓存最后请求时间
func GetTaskRegPointSameMacKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "task.reg.point.same.mac.key"
	return key
}

// 缓存最后请求时间
func GetDataBindPointSameIPKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "data.bind.point.same.ip.key"
	return key
}

// 缓存最后请求时间
func GetDataBindPointSameMacKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "data.bind.point.same.mac.key"
	return key
}

// 缓存用户任务积分奖励领取记录
func GetTaskPointRewardGetRecordKey(appId int32, weekDate int64, userId uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("task.point.reward.get.record.%d.%d.%d", appId, weekDate, userId)
	return key
}
