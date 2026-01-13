package frame

import (
	"fmt"
	"os"
	"sync"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type MgoSvrCfg struct {
	Scheme       string
	Url          []string //localhost:27017
	SrvUrl       string
	Query        string
	User         string
	Password     string
	PoolNum      int
	MinPoolNum   int
	ConnIdleTime int
	Ssl          bool
	SyncTask     bool
}

func (mcfg *MgoSvrCfg) GetUrl() string {
	add := ""
	for id, url := range mcfg.Url {
		if id == 0 {
			add = url
		} else {
			add += fmt.Sprintf(",%s", url)
		}
	}
	url := fmt.Sprintf("%s:%s@%s", mcfg.User, mcfg.Password, add)
	if mcfg.Query != "" {
		url = url + "?" + mcfg.Query
	}
	return url
}

// "user:pass@localhost:27017"
type MgoCfg struct {
	IsEncry         bool      // 是否加密数据库信息 0不加密 1加密
	Real            MgoSvrCfg // 运行库
	Conf            MgoSvrCfg // 配置库
	ImageRepository MgoSvrCfg // 镜像库
	Top             MgoSvrCfg // top库
	Log             MgoSvrCfg // log库
}

var mgo_configs = &sync.Map{}

func GetMgoCoinfig(plat_ids ...int32) *MgoCfg {
	plat_id := _Plat_id
	if len(plat_ids) > 0 {
		plat_id = plat_ids[0]
	}
	if val, ok := mgo_configs.Load(plat_id); ok {
		if cfg, ok := val.(*MgoCfg); ok {
			return cfg
		}
	}
	LoadMgoConfig(plat_id)
	if val, ok := mgo_configs.Load(plat_id); ok {
		if cfg, ok := val.(*MgoCfg); ok {
			return cfg
		}
	}
	return nil
}
func GetMongoRealConnUrl(plat_id int32) string {
	add := ""
	mgo_config := GetMgoCoinfig(plat_id)
	if mgo_config == nil {
		return ""
	}
	for id, url := range mgo_config.Real.Url {
		if id == 0 {
			add = url
		} else {
			add += fmt.Sprintf(",%s", url)
		}
	}
	url := fmt.Sprintf("%s:%s@%s", mgo_config.Real.User, mgo_config.Real.Password, add)
	if mgo_config.Real.Query != "" {
		url = url + "?" + mgo_config.Real.Query
	}
	return url
}

func LoadMgoConfig(plat_id int32) error {
	newConf := &MgoCfg{}

	fkey := "MgoConfig.toml"
	filename := GetGlobalConfigDir(plat_id) + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Errorf("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		return err
	}
	if newConf.Conf.User == "" {
		newConf.Conf = newConf.Real
	}
	if newConf.ImageRepository.User == "" {
		newConf.ImageRepository = newConf.Real
	}
	if newConf.Top.User == "" {
		newConf.Top = newConf.Real
	}
	if newConf.Log.User == "" {
		newConf.Log = newConf.Real
	}
	if newConf.IsEncry {
		DecodeConf(&newConf.Real)
		DecodeConf(&newConf.Conf)
		DecodeConf(&newConf.ImageRepository)
		DecodeConf(&newConf.Top)
		DecodeConf(&newConf.Log)
	}
	mgo_configs.Store(plat_id, newConf)
	logs.Print("Load MongoConfig", plat_id, filename, "成功:", newConf)
	return nil
}

func DecodeConf(conf *MgoSvrCfg) {
	if conf.Password != "" {
		conf.Password = Decrypt(conf.Password)
	}
	if conf.User != "" {
		conf.User = Decrypt(conf.User)
	}
}
