// goServer.go 框架主逻辑，包括注册Rpc/Seed、Run等接口
package frame

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"
	"github.com/aiden2048/pkg/utils/baselib"
)

const DEFAULT_RPC_REQUEST_SECONDS = 5
const DEFAULT_HTTP_REQUEST_SECONDS = DEFAULT_RPC_REQUEST_SECONDS * 3 * time.Second

var G_CloseChan chan int
var g_deferFuncs []func()

var _debug_mode = false
var _dev_mode = false
var _mix_mode = false
var once = &sync.Once{}

type FrameOption struct {
	Port             int
	Svrid            int  //-1000的时候,有系统根据服务器机器名编号
	DisableNats      bool // 禁用 nats
	DisableRpcx      bool //禁用rpcx
	DisableMgo       bool //不启用Mgo
	EnableMysql      bool //不启用mysql
	EnableMixServer  bool //是否开通大混服
	EnableAllAreaMix bool // 是否允许在全区域通服(plat_id<1000)启动
	EnableAppUid     bool //是否启用假ID

	XClientPoolSize  int
	EventHandlerNum  int //监听事件通知起的线程数, 默认是0, 不限制
	DisableHttpProxy bool
	AllowCrossServer bool // 是否允许进程跨服访问
}

var defFrameOption = &FrameOption{}

func init() {

	fmt.Printf("\n+++++===============+++++\n")

	for i := 0; i < len(os.Args); i++ {
		if strings.ToLower(os.Args[i]) == "-d" || strings.ToLower(os.Args[i]) == "dbg" {
			_debug_mode = true
			break
		}
	}

	for i := 0; i < len(os.Args); i++ {
		if strings.ToLower(os.Args[i]) == "dev" || strings.ToLower(os.Args[i]) == "dev" {
			_dev_mode = true
			break
		}
	}

	for i := 0; i < len(os.Args); i++ {
		if strings.ToLower(os.Args[i]) == "mix" {
			_mix_mode = true
			break
		}
	}

	fmt.Printf("_debug_mode = %t, _dev_mode:%t, _mix_mode = %t\n", _debug_mode, _dev_mode, _mix_mode)
}

func IsTestServer() bool {
	return _global_config.IsTestServer
}

func IsDebug() bool {
	return _debug_mode || _dev_mode /*|| _global_config.IsTestServer*/
}

func IsDev() bool {
	return _dev_mode
}
func IsMix() bool {
	return _mix_mode
}
func IsTool(t string) bool {
	for i := 0; i < len(os.Args); i++ {
		if strings.ToLower(os.Args[i]) == t {
			return true
		}
	}
	return false
}

var closeLock = sync.Mutex{}
var flagClose = false

func CloseServer() {
	closeLock.Lock()
	if !flagClose {
		logs.Infof("CloseServer")
		log.Println("CloseServer")
		close(G_CloseChan)
		flagClose = true
	}
	closeLock.Unlock()
}

func IsStop() <-chan int {
	return G_CloseChan
}

func Init() error {
	G_CloseChan = make(chan int, 0)
	if !defFrameOption.DisableRpcx {
		if err := initRpcxServer(); err != nil {
			logs.Errorf("initRpcxServer error:%s", err)
			logs.Error("Run exit")
			return err
		}
	}

	return nil
}
func printMemStats() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	logs.Infof("PrintMemStats ===================Begin================")
	logs.Infof("mem.Alloc: %d, %d Mb", mem.Alloc, mem.Alloc/1024/1024)
	logs.Infof("mem.TotalAlloc: %d, %d Mb", mem.TotalAlloc, mem.TotalAlloc/1024/1024)
	logs.Infof("mem.HeapAlloc: %d, %d Mb", mem.HeapAlloc, mem.HeapAlloc/1024/1024)
	logs.Infof("mem.HeapSys: %d, %d Mb", mem.HeapSys, mem.HeapSys/1024/1024)

	logs.Infof("MemStats: %+v", &mem)
	logs.Infof("PrintMemStats ===================End================")
}

func ProfileToFile(file string, profileTime int) {
	logs.Infof("Start CPUProfile")
	if err := baselib.ProfileToFile(file+"_cpu", profileTime); err != nil {
		logs.Infof("Create file %s failed, error: %+v", file+"_cpu", err)
	}
	logs.Infof("Finish CPUProfile")
	logs.Infof("Start Memory Profile")
	if err := baselib.ProfileHeapMemoryToFile(file + "_memory"); err != nil {
		logs.Infof("Create file %s failed, error: %+v", file+"_memory", err)
	}
	logs.Infof("Finish Memory Profile")
}

func Defer(f func()) {

	g_deferFuncs = append(g_deferFuncs, f)
	logs.Infof("append Defer %+v", g_deferFuncs)
}

func ForceStartNatsService() {
	//reportAllServices("start")
	time.Sleep(2 * time.Second)
}

func onStopServer([]byte) {
	baselib.SendStopSignal()
}

func GetFrameOption() *FrameOption {
	return defFrameOption
}

func Run(appInit func() error) {
	defer Stop() // 清理退出操作

	if err := Init(); err != nil {
		fmt.Println("====================================================================")
		fmt.Println("start server failed: ", err.Error())
		fmt.Println("====================================================================")
		return
	}
	//注册监听配置通知
	//去掉config.servername方式, 改成config.allgoserver.servername
	//ListenConfig(GetServerName(), baselib.ReloadServerConfig)
	ListenConfig("allgoserver", baselib.ReloadServerConfig)
	//stopKey := fmt.Sprintf("stop_server_%s_%d", GetServerName(), GetServerID())
	ListenConfig("stop_server", onStopServer)
	logs.Infof(">>>>>>>>>>>>>>> start  ================")
	runinfo := fmt.Sprintf("%d %s %d %s ... start\n", os.Getpid(),
		GetServerName(), GetServerID(), time.Now().Format(utils.TimeFormat_Full))
	//fmt.Printf("================Start Server %s as id %d================\n", GetServerName(), GetServerID())
	err := ioutil.WriteFile("run.id", []byte(runinfo), 0666)
	logs.Infof(">>>>>>>>>>>>>>> Start Server %s as id %d  ,err:%+v================\n", GetServerName(), GetServerID(), err)
	go func() {
		var startTime time.Time
		startTime = time.Now()
		logs.Infof(">>>>>>>>>>>>>>> Start Run %s-%d at %s ================", GetServerName(), GetServerID(), startTime.Format(baselib.TimeFormatMilli))

		//应用初始化接口
		if appInit != nil {
			err := appInit()
			if err != nil {
				fmt.Println("start appInit failed: ", err.Error())
				logs.Errorf("start appInit failed: %+v", err.Error())
				log.Fatalf("start appInit failed: %+v", err.Error())
				return
			}
		}
		//正式启动rpc_nats服务
		//reportAllServices("start")

		//启动rpcx服务
		if !defFrameOption.DisableRpcx {
			e := startRpcServer()
			if e != nil {
				fmt.Println("startRpcServer rpc server failed: ", e.Error())
			}

		}
		now := time.Now()
		fmt.Printf("\n>>>>>>>>>>>>>>> Succ Run %s-%d at %s,cost: %d ms <<<<<<<<<<<<<<<\n",
			GetServerName(), GetServerID(), now.Format(baselib.TimeFormatMilli), now.Sub(startTime)/time.Millisecond)
		fmt.Printf("FrameOption: %+v\n", defFrameOption)
		fmt.Printf(">>>>>>>>>>>>>>> <<<<<<<<<<<<<<<\n\n\n")
		logs.Importantf(">>>>>>>>>>>>>>> Succ Run %s-%d at %s,cost: %d ms ================",
			GetServerName(), GetServerID(), now.Format(baselib.TimeFormatMilli), now.Sub(startTime)/time.Millisecond)
	}()

	sig := baselib.NewSignalHandler()
	timewait := time.NewTicker(200 * time.Millisecond)
	defer timewait.Stop()
For:
	for {
		select {
		case <-sig.ReloadSignal():
			fmt.Println("Receive Reload Signal, Reloading")
			logs.Important("Receive Reload Signal, Reloading")
			_ = CheckSysStatus()
		case <-sig.UnloadSignal():
			fmt.Println("Receive Unload Signal, UnloadAllSubject")
			logs.Important("Receive Unload Signal, UnloadAllSubject")
			UnloadAllSubject()
			stopRpcxServer()
			//stopRpcxV3Server()
		case <-sig.StopSignal():
			fmt.Println("Receive Stop Signal, Stopping")
			logs.Important("Receive Stop Signal, Stopping")
			break For
		case <-sig.ProfSignal():
			if fileutil.Exist("profile") {
				_ = os.Remove("profile")
				fmt.Println("Receive Prof Signal, Start Profile")
				logs.Important("Receive Prof Signal, Start Profile")
				printMemStats()
				go ProfileToFile("profile", 20)
			} else {
				log.Println("Receive Stop Signal, Stopping")
				logs.Important("Receive Stop Signal, Stopping")
				ManualStop()
				break For
			}
		case <-timewait.C:
			if g_server_is_stop {
				log.Printf("g_server_is_stop, Stopping")
				logs.Important("g_server_is_stop, Stopping")
				break For
			}
		}
	}
	CloseServer()

	for _, f := range g_deferFuncs {
		if f != nil {
			f()
		}
	}
	once.Do(stop)
	return
}

func stop() {
	startTime := time.Now()
	logs.Importantf(">>>>>>>>>>>>>>> Start Stop %s-%d at %s ================", GetServerName(), GetServerID(), startTime.Format(baselib.TimeFormatMilli))
	log.Printf(">>>>>>>>>>>>>>> Start Stop %s-%d at %s ================\n", GetServerName(), GetServerID(), startTime.Format(baselib.TimeFormatMilli))

	// 必须按顺序来，先将nats的订阅全部取消, 使得新请求不会再过来
	UnloadAllSubject()
	stopRpcxServer()
	//stopRpcxV3Server()

	//log.Printf("UnloadAllSubject\n")
	//延迟一秒, 处理完所有的消息
	time.Sleep(1000 * time.Millisecond)
	// 然后处理剩下的请求之后，停止Service的协程
	//StopService()
	log.Printf("Server %s-%d stop!!!\n", GetServerName(), GetServerID())
	logs.Infof("Server %s-%d stop!!!", GetServerName(), GetServerID())
	// 最后Flush日志缓存
	logs.CloseAllLogAdapter() // 注册清理操作，将并发管道中还没有打印的日志全部打印完才退出
	log.Printf("CloseAllLogAdapter\n")
	//logs.CloseAllSyslogAdapter()
	//buf, _ := ioutil.ReadFile("run.id")
	//runinfo := fmt.Sprintf("%s\n%d %s %d %s ... stop\n", buf, os.Getpid(),
	//	GetServerName(), GetServerID(), time.Now().Format(utils.TimeFormat_Full))
	//ioutil.WriteFile("run.id", []byte(runinfo), os.ModeAppend)

	now := time.Now()
	logs.Infof(">>>>>>>>>>>>>>> Succ Stop %s-%d at %s,cost: %d ms ================",
		GetServerName(), GetServerID(), now.Format(baselib.TimeFormatMilli), now.Sub(startTime)/time.Millisecond)
	log.Printf(">>>>>>>>>>>>>>> Succ Stop %s-%d at %s,cost: %d ms ================\n",
		GetServerName(), GetServerID(), now.Format(baselib.TimeFormatMilli), now.Sub(startTime)/time.Millisecond)

}

func Stop() {
	//once.Do(stop)
}

func ManualStop() {
	// 必须按顺序来，先将nats的订阅全部取消, 使得新请求不会再过来

	logs.Infof("Begin ManualStop ... ")
	log.Println("Begin ManualStop ... ")
	UnloadAllSubject()
	//stopRpcxServer()
	//stopRpcxV3Server()
	//CloseServer()
	stopRegistryPlugin()
	log.Println("End ManualStop")
	logs.Infof("End ManualStop")
}
