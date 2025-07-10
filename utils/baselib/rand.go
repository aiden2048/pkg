package baselib

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 获取随机数，返回范围：[0, 无穷大)
func GetRandInt() int {
	return rand.Int()
}

// 获取随机数，返回范围：[0, n-1]
func GetRandIntn(n int) int {
	if n <= 0 {
		return 0
	} else {
		return rand.Intn(n)
	}
}

// 获取随机数， 返回范围：[min, max-1]
func GetRandIntRange(min, max int) int {
	tmpMin, tmpMax := min, max
	if max < min {
		tmpMin, tmpMax = max, min
	}

	if tmpMax <= 0 {
		return 0
	} else if tmpMin < 0 {
		tmpMin = 0
	}

	return GetRandIntn(tmpMax-tmpMin) + tmpMin
}

// 获取uin
func GetRandUin() uint64 {
	return uint64(rand.Int63n(4200000000) + 10000)
}

// 获取指定长度的随机字节数组，每个字节取值范围在0x00-0xff
func GetRandBytes(length int) []byte {
	if length <= 0 {
		return nil
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = byte(GetRandInt() % 256)
	}
	return b
}

const alphaCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphaCharsetLen = len(alphaCharset)

// 获取指定长度的随机字符串，字符串内容由大小写字母构成
func GetRandAlphaStr(length int) string {
	if length <= 0 {
		return ""
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = alphaCharset[GetRandIntn(alphaCharsetLen)]
	}
	return string(b)
}

const lowerAlphaCharset = "abcdefghijklmnopqrstuvwxyz"
const lowerAlphaCharsetLen = len(lowerAlphaCharset)
//获取全部为小写的随机字符串
func GetRandLowerAlphaStr(length int) string {
	if length <= 0 {
		return ""
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = lowerAlphaCharset[GetRandIntn(lowerAlphaCharsetLen)]
	}
	return string(b)
}
