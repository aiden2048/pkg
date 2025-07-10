package runtime

import (
	"github.com/petermattis/goid"
)

func GetGpid() int64 {
	return goid.Get()
}
