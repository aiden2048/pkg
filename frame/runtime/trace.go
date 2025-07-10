package runtime

import (
	"encoding/binary"
	"math/rand"
	"os"
	"time"

	"github.com/aiden2048/pkg/frame/session_def"
	"github.com/aiden2048/pkg/utils/baselib"
)

var TraceMap = baselib.Map{}

const (
	chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789xy"
)

var pid = uint16(0)

func GenTraceStrId() string {
	if pid == 0 {
		pid = uint16(os.Getpid())
	}
	var buf [15]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(time.Now().UnixNano()))
	binary.LittleEndian.PutUint16(buf[8:], pid)
	binary.LittleEndian.PutUint32(buf[10:], uint32(rand.Int()))
	//for i := 1; i < 7; i++ {
	//	x := buf[i]
	//	buf[i] = buf[14-i]
	//	buf[14-i] = x
	//}
	// 3 byte -> 4 char
	var s [20]byte
	d := 0
	for i := 0; i < 15; {
		val := uint(buf[i])<<16 | uint(buf[i+1])<<8 | uint(buf[i+2])
		i += 3
		s[d] = chars[val>>18&0x3F]
		s[d+1] = chars[val>>12&0x3F]
		s[d+2] = chars[val>>6&0x3F]
		s[d+3] = chars[val&0x3F]
		d += 4
	}
	return string(s[:])
}

func GetOrCreateTrace(sesses ...*session_def.Session) *session_def.STrace {
	gpid := GetGpid()
	traceAny, ok := TraceMap.Load(gpid)
	if ok {
		return traceAny.(*session_def.STrace)
	}

	trace := &session_def.STrace{}
	trace.TraceId = GenTraceStrId()
	if len(sesses) > 0 {
		sess := sesses[0]
		trace.Sess = sess
		trace.AppId, trace.Uid = sess.GetAppID(), sess.GetUid()
	}
	TraceMap.Store(gpid, trace)
	return trace
}

func GetTrace(sesses ...*session_def.Session) *session_def.STrace {
	gpid := GetGpid()
	traceAny, ok := TraceMap.Load(gpid)
	if ok {
		return traceAny.(*session_def.STrace)
	}
	trace := &session_def.STrace{}
	trace.TraceId = GenTraceStrId()
	if len(sesses) > 0 && sesses[0] != nil {
		trace.Sess = sesses[0]
		trace.AppId, trace.Uid = sesses[0].GetAppID(), sesses[0].GetUid()
	}
	// trace.Begin = utils.GetCallFile(2)
	return trace
}

func CreateTrace(sess *session_def.Session) (*session_def.STrace, int64) {
	gpid := GetGpid()
	/*	traceid, ok := TraceMap.Load(gpid)
		if ok {
			tid, _ := traceid.(*session_def.STrace)
			if tid != nil {
				return tid, ""
			}

		}*/
	trace := &session_def.STrace{
		Sess: sess,
	}
	trace.TraceId = GenTraceStrId()
	// trace.Begin = utils.GetCallFile(2)
	if sess != nil {
		trace.AppId, trace.Uid = sess.GetAppID(), sess.GetUid()
	}
	TraceMap.Store(gpid, trace)
	return trace, gpid
}

func StoreTrace(trace *session_def.STrace, sesses ...*session_def.Session) int64 {
	if trace == nil {
		return 0
	}
	if len(sesses) > 0 {
		trace.Sess = sesses[0]
	}
	gpid := GetGpid()
	TraceMap.Store(gpid, trace)
	return gpid
}

func RemoveTrace(gpid int64) {
	TraceMap.Delete(gpid)
}

func AutoRemoveTrace() {
	RemoveTrace(GetGpid())
}
