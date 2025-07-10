// +build darwin

// 系统信号处理基础库
package baselib

import (
	"os"
	"os/signal"
	"syscall"
)

type SignalHandler struct {
	stop   chan os.Signal
	reload chan os.Signal
	prof   chan os.Signal
	unload chan os.Signal
}

//信号(Signal)是Linux, 类Unix和其它POSIX兼容的操作系统中用来进程间通讯的一种方式。一个信号就是一个异步的通知，发送给某个进程，或者同进程的某个线程，告诉它们某个事件发生了。
//当信号发送到某个进程中时，操作系统会中断该进程的正常流程，并进入相应的信号处理函数执行操作，完成后再回到中断的地方继续执行。
//如果目标进程先前注册了某个信号的处理程序(signal handler),则此处理程序会被调用，否则缺省的处理程序被调用。
func NewSignalHandler() *SignalHandler {
	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGUSR1) //用户保留
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL) //SIGINT 用户发送INTR字符(Ctrl+C)触发 SIGTERM 结束程序(可以被捕获、阻塞或忽略) SIGKILL 无条件结束程序(不能被捕获、阻塞或忽略)
	prof := make(chan os.Signal, 1)
	signal.Notify(prof, syscall.SIGUSR2) //用户保留
	unload := make(chan os.Signal, 1)
	signal.Notify(unload, syscall.SIGHUP) //终端控制进程结束(终端连接断开)
	sig := &SignalHandler{stop: stop, reload: reload, prof: prof, unload: unload}
	return sig
}

func (sig *SignalHandler) ReloadSignal() <-chan os.Signal {
	return sig.reload
}

func (sig *SignalHandler) StopSignal() <-chan os.Signal {
	return sig.stop
}

func (sig *SignalHandler) ProfSignal() <-chan os.Signal {
	return sig.prof
}

func (sig *SignalHandler) UnloadSignal() <-chan os.Signal {
	return sig.unload
}

func RegisterStopFunc(f func()) {
	sig := NewSignalHandler()
	go func() {
		for {
			select {
			case <-sig.StopSignal():
				f()
			}
		}
	}()
}

func SendStopSignal() {
	sendSignal(syscall.SIGTERM)
}

func SendReloadSignal() {
	sendSignal(syscall.SIGUSR1)
}

func sendSignal(signal syscall.Signal) {
	syscall.Kill(syscall.Getpid(), signal)
}
