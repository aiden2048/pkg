package frame

import (
	"log"
	"os"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils/baselib"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
)

var g_server_is_stop = false

// 根据bin目录下touch文件, 在reload的时候检查各个状态
func CheckSysStatus() error {
	if fileutil.Exist("clear") {
		os.Remove("clear")
		log.Printf("Server want to clear rpcx from reload\n")
		logs.Trace("Server want to clear rpcx from reload\n")

		ReclearRpcxClient()
	}
	if fileutil.Exist("unload") {
		os.Remove("unload")
		log.Printf("Server want to unload from reload\n")
		logs.Trace("Server want to unload from reload\n")
		//baselib.Stop()
		ManualStop()
		//time.Sleep(2 * time.Second)
	}

	if fileutil.Exist("stop") {
		os.Remove("stop")
		log.Printf("Server want to stop from reload\n")
		logs.Trace("Server want to stop from reload\n")
		baselib.Stop()

		ManualStop()
		time.Sleep(time.Second)
		stopRpcxServer()
		CloseServer()
		g_server_is_stop = true

	}

	return nil
}
