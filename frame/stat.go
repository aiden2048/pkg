package frame

import (
	"fmt"
	"os"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"

	"github.com/aiden2048/pkg/frame/stat"

	"time"
)

var hostName string

func init() {

	// 消息注册
	hostName, _ = os.Hostname()
}
func ReportCallRpcStat(serviceName string, funcName string, result int32, processTime time.Duration) {
	key := fmt.Sprintf("rp:call-nats-%s-%s", serviceName, funcName)
	stat.ReportStat(key, int(result), processTime)
}
func ReportDoRpcStat(serviceName string, funcName string, result int32, processTime time.Duration) {
	key := fmt.Sprintf("rp:do-nats-%s-%s", funcName, serviceName)
	stat.ReportStat(key, int(result), processTime)
}
func ReportStat(object string, funcName string, result int, processTime time.Duration) {
	key := fmt.Sprintf("rp:%s-%s", object, funcName)
	stat.ReportStat(key, int(result), processTime)
}

/*
func statisicDataReport(req *monitor.StatisticReqMsgPara) (*monitor.StatisticRspMsgPara, error) {
b := &msg.PbMsg{req}
	// SendAndRecv
	session := NewEmptySession()
	rpc := NewRpc(session)
	rsp, err := rpc.NatsCall(session.GetUin(), "monitor", "StatisicDataReport", rb, 2)
	if err != nil {
		logs.Infof("monitor StatisicDataReport failed:%s", err)
		return nil, errors.New("NatsCall Error")
	}
	// 转成PbMsg
	if rsp, ok := rsp.(*msg.PbMsg); ok {
		// 得到PbMsg里的Pb结构
		if pb, ok := rsp.Pb.(*monitor.StatisticRspMsgPara); ok {
			return pb, nil
		} else {
			return nil, errors.New("Rsp Error") //
		}
	} else {
		logs.Infof("Recv Error Msg")
		return nil, errors.New("Rsp Error") //
	}
}
*/

func additionalMsgStat(key string, value *stat.MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration) {
	if GetNatsConn() == nil {
		logs.Importantf("nats 还没起来， 不能上报")
		return
	}
	// 组包上报统计数据
	evt := &ReportStatMsgEvent{}
	if err := utils.StructConv(value, evt); err != nil {
		logs.Importantf("err:%v", err)
		return
	}
	evt.AvgProcessTime = int64(avgProcessTime)
	evt.AvgSuccProcessTime = int64(avgSuccProcessTime)
	evt.Host = fmt.Sprintf("%d:%s", GetPlatformId(), hostName)
	evt.Key = key
	evt.SvrName = GetServerName()
	evt.SvrID = GetServerID()
	evt.PlatId = GetPlatformId()
	evt.Send()
}
