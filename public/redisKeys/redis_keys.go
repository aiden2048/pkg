package redisKeys

import (
	"fmt"
)

func GenCommonRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf(format, a...)
	return k
}

func GenEcsRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_ECS
	k.Key = fmt.Sprintf(k.Name+"."+format, a...)
	return k
}

func GenUserRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_USER
	k.Key = fmt.Sprintf(k.Name+"."+format, a...)
	return k
}

func GenSessionRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_SESSION
	k.Key = fmt.Sprintf(k.Name+"."+format, a...)
	return k
}

func GenRiskRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_RISK
	k.Key = fmt.Sprintf(k.Name+"."+format, a...)
	return k
}

func GenPayRedisKey(format string, a ...any) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_PAY
	k.Key = fmt.Sprintf(k.Name+"."+format, a...)
	return k
}
