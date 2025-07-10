package frame

import (
	"fmt"
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"
	"github.com/aiden2048/pkg/utils/maddr"

	"github.com/BurntSushi/toml"
)

type RedisSectionConfig struct {
	LocalIndex int
	IdleCount  int //空闲连接数上线
	Servers    []string
	Password   string
}
type RedisCfg struct {
	SelectLocal bool
	Redis       map[string]*RedisSectionConfig
}

var redis_config = &RedisCfg{}

func GetRedisConfig() *RedisCfg {
	return redis_config
}
func (cfg *RedisCfg) GetRedisNode(sec string) *RedisSectionConfig {
	if cfg == nil || len(cfg.Redis) == 0 {
		return nil
	}

	redisSec, ok := cfg.Redis[sec]
	if !ok {
		redisSec, ok = cfg.Redis["default"]
	}

	if !ok || len(redisSec.Servers) == 0 {
		return nil
	}
	return redisSec
}

func LoadRedisConfig() (error, map[string]bool) {
	bret := make(map[string]bool)

	newConf := &RedisCfg{}
	fkey := "RedisConfig.toml"
	filename := "../GlobalConfig/" + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.LogError("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		if err := LoadConfigFromMongo(fkey, newConf); err != nil {
			log.Printf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			logs.LogError("LoadConfigFromMongo[%s]: %+v", fkey, err)
			return err, bret
		}
	}
	if len(newConf.Redis) == 0 {
		err := fmt.Errorf("加载redis配置失败, 没有任何redis配置%s", newConf)
		log.Printf(err.Error())
		logs.LogError(err.Error())
		return err, bret
	}
	if v, ok := newConf.Redis["default"]; !ok || len(v.Servers) == 0 {
		err := fmt.Errorf("加载redis配置失败, 没有默认的redis配置%s", newConf)
		log.Printf(err.Error())
		logs.LogError(err.Error())
		return err, bret
	}
	for k, _ := range newConf.Redis {
		if _, ok := redis_config.Redis[k]; !ok {
			bret[k] = true
		}
	}
	for k, olds := range redis_config.Redis {
		news, ok := newConf.Redis[k]
		if ok {
			bret[k] = true
		} else {
			for _, s := range olds.Servers {
				if !utils.InArray(news.Servers, s) {
					bret[k] = true
					break
				}
			}
		}
	}

	if newConf.SelectLocal {
		localIps := maddr.GetLocalIps(nil)
		h, e := os.Hostname()
		if e == nil {
			localIps = append(localIps, h)
		}

		logs.Trace("LoadRedisConfig GetLocalIps:%+v", localIps)

		for _, redisSec := range newConf.Redis {
			redisSec.LocalIndex = -1
			for i := 0; i < len(redisSec.Servers); i++ {
				if utils.StringHasPrefix(redisSec.Servers[i], localIps) {
					redisSec.LocalIndex = i
					break
				}
			}
		}
	}

	logs.Trace("LoadRedisConfig:%+v", newConf)
	redis_config = newConf
	return nil, bret
}
