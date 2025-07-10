package frame

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"time"
)

/**
post数据加解密
*/

func DecryptPostData(msg string) ([]byte, error) {
	if len(msg) < 32 {
		return nil, errors.New("DecryptPostData")
	}
	msg = msg[0:len(msg)-32] + msg[len(msg)-16:]
	msg = msg[16:]
	return base64.StdEncoding.DecodeString(msg)
}
func EncryptPostData(msg []byte) string {
	encoded := base64.StdEncoding.EncodeToString(msg)
	str1 := getRandomString(16)
	encoded = str1 + encoded
	str2 := getRandomString(16)
	encoded = encoded[0:len(encoded)-16] + str2 + encoded[len(encoded)-16:]
	return encoded
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 获取指定长度的随机字符串
func getRandomString(length int) string {
	// 创建一个切片来存储随机字符
	result := make([]byte, length)

	// 使用当前时间的 Unix 时间戳作为种子，确保每次随机生成不同
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 随机从字母表中选择字符
	for i := 0; i < length; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}

	// 将字节切片转换成字符串并返回
	return string(result)
}
