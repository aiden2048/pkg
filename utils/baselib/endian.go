package baselib

import(
	"bytes"
	"encoding/binary"
	"errors"
	//"fmt"
	"unsafe"
)

const INT_SIZE int = int(unsafe.Sizeof(0))

// 判断本机字节序是否为大端序
func IsSystemBigEndian() bool {
	var i int = 0x1
	bs := (*[INT_SIZE]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		return true
	} else {
		return false
	}
}

// 大端序转换为小端序
// NOTE: 目前只支持golang基本数据类型
func BigEndianToLittleEndian(v interface{}) (interface{}, error) {
	switch v.(type) {
	case uint64 : {
		data, ok := v.(uint64)
		if !ok {
			return nil, errors.New("Not uint64")
		} else {
			// 声明返回值类型
			var rtn uint64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}


	case int64 : {
		data, ok := v.(int64)
		if !ok {
			return nil, errors.New("Not int64")
		} else {
			// 声明返回值类型
			var rtn int64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}


	case uint32 : {
		data, ok := v.(uint32)
		if !ok {
			return nil, errors.New("Not uint32")
		} else {
			// 声明返回值类型
			var rtn uint32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case int32 : {
		data, ok := v.(int32)
		if !ok {
			return nil, errors.New("Not int32")
		} else {
			// 声明返回值类型
			var rtn int32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case uint16 : {
		data, ok := v.(uint16)
		if !ok {
			return nil, errors.New("Not uint16")
		} else {
			// 声明返回值类型
			var rtn uint16

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case int16 : {
		data, ok := v.(int16)
		if !ok {
			return nil, errors.New("Not int16")
		} else {
			// 声明返回值类型
			var rtn int16

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case uint8 : {
		data, ok := v.(uint8)
		if !ok {
			return nil, errors.New("Not uint8")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case int8 : {
		data, ok := v.(int8)
		if !ok {
			return nil, errors.New("Not uint8")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case bool : {
		data, ok := v.(bool)
		if !ok {
			return nil, errors.New("Not bool")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case float32 : {
		data, ok := v.(float32)
		if !ok {
			return nil, errors.New("Not float32")
		} else {
			// 声明返回值类型
			var rtn float32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case float64 : {
		data, ok := v.(float64)
		if !ok {
			return nil, errors.New("Not float32")
		} else {
			// 声明返回值类型
			var rtn float64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.BigEndian, data)
			binary.Read(buffer, binary.LittleEndian, &rtn)

			return rtn, nil
		}
	}

	case int, uint : {
		return v, errors.New("The byte size of int and uint in golang depend on your system!")
	}

	case *int8, *int16, *int32, *int64, *uint8, *uint16, *uint32, *uint64, *bool : {
		return v, errors.New("Function do not support pointer type right now")
	}

	default: {
		// 其它数据类型
		return v, nil
	}

	}
}

// 小端序转换为大端序
func LittleEndianToBigEndian(v interface{}) (interface{}, error) {
	switch v.(type) {
	case uint64 : {
		data, ok := v.(uint64)
		if !ok {
			return nil, errors.New("Not uint64")
		} else {
			// 声明返回值类型
			var rtn uint64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case int64 : {
		data, ok := v.(int64)
		if !ok {
			return nil, errors.New("Not int64")
		} else {
			// 声明返回值类型
			var rtn int64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case uint32 : {
		data, ok := v.(uint32)
		if !ok {
			return nil, errors.New("Not uint32")
		} else {
			// 声明返回值类型
			var rtn uint32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case int32 : {
		data, ok := v.(int32)
		if !ok {
			return nil, errors.New("Not int32")
		} else {
			// 声明返回值类型
			var rtn int32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case uint16 : {
		data, ok := v.(uint16)
		if !ok {
			return nil, errors.New("Not uint16")
		} else {
			// 声明返回值类型
			var rtn uint16

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case int16 : {
		data, ok := v.(int16)
		if !ok {
			return nil, errors.New("Not int16")
		} else {
			// 声明返回值类型
			var rtn int16

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case uint8 : {
		data, ok := v.(uint8)
		if !ok {
			return nil, errors.New("Not uint8")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case int8 : {
		data, ok := v.(int8)
		if !ok {
			return nil, errors.New("Not uint8")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case bool : {
		data, ok := v.(bool)
		if !ok {
			return nil, errors.New("Not bool")
		} else {
			// 1字节无须字节序转换
			return data, nil
		}
	}

	case float32 : {
		data, ok := v.(float32)
		if !ok {
			return nil, errors.New("Not float32")
		} else {
			// 声明返回值类型
			var rtn float32

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case float64 : {
		data, ok := v.(float64)
		if !ok {
			return nil, errors.New("Not float32")
		} else {
			// 声明返回值类型
			var rtn float64

			buffer := new(bytes.Buffer)
			binary.Write(buffer, binary.LittleEndian, data)
			binary.Read(buffer, binary.BigEndian, &rtn)

			return rtn, nil
		}
	}

	case int, uint : {
		return v, errors.New("The byte size of int and uint in golang depend on your system!")
	}

	case *int8, *int16, *int32, *int64, *uint8, *uint16, *uint32, *uint64, *bool : {
		return v, errors.New("Function do not support pointer type right now")
	}

	default: {
		// 其它数据类型
		return v, nil
	}

	}
}

