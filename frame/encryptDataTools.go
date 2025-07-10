package frame

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/aiden2048/pkg/frame/logs"
)

/**
post数据加解密
*/

/**
客户端上报数据解密，可以做web和APP区分
AES_KEY_FORROUTE
AesPostKey
*/

// func DecryptPostData(msg []byte) (int, []byte) {

// 	if bytes.HasPrefix(msg, []byte("s:")) {
// 		return 0, DecryptClientBytes(0, msg[2:])
// 	} else if bytes.HasPrefix(msg, []byte("s1:")) {
// 		return 1, DecryptClientBytes(1, msg[3:])
// 	}
// 	return -1, msg
// }

/*
*
http/ws下发数据加密，可以做web和APP区分
AES_KEY_FORCLIENT
AesClientKey
*/
func EncryptResponse(s []byte) []byte {
	return EncryptForClientBytesNoPadding(s)
}

/*
*
http/ws下发数据加密，可以做web和APP区分
AES_KEY_FORCLIENT
AesClientKey
*/
func EncryptResponseStr(s string, key ...string) string {
	return EncryptForClientNoPadding(s, key...)
}

/*
*
解密客户端上传手机号码
AES_KEY_FORCLIENT
AesClientKey
*/
func DecodePhoneCode(phone string) string {
	if strings.HasPrefix(phone, "r:") {
		return RSADecrypt(phone[2:])
	} else if strings.HasPrefix(phone, "p:") {
		return phone[2:]
	} else {
		return DecryptForClientPadding(phone)
	}

}

/*
*
解密客户端上传邮箱
AES_KEY_FORCLIENT
AesClientKey
如果不带，则默认使用原始string
*/
func DecodeEmail(email string) string {
	if strings.HasPrefix(email, "r:") {
		return RSADecrypt(email[2:])
	} else if strings.HasPrefix(email, "p:") {
		return email[2:]
	} else {
		return email
	}
}

/*
*
解密客户端数据
AES_KEY_FORCLIENT
AesClientKey
*/
func DecodeStrNoPadding(str string, key ...string) string {
	return DecryptForClientNoPadding(str, key...)
}

// RSA解密
// cipherText 需要解密的base64数字
func RSADecrypt(base64Txt string) string {
	cipherText, decodeErr := base64.StdEncoding.DecodeString(base64Txt)
	if decodeErr != nil {
		logs.Errorf("rsa decide error:%+v, base64Txt:%+v", decodeErr, base64Txt)
		return ""
	}
	block, _ := pem.Decode(RSA_PHONE_PRI_KEY)
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logs.Errorf("rsa decrypt error:%+v, base64Txt:%+v", decodeErr, base64Txt)
		return ""
	}
	//对密文进行解密
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	//返回明文
	return string(plainText)
}

// RSA加密
// plainText 要加密的数据
// 返回base64string数据
func RSAEncrypt(plainText []byte) string {
	block, _ := pem.Decode(RSA_PHONE_PUB_KEY)
	//x509解码

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	//类型断言
	publicKey := publicKeyInterface.(*rsa.PublicKey)
	//对明文进行加密
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plainText)
	if err != nil {
		panic(err)
	}
	cipherText := base64.StdEncoding.EncodeToString(cipherBytes)
	//返回密文
	return cipherText
}
