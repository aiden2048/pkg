package redisKeys

import (
	"fmt"
	"strings"
)

/*
服务pay redis keys集合
*/

//必须是 pay. 开头
func InPayKey(key string) bool {
	if strings.Index(key, "pay.") != 0 {
		return false
	}
	return true
}

/// 订单响应信息
func GetPayOrderResponseRedisKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.order.response.%s", orderkey)
	return tmp
}

/// 订单信息
func GetPayOrderParamsRedisKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.payorder.params.%s", orderkey)
	return tmp
}

/// 订单信息
func GetNoPayOrderInfoKey(app_id int32, user_id uint64, pay_key string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.payorder.info.key.%d.%d.%s", app_id, user_id, pay_key)
	return tmp
}

/// 订单信息
func GetUserNowPayOrderIDKey(app_id int32, user_id uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.now.payorder.id.key.%d.%d", app_id, user_id)
	return tmp
}

/// 代付订单信息
func GetProxyPayOrderParamsKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.proxypayorder.params.%s", orderkey)
	return tmp
}

/// 代付订单信息
func GetProxyPayOrderVerfyKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.proxypayorder.params.%s", orderkey)
	return tmp
}

// 跳转页面缓存
func GetPayOrderCallBackKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.payorder.callback.%s", orderkey)
	return tmp
}

/// 订单信息
func GetPayLockRedisKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.pay.lock.%s", orderkey)
	return tmp
}

/// 订单查询频率锁定
func GetPayOrderQueryLockRedisKey(orderkey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.order.query.lock.%s", orderkey)
	return tmp
}

/// 订单提醒
func GetPayRemindLockRedisKey(orderKey string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_PAY
	tmp.Key = fmt.Sprintf("pay.pay.remind.lock.%s", orderKey)
	return tmp
}
