package baselib

import (
	"sync"
)

var pool = &sync.Pool{New: func() interface{} {
	return []byte{}
}}

func GetTempBuffer() []byte {
	return pool.Get().([]byte)
}

func FreeTempBuffer(data []byte) {
	pool.Put(data)
}
