package redisKeys

import "fmt"

// key 不可以改 也不可以丢弃值得
// ------------core----------------
func GenRedisIdFactoryKey() *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = "global.redis_id_factory"
	return k
}

func GetUserCodeTmpOrderKey(appId int32, uid uint64) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("user.data.code.tmp.order.%d.%d", appId, uid)
	return k
}

func GetUserCodeWinPointKey(appId int32, uid uint64) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("user.code.win.point.%d.%d", appId, uid)
	return k
}

func GetCodeOrderPointKey(_id string) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("user.code.order.%s", _id)
	return k
}

func GetArchiveDataKey() *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("archive.data.key")
	return k
}
