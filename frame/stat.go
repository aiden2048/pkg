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
