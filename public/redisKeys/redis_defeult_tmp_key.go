package redisKeys

import (
	"fmt"
)

const (
	tmpRedisKey = "tmp.data"
)

//随时可以 丢弃得key

// 发送验证码 手机锁
func GenSendPhoneLockKey(phoneCode string, phone string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "send.sms.phone.one.lock." + phoneCode + "." + phone
	return key
}

// 发送验证码 一天 锁
func GenSendPhoneOneDayLockKey(phoneCode string, phone string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "send.sms.phone.one.day.lock." + phoneCode + "." + phone
	return key
}

// 记录号码上次的通道
func GetSendSmsPrevId(phone string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "send.sms.phone.one.day.lock." + phone
	return key
}

// 短地址top缓存
func GenUrlTopShortKey(url string) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = "topadmin.short." + url
	return key
}

func GetCommonKey(key string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = key
	return tmp
}

// ------------------------------------------------------------------------------------------------------------------------------------------------
// 修改记录
func GetUpdateGameInfo(app_id int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("default.manager.update.game.info.%d", app_id)
	return key
}

// job_no users
func GetGobNoUsersMapInfo(app_id int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("default.manager.get.gobno.users.%d", app_id)
	return key
}

func GetSeoPayUserKey(app_id int32) *RedisKeys {
	key := &RedisKeys{}
	key.Name = REDIS_INDEX_COMMON
	key.Key = fmt.Sprintf("report.user.first.pay.user.%d", app_id)
	return key
}

// / 补助金用户领取次数
func GenUserSubsidyPointTimesRedisKey(appid int32, date string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.subsidypoint.times.%d.%s", appid, date)
	return tmp
}

// / 补助金领取锁
func GenUserSubsidyPointLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.subsidypoint.lock.%d", uid)
	return tmp
}

// / 充值红包用户领取次数
func GenUserDoRechargeRedTimesRedisKey(appid int32, date int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.rechargered.times.%d.%d", appid, date)
	return tmp
}

// / 充值红包领取锁
func GenUserDoRechargeRedLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.rechargered.lock.%d", uid)
	return tmp
}

// / VIP奖励领取锁
func GenUserVIPRewardLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.vip.reward.lock.%d", uid)
	return tmp
}

// / VIP奖励查询锁
func GenCheckUserVIPRewardLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.check.user.vip.reward.lock.%d", uid)
	return tmp
}

/*
*
领取VIP的记录
*/
func GenVipRewardReviceKey(appId int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.vip.reward.revice.%d.%d", appId, uid)
	return tmp
}

// / 落地页同步进度
func GenLandPageMakeScheduleRedisKey(uuid string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("default.topadmin.landpage.makeschedule.%s", uuid)
	return tmp
}

// / 系统命令输出缓存
func GenLandPageCommandOutRedisKey(uuid string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("default.topadmin.landpage.commandout.%s", uuid)
	return tmp
}

// / 域名告警过滤
func GenNoWarnDomainRedisKey(domain string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("default.topcloud.domain.nowarn.%s", domain)
	return tmp
}

// IP所在地缓存信息
func GenIpGeoKey() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("tmp.data.ip.geo")
	return tmp
}

// / 倍投策略执行次数
func GenDBetPloyNumRedisKey(idate int64, appId int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("task.doublebet.ploy.no.%d.%d", appId, idate)
	return tmp
}

// / 倍投策略执行概率缓存
func GenDBetPloyRateRedisKey(k string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("task.doublebet.ploy.rate.%s", k)
	return tmp
}

// / 每日礼金领取次数
func GenUserDailyGiftMoneyDayTimesRedisKey(appid int32, date int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.daily.gift.money.day.times.%d.%d", appid, date)
	return tmp
}

// / 每周礼金领取次数
func GenUserDailyGiftMoneyWeekTimesRedisKey(appid int32, date int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.daily.gift.money.week.times.%d.%d", appid, date)
	return tmp
}

// / 每月礼金领取次数
func GenUserDailyGiftMoneyMonthTimesRedisKey(appid int32, date int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.daily.gift.money.month.times.%d.%d", appid, date)
	return tmp
}

// / 每日礼金领取锁
func GenUserDailyGiftMoneyLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.daily.gift.money.lock.%d", uid)
	return tmp
}

// / 保险金购买锁
func GenUserBuyInsuranceLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appface.user.insurance.buy.lock.%d", uid)
	return tmp
}

// / 限时首充返利领取锁
func GenUserLimitFirstPayBackLockRedisKey(uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("active.user.limit.first.pay.back.lock.%d", uid)
	return tmp
}

// / 用户留存缓存
func GenStayUserSummaryListRedisKey(appid int32, sid, date string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.stay.user.summary.list.%d.%s.%s", appid, sid, date)
	return tmp
}
func GenStayUserSummaryListLockRedisKey(appid int32, sid, date string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("manager.stay.user.summary.list.lock.%d.%s.%s", appid, sid, date)
	return tmp
}

// / 用户vip月礼金领取
func GetRedisUserMonthVipPointKey(appid int32, month string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.user.month.vip.point.%d.%s", InfoForeverKey, appid, month)
	return tmp
}

// / 用户vip周礼金领取
func GetRedisUserWeekVipPointKey(appid int32, week string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.user.week.vip.point.%d.%s", InfoForeverKey, appid, week)
	return tmp
}

// / 用户周打码量
func GetRedisUserWeekCodePointKey(appid int32, week int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.user.week.code.point.%d.%d", InfoForeverKey, appid, week)
	return tmp
}

// / 用户月打码量
func GetRedisUserMonthCodePointKey(appid int32, month int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.user.month.code.point.%d.%d", InfoForeverKey, appid, month)
	return tmp
}

// aff 请求数据缓存KEY
func GetAffRequestDataKey(key string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.aff.request.data.key.%s", InfoForeverKey, key)
	return tmp
}

// aff 请求数据缓存KEY
func GetAffRequestLenKey(key string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("%s.aff.request.len.key.%s", InfoForeverKey, key)
	return tmp
}

// gameStatusData key
func GetGameStatusDataKey(appId, gid int32, level string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.status.data.%d.%d.%s", appId, gid, level)
	return tmp
}

// gameStatusData prev key
func GetGameStatusDataPrevKey(appId, gid int32, level string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.prev.status.data.%d.%d.%s", appId, gid, level)
	return tmp
}

// change game data lock key
func ChangeGameDataLockKey(appId, gid int32, level string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.change.game.lock.%d.%d.%s", appId, gid, level)
	return tmp
}

// change game data lock key
func GetBatchCfgLockKey(platid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.batch.cfg.game.lock.%d", platid)
	return tmp
}

// 配置是否重载过
func GetBatchCfgReloadKey() *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.batch.cfg.reload")
	return tmp
}

// change game data key
func ChangeGameDataKey(appId, gid int32, level string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.temp.change.data.%d.%d.%s", appId, gid, level)
	return tmp
}

// minesG bet key
func GetMinesBetKey(appId, gid int32, level string, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("minesG.bet.uid.%d.%d.%s.%d", appId, gid, level, uid)
	return tmp
}

// quickG bet key
func GetQuickGBetKey(appId, gid int32, level string, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("quickG.bet.uid.%d.%d.%s.%d", appId, gid, level, uid)
	return tmp
}

func GetQuickGameSeedKey(appId, gid int32, level string, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("quickG.game.seed.%d.%d.%s.%d", appId, gid, level, uid)
	return tmp
}

func GetSingleGameCurrRateKey(appId, gid int32, level string, uid uint64, pointType int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("singleG.game.currRate.%d.%d.%s.%d.%d", appId, gid, level, uid, pointType)
	return tmp
}

// SingleG bet key
func GetSingleGBetKey(appId, gid int32, level string, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("SingleG.bet.uid.%d.%d.%s.%d", appId, gid, level, uid)
	return tmp
}

func GetRedisUserDataInfoKey(appid int32) string {
	return fmt.Sprintf("data.user.vip.data.record.%d", appid)
	//tmp := &RedisKeys{}
	//tmp.Name = REDIS_INDEX_COMMON
	//tmp.Key = fmt.Sprintf("data.user.vip.data.record.%d", appid)
	//return tmp
}

func GetRedisUserDataTmpInfoKey(appid int32, uid, idate int64) string {
	return fmt.Sprintf("data.user.vip.data.tmp.record.%d.%d.%d", appid, uid, idate)
}

// 返水奖励奖金详情
func GetRedisUserGameRebateBonusDetailKey(appId int32, userId uint64, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("user.game.rebate.bonus.detail.%d.%d.%d", appId, userId, idate)
	return tmp
}

// 返水配置记录
func GetRedisGameRebateConfigRecordKey(appId int32, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.rebate.config.record.%d.%d", appId, idate)
	return tmp
}

// 最近游戏记录
func GetRedisRecentGameListRecordKey(appId int32, uid uint64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("game.recent.gamelist.record.%d.%d", appId, uid)
	return tmp
}

// 忘记密码时的key
func GetRedisForgetPwdKey(appId int32, key string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("appserver.forget.pwd.%d.%s", appId, key)
	return tmp
}

// 充值提现上报sdk信息  ty：1 充值 2 提现
func GetRedisWealthAppSdkDataKey(appId, ty int32, order_id string) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("wealth.app_sdk_data.%d.%d.%s", appId, ty, order_id)
	return tmp
}

func GenLockRedisKey(app_id, uid uint64, key any) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("lock.%d.%d.%v", app_id, uid, key)
	return tmp
}

// 玩家中奖记录列表
func GetAllUserGameWinListKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("all.game.win.list.%d", appid)
	return tmp
}

func SingleGameFreePrizeCount(appId, gid int32, uid uint64, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("singleG.free.prize.count.%d.%d.%d.%d", appId, gid, uid, idate)
	return tmp
}

// 标记当天已经记录登录活跃
func GetUserTodayLoginFlagKey(appId int32, idate int64) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("user.today.login.activity.%d.%d", appId, idate)
	return tmp
}

// 玩家中奖记录列表
func GetAllUserGameWinStrKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("all.game.win.str.%d", appid)
	return tmp
}

// 玩家待同步uid集合
func GetSynvVipCoinUidKey(appid int32) *RedisKeys {
	tmp := &RedisKeys{}
	tmp.Name = REDIS_INDEX_COMMON
	tmp.Key = fmt.Sprintf("vip.sync.vipcoin.uids.%d", appid)
	return tmp
}
