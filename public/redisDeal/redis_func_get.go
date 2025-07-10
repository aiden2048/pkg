package redisDeal

import (
	"context"
	"reflect"
	"time"

	"github.com/aiden2048/pkg/public/redisKeys"
)

func RedisDoMGet(list interface{}, keys ...*redisKeys.RedisKeys) map[string]string {
	if len(keys) == 0 {
		return nil
	}

	typ := reflect.TypeOf(list)
	if typ.Kind() != reflect.Pointer {
		return nil
	}

	typ = typ.Elem()
	if typ.Kind() != reflect.Slice {
		return nil
	}

	var (
		redisName  string
		keyStrList []string
	)
	for _, key := range keys {
		if key.Name != "" && redisName == "" {
			redisName = key.Name
		}
		keyStrList = append(keyStrList, key.Key)
	}
	conn := GetRedisPool(redisName)

	start := time.Now()
	cmd := conn.MGet(context.TODO(), keyStrList...)
	values, err := cmd.Result()

	logRedisCommand(redisName, err, "MGET", keyStrList)
	duration := time.Since(start)
	reportRedisStats(redisName, "MGET", err != nil, duration)
	retMap := map[string]string{}
	for i, val := range values {
		if val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok {
			continue
		}
		retMap[keyStrList[i]] = strVal

	}

	return retMap
}
