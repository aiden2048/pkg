package configCache

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/aiden2048/pkg/utils/baselib"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"
	jsoniter "github.com/json-iterator/go"
)

var ReloadTime = sync.Map{}
var WaitReloadKey = sync.Map{}

/*
// 只有一个表需要通知的时候, 继续调用老的接口名

	func NotifyReloadConfigCache(key string) {
		//NotifyReloadConfigCacheSingleKey(key)
		NotifyReloadConfigCacheMutilKey(key)
	}

// top通知更新多个表的时候, 调用这个接口

	func NotifyReloadConfigCaches(keys []string) {
		if len(keys) == 0 {
			return
		}
		for _, key := range keys {
			NotifyReloadConfigCacheSingleKey(key)
		}

		NotifyReloadConfigCacheMutilKey(keys...)
	}

	func NotifyReloadConfigCacheSingleKey(key string) {
		go func(k string) {
			time.Sleep(3 * time.Second)
			frame.NotifyReloadConfig(ConfigCacheKey, k)
		}(key)
	}
*/
func NotifyReloadConfigCacheMutilKey(keys ...string) {
	go func(k []string) {
		time.Sleep(3 * time.Second)
		frame.NotifyReloadConfig(ConfigCacheKeys, k)
	}(keys)
}

var _reloadConfigHandler func(data []string)

func SetReloadConfigHandler(h func([]string)) {
	_reloadConfigHandler = h
}
func reloadConfigHandler(data []byte) {

	var keys []string
	err := jsoniter.Unmarshal(data, &keys)
	if err != nil {
		var key string
		jsoniter.Unmarshal(data, &key)
		keys = append(keys, key)
	}

	//每个进程随机等待下, 避免同时读
	time.Sleep(time.Second * time.Duration(rand.Int63n(10)))

	//如果包含了 c_environment, 直接刷新整个进程
	if utils.InArray(keys, "c_environment") {
		logs.Print("config_reload 收到重载通知", keys, "包含 c_environment , 直接重载整个进程")
		logs.Console("config_reload 收到重载通知", keys, "包含 c_environment , 直接重载整个进程")
		baselib.ReloadServerConfig(data)
		return
	}
	logs.Print("config_reload 收到重载通知", keys)
	if _reloadConfigHandler != nil {
		logs.Print("config_reload 收到重载通知", keys, "托管给业务自己处理了")
		logs.Console("config_reload 收到重载通知", keys, "托管给业务自己处理了")
		_reloadConfigHandler(keys)
	} else {
		ReloadConfig(keys)
	}

}

type ConfigInterfaceList []ConfigInterface

func (p ConfigInterfaceList) Len() int { return len(p) }
func (p ConfigInterfaceList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ConfigInterfaceList) Less(i, j int) bool {
	return p[i].GetPower() > p[j].GetPower()
}
func checkIsMyConfig(c ConfigInterface, key string) bool {
	for _, v := range c.ListenTables() {
		if v == "*" || v == key {
			return true
		}
	}
	return false
}
func ReloadConfig(keys []string) {
	start := time.Now()
	//now := start.Unix()
	_locker.Lock()
	ccs := make(map[string]ConfigInterface, len(_configInstances))

	logs.Trace("============================Begin ReloadConfigCache for keys:%+v============================", keys)
	logs.Console("Begin ReloadConfigCache for keys", keys)
	logs.WriteBill("root_config", "=================Begin ReloadConfigCache for keys:%+v=================", keys)
	for k, c := range _configInstances {
		for _, kk := range keys {
			if checkIsMyConfig(c, kk) {
				// 过滤间隔时间内的表
				/*	rtime, ok := ReloadTime.Load(k + kk)
					if ok {
						lastTime, ok := rtime.(int64)
						if ok {
							if now-lastTime < c.Intervals() {
								WaitReloadKey.Store(kk, now)
								logs.Trace("============================Stop ReloadConfigCache for k:%+v, table:%+v,intervals:%+d,need:%d============================", k, kk, now-lastTime, c.Intervals())
								continue
							}
						}
					}
					ccs[k] = c
					ReloadTime.Store(k+kk, now)*/

				ccs[k] = c
				break
			}

		}

	}
	_locker.Unlock()
	if len(ccs) == 0 {
		return
	}
	//ks := []string{"AppsConfig"}
	//for _, k := range ks {
	//	if c, ok := ccs[k]; ok {
	//		cc, _ := c.ReloadConfig()
	//		delete(ccs, k)
	//		logs.Trace("ReloadConfigCache 配置:%s 更换内存对象从%p->%p", k, c, cc)
	//	}
	//}
	clist := make(ConfigInterfaceList, 0, len(ccs))
	for _, c := range ccs {
		clist = append(clist, c)
	}
	sort.Sort(clist)
	for _, c := range clist {
		k := c.GetConfigKey()
		logs.Trace("==============ReloadConfigCache [%s:%d]  listen:%+v==============", k, c.GetPower(), c.ListenTables())
		logs.Console(fmt.Sprintf("==============ReloadConfigCache [%s:%d]  listen:%+v==============", k, c.GetPower(), c.ListenTables()))
		cc, err := c.ReloadConfig()
		if err != nil {
			//log.Panicf("config instance: %s LoadConfig:%+v", key, err)
			logs.PrintErr("ReloadConfigCache 严重错误,进程可能要挂掉", "加载配置失败", k, err.Error(), "继续保留原配置........")
			continue
		}
		logs.Trace("ReloadConfigCache 配置:%s 更换内存对象从%p->%p", k, c, cc)
		if c != cc {
			_locker.Lock()
			cc.SetConfigKey(k)
			//cc.SetPower(c.GetPower())
			_configInstances[k] = cc
			_locker.Unlock()
		}
		logs.Trace("==============end ReloadConfigCache [%s]  listen:%+v==============", k, c.ListenTables())
		logs.Console(fmt.Sprintf("==============end ReloadConfigCache [%s]  listen:%+v==============", k, c.ListenTables()))
	}
	/*

		for k, c := range ccs {
			logs.Trace("==============ReloadConfigCache [%s]  listen:%+v==============", k, c.ListenTables())
			cc, err := c.ReloadConfig()
			if err != nil {
				//log.Panicf("config instance: %s LoadConfig:%+v", key, err)
				logs.PrintErr("ReloadConfigCache 严重错误,进程可能要挂掉", "加载配置失败", k, err.Error(), "继续保留原配置........")
				continue
			}
			logs.Trace("ReloadConfigCache 配置:%s 更换内存对象从%p->%p", k, c, cc)
			if c != cc {
				_locker.Lock()
				_configInstances[k] = cc
				_locker.Unlock()
			}

			logs.Trace("==============end ReloadConfigCache [%s]  listen:%+v==============", k, c.ListenTables())
		}
	*/
	logs.Trace("============================End ReloadConfigCache for keys:%+v============================", keys)
	logs.Console(fmt.Sprintf("============================End ReloadConfigCache for keys:%+v============================", keys))
	logs.WriteBill("root_config", "=================End ReloadConfigCache for keys:%+v, cost:%d=================", keys, time.Now().Sub(start)/time.Millisecond)
}

func GetListenKeys() []string {
	ss := []string{}
	for _, c := range _configInstances {
		ss = append(ss, c.ListenTables()...)
	}
	return ss
}

//
//func tickerInit() {
//	go func() {
//		ticker := time.NewTicker(1 * time.Minute)
//		defer ticker.Stop()
//		for {
//			select {
//			case <-ticker.C:
//				go WaitReload()
//			}
//		}
//	}()
//}
//func WaitReload() {
//	wrMap := WaitReloadKey
//	WaitReloadKey = sync.Map{}
//	var keys []string
//	wrMap.Range(func(key, value any) bool {
//		keys = append(keys, fmt.Sprintf("%v", key))
//		return true
//	})
//	if len(keys) > 0 {
//		ReloadConfig(keys)
//	}
//}
