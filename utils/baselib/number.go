package baselib

import (
	"reflect"
	"strconv"
)

func ForceToInt(v interface{}) (int ) {
	switch v.(type) {
	case string:
		i, _ := strconv.Atoi(v.(string))
		return int(i)
	case float32, float64:
		return int(reflect.ValueOf(v).Float())
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint())
	}
	return 0
}