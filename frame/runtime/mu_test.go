package runtime

import (
	"fmt"
	"sync"
	"testing"
)

func TestMulElemMuFactory_MakeOrGetSpecElemMu(t *testing.T) {
	factory := NewMulElemMuFactory()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mu := factory.MakeOrGetSpecElemMu("mu1")
			mu.Lock()
			fmt.Println("mu1", i)
			mu.Unlock()
		}(i)
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			wg.Done()
			mu := factory.MakeOrGetSpecElemMu("mu2")
			mu.Lock()
			fmt.Println("mu2", i)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
}
