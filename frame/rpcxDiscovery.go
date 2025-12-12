package frame

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rpcxio/libkv/store"
	etcdclient "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils/baselib"
)

// ============call==================

// 缓存所有的Rpc客户端
var _allRpcxClients = struct {
	lock    sync.Mutex
	clients map[string]*XClientPool
}{}

// 缓存所有的Rpc客户端
var _allRpcxEtcdStore = struct {
	lock   sync.Mutex
	stores map[string]store.Store
}{}

func CheckRpcxService(platId int32, sname string, svrid int32, fname string) bool {
	if platId <= 0 || platId == GetPlatformId() {
		return IsUseRpcx()
	}
	if !defFrameOption.EnableMixServer && !defFrameOption.AllowCrossServer {
		logs.PrintError("CheckRpcxService Plat", platId, sname, svrid, fname, "本进程不同服, 不能跨服访问")
		return false
	}
	return true
}

var dis_cache = &baselib.Map{}

func createDiscovery(addr []string, basePath string, platId int32, sname string, ver string) (client.ServiceDiscovery, error) {
	var d client.ServiceDiscovery
	var err error
	cache_key := fmt.Sprintf("%s/%s%s/%s", strings.Join(addr, ","), ver, basePath, sname)
	v, ok := dis_cache.Load(cache_key)
	if ok {
		d, _ = v.(client.ServiceDiscovery)
		logs.PrintImportant("XClient createDiscovery", ver, " Get from Cache", cache_key, d)
		logs.Console("XClient createDiscovery", ver, " Get from Cache", cache_key, d)
	} else /* if sname == "/"*/ {

		if strings.ToUpper(ver) == "V3" {
			d, err = etcdclient.NewEtcdV3Discovery(basePath, sname, addr, true, nil)
		} else {
			d, err = etcdclient.NewEtcdDiscovery(basePath, sname, addr, true, nil)
		}
		if err != nil {
			err = errors.New(fmt.Sprintf("XClient Discovery %s:%s platId:%d %s", ver, cache_key, platId, err.Error()))
			logs.PrintImportant(" 创建发现服务失败 NewEtcdDiscovery:", "key", cache_key, "err", err, "centerAddr", addr)
			return nil, err
		}
		dis_cache.Store(cache_key, d)
		logs.PrintImportant("XClient Discovery", ver, " Store Create And Store", cache_key, d)
		logs.Console("XClient Discovery", ver, " Store Create  And Store", cache_key, d)
	}
	// v3版本
	//logs.Console("Start XClient Discovery", ver, addr, platId, sname, "key", cache_key, d.GetServices())
	if d != nil && len(d.GetServices()) > 0 {
		logs.Console("XClient createDiscovery", ver, " succ:", "key", cache_key, "centerAddr", addr, d.GetServices())
		return d, nil
	}

	if d != nil && len(d.GetServices()) <= 0 {
		err = errors.New(fmt.Sprintf("XClient Discovery %s:%s platId:%d No-Services", ver, cache_key, platId))
		logs.PrintImportant(" 创建发现服务失败 GetServices:", "key", cache_key, "err", err, "centerAddr", addr)
	}
	//d.Close()

	if err == nil {
		err = errors.New(fmt.Sprintf("XClient Discovery %s:%s platId:%d Unknown", ver, cache_key, platId))
		logs.PrintImportant(" 创建发现服务失败 GetServices:", "key", cache_key, "err", err, "centerAddr", addr)
	}
	logs.PrintImportant("XClient createDiscovery", ver, "key", cache_key, "err", err, "centerAddr", addr)
	logs.Console("XClient createDiscovery", ver, "key", cache_key, "err", err, "centerAddr", addr)
	return nil, err
}
func createDiscoveryV3(platId int32, sname string) (client.ServiceDiscovery, error) {

	centerAddr := GetEtcdConfig().GetCenterEtcdAddr(platId)
	if len(centerAddr) == 0 {
		return nil, errors.New("No-EtcdV3-Addr")
	}

	var d client.ServiceDiscovery
	var err error
	// etcd v3
	basePath := fmt.Sprintf("%s/%d", V3BasePath, platId) //先找下本组也没有//GetDiscoverBasePath(sname)
	if d, err = createDiscovery(centerAddr, basePath, platId, sname, "V3"); err == nil {
		logs.PrintImportant("XClient createDiscovery 创建发现服务V3:", "plat", platId, "service", sname, "etcd", centerAddr, basePath, d.GetServices())
		return d, err
	}
	basePath = V3BasePath
	if d, err = createDiscovery(centerAddr, basePath, platId, sname, "V3"); err == nil {
		logs.PrintImportant("XClient createDiscovery 创建发现服务V3:(通服进程)", "plat", platId, "service", sname, "etcd", centerAddr, basePath, d.GetServices())
		return d, err
	}

	return nil, err
}

func getDiscovery(platId int32, sname string) (client.ServiceDiscovery, error) {

	dis, err := createDiscoveryV3(platId, sname)
	if dis != nil {
		return dis, nil
	}

	return nil, err
}

// 选取rpcx client
func GetXClient(uid uint64, platId int32, sname string) client.XClient {
	return getXClient(uid, platId, sname)
}
func getXClient(uid uint64, platId int32, sname string) client.XClient {
	if platId <= 0 {
		platId = GetPlatformId()
	}
	_allRpcxClients.lock.Lock()
	defer _allRpcxClients.lock.Unlock()
	key := fmt.Sprintf("%d.%s", platId, sname)
	if v, ok := _allRpcxClients.clients[key]; ok {

		if v == nil || v.discovery == nil || len(v.discovery.GetServices()) == 0 {
			//删除缓存的pool, 重新创建
			delete(_allRpcxClients.clients, key)
			if v != nil {
				//不能停掉发现者
				//v.discovery.Close()
				v.Close()
				logs.Print("XClient Pool:", platId, sname, "发现者", v.discovery, "没有服务了, 重新选择服务")
				logs.Console("XClient Pool:", platId, sname, "发现者", v.discovery, "没有服务了, 重新选择服务")
			} else {
				logs.Print("XClient Pool:", platId, sname, "发现者", v, "没有服务了, 重新选择服务")
				logs.Console("XClient Pool:", platId, sname, "发现者", v, "没有服务了, 重新选择服务")
			}
		} else {
			c := v.Get(uid)
			if IsDebug() {
				logs.Print("Get XClient, Uid", uid, "plat", platId, "Sname", sname,
					"pool", key, fmt.Sprintf("%p", v), "Client", fmt.Sprintf("%p=%+v", c, c), "discovery", v.discovery)
				//logs.Importantf("Client:%+v", c)
			} else {
				logs.PrintDebug("Get XClient, Uid", uid, "plat", platId, "Sname", sname,
					"pool", key, fmt.Sprintf("%p", v), "Client", "discovery", fmt.Sprintf("%p", v.discovery))
			}

			return c
		}
	}

	var d client.ServiceDiscovery
	d, err := getDiscovery(platId, sname)
	if d == nil {
		b := make([]byte, 64<<10)
		b = b[:runtime.Stack(b, false)]
		if CheckPrintRpcxErr(sname) {
			logs.Errorf("XClient getDiscovery for Plat:%d, uid:%d, Key:%s, err:%+v, stack:%s", platId, uid, key, err, b)
		} else {
			logs.Importantf("XClient getDiscovery for Plat:%d, uid:%d, Key:%s, err:%+v, stack:%s", platId, uid, key, err, b)
		}
		return nil
	}
	option := client.DefaultOption
	option.Heartbeat = true
	option.HeartbeatInterval = time.Second
	psize := GetEtcdConfig().XClientPoolSize
	if defFrameOption.XClientPoolSize > 0 {
		psize = defFrameOption.XClientPoolSize
	}
	failMode := client.Failfast
	//selectMode := client.WeightedICMP
	selectMode := client.RandomSelect
	if strings.HasPrefix(sname, "conn.") {
		failMode = client.Failtry
		selectMode = client.WeightedICMP
	}

	pool := NewXClientPool(psize, sname, failMode /*client.Failfast*/, selectMode /*client.RandomSelect*/, d, option)
	_allRpcxClients.clients[key] = pool
	logs.Importantf("Start XClient Pool for Uid:%d Plat:%d, Sname:%s, Key:%s, Pool:%p=%+v",
		uid, platId, sname, key, pool, pool)
	c := pool.Get(uid)
	logs.Importantf("Get XClient for Uid:%d Plat:%d, Sname:%s, Client:%p=%+v, discovery:%p=%+v",
		uid, platId, sname, c, c, d, d.GetServices())

	//logs.LogDebug("getXClient, uid:%d, sname:%s, client:%p=%+v", uid, sname, c, c)
	return c
}
