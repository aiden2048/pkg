package frame

import (
	"sync"
	"time"
)

// global pool of *time.Timer's. can be used by multiple goroutines concurrently.
var globalTimerPool timerPool

type Timer struct {
	*time.Timer
	d time.Duration
}

func (t *Timer) Reset() {
	if t == nil || t.Timer == nil {
		return
	}
	t.Timer.Reset(t.d)
}

// timerPool provides GC-able pooling of *time.Timer's.
// can be used by multiple goroutines concurrently.
type timerPool struct {
	p sync.Pool
}

// Get returns a timer that completes after the given duration.
func (tp *timerPool) Get(d time.Duration) *Timer {
	if t, _ := tp.p.Get().(*Timer); t != nil {
		t.Timer.Reset(d)
		return t
	}
	t := &Timer{Timer: time.NewTimer(d), d: d}
	return t
}

// Put pools the given timer.
//
// There is no need to call t.Stop() before calling Put.
//
// Put will try to stop the timer before pooling. If the
// given timer already expired, Put will read the unreceived
// value if there is one.
func (tp *timerPool) Put(t *Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	tp.p.Put(t)
}

func GetTimer(d time.Duration) *Timer {

	return globalTimerPool.Get(d)
}
func PutTimer(t *Timer) {
	globalTimerPool.Put(t)
}
