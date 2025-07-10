package utils

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

var g_MuLocker sync.Mutex
var g_FileCache map[string][]byte

func LoadJsonFileUseCache(path string) (ret_buffer []byte, err error) {
	g_MuLocker.Lock()
	defer g_MuLocker.Unlock()
	hasCache := false
	if g_FileCache == nil {
		g_FileCache = make(map[string][]byte)
	} else {
		ret_buffer, hasCache = g_FileCache[path]
	}
	if !hasCache {
		ret_buffer, err = LoadJsonFile(path)
		if err != nil {
			return
		}
		g_FileCache[path] = ret_buffer
	}

	return

}

func ClearFileCache() {
	g_MuLocker.Lock()
	defer g_MuLocker.Unlock()
	g_FileCache = make(map[string][]byte)
}
func LoadJsonFile(path string) (ret_buffer []byte, err error) {
	config_file, err := os.Open(path)
	if err != nil {
		//logs.LogError("Failed to open config file '%s': %s\n", path, err)
		errstr := fmt.Sprintf("Failed to open config file '%s': %s\n", path, err)
		return nil, errors.New(errstr)
	}

	fi, _ := config_file.Stat()
	if size := fi.Size(); size > (1024 * 1024 * 10) {
		errstr := fmt.Sprintf("config file (%d) size exceeds reasonable limit (%d) - aborting", size, 1024*1024*10)
		return nil, errors.New(errstr)
	}

	if fi.Size() == 0 {
		//logs.LogError("config file (%q) is empty, skipping", path)
		errstr := fmt.Sprintf("config file (%q) is empty, skipping", path)
		return nil, errors.New(errstr)
	}

	buffer := make([]byte, fi.Size())
	_, err = config_file.Read(buffer)

	//buffer, err = StripComments(buffer) //去掉注释
	if err != nil {
		//	logs.LogError("Failed to strip comments from json: %s\n", err.Error())
		errstr := fmt.Sprintf("Failed to strip comments from json: %s\n", err.Error())
		return nil, errors.New(errstr)
	}

	return buffer, nil
	//buffer = []byte(os.ExpandEnv(string(buffer))) //特殊
}
