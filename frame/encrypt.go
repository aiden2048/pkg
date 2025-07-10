package frame

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"

	"github.com/aiden2048/pkg/frame/logs"

	"strconv"
	"strings"
)

var decimalDict = "fgHicdEyzKLMnOpqjUvWXAbRst"

const (
	/**
	服务端内部加解密，不需要跟任何人同步
	*/
	AES_KEY = "1AiOOIBOBU2leghSWsRR50fVYEstyhM="
	AES_IV  = "hceaV12V/+6K9ih6"

	/**
	对接外部钱包和三方游戏token验证使用的加密，不需要把秘钥给到任何人
	*/
	AES_KEY_FORWALLET = "TUTOZ3nsL2PD3UcW1iMHON2loYpsK545" // 钱包
	AES_IV_FORWALLET  = "nTcpqxsTCjrYIQr1"                 // 钱包

	/**
	下发数据给客户端加密，需要同步给客户端 : 下发数据、手机号码加解密、data文件加密
	*/
	AES_KEY_FORCLIENT = "tx6pWYrj68vo1HctJZaC5D7j9pPAzfo=" // 客户端
	AES_IV_FORCLIENT  = "RDlhO/GW+OWQULFm"                 // 客户端

	/**
	客户端POST数据加解密，需要同步给客户端: 客户端POST数据
	*/
	AES_KEY_FORROUTE = "k8453279789jhlouyuf2098767787277" // 路由
	AES_IV_FORROUTE  = "kjuf20987hlouy67"                 // 路由
)

// 手机号加密key
var RSA_PHONE_PRI_KEY = []byte(`
`)
var RSA_PHONE_PUB_KEY = []byte(`

`)

func getAesClientKey() string {
	if GetAesClientKey() != "" {
		return GetAesClientKey()
	}
	return AES_KEY_FORCLIENT
}

func getAesClientIv() string {
	if GetAesClientIv() != "" {
		return GetAesClientIv()
	}
	return AES_IV_FORCLIENT
}

func getAesPostKey(ikey int) string {
	if GetAesPostKey(ikey) != "" {
		return GetAesPostKey(ikey)
	}
	return AES_KEY_FORROUTE
}

func getAesPostIv(ikey int) string {
	if GetAesPostIv(ikey) != "" {
		return GetAesPostIv(ikey)
	}
	return AES_IV_FORROUTE
}

// 10进制转任意进制
func EncodeAppid(appid int32) string {
	new_num_str := ""
	var remainder int32
	var remainder_string string
	nn := int32(26)
	num := appid
	for num != 0 {
		remainder = num % nn
		if nn > remainder && remainder >= 0 {
			remainder_string = decimalDict[remainder : remainder+1]
		} else {
			remainder_string = strconv.Itoa(int(remainder))
		}
		new_num_str = remainder_string + new_num_str
		num = num / nn
	}
	return new_num_str
}
func DecodeAppid(appstr string) int32 {
	appid := 0
	ll := len(appstr)
	dl := len(decimalDict)
	for nn := 0; nn < ll; nn++ {
		n := strings.Index(decimalDict, appstr[nn:nn+1])
		appid = appid*dl + n

	}
	return int32(appid)
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	if length <= 0 {
		return origData
	}
	unPadding := int(origData[length-1])
	if unPadding <= 0 {
		unPadding = 1
	}
	if length-unPadding <= 0 {
		return origData
	}
	return origData[:(length - unPadding)]
}
func pKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	// 填充
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// 加密
func EncryptBytesByKey(s []byte, aesKey, aesIV string, isPadding bool) []byte {
	//logs.LogDebug("EncryptBytesByKey s:%s, key:%s, iv:%s", string(s), aesKey, aesIV)
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		logs.LogDebug("NewCipher err :%s", err.Error())
		return nil
	}
	cfb := cipher.NewCBCEncrypter(block, []byte(aesIV))

	var plainBytes []byte
	if !isPadding {
		padding := cfb.BlockSize() - len(s)%cfb.BlockSize()
		padding = padding % cfb.BlockSize()
		padtext := bytes.Repeat([]byte{' '}, padding) // make([]byte, padding)
		//fmt.Printf("padding: %d", padding)
		//for i := 0; i < padding; i++ {
		//	padtext[i] = ' '
		//}

		plainBytes = append(s, padtext...)
	} else {

		plainBytes = pKCS7Padding(s, cfb.BlockSize())

	}

	enBytes := make([]byte, len(plainBytes))
	cfb.CryptBlocks(enBytes, plainBytes)
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(enBytes)))
	base64.StdEncoding.Encode(buf, enBytes)
	return buf
}

// 加密
func EncryptStringByKey(str string, aesKey, aesIV string, isPadding bool) string {
	return string(EncryptBytesByKey([]byte(str), aesKey, aesIV, isPadding))
}

// 解密
func DecryptBytesByKey(s []byte, aesKey, aesIV string, isPadding bool) []byte {
	//	logs.LogDebug("DecryptBytesByKey s:%s, key:%s, iv:%s", string(s), aesKey, aesIV)
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	n, err := base64.StdEncoding.Decode(dbuf, []byte(s))
	if n < aes.BlockSize || n%aes.BlockSize != 0 {
		return nil
	}
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil
	}
	cfb := cipher.NewCBCDecrypter(block, []byte(aesIV))
	plaintxt := make([]byte, n)
	cfb.CryptBlocks(plaintxt, dbuf[:n])

	var plainBytes []byte
	if !isPadding {
		plainBytes = []byte(strings.TrimRight(string(plaintxt), " "))
	} else {
		plainBytes = pKCS7UnPadding(plaintxt)
	}

	return plainBytes
}

// 解密
func DecryptStringByKey(s string, aesKey, aesIV string, isPadding bool) string {
	return string(DecryptBytesByKey([]byte(s), aesKey, aesIV, isPadding))
}

func EncryptWithKeyNoPadding(str string, aesKey, aesIV string) string {
	return EncryptStringByKey(str, aesKey, aesIV, false)
}
func DecryptWithKeyNoPadding(str string, aesKey, aesIV string) string {
	return DecryptStringByKey(str, aesKey, aesIV, false)
}

// 加密, 服务端内部使用
func Encrypt(str string) string {
	return EncryptStringByKey(str, AES_KEY, AES_IV, false)
}
func Decrypt(s string) string {
	return DecryptStringByKey(s, AES_KEY, AES_IV, false)
}

// 客户端信息加密
func EncryptForClientNoPadding(str string, key ...string) string {
	k := getAesClientKey()
	if len(key) > 0 && key[0] != "" {
		k = key[0]
	}
	iv := getAesClientIv()
	if len(key) > 1 && key[1] != "" {
		iv = key[1]
	}
	return EncryptStringByKey(str, k, iv, false)
}

func EncryptForClientBytesNoPadding(s []byte) []byte {
	return EncryptBytesByKey(s, getAesClientKey(), getAesClientIv(), false)
}

// 客户端信息解密
func DecryptForClientNoPadding(str string, key ...string) string {
	k := getAesClientKey()
	if len(key) > 0 && key[0] != "" {
		k = key[0]
	}
	iv := getAesClientIv()
	if len(key) > 1 && key[1] != "" {
		iv = key[1]
	}
	return DecryptStringByKey(str, k, iv, false)
}

func EncryptForClientPadding(str string) string {
	return EncryptStringByKey(str, getAesClientKey(), getAesClientIv(), true)
}

func DecryptForClientPadding(str string) string {
	return DecryptStringByKey(str, getAesClientKey(), getAesClientIv(), true)
}

// 注意, 这里用的是另外一个key
func EncyptForClientNoPadding(ikey int, str string) string {
	return EncryptStringByKey(str, getAesPostKey(ikey), getAesPostIv(ikey), true)
}

// 解密客户端加密的数据
// 注意, 这里用的是另外一个key
func DecryptClientBytes(ikey int, msg []byte) []byte {
	return DecryptBytesByKey(msg, getAesPostKey(ikey), getAesPostIv(ikey), true)
}

func md5Sum(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
func ClientChecksum(checkKey, str string) string {
	str = md5Sum(str)
	str = str[:len(str)/2]
	checkKey = md5Sum(checkKey)
	checkKey = checkKey[len(checkKey)/2:]
	data := md5Sum(str + checkKey)
	data = data[5:15]

	return data
}

//----------------------------------------------------------------------------------
//加密
// func EncryptBufferByKey(s []byte, aesKey, aesIV string, isPadding bool) []byte {
//
// 	block, err := aes.NewCipher([]byte(aesKey))
// 	if err != nil {
// 		logs.LogDebug("NewCipher err :%s", err.Error())
// 		return nil
// 	}
// 	cfb := cipher.NewCBCEncrypter(block, []byte(aesIV))
//
// 	var plainBytes []byte
// 	if !isPadding {
// 		padding := cfb.BlockSize() - len(s)%cfb.BlockSize()
// 		padding = padding % cfb.BlockSize()
// 		padtext := bytes.Repeat([]byte{' '}, padding) // make([]byte, padding)
// 		//fmt.Printf("padding: %d", padding)
// 		//for i := 0; i < padding; i++ {
// 		//	padtext[i] = ' '
// 		//}
//
// 		plainBytes = append(s, padtext...)
// 	} else {
//
// 		plainBytes = pKCS7Padding(s, cfb.BlockSize())
//
// 	}
//
// 	enBytes := make([]byte, len(plainBytes))
// 	cfb.CryptBlocks(enBytes, plainBytes)
// 	return enBytes
// }
//
// // 解密
// func DecryptBufferByKey(sbuff []byte, aesKey, aesIV string, isPadding bool) []byte {
//
// 	n := len(sbuff)
// 	if n < aes.BlockSize || n%aes.BlockSize != 0 {
// 		return nil
// 	}
// 	block, err := aes.NewCipher([]byte(aesKey))
// 	if err != nil {
// 		return nil
// 	}
// 	cfb := cipher.NewCBCDecrypter(block, []byte(aesIV))
// 	plaintxt := make([]byte, n)
// 	cfb.CryptBlocks(plaintxt, sbuff[:n])
//
// 	var plainBytes []byte
// 	if !isPadding {
// 		plainBytes = []byte(strings.TrimRight(string(plaintxt), " "))
// 	} else {
// 		plainBytes = pKCS7UnPadding(plaintxt)
// 	}
//
// 	return plainBytes
// }
//
// func EncryptClientByffer(s []byte) []byte {
// 	return EncryptBufferByKey(s, getAesClientKey(), getAesClientIv(), false)
// }
//
// func DecryptClientByffer(s []byte) []byte {
// 	return DecryptBufferByKey(s, getAesClientKey(), getAesClientIv(), false)
// }
