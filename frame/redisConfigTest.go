package frame

import (
	"fmt"
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"

	"github.com/aiden2048/pkg/utils"
	"github.com/aiden2048/pkg/utils/maddr"
)

type RedisSectionConfig2 struct {
	LocalIndex int
	Servers    []string
	Password   string
}
type RedisCfg2 struct {
	SelectLocal bool
	Redis       map[string]*RedisSectionConfig2
}

var redis_config2 = &RedisCfg2{}

func GetRedisConfig2() *RedisCfg2 {
	return redis_config2
}
func (cfg *RedisCfg2) GetRedisNode(sec string) *RedisSectionConfig2 {
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

func LoadRedisConfig2() (error, map[string]bool) {
	bret := make(map[string]bool)

	newConf := &RedisCfg2{}
	fkey := "RedisConfig.toml"
	filename := "../GlobalConfig/" + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Errorf("DecodeFile:%s failed:%s", fkey, err.Error())
		}
		if err := LoadConfigFromMongo(fkey, newConf); err != nil {
			log.Printf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			logs.Errorf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			return err, bret
		}
	}
	if len(newConf.Redis) == 0 {
		err := fmt.Errorf("加载redis配置失败, 没有任何redis配置")
		log.Printf(err.Error())
		logs.Errorf(err.Error())
		return err, bret
	}
	if v, ok := newConf.Redis["default"]; !ok || len(v.Servers) == 0 {
		err := fmt.Errorf("加载redis配置失败, 没有默认的redis配置")
		log.Printf(err.Error())
		logs.Errorf(err.Error())
		return err, bret
	}
	for k, _ := range newConf.Redis {
		if _, ok := redis_config2.Redis[k]; !ok {
			bret[k] = true
		}
	}
	for k, olds := range redis_config2.Redis {
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

		logs.Infof("LoadRedisConfig GetLocalIps:%+v", localIps)

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

	logs.Infof("LoadRedisConfig:%+v", newConf)
	redis_config2 = newConf
	return nil, bret
}
