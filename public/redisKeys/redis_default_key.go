package redisKeys

import (
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/utils"
)

//key 值可以丢弃。 但是会有一点小影响  过后不影响

// ------------online ----------------
// 在游戏的用户数
func GenUserInGameSet(appid int32) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("online.user.ingame.set.%d", appid)
	return k
}

// 在线用户数统分钟计锁
func GenStatMinuteOnlineUserLockKey() *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprint("GenStatMinuteOnlineUserLockKey.lock")
	return k
}

// 在线用户数
func GenUserTotalOnlineSet(appid int32, gameid int32) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("online.user.total.set.%d.%d", appid, gameid)
	return k
}

// ----robot---
// 当然所有机器人机器人
func GetRichAllRobotInfoKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("robot.info.list")
	return key
}

// 当然所有机器人uid对应nanme
func GetRichAllRobotNameInfoKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("robot.all.name.info.list")
	return key
}

// 短信发送记录
func GetSmsNumInfoKey(month int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.sms.send.record.%d", InfoForeverKey, month)
	return key
}

/*
*
第三方上分记录
*/
func GetTransferSaveTmpKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.user.transfersave.%d", InfoForeverKey, appId)
	return key
}

/*
*
沙巴最后的注单version key
*/
func GetShabaVersionKey(appId, gid int32, platform string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.shaba.version.key.%d.%d%s", InfoForeverKey, appId, gid, platform)
	return key
}
func GetLastTimeKeyKey(appId, gid int32, platform string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.last.key.%d.%d%s", InfoForeverKey, appId, gid, platform)
	return key
}
func GetBGSOFTAuthTokenKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.bgsoft.autk.token.key", InfoForeverKey)
	return key
}
func GetSkywindAuthTokenKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.skywind.autk.token.key", InfoForeverKey)
	return key
}
func GetSkywindSeamlessAuthTokenKey() *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.skywindseamless.autk.token.key", InfoForeverKey)
	return key
}
func GetGameAgentIncrKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.incr.key.%d", InfoForeverKey, appId)
	return key
}

func GetInfinGameKeyToGuidKey(token string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.infingame.key.guid.%s", InfoForeverKey, token)
	return key
}

func GetEasyGameKeyByUid(token string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.easygame.key.uid.%s", InfoForeverKey, token)
	return key
}

func GetEasyGameKeyByAgent(token string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.easygame.key.agent.%s", InfoForeverKey, token)
	return key
}
func GetEasyGameTmpKeyByAgent(token string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.easygame.tmpkey.agent.%s", InfoForeverKey, token)
	return key
}

func GetEasyGameOfUriKey(token string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.easygame.uri.key.uid.%s", InfoForeverKey, token)
	return key
}

func GetGameAgentSBOKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.sbo.key.%d", InfoForeverKey, appId)
	return key
}

func GetMwDomainKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.mw.domain.key.%d", InfoForeverKey, appId)
	return key
}

/*
*
第三方上分记录时间
*/
func GetTransferSaveTimeKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.user.transfer.time.%d", InfoForeverKey, appId)
	return key
}

/*
*
第三方上分记录游戏ID
*/
func GetTransferSaveGidKey(appId int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.user.transfer.gid.%d", InfoForeverKey, appId)
	return key
}

func GetGameAgentCheckSaveKey(appId, gid int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.check.point.save.%d.%d", InfoForeverKey, appId, gid)
	return key
}
func GetGameAgentCheckTakeKey(appId, gid int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("%s.gameAgent.check.point.take.%d.%d", InfoForeverKey, appId, gid)
	return key
}

/*
*
游戏数据记录缓存
*/
func GetGameRewardData(appId, gid int32, rule, level string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("task.game.reward.data.%d.%d.%s.%s", appId, gid, rule, level)
	return key
}

// 倍投玩家活跃记录
func GetDoubleActiveUserKey(appid int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("task.user.active.data.%d.%d", appid, idate)
	return key
}

// /////用户id与注册时间
func GenUserRegTimeRedisKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("register.user.time.zset.%d", appId)
	return tmp
}
func GenUserRegTimeLoatTimeRedisKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("register.user.load.time.%d", appId)
	return tmp
}

/*
*
玩家游戏时长记录
*/
func UserInGameTimeKey(appId int32, idate int64) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("online.user.in.game.time.%d.%d", appId, idate)
	return key
}
func GenTopAdminLoginPhoneCodeRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("topadmin.login.phone.code.%d", uid)
	return tmp
}

// 今日是否已经请求私人跑马灯
func GetPrivateMaqueeKey(appid int32, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.private.maquee.req.%d.%d", appid, idate)
	return tmp
}

// 今日是否已经推送私人跑马灯
func GetPushPrivateMaqueeKey(appid int32, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.private.maquee.push.%d.%d", appid, idate)
	return tmp
}

// 用户打码量
func GetAppUserLavePutCodeKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.app.user.lave.put.code.key.%d", appid)
	return tmp
}

// 用户上一次提现打码倍数
func GetAppUserLastPutPenKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.app.user.last.put.pen.key.%d", appid)
	return tmp
}

// 新元缓存信息
func GetXyCacheExtInfoKey(orderNo string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.app.xy.cache.info.key.%s", orderNo)
	return tmp
}

// 用户今日最大充值
func GetAppUserDayMaxPayKey(appid int32, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.app.user.max.pay.key.%d.%d", appid, idate)
	return tmp
}

func GetAppUserFirstForeverKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.app.user.first.forever.pay.%d", appId)
	return tmp
}

func GetAppUserFirstKey(appId int32, uid int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.app.user.first.pay.info.%d.%d", appId, uid)
	return tmp
}

func GetSyncAbnormalStock() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.root.sync.abnotmal.stock")
	return tmp
}

func GetRiseFallBtcMarketUsdtKey(gid int32, level, market string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("gameAgent.rise.fall.price.%d.%s.%s", gid, level, market)
	return tmp
}

func GetRiseFallBtcUsdtKey(gid int32, level string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("gameAgent.rise.fall.price.%d.%s", gid, level)
	return tmp
}

// 礼包码rule_id计数
func GetRedisGiftCodeRuleIdIncrKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.giftcode.ruleid.incr.%d", appId)
	return tmp
}

// 礼包码group_id计数
func GetRedisGiftCodeGroupIdIncrKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.giftcode.groupid.incr.%d", appId)
	return tmp
}

// 礼包码item_id计数
func GetRedisGiftCodeItemIdIncrKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.giftcode.itemid.incr.%d", appId)
	return tmp
}

// 玩家重复请求下分
func GetUserReplyTakeKey(appId int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("gameAgent.user.reply.take.%d.%d", appId, uid)
	return tmp
}

// 优惠道具配置id计数
func GetRedisPromoItemCfgIdIncrKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.promoitem.cfg.id.incr.%d", appId)
	return tmp
}

// 优惠道具id计数
func GetRedisUserPromoItemIdIncrKey(appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.user.promoitem.id.incr.%d", appId)
	return tmp
}

// robot free
func GetRobotFreeUids() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("robot.free.uids")
	return tmp
}

// robot prev uid start
func GetRobotFreePrevStart() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("robot.free.prev.start")
	return tmp
}

// PushMsgToUser
func GetPushMsgToUser() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = "PushMsgToUser"
	return tmp
}

// 新版注册用户id
func GenAppUserNewSet(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appserver.app.user.set.%d", appid)
	return tmp
}

//*******************在线时长统计**************************

// app在线用户集合
func GenAppUserOnlineSet(appid int32) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("online.app.user.set.%d", appid)
	return k
}

// 用户实时子游戏信息
func GenUserGidHash(appid int32) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("online.user.gid.hash.%d", appid)
	return k
}

// 子游戏实时在线人数
func GenAppGidOnlineHash(appid int32) *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("online.app.gid.hash.%d", appid)
	return k
}

// 收藏游戏记录
func GetRedisCollectGameListRecordKey(appId int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.collect.gamelist.record.%d.%d", appId, uid)
	return tmp
}

// 转盘轮播list
func GetAllUserNewZhuanPanListKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("all.zhuan.pan.list.new.%d", appid)
	return tmp
}

// 索引上报队列key
func GenMongoIndexReportSetKey() *RedisKeys {
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("common.mongo.index.set.%d", frame.GetPlatformId())
	return k
}

// 每个表名，每类查询次数
func GenMongoQueryHashKey(dbName, tbName string, idate int64) *RedisKeys {
	if idate == 0 {
		idate = utils.GetIDate(time.Now().Unix())
	}
	k := &RedisKeys{}
	k.Name = REDIS_INDEX_COMMON
	k.Key = fmt.Sprintf("common.mongo.query.hash.%s.%s.%d", dbName, tbName, idate)
	return k
}
