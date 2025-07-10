package redisDeal

import (
	"context"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/aiden2048/pkg/public/redisKeys"

	jsoniter "github.com/json-iterator/go"
)

// 发布订阅
// 发送消息
// PUBLISH
func RedisDoPUBLISH(key *redisKeys.RedisKeys, data interface{}) error {
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
	conn := GetRedisPool(key.Name)
	startTime := time.Now()
	cmd := conn.Publish(context.TODO(), key.Key, str)
	_, err := cmd.Result()
	logRedisCommand(key.Name, err, "PUBLISH", key.Key, str)
	duration := time.Since(startTime)
	reportRedisStats(key.Name, "PUBLISH", err != nil, duration)
	return err
}
