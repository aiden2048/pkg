package logs

import (
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame/logs/logger"
	"github.com/aiden2048/pkg/frame/logs/logger/writer"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/utils"
)

var (
	SrvName                        string
	defaultCfgLoader               *logger.ConfLoader
	defaultLogger, exceptionLogger *logger.Logger
)

func InitDefaultCfgLoader(srvName string, cfgFile string, defaultLogCfg *logger.LogConf) error {
	var err error
	defaultCfgLoader, err = logger.NewConfLoader(cfgFile, 10, defaultLogCfg)
	if err != nil {
		return err
	}

	SrvName = srvName

	return nil
}

func MustDefaultCfgLoader() *logger.ConfLoader {
	if defaultCfgLoader == nil {
		panic("defaultCfgLoader not init")
	}
	return defaultCfgLoader
}

func GetDefaultCfgLoader() (*logger.ConfLoader, bool) {
	if defaultCfgLoader == nil {
		return nil, false
	}

	return defaultCfgLoader, true
}

func InitDefaultLogger(alertFunc writer.AlertFunc) error {
	cfgLoader := MustDefaultCfgLoader()
	cfg := cfgLoader.GetConf()
	var err error
	defaultLogger, err = InitWithAlertFileLogger(SrvName, serverID, cfg.File.DefaultLogDir, SrvName, 6, cfgLoader, alertFunc)
	if err != nil {
		return err
	}
	return nil
}

func InitExceptionLogger() error {
	cfgLoader := MustDefaultCfgLoader()
	cfg := cfgLoader.GetConf()
	var err error
	exceptionLogger, err = InitFileLogger(SrvName, cfg.File.ExceptionLogDir, "error", 5, cfgLoader)
	if err != nil {
		return err
	}
	return nil
}

func OpenNewFileByByDateHour(writer *writer.FileWriter, _ *time.Time, _ bool) (string, bool) {
	fileName := writer.GetFilePrefix() + time.Now().Format("2006010215") + "_" + fmt.Sprintf("%d", serverID) + ".txt"

	if writer.GetCurFileName() != fileName {
		return fileName, true
	}

	return "", false
}

func InitFileLogger(srvName string, baseDir, filePrefix string, skipCall int, cfgLoader *logger.ConfLoader) (*logger.Logger, error) {
	writerCfg := &writer.FileWriterConf{
		ModuleName:               srvName,
		FilePrefix:               filePrefix,
		BaseDir:                  baseDir,
		SkipCall:                 skipCall,
		LogCfgLoader:             cfgLoader,
		CheckFileFullIntervalSec: 10,
		BufChanLen:               100000,
		CheckTimeToOpenNewFile:   OpenNewFileByByDateHour,
		OnLogErr: func(err error) {
			fmt.Println(err)
		},
	}
	fileWriter, err := writer.NewFileWriter(writerCfg)
	if err != nil {
		return nil, err
	}
	runtime.Go(func() {
		fileWriter.Loop()
	})
	var fileLogger *logger.Logger
	fileLogger = logger.NewLogger(fileWriter)
	return fileLogger, nil
}

func InitWithAlertFileLogger(srvName string, srvId int32, baseDir, filePrefix string, skipCall int, cfgLoader *logger.ConfLoader, alertFunc writer.AlertFunc) (*logger.Logger, error) {
	writerCfg := &writer.FileWriterConf{
		ModuleName:               srvName,
		FilePrefix:               filePrefix,
		BaseDir:                  baseDir,
		SkipCall:                 skipCall,
		LogCfgLoader:             cfgLoader,
		CheckFileFullIntervalSec: 10,
		BufChanLen:               100000,
		CheckTimeToOpenNewFile:   OpenNewFileByByDateHour,
		OnLogErr: func(err error) {
			fmt.Println(err)
		},
	}
	fileWriter, err := writer.NewFileWriter(writerCfg)
	if err != nil {
		return nil, err
	}
	runtime.Go(func() {
		fileWriter.Loop()
	})
	var withAlertLogger *logger.Logger
	withAlertLogger = logger.NewLogger(writer.NewWithAlertWriter(fileWriter, cfgLoader, alertFunc))
	return withAlertLogger, nil
}

func OnExit() {
	if defaultLogger != nil {
		if err := defaultLogger.Flush(); err != nil {
			fmt.Println("flush default logger, err:", err)
		}
	}

	for billName, billLogger := range billLoggerFactory.loggerMap {
		if err := billLogger.Flush(); err != nil {
			fmt.Println("flush bill logger:", billName, "err:", err)
		}
	}
}

func SetLogConfig(cfg *logger.LogConf) {
	if cfgLoader, ok := GetDefaultCfgLoader(); ok {
		cfgLoader.SetDefaultLogConf(cfg)
	}
}

func GetLevel() logger.Level {
	if cfgLoader, ok := GetDefaultCfgLoader(); ok {
		return cfgLoader.GetConf().File.GetLevel()
	}

	return logger.LevelDebug
}

func fmtMsgForPrint(v ...interface{}) string {
	msg := "" // fmt.Sprintf("[%s.%d]|", serverName, serverID)
	for _, a := range v {
		msg = msg + " -- " + utils.AutoToString(a)
	}
	return msg
}

func IsDebug() bool {
	if GetLevel() == logger.LevelDebug {
		return true
	}
	return runtime.IsDebug()
}

func Debug(content interface{}) {
	defaultLogger.Debug(content)
}

func Info(content interface{}) {
	defaultLogger.Info(content)
}

func Warn(content interface{}) {
	defaultLogger.Warn(content)
}

func Important(content interface{}) {
	defaultLogger.Important(content)
}

func Error(content interface{}) {
	defaultLogger.Error(content)
}

func Panic(content interface{}) {
	defaultLogger.Panic(content)
}

func Fatal(content interface{}) {
	defaultLogger.Fatal(content)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Importantf(format string, args ...interface{}) {
	defaultLogger.Importantf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func PrintDebug(args ...interface{}) {
	defaultLogger.Debug(fmtMsgForPrint(args...))
}

func PrintInfo(args ...interface{}) {
	defaultLogger.Info(fmtMsgForPrint(args...))
}

func PrintImportant(args ...interface{}) {
	defaultLogger.Important(fmtMsgForPrint(args...))
}

func PrintWarn(args ...interface{}) {
	defaultLogger.Warnf(fmtMsgForPrint(args...))
}

func PrintError(args ...interface{}) {
	defaultLogger.Error(fmtMsgForPrint(args...))
}

func PrintPanic(args ...interface{}) {
	defaultLogger.Panicf(fmtMsgForPrint(args...))
}

func PrintFatal(args ...interface{}) {
	defaultLogger.Fatal(fmtMsgForPrint(args...))
}

func WriteBySkipCall(level logger.Level, skipCall int, format string, args ...interface{}) {
	if err := defaultLogger.WriteBySkipCall(level, skipCall, append([]interface{}{format}, args...)...); err != nil {
		fmt.Println(err)
	}
}
