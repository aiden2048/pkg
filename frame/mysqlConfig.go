package frame

import (
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type MysqlSvrCfg struct {
	Host     string
	Port     int32
	User     string
	Password string
	Param    map[string]string
}
type MysqlCfg struct {
	IsEncry  bool // 是否加密数据库信息 0不加密 1加密
	Param    map[string]string
	Real     MysqlSvrCfg
	ReadOnly []*MysqlSvrCfg
}

var mysql_config = &MysqlCfg{}

func SetMysqlConfig(cfg *MysqlCfg) {
	mysql_config = cfg
}

func GetMysqlConfig() *MysqlCfg {
	return mysql_config
}

func LoadMysqlConfig() error {
	new_cfg := &MysqlCfg{}
	fkey := "MysqlConfig.toml"
	configPath := GetGlobalConfigDir() + fkey
	if _, err := os.Stat(configPath); err != nil {
		logs.Errorf("Load configPath: %+v", err)
		return err
	}
	if _, err := toml.DecodeFile(configPath, new_cfg); err != nil {
		logs.Errorf("Load configPath: %+v", err)
		if err := LoadConfigFromMongo(fkey, new_cfg); err != nil {
			log.Printf("LoadConfig From Mongo[%s]: %+v", fkey, err)
			logs.Errorf("LoadConfig From Mongo[%s]: %+v", fkey, err)
			return err
		}
	}
	logs.Infof("LoadMysqlConfig:%+v", new_cfg)
	if new_cfg.IsEncry {
		new_cfg.Real.Host = Decrypt(new_cfg.Real.Host)
		new_cfg.Real.User = Decrypt(new_cfg.Real.User)
		new_cfg.Real.Password = Decrypt(new_cfg.Real.Password)
		//new_cfg.Real.Port = encrypt.Decrypt(new_cfg.Real.Port )
		for _, r := range new_cfg.ReadOnly {
			r.Host = Decrypt(r.Host)
			r.User = Decrypt(r.User)
			r.Password = Decrypt(r.Password)
		}
	}
	if len(new_cfg.Param) > 0 {
		new_cfg.Real.Param = new_cfg.Param
		for _, r := range new_cfg.ReadOnly {
			r.Param = new_cfg.Param
		}
	}
	mysql_config = new_cfg

	return nil
}
