package frame

import (
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"

	"github.com/BurntSushi/toml"
)

type TGlobalConfig struct {
	PlatformID      int32  //平台ID
	EnableMixServer int32  //该组是否可以启用大通服
	MixSuffix       string //通服的分组前缀
	IsTestServer    bool   //是否测试服

	//HttpProxyAddr  []string // 内部http服务器地址
	ClientCheckKey string //客户端校验用的key
	WebCheckKey    string //WEB客户端校验用的key
	//
	//MgoGlobalName    string // 是否是 mg_global 名字
	//TopManagerDbName string // 是否是 topmanager_data 名字 支付配置

	CheckUriSum bool
	WhiteIps    []string // 客户端Ip白名单

	TestUids       []int64 //测试专用uid
	EcryptResponse bool    // 加密接口回包是否加密

	RiskAlarmToken string // 虚拟手机号告警token
	RiskAlarmCid   string // 虚拟手机号群

	/** 加解密key配置 **/
	AesClientKey     string // 客户端加密KEY
	AesClientIv      string // 客户端加密IV
	AesPostKey       string // 解密上传数据KEY
	AesPostIv        string //  解密上传数据IV
	UseSelfHttpProxy bool   // 是否使用本组的httpProxy
}

var _global_config = &TGlobalConfig{}

func GetGlobalConfig() *TGlobalConfig {
	return _global_config
}

// 客户端加密KEY
func GetAesClientKey() string {
	if GetGlobalConfig() == nil {
		return ""
	}
	return GetGlobalConfig().AesClientKey
}

// 客户端加密IV
func GetAesClientIv() string {
	if GetGlobalConfig() == nil {
		return ""
	}
	return GetGlobalConfig().AesClientIv
}

// 解密上传数据KEY
func GetAesPostKey(ikey int) string {
	if GetGlobalConfig() == nil {
		return ""
	}
	return GetGlobalConfig().AesPostKey
}

// 解密上传数据IV
func GetAesPostIv(ikey int) string {
	if GetGlobalConfig() == nil {
		return ""
	}
	return GetGlobalConfig().AesPostIv
}

// 使用httpProxy的地址
/*func GetHttpProxyAddr() []string {
	if GetGlobalConfig() == nil || len(GetGlobalConfig().HttpProxyAddr) <= 0 {
		return []string{}
	}
	return GetGlobalConfig().HttpProxyAddr
}
*/
// 判断是否使用httpProxy
func CheckUseHttpProxy() bool {
	return !GetFrameOption().DisableHttpProxy
}

func IsTestUid(uid uint64) bool {
	return utils.InArray(_global_config.TestUids, int64(uid))
}
func LoadGlobalConfig() error {
	newConf := &TGlobalConfig{}
	fkey := "GlobalConfig.toml"
	filename := "../GlobalConfig/" + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.LogError("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		if err := LoadConfigFromMongo(fkey, newConf); err != nil {
			log.Printf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			logs.LogError("LoadConfigFromMongo[%s]: %+v", fkey, err)
			return err
		}
	}
	if newConf.ClientCheckKey == "" {
		newConf.ClientCheckKey = "lV22El4V8sOOe/7cmjR1Qg=="
	}
	if newConf.WebCheckKey == "" {
		newConf.WebCheckKey = "LDlWBqBzC3sKfc1/YtisSw=="
	}

	if newConf.AesClientKey == "" { // 客户端加密KEY
		newConf.AesClientKey = "U41U5FeIj0uhOKnzgWZ1MBBC6iEUF2DZ"
	}
	if newConf.AesClientIv == "" { // 客户端加密IV
		newConf.AesClientIv = "4Y535fZnkQ57BVoq"
	}
	if newConf.AesPostKey == "" { // 解密上传数据KEY
		newConf.AesPostKey = "8dw/JfjjoMs0dzVGOX2ntb1iw2k9+JD4"
	}
	if newConf.AesPostIv == "" { //  解密上传数据IV
		newConf.AesPostIv = "ZGdIobme/Sb4Idwg"
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

// 客户端 s 加解密key, 区分 web和APP
// 用途: 1. uri里面加密字段 s 的生成, 2.登录时候sign的生成
func GetClientCheckKey(osType int32) string {
	if GetGlobalConfig() == nil {
		return "lV22El4V8sOOe/7cmjR1Qg=="
	}
	if osType == OsTypeWeb {
		return GetGlobalConfig().WebCheckKey
	}
	return GetGlobalConfig().ClientCheckKey
}

const (
	OsTypeUnKnow = 0 //未知
	OsTypeWeb    = 1 //web
	OsTypeAos    = 2 //aos
	OsTypeIos    = 3 //ios
	OsTypeAosWeb = 4 //安卓web
	OsTypeIosWeb = 5 //ios Web
)
