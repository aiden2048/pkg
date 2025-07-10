package cache

import (
	"reflect"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/errorMsg"
	"github.com/aiden2048/pkg/public/redisDeal"
	"github.com/aiden2048/pkg/public/redisKeys"
	jsoniter "github.com/json-iterator/go"
)

func GetAndRedisCache(rKey *redisKeys.RedisKeys, obj interface{}, fn func() (interface{}, *errorMsg.ErrRsp), ttl int64) *errorMsg.ErrRsp {
	s, err := redisDeal.RedisDoGetStrV2(rKey)
	if err != nil {
		logs.Errorf("err:%v", err)
		return errorMsg.RspError.Return(err)
	}

	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Pointer && typ.Kind() != reflect.Ptr {
		return errorMsg.RspError.Return("type of obj must be pointer")
	}

	if s != "" {
		switch obj.(type) {
		case *string:
			obj = &s
			return nil
		}

		if err := jsoniter.UnmarshalFromString(s, obj); err == nil {
			return nil
		}
	}

	resp, fnErr := fn()
	if fnErr != nil {
		return fnErr
	}

	obj = resp

	_ = redisDeal.RedisDoSet(rKey, resp, ttl)

	return nil
}
