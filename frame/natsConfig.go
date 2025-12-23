package frame

import (
	"log"
	"os"
	"sort"

	"github.com/aiden2048/pkg/frame/logs"

	"github.com/BurntSushi/toml"
)

type NatsConfig struct {
	PlatId   int32
	IsMix    bool
	User     string
	Password string
	Servers  []string
}
type AllNatsConfig struct {
	AllNats []*NatsConfig
}

var natsConfig = &AllNatsConfig{}

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

	for _, k := range newConf.AllNats {

		needStart := false

		if IsMix() {
			if k.IsMix && k.PlatId/100 == GetPlatformId()/100 {
				needStart = true
			}
		} else {
			if (k.IsMix && k.PlatId/100 == GetPlatformId()/100) ||
				k.PlatId == GetPlatformId() {
				needStart = true
			}
		}

		if !needStart {
			continue
		}

		sort.Strings(k.Servers)

		if err := StartNatsService(k); err != nil {
			logs.Errorf("Start NatsService Failed:%s Config:%+v", err.Error(), k)
			return err
		}

		logs.Infof("Save Nats Config:%+v", k)
	}
	natsConfig = newConf
	return nil
}
