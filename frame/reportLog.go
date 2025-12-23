package frame

import (
	"context"
	"os"
	"sync"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/frame/stat"
	"github.com/aiden2048/pkg/utils"
	"github.com/aiden2048/pkg/utils/baselib"
	"github.com/alitto/pond"
	jsoniter "github.com/json-iterator/go"
	"github.com/smallnest/rpcx/client"
)

type TLogReport struct {
	ServerHost string `json:"server_host"`
	Pid        int    `json:"pid"`
	LogType    string `json:"log_type"`
	LogTag     string `json:"log_tag"`
	LogMsg     string `json:"log_msg"`
}

var monitorDiscovery client.ServiceDiscovery
var locker = sync.Mutex{}
var logWorkerPool = pond.New(1, 100)

var isReporting = baselib.Map{}

func ReportLog(ltype, ltag, lmsg string) {
	// 组包上报统计数据
	if GetNatsConn(GetPlatformId()) == nil {
		logs.Importantf("nats 还没起来， 不能上报")
		return
	}
	evt := &ReportLogMsgEvent{}
	evt.ServerHost = hostName
	evt.Pid = os.Getpid()
	evt.LogType = ltype
	evt.LogTag = ltag
	evt.LogMsg = lmsg
	evt.SvrName = GetServerName()
	evt.SvrID = GetServerID()
	evt.Send()
}

func SendToMonitor(sname, funcName string, req interface{}) {

	xclient := getXClient(0, GetPlatformId(), sname)
	if xclient == nil {
		return
	}
	reqBuf := &TRpcxMsg{}
	var err error
	reqBuf.Data, err = jsoniter.Marshal(req)
	if err != nil {
		logs.Importantf("SendToMonitor Marshal failed:%s", utils.AutoToString(err))
		return
	}
	err = xclient.Call(context.TODO(), funcName, reqBuf, nil)

	if err != nil {
		logs.Importantf("SendToMonitor mClient.Call failed:%s", utils.AutoToString(err))
		return
	}
}

func ReportToMonitor(funcName string, req interface{}) {
	if GetServerName() == "monitor" {
		return
	}
	gpid := runtime.GetGpid()
	if _, ok := isReporting.Load(gpid); ok {
		logs.PrintImportant("thread: %d is reporting", gpid)
		return
	}
	logWorkerPool.TrySubmit(func() {
		gpid := runtime.GetGpid()
		isReporting.Store(gpid, true)
		SendToMonitor("monitor", funcName, req)
		isReporting.Delete(gpid)
	})
}

func ReportBillStat(billName string) {
	stat.ReportStat("bill:"+billName, 0, 0)
}
