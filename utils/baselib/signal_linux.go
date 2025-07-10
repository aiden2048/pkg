// +build !windows

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

func NewSignalHandler() *SignalHandler {
	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGUSR1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	prof := make(chan os.Signal, 1)
	signal.Notify(prof, syscall.SIGUSR2)
	unload := make(chan os.Signal, 1)
	signal.Notify(unload, syscall.SIGHUP)
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
