package logs

import (
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/frame/session_def"
)

func GetOrCreateTraceId(sesses ...*session_def.Session) *session_def.STrace {
	return runtime.GetOrCreateTrace(sesses...)
}

func GetTraceId(sesses ...*session_def.Session) *session_def.STrace {
	return runtime.GetTrace(sesses...)
}

func CreateTraceId(sess *session_def.Session) (*session_def.STrace, int64) {
	return runtime.CreateTrace(sess)
}

func StoreTraceId(traceId *session_def.STrace, sesses ...*session_def.Session) int64 {
	return runtime.StoreTrace(traceId, sesses...)
}

func RemoveTraceId(gpid int64) {
	runtime.RemoveTrace(gpid)
}

func AutoRemoveTraceId() {
	runtime.AutoRemoveTrace()
}
