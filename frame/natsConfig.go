package frame

import (
	"log"
	"os"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"

	"github.com/BurntSushi/toml"
)

type NatsConfig struct {
	PlatId int32
	IsMix  bool
	//Url      string
	User     string
	Password string
	//Token    string
	Timeout int
	Secure  bool
	RootCa  string
	Servers []string
}
type AllNatsConfig struct {
	*NatsConfig
	AllNats []*NatsConfig
	//mapNats map[int32]*NatsConfig
	//CenterNats    []*NatsConfig // 每个区域的中心节点
	//centerMapNats map[int32]*NatsConfig
	OnlyNats bool
}

func (c *AllNatsConfig) GetNatsConfig(platId int32) *NatsConfig {
	if c == nil {
		return nil
	}
	//if !defFrameOption.EnableMixServer || len(c.AllNats) == 0 || platId == 0 || platId == GetPlatformId() {
	//	return c.NatsConfig
	//}
	//if a, ok := c.mapNats[platId]; ok {
	//	return a
	//}
	return c.NatsConfig
}

var natsConfig = &AllNatsConfig{}

func GetNatsConfig() *NatsConfig {
	return natsConfig.GetNatsConfig(0)
}

func checkNatsConfigEqual(nConfig *NatsConfig, mConfig *NatsConfig) bool {
	if nConfig == nil || mConfig == nil {
		return false
	}
	//if nats_config.Url != nConfig.Url {
	//	return false
	//}
	//for _, c := range nConfig.Servers {
	//	if !utils.InArrayString(nats_config.Servers, c) {
	//		return false
	//	}
	//}
	for _, c := range nConfig.Servers {
		if !utils.InArray(mConfig.Servers, c) {
			return false
		}
	}
	return true
}
func LoadNatsConfig() error {
	newConf := &AllNatsConfig{}

	fkey := "NatsConfig.toml"
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

	//newConf.mapNats = make(map[int32]*NatsConfig, len(newConf.AllNats))
	for _, k := range newConf.AllNats {
		//if GetPlatformId() >= 1000 && GetPlatformId()/1000 != k.PlatId/1000 {
		//	logs.Print("AllNats Nats", k, "跟本组不通服, 不加入到通服配置")
		//	continue
		//}
		if k.Timeout <= 0 {
			k.Timeout = 0
		}
		if k.User == "" {
			k.User = newConf.User
		}

		if k.Password == "" {
			k.Password = newConf.Password
		}

		//newConf.mapNats[k.PlatId] = k
		if k.PlatId == GetPlatformId() {
			newConf.NatsConfig = k
			logs.Print("读取本组 Nats 配置", k)
		}
	}

	if newConf.GetNatsConfig(0) != nil {
		if GetNatsConn() == nil || !checkNatsConfigEqual(natsConfig.GetNatsConfig(0), newConf.GetNatsConfig(0)) {
			if err := StartNatsService(newConf.GetNatsConfig(0)); err == nil {
				logs.Infof("Save Nats Config:%+v", newConf.GetNatsConfig(0))
			} else {
				logs.Errorf("Start NatsService Failed:%s", err.Error())
				logs.Infof("New Nats Config:%+v can not be connected, keep config:%+v", newConf.GetNatsConfig(0), natsConfig.GetNatsConfig(0))
				return err
			}
		}
	} else {
		logs.Print("plat", GetPlatformId(), "本组 没有配置Nats,++++++++++++++++++")
	}
	natsConfig = newConf
	return nil
}
