package logs

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strings"

	"github.com/aiden2048/pkg/frame/logs/logger"
	runtimex "github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/utils"
)

const (
	MaxDays                     = 3
	MaxFileSizeBytes            = 1024 * 1024 * 200
	LogDebugBeforeFileSizeBytes = 1024 * 1024 * 100
	DebugMsgMaxLen              = 1500
)

var serverName = "Server"
var serverID = int32(1)
var logBasePath = "../"

// ==========1.所有的 LogAdapter 的初始化逻辑（包加载时自动执行）==========
var reportFunc func(typ string, tag string, msg string)

func reportFuncWrapper(msg *logger.Msg) {
	if reportFunc == nil {
		return
	}

	if err := exceptionLogger.GetWriter().WriteMsg(msg); err != nil {
		fmt.Println(err)
	}

	_, file, line, ok := runtime.Caller(msg.SkipCall - 1)
	if !ok {
		file = "???"
		line = 0
	}
	_, fileName := path.Split(file)
	if ss := strings.Split(file, "/"); len(ss) > 2 {
		file = strings.Join(ss[len(ss)-2:], "/")
	}

	tag := fmt.Sprintf("%s:%d", fileName, line)

	reportFunc(logger.LevelToStrMap[msg.Level], tag, msg.Formatted)
}

func InitServer(sn string, sid int32, lp string, isTestServer bool, rfunc func(string, string, string), onBillFunc func(billName string)) {
	serverName = sn
	serverID = sid
	logBasePath = lp
	realLogBasePath := GetLogBasePath()

	maxDays := MaxDays
	if runtimex.IsDebug() {
		maxDays = 1
	}
	err := InitDefaultCfgLoader(sn, "../LocalConfig/log.toml", &logger.LogConf{
		File: logger.FileLogConf{
			MaxFileSizeBytes:            MaxFileSizeBytes,
			LogDebugBeforeFileSizeBytes: LogDebugBeforeFileSizeBytes,
			LogInfoBeforeFileSizeBytes:  -1,
			DebugMsgMaxLen:              DebugMsgMaxLen,
			DefaultLogDir:               realLogBasePath + "/log",
			ExceptionLogDir:             strings.TrimRight(lp, "/") + "/error/log",
			BillLogDir:                  realLogBasePath + "/bills",
			StatLogDir:                  realLogBasePath + "/stat",
			Level:                       logger.LevelToStrMap[logger.LevelDebug],
			FileMaxRemainDays:           maxDays,
			CompressFrequentHours:       24,
			IsTestServer:                isTestServer,
		},
		AlertLevel: logger.LevelToStrMap[logger.LevelWarn],
	})
	if err != nil {
		log.Fatalf("logs.InitServer--logV2.InitDefaultCfgLoader err:%v", err)
		return
	}

	if err = InitDefaultLogger(reportFuncWrapper); err != nil {
		log.Fatalf("logs.InitServer--logV2.InitDefaultLogger err:%v", err)
		return
	}

	if err = InitExceptionLogger(); err != nil {
		log.Fatalf("logs.InitServer--logV2.InitExceptionLogger err:%v", err)
		return
	}

	reportFunc = rfunc

	OnBill(onBillFunc)

	playerLoggerMng = newPlayerLoggerMng()

	Trace("================Start Server %s as id %d, logPath:%s================", serverName, serverID, realLogBasePath)
}

func SetServerId(sid int32) {
	if sid == serverID {
		return
	}
	serverID = sid
	Print("Server", serverName, serverID, "startLog")
}

// CloseAllLogAdapter ==========2.所有的 LogAdapter 的清理逻辑（业务逻辑手动调用）==========
func CloseAllLogAdapter() {
	OnExit()
}

func GetLogBasePath() string {
	return logBasePath + serverName
}

func CheckPrintLogLevel(level logger.Level) bool {
	if GetLevel() >= level {
		return true
	}

	if IsDebug() {
		return true
	}

	return false
}

func CheckLogDebug() bool {
	return CheckPrintLogLevel(logger.LevelDebug)
}

const v1AdapterSkipCall = 6

// Deprecated: replace by Importantf
func Trace(format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelImportant, v1AdapterSkipCall, format, v...)
}

func Print(v ...interface{}) {
	WriteBySkipCall(logger.LevelImportant, v1AdapterSkipCall, fmtMsgForPrint(v...))
}

// Deprecated: replace by Debugf
func LogDebug(format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelDebug, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Infof
func LogInfo(format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelInfo, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Errorf
func LogError(format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelError, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by PrintError
func PrintErr(v ...interface{}) {
	msg := fmt.Sprintf("[%s.%d]|", serverName, serverID)
	for _, a := range v {
		msg = msg + " -- " + utils.AutoToString(a)
	}
	WriteBySkipCall(logger.LevelError, v1AdapterSkipCall, msg)
}

// Deprecated: replace by Warnf
func LogWarn(format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelWarn, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Bill
func WriteBill(tag string, format string, v ...interface{}) {
	BillBySkipCall(v1AdapterSkipCall-1, tag, format, v...)
}

func Console(v ...interface{}) {
	PrintDebug(v...)
	if !runtimex.IsDebug() && !runtimex.IsDev() {
		return
	}
	msg := fmt.Sprintf("[%s.%d]|", serverName, serverID)
	for _, a := range v {
		msg = msg + " -- " + utils.AutoToString(a)
	}
	log.Printf("Console:%s\n", msg)

}
