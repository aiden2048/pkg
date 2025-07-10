package runtime

import (
	"fmt"
	"sync"
	"testing"
)

func TestGetGpid(t *testing.T) {
	gid := GetGpid()
	fmt.Println(gid)

	var mu sync.Mutex
	m := map[int64]int{}
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			gid := GetGpid()
			fmt.Println("gid 1:", gid)
			mu.Lock()
			defer mu.Unlock()
			c := m[gid]
			c++
			m[gid] = c
		}()
	}
	wg.Wait()
	fmt.Println(m)
}
