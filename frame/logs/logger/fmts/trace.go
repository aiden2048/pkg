package fmts

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/aiden2048/pkg/frame/logs/logger"
	runtimex "github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/utils"
)

var _ logger.Formatter = (*TraceFormatter)(nil)

func NewTraceFormatter(moduleName string, skipCall int, formatType Format, disabledStdoutColor bool, cfgLoader *logger.ConfLoader) *TraceFormatter {
	return &TraceFormatter{
		skipCall:            skipCall,
		formatType:          formatType,
		moduleName:          moduleName,
		cfgLoader:           cfgLoader,
		disabledStdoutColor: disabledStdoutColor,
	}
}

type TraceFormatter struct {
	skipCall            int
	formatType          Format
	moduleName          string
	cfgLoader           *logger.ConfLoader
	disabledStdoutColor bool
}

func (f *TraceFormatter) GetSkipCall() int {
	return f.skipCall
}

func (f *TraceFormatter) SetSkipCall(skipCall int) {
	f.skipCall = skipCall
}

func (f *TraceFormatter) Copy() logger.Formatter {
	return NewTraceFormatter(f.moduleName, f.skipCall, f.formatType, f.disabledStdoutColor, f.cfgLoader)
}

func (f *TraceFormatter) Sprintf(level logger.Level, stdoutColor logger.Color, format string, args ...interface{}) (string, error) {
	levelStr, err := logger.TransferLevelToStr(level)
	if err != nil {
		return "", err
	}

	var colorStdoutStart, colorStdoutEnd string
	if !f.disabledStdoutColor && stdoutColor != logger.ColorNil {
		colorStdoutStart, err = logger.GetColorStdout(stdoutColor)
		if err != nil {
			return "", err
		}

		colorStdoutEnd = logger.ColorToStdoutMap[logger.ColorNil]
	}

	pc, callFile, callLine, ok := runtime.Caller(f.skipCall)
	var callFuncName string
	if ok {
		callFuncName = runtime.FuncForPC(pc).Name()
	}

	_, fileName := path.Split(callFile)

	now := time.Now()
	rawFormatted := strconv.Quote(fmt.Sprintf(format, args...))
	rawFormatted = rawFormatted[1 : len(rawFormatted)-1]

	trace := runtimex.GetTrace()

	if level >= logger.LevelWarn && trace.Request != nil {
		rawFormatted += "------request msg------" + utils.AutoToString(trace.Request)
	}

	switch level {
	case logger.LevelDebug:
		debugMsgMaxLen := f.cfgLoader.GetConf().File.DebugMsgMaxLen
		if debugMsgMaxLen > 0 && int32(utf8.RuneCountInString(rawFormatted)) > debugMsgMaxLen {
			rawFormatted = string([]rune(rawFormatted)[:debugMsgMaxLen]) + "..."
		}
	case logger.LevelInfo:
		infoMsgMaxLen := f.cfgLoader.GetConf().File.InfoMsgMaxLen
		if infoMsgMaxLen > 0 && int32(utf8.RuneCountInString(rawFormatted)) > infoMsgMaxLen {
			rawFormatted = string([]rune(rawFormatted)[:infoMsgMaxLen]) + "..."
		}
	}

	switch f.formatType {
	case FormatText:
		//return fmt.Sprintf(
		//	"[%s.%04d] [%s] [%s][%d:%d] %s %s:%s:%d %s\n",
		//	now.Format("2006-01-02 15:04:05"),
		//	now.Nanosecond()/100000,
		//	f.moduleName,
		//	trace.TraceId,
		//	trace.AppId,
		//	trace.Uid,
		//	levelStr,
		//	callFuncName,
		//	fileName,
		//	callLine,
		//	rawFormatted,
		//), nil
		return fmt.Sprintf(
			"[%s.%04d] [%s] [%s][%d:%d] %s%s %s:%s:%d %s%s \n",
			now.Format("2006-01-02 15:04:05"),
			now.Nanosecond()/100000,
			f.moduleName,
			trace.TraceId,
			trace.AppId,
			trace.Uid,
			colorStdoutStart,
			levelStr,
			callFuncName,
			fileName,
			callLine,
			colorStdoutEnd,
			rawFormatted,
		), nil
	default:
		return "", errors.New("not support log format")
	}
}
