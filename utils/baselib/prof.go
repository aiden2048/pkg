package baselib

import (
	"os"
	"runtime/pprof"
	"time"
)

func ProfileToFile(file string, profileTime int) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	pprof.StartCPUProfile(f) // 开始cpu profile，结果写到文件f中
	time.Sleep(time.Duration(profileTime) * time.Second)
	pprof.StopCPUProfile() // 结束profile
	//pprof.WriteHeapProfile(f)


	return nil
}

func ProfileHeapMemoryToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	pprof.WriteHeapProfile(f) // 开始堆内存 profile，结果写到文件f中
	return nil
}

func PrintStackToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	pprof.Lookup("goroutine").WriteTo(f, 2) // 将当前函数栈，结果写到文件f中
	return nil
}
