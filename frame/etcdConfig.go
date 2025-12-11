package frame

import (
	"log"
	"os"
	"sort"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type etcdAddr struct {
	IsMix    bool
	PlatId   int32
	EtcdAddr []string
}

type EtcdConfig struct {
	IpBlocks []string // ip前缀

	UpdateTime      int64 //服务上报更新时间间隔
	XClientPoolSize int

	//中心节点配置
	CenterEtcdAddrs    []etcdAddr
	MapCenterEtcdAddrs map[int32]etcdAddr
}

func (c *EtcdConfig) GetCenterEtcdAddr(platId int32) []string {
	if c == nil {
		return nil
	}
	if a, ok := c.MapCenterEtcdAddrs[platId]; ok {
		return a.EtcdAddr
	}
	return nil
}

func (c *EtcdConfig) IsRpcxOnly() bool {
	if c == nil {
		return false
	}
	if GetGlobalConfig().IsTestServer {
		return false
	}
	return true
}

var etcdConfig = &EtcdConfig{}

func GetEtcdConfig() *EtcdConfig {
	return etcdConfig
}

func LoadEtcdConfig() error {
	newConf := &EtcdConfig{}

	fkey := "EtcdConfig.toml"
	filename := GetGlobalConfigDir() + fkey
	_, err := toml.DecodeFile(filename, newConf)
	if err != nil {
		if !os.IsNotExist(err) {
			logs.Errorf("DecodeFile:%s failed:%s", filename, err.Error())
		}
		if err := LoadConfigFromMongo(fkey, newConf); err != nil {
			log.Printf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			logs.Errorf("LoadConfigFromMongo[%s]: %+v", fkey, err)
			return err
		}
	}

	if newConf.UpdateTime <= 0 {
		newConf.UpdateTime = 30
	}
	if newConf.XClientPoolSize <= 0 {
		newConf.XClientPoolSize = 32
	}

	newConf.MapCenterEtcdAddrs = make(map[int32]etcdAddr, len(newConf.CenterEtcdAddrs))
	for _, k := range newConf.CenterEtcdAddrs {
		if k.PlatId <= 0 || len(k.EtcdAddr) == 0 {
			logs.Print("MapCenterEtcdAddrs Etcd", k, "配置错误，请检查")
			continue
		}
		if GetPlatformId() >= 1000 && k.IsMix && k.PlatId != GetPlatformId() {
			logs.Print("MapCenterEtcdAddrs Etcd", k, "这是其他通服组etcd, 不需要连")
			continue
		}
		if !GetFrameOption().EnableAllAreaMix && GetPlatformId() >= 1000 && GetPlatformId()/1000 != k.PlatId/1000 {
			logs.Print("MapCenterEtcdAddrs Etcd", k, "跟本组不通服, 不加入到通服配置")
			continue
		}
		sort.Strings(k.EtcdAddr)
		logs.Print("MapCenterEtcdAddrs 添加center Etcd配置", k)
		newConf.MapCenterEtcdAddrs[k.PlatId] = k
	}

	etcdConfig = newConf
	logs.Print("LoadEtcdConfig", newConf)

	OnEtcdReload()
	//OnEtcdV3Reload()
	return nil
}

func CheckPrintRpcxErr(sname string) bool {
	if sname == "monitor" {
		logs.Print("CheckPrintRpcxErr false", GetServerName(), sname)
		return false
	}
	if GetServerName() != "conn" {
		logs.Print("CheckPrintRpcxErr", GetServerName(), sname)
		return true
	}
	return false
}
