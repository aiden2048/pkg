package frame

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
)

// rpcx etcdV3注册
var registryPluginV3 []*serverplugin.EtcdV3RegisterPlugin

var rpcxV3Started bool
var listenerV3 net.Listener
var rpcxV3Server *server.Server
var addressV3 string

var allPlatsV3 = sync.Map{}

var regsV3 []*regFunc

const V3BasePath = "/V3Rpcx"

func GetRegistEtcdBasePathV3(s string) string {
	key := V3BasePath
	//如果进程是通服的, 不能加上platid
	if !GetFrameOption().EnableMixServer {
		key = fmt.Sprintf("%s/%d", V3BasePath, GetPlatformId())
	}
	log.Printf("GetRegistEtcdBasePathV3 %s isMix:%t, 注册到key:%s\n", GetServerName(), GetFrameOption().EnableMixServer, key)
	return key
}
func addEtcdV3RgistryPlugin(serv *server.Server) error {
	if serv == nil {
		logs.Importantf("进程还没初始化完成, 忽略重载.......")
		return errors.New("serv is nil")
	}
	{
		centerEtcdAddr := GetEtcdConfig().GetCenterEtcdAddr(GetPlatformId())
		if len(centerEtcdAddr) == 0 {
			logs.PrintImportant("there is no v3 etcd config")
			return nil
		}
		addrs := strings.Join(centerEtcdAddr, ",")
		_, ok := allPlatsV3.Load(addrs)
		if !ok && len(centerEtcdAddr) > 0 {
			plugin := &serverplugin.EtcdV3RegisterPlugin{
				ServiceAddress: "tcp@" + address,
				EtcdServers:    centerEtcdAddr,
				BasePath:       GetRegistEtcdBasePathV3(GetServerName()),
				Metrics:        metrics.NewRegistry(),
				Services:       nil,
				UpdateInterval: time.Second * time.Duration(GetEtcdConfig().UpdateTime),
				//	Options:        &store.Config{ConnectionTimeout: 3 * time.Second},
			}

			err := plugin.Start()
			if err != nil {
				logs.PrintError("addEtcdV3RgistryPlugin LocalPlat", centerEtcdAddr, plugin, "Failed", err)
				return err
			}
			logs.Print("addEtcdV3RgistryPlugin LocalPlat", centerEtcdAddr, plugin)
			serv.Plugins.Add(plugin)
			registryPluginV3 = append(registryPluginV3, plugin)
			allPlatsV3.Store(addrs, 1)
		} else if !ok {
			logs.PrintImportant("本地未配置Center ETCD地址, 未能启动ETCD注册服务", "platformId", GetPlatformId())
		}
		logs.Importantf("addEtcdV3RgistryPlugin, defFrameOption.EnableMixServer:%+v", defFrameOption.EnableMixServer)

	}

	//该进程支持大混服, 要链接其他的etcd
	if defFrameOption.EnableMixServer {

		logs.Importantf("PlatId:%d,  addEtcdV3RgistryPlugin start 本进程启用互通模式,需要链接其他的Etcd:%+v", GetPlatformId(), GetEtcdConfig().MapCenterEtcdAddrs)
		for _, addr := range GetEtcdConfig().MapCenterEtcdAddrs {
			if addr.PlatId == GetPlatformId() {
				continue
			}
			if len(addr.EtcdAddr) == 0 {
				logs.PrintError("互通的etcd配置错误, addr:%+v", addr)
				continue
			}
			// 相同的添加过就不再添加
			platAddrs := strings.Join(addr.EtcdAddr, ",")
			if _, ok := allPlatsV3.Load(platAddrs); ok {
				continue
			}
			allPlatsV3.Store(platAddrs, 1)

			plugin := &serverplugin.EtcdV3RegisterPlugin{
				ServiceAddress: "tcp@" + address,
				EtcdServers:    addr.EtcdAddr,
				BasePath:       GetRegistEtcdBasePathV3(GetServerName()),
				Metrics:        metrics.NewRegistry(),
				Services:       nil,
				UpdateInterval: time.Second * time.Duration(GetEtcdConfig().UpdateTime),
				//	Options:        &store.Config{ConnectionTimeout: 3 * time.Second},
			}

			err := plugin.Start()
			if err != nil {
				logs.PrintError("addEtcdV3RgistryPlugin Plat:", addr, plugin, "Failed", err)
				return err
			}
			serv.Plugins.Add(plugin)
			registryPluginV3 = append(registryPluginV3, plugin)
			logs.Print("addEtcdV3RgistryPlugin Plat:", addr, plugin)

			//这是重载配置的时候, 新加etcd 配置
			if rpcxStarted {
				OnAddOnePlugin(plugin)
			}
		}
		logs.Importantf("PlatId:%d, addEtcdV3RgistryPlugin end 本进程启用互通模式,需要链接其他的Etcd", GetPlatformId())
	}
	return nil
}

func OnAddOnePlugin(plugin server.RegisterFunctionPlugin) {
	logs.Print("添加新Plugin", plugin, "OnAddOnePlugin 重新注册:", regs)
	if plugin == nil {
		return
	}
	for _, reg := range regs {
		_ = plugin.RegisterFunction(reg.Sevice, reg.Name, reg.Fn, reg.Metadata)
	}
}
