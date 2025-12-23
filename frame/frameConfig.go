package frame

import (
	"fmt"
	"os"

	"github.com/aiden2048/pkg/utils"

	logger2 "github.com/aiden2048/pkg/frame/logs/logger"

	runtimex "github.com/aiden2048/pkg/frame/runtime"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"

	"github.com/smallnest/rpcx/share"

	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/stat"
	"github.com/aiden2048/pkg/utils/baselib"

	"github.com/BurntSushi/toml"
)

type RunConfig struct {
	ServerName string
	ServerID   int32
	Port       int
}

// type FrameConfig struct {
// 	LogConf *logger2.LogConf
// }

// var config = NewDefaultFrameConfig()
var server_config = RunConfig{ServerName: "goServer", ServerID: 0}
var GameNames = []string{"eGame", "flyGame" /*"quickG", "minesG", "singleG",*/, "gchat", "dataWatch"}

func GetServerConfig() *RunConfig {
	return &server_config
}

func GetServerName() string {
	return server_config.ServerName
}

func GetServerID() int32 {
	return server_config.ServerID
}
func GetServerPort() int {
	return server_config.Port
}
func GetStrServerID() string {
	return fmt.Sprintf("%d", server_config.ServerID)
}

func GetPlatformId() int32 {
	return GetGlobalConfig().PlatformID
}

func GetCallTimeout() int32 {
	return defFrameOption.RpcCallTimeout
}

func GetCallHttpTimeout() int32 {
	return defFrameOption.RpcCallTimeout * 3
}

func IsLogTable() bool {
	return defFrameOption.LogConf.GameFrame.TraceAllTable
}

func GetRpcCallTimeout() time.Duration {
	return time.Duration(defFrameOption.RpcCallTimeout) * time.Second
}

func IsUseRpcx() bool {
	if defFrameOption.EnableMixServer {
		return true
	}
	newConf := GetEtcdConfig()
	if len(newConf.GetCenterEtcdAddr(GetPlatformId())) == 0 {
		return false
	}
	return true
}

func InitConfig(svrName string, opt ...*FrameOption) error {
	// 初始化
	if len(opt) > 0 {
		defFrameOption = opt[0]
	}
	cpun := runtime.NumCPU()
	if cpun >= 16 {
		cpun = 16
		runtime.GOMAXPROCS(cpun)
	}

	// 加载框架配置
	server_config.ServerName = svrName
	server_config.Port = defFrameOption.Port
	if defFrameOption.Svrid <= 0 {
		if len(os.Args) > 1 {
			for _, arg := range os.Args {
				sid, err := strconv.Atoi(arg)
				if err == nil && sid > 0 {
					server_config.ServerID = int32(sid)
					break
				}
			}
		}
		//进程ID不超过九位数,所以做了组ID精简
		if server_config.ServerID <= 0 {

			nPlatformId := int64(_global_config.PlatformID) * 100000
			if utils.InArray(GameNames, GetServerName()) {
				server_config.ServerID = int32(nPlatformId)
			} else {
				ip := baselib.GetLocalIP()
				ips := strings.Split(ip, ".")
				if len(ips) > 3 {
					nPlatformId += utils.StrToInt64(ips[3])
				}
				server_config.ServerID = int32(nPlatformId)
			}

		}

	} else {
		server_config.ServerID = int32(defFrameOption.Svrid)
	}

	sleepTime := time.Millisecond * 20
	os := runtime.GOOS
	if os == "windows" || os == "darwin" {
		logs.InitServer(server_config.ServerName, server_config.ServerID, "../logs/", IsTestServer(), ReportLog, ReportBillStat)
	} else {
		logs.InitServer(server_config.ServerName, server_config.ServerID, "/data/logs/", IsTestServer(), ReportLog, ReportBillStat)
	}
	// 根据bin目录下touch文件, 在reload的时候检查各个状态
	baselib.RegisterReloadFunc(CheckSysStatus)
	// 加载启动配置
	if err := LoadBootConfig(); err != nil {
		logs.Errorf("InitConfig error:%s", err)
		logs.Errorf("Run exit")
		//等待一会, 让日志打印出去
		time.Sleep(sleepTime)
		return err
	}
	if defFrameOption.EnableMixServer && !GetGlobalConfig().EnableMixServer {
		logs.Errorf("platid:%d 本组不能启用MixServer, 如需启用检查下Platform.tom配置,强制改成单组服务", GetPlatformId())
		fmt.Printf("platid:%d 本组不能启用MixServer, 如需启用检查下Platform.tom配置,强制改成单组服务\n", GetPlatformId())
		defFrameOption.EnableMixServer = false
	}
	//给top组专用的, 因为top的组ID<100, 如果不强制指定EnableAllAreaMix,不能启用通服
	if defFrameOption.EnableMixServer && GetGlobalConfig().PlatformID < 1000 && !defFrameOption.EnableAllAreaMix {
		logs.Errorf("platid:%d 本组不能启用MixServer,必须开启EnableMixServer", GetPlatformId())
		fmt.Printf("platid:%d 本组不能启用MixServer, 必须开启EnableMixServer\n", GetPlatformId())
		defFrameOption.EnableMixServer = false
	}
	if defFrameOption.EnableMixServer {
		defFrameOption.AllowCrossServer = true
	}
	if err := LoadFrameConfig(); err != nil {
		//等待一会, 让日志打印出去
		time.Sleep(sleepTime)
		return err
	}
	baselib.RegisterReloadFunc(LoadGlobalConfig)
	baselib.RegisterReloadFunc(LoadFrameConfig)
	logs.Infof("Init Server GOMAXPROCS:%d, NumCpu:%d", cpun, runtime.NumCPU())
	logs.Infof("Init Server %+v, FrameConfig:%+v", server_config, defFrameOption)

	if err := LoadSystemConfig(); err != nil {
		//等待一会, 让日志打印出去
		time.Sleep(sleepTime)
		return err
	}
	baselib.RegisterReloadFunc(LoadSystemConfig)
	logs.SetServerId(GetServerID())

	stat.SetAdditionMsgReport(GetServerName(), additionalMsgStat)
	return nil
}

func LoadBootConfig() error {
	if defFrameOption.EnableMysql {
		err := LoadMysqlConfig()
		if err != nil {
			return err
		}
		logs.Infof("connect to Mysql")
	}
	if !defFrameOption.DisableMgo {
		err := LoadMgoConfig()
		if err != nil {
			return err
		}
		logs.Infof("connect to Mgo")
	}

	if err := LoadGlobalConfig(); err != nil {
		return err
	}
	return nil
}

func LoadFrameConfig() error {
	if !defFrameOption.DisableRpcx {
		if err := LoadEtcdConfig(); err != nil {
			return err
		}
	}
	if err := LoadNatsConfig(); err != nil {
		return err
	}
	logs.Infof("Read Etcd Config:%+v", etcdConfig)
	if defFrameOption.RpcCallTimeout <= 0 {
		defFrameOption.RpcCallTimeout = DEFAULT_RPC_REQUEST_SECONDS
	}

	logConf, err := _loadLogConfig()
	if err == nil && logConf != nil {
		defFrameOption.LogConf = logConf
	} else {
		defFrameOption.LogConf = &logger2.LogConf{
			File: logger2.FileLogConf{
				MaxFileSizeBytes:            logs.MaxFileSizeBytes,
				LogDebugBeforeFileSizeBytes: logs.LogDebugBeforeFileSizeBytes,
				LogInfoBeforeFileSizeBytes:  -1,
				DebugMsgMaxLen:              logs.DebugMsgMaxLen,
				FileMaxRemainDays:           3,
				CompressFrequentHours:       24,
				Level:                       logger2.LevelToStrMap[logger2.LevelDebug],
				IsTestServer:                IsTestServer(),
			},
			AlertLevel: logger2.LevelToStrMap[logger2.LevelWarn],
		}

	}

	LoadLogConfig(defFrameOption.LogConf)

	share.Trace = defFrameOption.LogConf.LogRpcx
	return nil
}

func LoadSystemConfig() error {

	return nil
}

func LoadLogConfig(cfg *logger2.LogConf) {
	if fileutil.Exist("debug_log") {
		logs.Important("exits debug_log, force debug")
		runtimex.SetForceDebug(true)
	} else {
		logs.Important("not exits debug_log, not force debug")
		runtimex.SetForceDebug(false)
	}
	logs.SetLogConfig(cfg)
	// 加载UinDebug配置
	LoadUinConfig(cfg)
}

func _loadLogConfig() (*logger2.LogConf, error) {
	newConf := &logger2.LogConf{}
	filename := "../LocalConfig/log.toml"
	_, err := toml.DecodeFile(filename, newConf)
	if err == nil {
		logs.Infof("load LocalFile %s config: %+v", filename, newConf)
		return nil, err
	}
	return newConf, err
}

type UidTraceCfg struct {
	Uids []uint64
}

func LoadUinConfig(cfg *logger2.LogConf) error {
	logs.ClearTraceUid()
	logs.SetTraceAllUid(cfg.GameFrame.TraceAllUid)
	if cfg.GameFrame.TraceAllUid {
		logs.Infof("Log.TraceAllUid")
		//return nil
	}

	newConf := &UidTraceCfg{}
	fkey := "trace_uids.toml"

	filename := fmt.Sprintf("../LocalConfig/%s", fkey)
	_, err := toml.DecodeFile(filename, newConf)
	if err == nil {
		for _, u := range newConf.Uids {
			if u > 10000 {
				logs.AddTraceUid(u)
				logs.Infof("AddTraceUid %d Local", u)
			}
		}
		return nil
	}

	err = LoadConfigFromMongo(fkey, newConf)
	if err != nil {
		logs.Infof("load Mysql %s config: %+v", fkey, newConf)
		return err
	}

	for _, u := range newConf.Uids {
		if u > 10000 {
			logs.AddTraceUid(u)
			logs.Infof("AddTraceUid %d Mysql", u)
		}
	}

	return nil
}
