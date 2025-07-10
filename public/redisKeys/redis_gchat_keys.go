package redisKeys

import "fmt"

// gchat

// 记录上次发消息时间
func GetUserLastChatTimeKey(appId int32, uid uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("gchat.user.last.%d.%d", appId, uid)
	return key
}

func GetGchatMsgKey(appId int32, level string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("gchat.user.msg.%d.%s", appId, level)
	return key
}
func GetGchatMsgId(appId int32, level string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("gchat.user.msg.id.%d.%s", appId, level)
	return key
}
