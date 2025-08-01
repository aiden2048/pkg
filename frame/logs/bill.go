package logs

import (
	"fmt"
	"sync"

	"github.com/aiden2048/pkg/frame/logs/logger"
	"github.com/aiden2048/pkg/utils"
)

var onBillFunc func(billName string)

func OnBill(fn func(billName string)) {
	onBillFunc = fn
}

func emitOnBill(billName string) {
	if onBillFunc != nil {
		onBillFunc(billName)
	}
}

var billLoggerFactory = &BillLoggerFactory{
	loggerMap: make(map[string]*logger.Logger),
}

type BillLoggerFactory struct {
	loggerMap map[string]*logger.Logger
	mu        sync.RWMutex
}

func (b *BillLoggerFactory) MustLogger(billName string) *logger.Logger {
	b.mu.RLock()
	billLogger, ok := billLoggerFactory.loggerMap[billName]
	if ok {
		b.mu.RUnlock()
		return billLogger
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	cfgLoader := MustDefaultCfgLoader()
	cfg := cfgLoader.GetConf()
	billLogger, err := InitFileLogger(SrvName+"-bill", cfg.File.BillLogDir, billName, 5, cfgLoader)
	if err != nil {
		panic(err)
	}
	billLoggerFactory.loggerMap[billName] = billLogger

	return billLogger
}

func Bill(billName string, format string, args ...interface{}) {
	billLoggerFactory.MustLogger(billName).Importantf(format, args...)
	emitOnBill(billName)
}

func PrintBill(billName string, args ...interface{}) {
	var msg string
	for _, arg := range args {
		msg = msg + " -- " + utils.AutoToString(arg)
	}
	billLoggerFactory.MustLogger(billName).Important(msg)
	emitOnBill(billName)
}

func BillBySkipCall(skipCall int, billName string, format string, args ...interface{}) {
	if err := billLoggerFactory.MustLogger(billName).WriteBySkipCall(logger.LevelImportant, skipCall, append([]interface{}{format}, args...)...); err != nil {
		fmt.Println(err)
	}
	emitOnBill(billName)
}

func PrintBillBySkipCall(skipCall int, billName string, args ...interface{}) {
	var msg string
	for _, arg := range args {
		msg = msg + " -- " + utils.AutoToString(arg)
	}
	if err := billLoggerFactory.MustLogger(billName).WriteBySkipCall(logger.LevelImportant, skipCall, msg); err != nil {
		fmt.Println(err)
	}
	emitOnBill(billName)
}
