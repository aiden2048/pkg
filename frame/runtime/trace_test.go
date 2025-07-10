package runtime

import (
	"sync"
	"testing"
)

func TestGenTraceStrId(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			t.Log("trace id ", i, GenTraceStrId())
		}(i)
	}
	wg.Wait()
}
