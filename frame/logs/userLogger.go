package logs

import (
	"sync"

	"github.com/aiden2048/pkg/frame/logs/logger"
	"github.com/aiden2048/pkg/frame/session_def"
)

type PlayerLoggerMng struct {
	mu          sync.Mutex
	traceAllUin bool
	traceUin    map[uint64]bool
}

var playerLoggerMng *PlayerLoggerMng

func GetPlayerLoggerMng() *PlayerLoggerMng {
	return playerLoggerMng
}

func newPlayerLoggerMng() *PlayerLoggerMng {
	plm := new(PlayerLoggerMng)
	plm.mu.Lock()
	plm.traceUin = make(map[uint64]bool)
	plm.mu.Unlock()
	return plm
}

func (plm *PlayerLoggerMng) ClearTraceUin() {
	plm.mu.Lock()
	defer plm.mu.Unlock()
	plm.traceUin = make(map[uint64]bool)
}

func (plm *PlayerLoggerMng) AddTraceUin(uin uint64) {
	plm.mu.Lock()
	defer plm.mu.Unlock()
	plm.traceUin[uin] = true
}

func (plm *PlayerLoggerMng) CheckUin(uin uint64) bool {
	if uin == 0 {
		return false
	}
	if plm.traceAllUin {
		return true
	}
	plm.mu.Lock()
	defer plm.mu.Unlock()
	return plm.traceUin[uin]
}

func (plm *PlayerLoggerMng) IsTraceUin(uin uint64) bool {
	if uin == 0 {
		return false
	}
	plm.mu.Lock()
	defer plm.mu.Unlock()
	return plm.traceUin[uin]
}

const MaxUserId = 10000000000

func GetUidFromAppUid(uid uint64) uint64 {
	return uid % MaxUserId
}

func IsNeedTraceUid(uin uint64) bool {
	if uin < 10000 {
		return false
	}
	uin = GetUidFromAppUid(uin)
	return /* CheckPrintLogLevel(LL_DEBUG) ||*/ playerLoggerMng.CheckUin(uin) || IsDebug()
}

func (plm *PlayerLoggerMng) WriteMsg(_ uint64, format string, v ...interface{}) {
	Importantf(format, v...)
}

func SetTraceAllUid(b bool) {
	playerLoggerMng.traceAllUin = b
	Trace("set traceAllUin %t", b)
}

func ClearTraceUid() {
	playerLoggerMng.ClearTraceUin()
}

func AddTraceUid(uin uint64) {
	playerLoggerMng.AddTraceUin(uin)
}

// Deprecated: replace by Debug
func LogUser(_ uint64, format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelDebug, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Debug
func PrintUser(_ uint64, v ...interface{}) {
	WriteBySkipCall(logger.LevelDebug, v1AdapterSkipCall, fmtMsgForPrint(v...))
}

// Deprecated: replace by Info
func LogInfoUser(_ uint64, format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelInfo, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Debugf
func LogAppUser(_ int32, _ uint64, format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelDebug, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Errorf
func LogUserError(_ uint64, format string, v ...interface{}) {
	WriteBySkipCall(logger.LevelError, v1AdapterSkipCall, format, v...)
}

// Deprecated: replace by Error
func PrintUserError(_ uint64, v ...interface{}) {
	WriteBySkipCall(logger.LevelError, v1AdapterSkipCall, fmtMsgForPrint(v...))
}

// Deprecated: replace by Debug
func PrintSess(_ *session_def.Session, v ...interface{}) {
	WriteBySkipCall(logger.LevelDebug, v1AdapterSkipCall, fmtMsgForPrint(v...))
}

// Deprecated: replace by Error
func PrintSessError(_ *session_def.Session, v ...interface{}) {
	WriteBySkipCall(logger.LevelError, v1AdapterSkipCall, fmtMsgForPrint(v...))
}
