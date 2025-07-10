package runtime

import (
	"fmt"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	GetOrCreateTrace()
	fmt.Println("trace:", GetTrace().TraceId)
	Go(func() {
		fmt.Println("runtime.Go trace:", GetTrace().TraceId)
	})
	go func() {
		fmt.Println("go trace:", GetTrace().TraceId)
	}()
	time.Sleep(time.Second)
}

func TestNewLimitedConcurErrGrp(t *testing.T) {
	grp := NewLimitedConcurErrGrp(10, true)
	grp.Go(func() error {
		time.Sleep(time.Second)
		fmt.Println("1")
		return nil
	})
	grp.Go(func() error {
		time.Sleep(time.Second * 2)
		fmt.Println("2")
		return nil
	})
	if err := grp.Wait(); err != nil {
		panic(err)
	}
}
