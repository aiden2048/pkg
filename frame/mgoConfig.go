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
	IsEncry bool      // 是否加密数据库信息 0不加密 1加密
	Real    MgoSvrCfg // 运行库
}

var mgo_config = &sync.Map{}

func GetMgoCoinfig(plat_id ...int32) *MgoCfg {
	pid := int32(_Plat_id)
	if len(plat_id) > 0 {
		pid = plat_id[0]
	}
	cfg, ok := mgo_config.Load(pid)
	if !ok {
		// 没加载，尝试加载
		LoadMgoConfig(plat_id...)
		cfg, ok := mgo_config.Load(pid)
		if !ok {
			return nil
		}
		return cfg.(*MgoCfg)
	}
	return cfg.(*MgoCfg)
}

func GetMongoRealConnUrl(plat_id ...int32) string {
	add := ""
	mcfg := GetMgoCoinfig(plat_id...)
	if mcfg == nil {
		return ""
	}
	for id, url := range mcfg.Real.Url {
		if id == 0 {
			add = url
		} else {
			add += fmt.Sprintf(",%s", url)
		}
	}
	url := fmt.Sprintf("%s:%s@%s", mcfg.Real.User, mcfg.Real.Password, add)
	if mcfg.Real.Query != "" {
		url = url + "?" + mcfg.Real.Query
	}
	return url
}

func LoadMgoConfig(plat_id ...int32) error {
	newConf := &MgoCfg{}

	fkey := "MgoConfig.toml"
	filename := GetGlobalConfigDir(plat_id...) + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Errorf("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		return err
	}

	if newConf.IsEncry {
		DecodeConf(&newConf.Real)
	}
	pid := int32(_Plat_id)
	if len(plat_id) > 0 {
		pid = plat_id[0]
	}
	mgo_config.Store(pid, newConf)
	logs.Print("Load MongoConfig", newConf)
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
