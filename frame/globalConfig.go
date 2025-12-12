package frame

import (
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type TGlobalConfig struct {
	PlatformID      int32  //平台ID
	EnableMixServer bool   //该组是否可以启用大通服
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
	filename := GetGlobalConfigDir() + fkey
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

	_global_config = newConf
	_global_config.PlatformID = int32(_Plat_id)
	if _global_config.MixSuffix == "" {
		_global_config.MixSuffix = "Mix"
	}
	logs.Debugf("_global_config:%+v", _global_config)

	return nil
}

// 客户端类型 1 安卓h5 2 安卓pwa 3 安卓apk 4 ios h5 5 ios pwa 6 ios app 7 windows 8 macos
const (
	OsTypeUnKnow = 0 //未知
	OsTypeAnH5   = 1 //安卓h5
	OsTypeAnPWA  = 2 //安卓pwa
	OsTypeAnApk  = 3 //安卓apk
	OsTypeiOSH5  = 4 //ios h5
	OsTypeiOSPWA = 5 //ios pwa
	OsTypeiOSApp = 6 //ios app
	OsTypeWin    = 7 //windows
	OsTypeMacOS  = 8 //macos
)
