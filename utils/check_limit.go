package utils

import (
	"fmt"
	"sync"
	"time"
)

type limit struct {
	t int64
	c int32
}

var localLimit map[string]*limit
var lm sync.Mutex

func init() {
	localLimit = make(map[string]*limit)

	go func() {
		timer := time.NewTicker(7200 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				lm.Lock()
				localLimit = make(map[string]*limit)
				lm.Unlock()
			}
		}
	}()
}

func CheckLimit30s(addr string, cmd string, now time.Time, lim int32) int32 {
	if lim <= 0 {
		return 0
	}
	key := fmt.Sprintf("%s.%s", addr, cmd)
	lm.Lock()
	defer lm.Unlock()
	t := now.Unix()
	if v, ok := localLimit[key]; ok {
		//判断一秒钟次数上限
		n := v.c - lim
		v.c++
		if t-v.t < int64(30) {
			if n >= 0 && n%10 == 0 {
				//重置当前时间, 保证一直在限制中
				v.t = t
				return v.c
			} else if n > 0 {
				//重置当前时间, 保证一直在限制中
				v.t = t
				return -1
			}

		} else {
			v.t = t
			v.c = 1
		}
	} else {
		l := &limit{t: t, c: 1}
		localLimit[key] = l
	}

	return 0
}
