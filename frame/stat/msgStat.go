package stat

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/logs/logger"
)

const (
	maxFileNum  = 5
	maxSize     = 100 * 1024 * 1024
	maxDays     = 7
	channelSize = 10 * 1024
)
const (
	Stat_Sucessfull = 0
	Stat_Failed     = 1
	Stat_Timeout    = 2
)

type ReportData struct {
	key         string        // msgid/rpcname
	report_type int           //0, 累加统计, 1: set 统计
	result      int           // succeed/failed/timeout
	processTime time.Duration // 处理耗时
}

type MsgStatData struct {
	Host          string `json:"host"`
	Key           string `json:"key"`
	Type          int32  `json:"type"`            //0:累计统计, 每次输出后清零, 1: 重置型统计, 每次输出后不清零
	TotalMsgNum   int64  `json:"total_msg_num"`   //!<消息处理总数
	SuccessMsgNum int32  `json:"success_msg_num"` //!<消息处理成功的个数
	FailMsgNum    int32  `json:"fail_msg_num"`    //!<消息处理失败的个数
	TimeoutMsgNum int32  `json:"timeout_msg_num"` //!<消息处理超时的个数

	SumProcessTime     time.Duration `json:"sum_process_time"`      //!<总处理耗时
	MaxProcessTime     time.Duration `json:"max_process_time"`      //!<最大处理耗时
	SumSuccProcessTime time.Duration `json:"sum_succ_process_time"` //!<成功请求-总处理耗时
	MaxSuccProcessTime time.Duration `json:"max_succ_process_time"` //!<成功请求-最大处理耗时

	AvgProcessTime     int64 `json:"avg_process_time"`
	AvgSuccProcessTime int64 `json:"avg_succ_process_time"`
}

type MsgStat struct {
	FileLogger        *logger.Logger
	statData          map[string]*MsgStatData
	statChan          chan *ReportData
	additionMsgReport func(key string, value *MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration)
}

var defaultMsgStat *MsgStat

func init() {

}
func getDefaultmsgStat() *MsgStat {
	if defaultMsgStat == nil {
		defaultMsgStat = NewMsgStat(filepath.Base(os.Args[0]), noAdditionMsgReport)
	}
	return defaultMsgStat
}
func noAdditionMsgReport(key string, value *MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration) {
}

func NewMsgStat(svrName string, additionMsgReport func(key string, value *MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration)) *MsgStat {
	m := &MsgStat{additionMsgReport: additionMsgReport}

	cfgLoader := logs.MustDefaultCfgLoader()
	var err error
	m.FileLogger, err = logs.InitFileLogger(
		svrName,
		cfgLoader.GetConf().File.StatLogDir,
		fmt.Sprintf("msgStat.%s.log", svrName),
		5,
		cfgLoader,
	)
	if err != nil {
		panic(err)
	}

	m.statData = map[string]*MsgStatData{}
	m.statChan = make(chan *ReportData, 65536)

	go m.RunStat()
	return m
}

func SetAdditionMsgReport(svrName string, reportFunc func(key string, value *MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration)) {
	defaultMsgStat = NewMsgStat(svrName, noAdditionMsgReport)
	getDefaultmsgStat().SetAdditionMsgReport(reportFunc)
}

func (m *MsgStat) SetAdditionMsgReport(reportFunc func(key string, value *MsgStatData, avgProcessTime time.Duration, avgSuccProcessTime time.Duration)) {
	m.additionMsgReport = reportFunc
}

func (m *MsgStat) Reset() {
	for key, v := range m.statData {
		if v.Type == 0 {
			m.statData[key] = &MsgStatData{}
		}
	}
}

func (m *MsgStat) RunStat() {
	lastPrintTime := time.Now()
	//tc := time.Tick(time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Sub(lastPrintTime) >= time.Minute {
				m.PrintAllStat()
				lastPrintTime = now
			}
		case data := <-m.statChan:
			if data.report_type == 0 {
				m.AddStat(data)
			} else {
				m.SetStat(data)
			}
		}
	}
}

func (m *MsgStat) AddStat(data *ReportData) {
	// 获取统计结点，不存在则插入
	key := data.key
	if _, ok := m.statData[key]; !ok {
		m.statData[key] = &MsgStatData{}
	}
	pStatData := m.statData[key]

	// 统计成功/失败/超时
	pStatData.TotalMsgNum++
	switch data.result {
	case 0:
		pStatData.SuccessMsgNum++
	case 1, -1:
		pStatData.FailMsgNum++
	case 2, -2:
		pStatData.TimeoutMsgNum++
	default:
	}

	// 统计处理耗时
	pStatData.SumProcessTime += data.processTime
	if pStatData.MaxProcessTime < (data.processTime) {
		pStatData.MaxProcessTime = (data.processTime)
	}
	if data.result == 0 {
		pStatData.SumSuccProcessTime += data.processTime
		if pStatData.MaxSuccProcessTime < data.processTime {
			pStatData.MaxSuccProcessTime = data.processTime
		}
	}
}

func (m *MsgStat) SetStat(data *ReportData) {
	// 获取统计结点，不存在则插入
	key := data.key
	if _, ok := m.statData[key]; !ok {
		m.statData[key] = &MsgStatData{}
	}
	pStatData := m.statData[key]

	// 统计成功/失败/超时
	pStatData.SuccessMsgNum = int32(data.result)
	if int64(data.result) > pStatData.TotalMsgNum {
		pStatData.TotalMsgNum = int64(data.result)
	}

}

func (m *MsgStat) PrintAllStat() {
	m.FileLogger.Importantf("=========MsgStat begin=========")
	for key, value := range m.statData {
		if value.TotalMsgNum <= 0 {
			continue
		}
		avgProcessTime := time.Duration(int64(value.SumProcessTime) / value.TotalMsgNum)
		avgSuccProcessTime := time.Duration(int64(value.SumSuccProcessTime) / value.TotalMsgNum)
		m.FileLogger.Importantf("%s: Success = %d, Fail = %d, Timeout = %d, Total = %d, MaxTime = %+v, AvgTime = %+v, TotalTime = %+v, MaxSuccTime = %+v, AvgSuccTime = %+v",
			key,
			value.SuccessMsgNum,
			value.FailMsgNum,
			value.TimeoutMsgNum,
			value.TotalMsgNum,
			value.MaxProcessTime,
			avgProcessTime,
			value.SumProcessTime,
			value.MaxSuccProcessTime,
			avgSuccProcessTime,
		)
		if m.additionMsgReport != nil {
			m.additionMsgReport(key, value, avgProcessTime, avgSuccProcessTime)
		}
	}
	m.FileLogger.Important("=========MsgStat end=========")

	m.Reset()
}

func (m *MsgStat) ReportStat(key string, t int, result int, processTime time.Duration) {
	data := &ReportData{key: key, report_type: t, result: result, processTime: processTime}
	m.statChan <- data
}

func ReportStat(key string, result int, processTime time.Duration) {
	getDefaultmsgStat().ReportStat(key, 0, result, processTime)
}

func ReportTotalStat(key string, result int) {
	getDefaultmsgStat().ReportStat(key, 1, result, 0)
}
