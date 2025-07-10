package strings

import(
    "fmt"
    "strconv"
)

// byte转换成C_string（以'\0'作为字符串结束的标志）
func ByteToCString(p []byte) string {
    for i := 0; i < len(p); i++ {
        if p[i] == 0 {
            return string(p[0:i])
        }
    }
    return string(p)
}

// 计算C_string风格字符串的长度（以'\0'作为字符串结束的标志）
func CStrLen(p []byte) int {
    for i := 0; i < len(p); i++ {
        if p[i] == 0 {
            return i
        }
    }
    return len(p)
}

// 二进制转为可读的字符串，eg. []byte{123,77} => "7B4D" 
func Bin2Str(p []byte) string {
    if len(p) == 0 {
        return ""
    }
    
    s := ""
    for i:=0; i<len(p); i++ {
        s += fmt.Sprintf("%02x", p[i])
    }
    return s
}

// 可读字符串转为二进制，eg. "7B4D" => []byte{123,77} 
func Str2Bin(s string) []byte {
    buf := []byte{}
    
    for i:=0; i<len(s); i+=2 {
        sNum := ""
        if i+1 >= len(s) {
            sNum = string(s[i]) + "0"
        } else {
            sNum = s[i:i+2]
        }
        if v,err := strconv.ParseUint(sNum, 16, 8); err == nil {
            buf = append(buf, uint8(v))
        } else {
            buf = append(buf, uint8(0))
        }
    }
    
    return buf
}

// 是否是16进制字符串
// 判断标准是：（1）只能有 0-9, A-F, a-f；（2）字母要么全大写，要么全小写
func IsHexStr(src string) bool {
    if len(src) <= 0 || len(src) % 2 != 0 {
        return false
    }

	hasUp := false
	hasLower := false
	for _,c := range src {
		if !( (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') ) {
			return false
		}

		if !hasUp {
			hasUp = (c >= 'A' && c <= 'F')
		}
		if !hasLower {
			hasLower = (c >= 'a' && c <= 'f')
		}
        
		if hasUp && hasLower {
			return false
		}
	}
	return true
}

// 16进制字符串=>10进制数（eg. "1F"  => 1*16+15 = 31）
// 16进制字符串中不含负号
func Hex2Int(src string)  (uint64, error) {
	return strconv.ParseUint(src, 16, 64)
}


