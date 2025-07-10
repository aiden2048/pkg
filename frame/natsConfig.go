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
		if k.User == "" {
			k.User = "goapp"
		}
		if k.Password == "" {
			k.Password = newConf.Password
		}
		if k.Password == "" {
			k.Password = "Rx7mTNwgBW"
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

	//if defFrameOption.EnableMixServer {
	//	logs.Print("本进程启用互通模式,需要链接其他的 mapNats:", newConf.mapNats)
	//	// 先清理配置不存在的连接
	//	clearnMixNatsConn(newConf)
	//	//newConf.mapNats = make(map[int32]*NatsConfig, len(newConf.mapNats))
	//	for _, addr := range newConf.mapNats {
	//		if addr.PlatId == GetPlatformId() {
	//			continue
	//		}
	//		if GetPlatformId() > 100 && addr.IsMix {
	//			// 不是top的通服组 不连接通服nats
	//			continue
	//		}
	//		if len(addr.Servers) == 0 {
	//			logs.PrintError("mapNats 互通的 Nats 配置错误, plat", addr)
	//			continue
	//		}
	//		logs.Print("mapNats 开始链接 其他组nats", addr)
	//		if getMixNatsConn(addr.PlatId) == nil || !checkNatsConfigEqual(natsConfig.GetNatsConfig(addr.PlatId), addr) {
	//			if err := StartMixNatsService(addr); err != nil {
	//				logs.PrintError("mapNats Start NatsService", addr, " Failed", err)
	//			}
	//		}
	//	}
	//}

	// 连接中心节点
	//centerNatsConfig := getCenterNatsConfig(GetPlatformId())
	//if centerNatsConfig != nil && len(centerNatsConfig.Servers) > 0 {
	//	logs.Print("centerNats 开始链接 centerNats", centerNatsConfig)
	//	if getCenterNatsConn(GetPlatformId()) == nil {
	//		if err := StartCenterNatsService(centerNatsConfig); err != nil {
	//			logs.PrintError("centerNats Start NatsService", centerNatsConfig, " Failed", err)
	//		}
	//	}
	//}

	natsConfig = newConf
	return nil
}
