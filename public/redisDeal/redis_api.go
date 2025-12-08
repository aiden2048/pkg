package redisDeal

import (
	"context"
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/stat"
	"github.com/aiden2048/pkg/public/redisKeys"

	//"github.com/gomodule/redigo/redis"
	"github.com/redis/go-redis/v9"
)

func RedisConnDo(key *redisKeys.RedisKeys, conn *redis.ClusterClient, commandName string, args ...interface{}) (*redis.Cmd, error) {
	start := time.Now()
	if !check(key) {
		logs.Errorf("key =%s 不属于 %s", key.Key, key.Name)
		return nil, fmt.Errorf("key =%s 不属于 %s", key.Key, key.Name)
	}
	allArgs := append([]interface{}{commandName}, args...)

	cmd := conn.Do(context.Background(), allArgs...)
	err := cmd.Err()

	logRedisCommand(commandName, err, args...)

	duration := time.Since(start)
	reportRedisStats(key.Name, commandName, err != nil && err != redis.Nil, duration)

	return cmd, err
}
func logRedisCommand(command string, err error, args ...interface{}) {
	if err != nil && err != redis.Nil {
		logs.Errorf("redis:%s, args:%+v, err:%v", command, args, err)
	}
	if command == "EXPIRE" {
		logs.Debugf("redis:%s, args:%+v, err:%v", command, args, err)
	} else {
		logs.Debugf("redis:%s, args:%+v, err:%v", command, args, err)
	}
}
func reportRedisStats(redisName, command string, isError bool, duration time.Duration) {
	ret := 0
	if isError {
		ret = 1
	}
	baseStat := fmt.Sprintf("rp:redisDeal.Do.%s.%s", redisName, command)
	stat.ReportStat(baseStat, ret, duration)

	if duration > time.Second {
		logs.Errorf("RedisTimeout %s 执行时间:%v > 1000ms: %s %+v", redisName, duration, command, ret)
		return
	}

	delayBuckets := []time.Duration{100, 200, 300, 400, 500}
	for _, bucket := range delayBuckets {
		if duration > bucket*time.Millisecond && duration <= (bucket+100)*time.Millisecond {
			stat.ReportStat(fmt.Sprintf("rp:redisDeal.DoTimeout.%s.Do.%d.%s", redisName, bucket, command), ret, duration)
			break
		}
	}
}
