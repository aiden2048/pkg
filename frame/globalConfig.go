package frame

import (
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type TGlobalConfig struct {
	PlatformID      int32  //平台ID
	EnableMixServer int32  //该组是否可以启用大通服
	MixSuffix       string //通服的分组前缀
	IsTestServer    bool   //是否测试服
}

var _global_config = &TGlobalConfig{}

func GetGlobalConfig() *TGlobalConfig {
	return _global_config
}

func LoadGlobalConfig() error {
	newConf := &TGlobalConfig{}
	fkey := "GlobalConfig.toml"
	filename := "../GlobalConfig/" + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Errorf("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		if err := LoadConfigFromMongo(fkey, newConf); err != nil {
			log.Printf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			logs.Errorf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			return err
		}
	}
	plat := _global_config.PlatformID
	mix := _global_config.EnableMixServer
	suffix := _global_config.MixSuffix

	isTest := _global_config.IsTestServer

	_global_config = newConf
	_global_config.PlatformID = plat
	_global_config.EnableMixServer = mix
	_global_config.MixSuffix = suffix
	if _global_config.MixSuffix == "" {
		_global_config.MixSuffix = "Mix"
	}
	_global_config.IsTestServer = isTest
	//if GetFrameOption().DisableHttpProxy == true {
	//	_global_config.HttpProxyAddr = nil
	//}
	logs.LogDebug("_global_config:%+v", _global_config)
	//if err := LoadProxyConfigs(); err != nil{
	//	return err
	//}
	return nil
}

const (
	OsTypeUnKnow = 0 //未知
	OsTypeWeb    = 1 //web
	OsTypeAos    = 2 //aos
	OsTypeIos    = 3 //ios
	OsTypeAosWeb = 4 //安卓web
	OsTypeIosWeb = 5 //ios Web
)
