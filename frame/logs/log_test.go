package logs

import (
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aiden2048/pkg/frame/logs/logger"
)

func TestCfgWrapper(t *testing.T) {
	type LogCfgWrapper struct {
		File       logger.FileLogConf
		GameFrame  logger.GameFrameLogConf //gameFrame专用
		AlertLevel string
		LogRpcx    bool

		logger.LogConf
	}

	type LogCfgWrapper2 struct {
		logger.LogConf
	}

	cfg2 := &logger.LogConf{}
	_, err := toml.DecodeFile("./test/log.toml", cfg2)
	if err != nil {
		panic(err)
	}
	t.Log(cfg2)

	cfg := &LogCfgWrapper{}
	_, err = toml.DecodeFile("./test/log.toml", cfg)
	if err != nil {
		panic(err)
	}
	t.Log(cfg)

	cfg3 := &LogCfgWrapper2{}
	_, err = toml.DecodeFile("./test/log.toml", cfg3)
	if err != nil {
		panic(err)
	}
	t.Log(cfg)
}

func TestBillLogger(t *testing.T) {
	err := InitDefaultCfgLoader("logV2Test", "./test/log.toml", &logger.LogConf{
		File: logger.FileLogConf{
			MaxFileSizeBytes: 100,
			Level:            "DBG",
			DefaultLogDir:    "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/log",
			BillLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/bill",
			StatLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/stat",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	Bill("bill-tag", "i am bumble bee, %s", "yep")
	PrintBill("bill-tag2", "i am bill 2", "yeah")
	OnExit()
}

func TestInitDefaultLogger(t *testing.T) {
	err := InitDefaultCfgLoader("logV2Test", "./test/log.toml", &logger.LogConf{
		File: logger.FileLogConf{
			MaxFileSizeBytes:            100,
			LogInfoBeforeFileSizeBytes:  -1,
			LogDebugBeforeFileSizeBytes: -1,
			DebugMsgMaxLen:              15,
			Level:                       "DBG",
			DefaultLogDir:               "/var/work/gitlabproj/afun-golang-oversea/dist.dev/logs/v2/log",
			BillLogDir:                  "/var/work/gitlabproj/afun-golang-oversea/dist.dev/logs/v2/bill",
			StatLogDir:                  "/var/work/gitlabproj/afun-golang-oversea/dist.dev/logs/v2/stat",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = InitDefaultLogger(nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100000; i++ {
		Infof("hello infof! my name is:bumble bee %d\r\n", i)
	}
	Debug("hello debug! my name is bumble bee")
	Debugf("hello debugf! my name is:%s", "bumble bee")
	PrintDebug("hello PrintDebug", "my name is bumble bee")
	Info("hello info! my name is bumble bee")
	Infof("hello infof! my name is:%s", "bumble bee")
	PrintInfo("hello PrintInfo", "my name is bumble bee")
	Important("hello important! my name is bumble bee")
	Importantf("hello importantf! my name is:%s", "bumble bee")
	PrintImportant("hello PrintImportant", "my name is bumble bee")
	Warn("hello warn! my name is bumble bee")
	Warnf("hello warnf! my name is:%s", "bumble bee")
	PrintWarn("hello PrintWarn", "my name is bumble bee")
	Error("hello error! my name is bumble bee\r\n")
	Errorf("hello errorf! my name is:%s", "bumble bee\r\n")
	//PrintError("hello PrintError", "my name is bumble bee")
	//Panic("hello panic! my name is bumble bee")
	//Panicf("hello panicf! my name is:%s", "bumble bee")
	//PrintPanic("hello PrintPanic", "my name is bumble bee")
	//Fatal("hello fatal! my name is bumble bee")
	//Fatalf("hello fatalf! my name is:%s", "bumble bee")
	//PrintFatal("hello PrintFatal", "my name is bumble bee")
	OnExit()
}

func TestInitDefaultCfgLoader(t *testing.T) {
	err := InitDefaultCfgLoader("logV2Test", "./test/log.toml", &logger.LogConf{
		File: logger.FileLogConf{
			MaxFileSizeBytes: 100,
			Level:            "DBG",
			DefaultLogDir:    "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/log",
			BillLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/bill",
			StatLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/stat",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 20)
}

func TestInitDefaultCfgLoaderWithEmptyCfgFile(t *testing.T) {
	err := InitDefaultCfgLoader("logV2Test", "", &logger.LogConf{
		File: logger.FileLogConf{
			MaxFileSizeBytes: 100,
			Level:            "DBG",
			DefaultLogDir:    "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/log",
			BillLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/bill",
			StatLogDir:       "../frame/afun-golang-pkg/afun-golang-oversea/dist.dev/logs/v2/stat",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 20)
}
