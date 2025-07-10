package redisKeys

import (
	"fmt"
	"strings"
)

/*
服务user redis keys集合
*/
func GetUserNickNameCountKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("user.uinfo.%d.nickname.count", appId)
	return tmp
}

// 必须是 user. 开头
func InUserKey(key string) bool {
	if strings.Index(key, "user.") != 0 {
		return false
	}
	return true
}

// /用户信息
func GenUserInfoRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.uinfo.%d.%d", appid, uid)
	return tmp
}

// 上级
func GenUserUpShareRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.up_share.%d.%d", appid, uid)
	return tmp
}

// 最顶级
func GenUserTopLevelRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.top_level.%d.%d", appid, uid)
	return tmp
}

// 下级  没用
func GenUserDownShareRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.down_share.%d.%d", appid, uid)
	return tmp
}

// 下级数量
func GenUserDownNumRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.down_num.%d.%d", appid, uid)
	return tmp
}

// 下级更新时间 不能设置ttl
func GenUserDownTimeRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.down_time.%d.%d", appid, uid)
	return tmp
}

// /账号信息
func GenUserAccountRedisKeyByAcc(appid int32, account string, userType int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.account_acc.%d.%s.%d", appid, account, userType)
	return tmp
}
func GenUserAccountRedisKeyByUid(appid int32, uid uint64, userType int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.account_uid.%d.%d.%d", appid, uid, userType)
	return tmp
}
func GenUserAccountRedisKeyInvalidAcc(appid int32, account string, userType int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.no_account_acc.%d.%s.%d", appid, account, userType)
	return tmp
}
func GenUserAccountRedisKeyInvalidUid(appid int32, uid uint64, userType int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.no_account_uid.%d.%d.%d", appid, uid, userType)
	return tmp
}

// /客服
func GenUserJobNoRedisKey(appid int32, jobNo int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.jobno.%d.%d", appid, jobNo)
	return tmp
}
func GenUserJobNoTimeRedisKey(appid int32, jobNo int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.jobno_time.%d.%d", appid, jobNo)
	return tmp
}

// /银行卡绑定信息
func GenUserBankInfoRedisKeyByAcc(appid int32, account string, bindType int32, bindSubType ...int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.bank_acc.%d.%s.%d", appid, account, bindType)
	if len(bindSubType) > 0 && bindSubType[0] > 0 {
		tmp.Key = fmt.Sprintf("user.bank_acc.%d.%s.%d.%d", appid, account, bindType, bindSubType[0])
	}
	return tmp
}
func GenUserBankInfoRedisKeyByUid(appid int32, uid uint64, bindType int32, bindSubType ...int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.bank_uid.%d.%d.%d", appid, uid, bindType)
	if len(bindSubType) > 0 && bindSubType[0] > 0 {
		tmp.Key = fmt.Sprintf("user.bank_uid.%d.%d.%d.%d", appid, uid, bindType, bindSubType[0])
	}
	return tmp
}
func GenUserBankInfoRedisKeyInvalidAcc(appid int32, account string, bindType int32, bindSubType ...int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.no_bank_acc.%d.%s.%d", appid, account, bindType)
	if len(bindSubType) > 0 && bindSubType[0] > 0 {
		tmp.Key = fmt.Sprintf("user.no_bank_acc.%d.%s.%d.%d", appid, account, bindType, bindSubType[0])
	}
	return tmp
}
func GenUserBankInfoRedisKeyInvalidUid(appid int32, uid uint64, bindType int32, bindSubType ...int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.no_bank_uid.%d.%d.%d", appid, uid, bindType)
	if len(bindSubType) > 0 && bindSubType[0] > 0 {
		tmp.Key = fmt.Sprintf("user.no_bank_uid.%d.%d.%d.%d", appid, uid, bindType, bindSubType[0])
	}
	return tmp
}
func GenUserLabelRedisKey(appid int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.labels.%d.%d", appid, uid)
	return tmp
}

// /////异步存db
func GenUserSyncDbRedisKey() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.redis_update.synctask")
	return tmp
}

// 新users進程用的，跟原本的 v1 混再一起似乎有bug，先分開
func GenUserSyncDbRedisKeyV2() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_USER
	tmp.Key = fmt.Sprintf("user.redis_update.synctaskV2")
	return tmp
}

// 用户任务积分信息(旧)
func GetAppAllUserTaskPointKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.all.all.user.task.point.%d", InfoForeverKey, appId)
	return key
}

// 用户任务积分信息，每周过期
func GetAppWeekUserTaskPointKey(appId int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("user.task.week.point.%d.%d", appId, idate)
	return key
}

// 任务详细
func GetUserTaskRecordInfoKey(appId int32, taskKey string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("user.task.record.info.%d.%s", appId, taskKey)
	return key
}

// 任务活动数据累加
func GetUserTaskIncKey(appId int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("user.task.inc.%d.%d", appId, idate)
	return key
}

// 用户自定义数据
func GetUserClientJsonDataKey(appId int32, userId uint64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("user.client.json.data.%d.%d", appId, userId)
	return key
}
