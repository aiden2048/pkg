package configCache

import (
	"log"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/aiden2048/pkg/frame"
)

const (
	ConfigCacheKey  = "Reload.ConfigCache"
	ConfigCacheKeys = "Reload.ConfigCaches"
)

type ConfigCache struct {
	Key string `json:"key"`
}
type ConfigInterface interface {
	LoadConfig() error
	ReloadConfig() (ConfigInterface, error)
	ListenTables() []string
	Intervals() int64 // 重载间隔
	SetConfigKey(string)
	GetConfigKey() string
	//SetPower(int)
	GetPower() int
}

var _configInstances map[string]ConfigInterface
var _locker sync.Mutex
var isListen bool

func init() {
	_configInstances = make(map[string]ConfigInterface)
	_locker = sync.Mutex{}
	//tickerInit()
}

func SetInstance(key string, val ConfigInterface) ConfigInterface {
	if !isListen {
		isListen = true

		frame.ListenConfig(ConfigCacheKeys, reloadConfigHandler)
		//frame.ListenConfig(ConfigCacheKey, reloadConfigHandler)
	}

	_locker.Lock()
	if inst, ok := _configInstances[key]; ok {
		_locker.Unlock()
		logs.Infof("GetInstance %s, Listen:%+v On Reload", key, val.ListenTables())
		start := time.Now()
		_ = inst.LoadConfig()
		tt := time.Since(start)
		//fmt.Printf("\nLoadConfig %s, Listen:%+v, cost:%d\n", key, val.ListenTables(), tt/time.Millisecond)
		logs.Infof("LoadConfig %s, Listen:%+v, cost:%d", key, inst.ListenTables(), tt/time.Millisecond)

		return inst
	}
	_locker.Unlock()

	val.SetConfigKey(key)
	logs.Infof("SetInstance %s, Listen:%+v On Load", key, val.ListenTables())
	start := time.Now()
	err := val.LoadConfig()
	if err != nil {
		//log.Panicf("config instance: %s LoadConfig:%+v", key, err)
		logs.PrintErr("严重错误,进程可能要挂掉", "加载配置失败", key, err.Error(), "继续保留原配置........")
		return val
	}
	tt := time.Since(start)
	//fmt.Printf("\nLoadConfig %s, Listen:%+v, cost:%d\n", key, val.ListenTables(), tt/time.Millisecond)
	logs.Infof("LoadConfig %s, Listen:%+v, cost:%d", key, val.ListenTables(), tt/time.Millisecond)

	_locker.Lock()
	_configInstances[key] = val
	_locker.Unlock()
	//oldVal, ok := _configInstances[key]
	//if ok {
	//
	//	start := time.Now()
	//	err := oldVal.LoadConfig()
	//	if err != nil {
	//		//log.Panicf("config instance: %s LoadConfig:%+v", key, err)
	//		logs.PrintErr("SetInstance", key, "加载配置失败........", err.Error())
	//	}
	//	tt := time.Since(start)
	//	fmt.Printf("\nLoadConfig %s, Listen:%+v, cost:%d\n", key, oldVal.ListenTables(), tt/time.Millisecond)
	//	logs.Infof("LoadConfig %s, Listen:%+v, cost:%d", key, oldVal.ListenTables(), tt/time.Millisecond)
	//
	//	_locker.Unlock()
	//	return oldVal
	//}

	return val
}

func GetInstance(key string) ConfigInterface {
	_locker.Lock()
	v, ok := _configInstances[key]
	_locker.Unlock()
	if !ok {
		log.Panicf("not find config instance: %s", key)
	}

	return v
}
