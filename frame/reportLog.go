package frame

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
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

/*
*
老的方式，publish nats，nats去掉通服之后不能使用
*/
func ReportLogWithNats(ltype, ltag, lmsg string) {
	if GetServerName() == "monitor" {
		return
	}
	//logs.LogDebug("ReportLog ltype:%s, ltag:%s, lmsg:%s", ltype, ltag, lmsg)
	// 组包上报统计数据
	value := &TLogReport{}
	value.ServerHost = hostName
	value.Pid = os.Getpid()
	value.LogType = ltype
	value.LogTag = ltag
	value.LogMsg = lmsg

	reqMsg := &NatsMsg{}
	reqMsg.Sess = Session{}
	reqMsg.MsgBody.Func = "ReportLog"
	reqMsg.MsgBody.Check = time.Now().Format("2006-01-02 15:04:05.999")

	reqMsg.MsgBody.Param, _ = jsoniter.Marshal(value)

	reqMsg.Sess.PlatId = GetPlatformId()
	reqMsg.Sess.SvrFE = GetServerName()
	reqMsg.Sess.SvrID = GetServerID()
	reqMsg.Sess.Time = time.Now().Unix()

	//var err error
	//if CheckRpcxService("monitor", -1, "ReportLog") {
	//	err = SendRpcx("monitor", "ReportLog", -1, reqMsg)
	//} else {
	//
	//	//	req.Sess.Route = fmt.Sprint("%s->%s.%d", req.Sess.Route, GetServerName(), GetServerID())
	//	subj := GenReqSubject("monitor", "ReportLog", -1)
	//	// data, err := jsoniter.Marshal(reqMsg)
	//	// if err != nil {
	//	// 	logs.Errorf("NatsCall Marshal req (%+v) failed:%s", reqMsg, err.Error())
	//	// 	return
	//	// }
	//	err = NatsPublish(g_natsConn, subj, reqMsg, false, nil)
	//}
	subj := GenReqSubject("monitor", "ReportLog", -1)
	// data, err := jsoniter.Marshal(reqMsg)
	// if err != nil {
	// 	logs.Errorf("NatsCall Marshal req (%+v) failed:%s", reqMsg, err.Error())
	// 	return
	// }
	logs.Print("NatsPublish ReportLog", "value", value, "reqMsg", reqMsg, "subj", subj)
	if gNatsconn != nil {
		data, _ := jsoniter.Marshal(reqMsg)
		err := gNatsconn.Publish(subj, data)
		logs.Debugf("NatsPublish Publish err:%+v, subj:%s, reqMsg:%+v", err, subj, reqMsg)
	}
}

/*
*
使用rpcx发error，etcd通服
*/
//
//func getMonitorDis(sname string) (client.ServiceDiscovery, error) {
//	var d client.ServiceDiscovery
//	var err error
//	// etcd v3
//	platId := GetPlatformId()
//	discoveryKey := sname
//	centerAddr := GetEtcdConfig().GetCenterEtcdAddr(platId)
//	d, err = etcdclient.NewEtcdV3Discovery(GetEtcdConfig().BasePath, discoveryKey, centerAddr, true, nil) // v3版本
//	if err != nil {
//		if d != nil {
//			d.Close()
//		}
//		logs.PrintInfo("Etcd Discovery v3 启动Etcd发现服务失败:", err.Error(), centerAddr, "platId", platId, "sname", sname, "discoveryKey", discoveryKey)
//		return nil, err
//	}
//	if len(d.GetServices()) <= 0 {
//		d.Close()
//		err = errors.New("No-Services")
//	}
//	return d, nil
//
//}

var monitorDiscovery client.ServiceDiscovery
var locker = sync.Mutex{}
var logWorkerPool = pond.New(1, 100)

//
//func SendToMonitor1(sname, funcName string, req interface{}) {
//	locker.Lock()
//	if monitorDiscovery == nil {
//		dis, err := createDiscoveryV3(GetPlatformId(), sname)
//		if dis == nil {
//			logs.Importantf("SendToMonitor failed:%s ---------- Msg:%s", utils.AutoToString(err), utils.AutoToString(req))
//			locker.Unlock()
//			return
//		}
//		monitorDiscovery = dis
//	}
//	if len(monitorDiscovery.GetServices()) <= 0 {
//		log.Printf("Monitor is closed\n")
//		monitorDiscovery = nil
//		locker.Unlock()
//		return
//	}
//	locker.Unlock()
//
//	option := client.DefaultOption
//	option.Heartbeat = true
//	option.HeartbeatInterval = time.Second
//
//	failMode := client.Failfast
//	//selectMode := client.WeightedICMP
//	selectMode := client.RandomSelect
//
//	mClient := client.NewXClient(sname, failMode /*client.Failfast*/, selectMode /*client.RandomSelect*/, monitorDiscovery, option)
//	defer func() {
//		_ = mClient.Close()
//	}()
//	reqBuf := &TRpcxMsg{}
//	var err error
//	reqBuf.Data, err = jsoniter.Marshal(req)
//	if err != nil {
//		logs.Importantf("SendToMonitor Marshal failed:%s", utils.AutoToString(err))
//		return
//	}
//	err = mClient.Call(context.TODO(), funcName, reqBuf, nil)
//
//	if err != nil {
//		logs.Importantf("SendToMonitor mClient.Call failed:%s", utils.AutoToString(err))
//		return
//	}
//}

var isReporting = baselib.Map{}

func ReportLog(ltype, ltag, lmsg string) {
	// 组包上报统计数据
	if GetNatsConn() == nil {
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
