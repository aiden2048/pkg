package redisDeal

import (
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/gomodule/redigo/redis"
)

const (
	ID_key_online_queue_seq = "online.msgseq"
)

// redis 自增ID
func GenRedisIncrId(src string, ininc ...int64) int64 {
	if src == "" {
		return -1
	}
	inc := int64(1)
	if len(ininc) > 0 && ininc[0] > 0 {
		inc = ininc[0]
	}
	key := redisKeys.GenRedisIdFactoryKey()
	seq, err := RedisDoHincrby(key, src, inc)
	if err != nil && err != redis.ErrNil {
		logs.LogError("redis RedisDoHincrby key:%s,src:%s failed:%s ", key, src, err.Error())
		return -1
	}

	return seq
}
