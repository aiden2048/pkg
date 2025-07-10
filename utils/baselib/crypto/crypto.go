package crypto

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

// md5 算法：返回结果的原始buffer
func Md5(data []byte) []byte {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	return md5Ctx.Sum(nil)
}

// md5 算法：返回结果的16进制字符串
func Md5Str(data []byte) string {
	return hex.EncodeToString(Md5(data))
}

// sha1 算法：返回结果的原始buffer
func Sha1(data []byte) []byte {
	h := sha1.New()
	return h.Sum(data)
}

func Sha1Base64(data string) string {
	h := sha1.New()
	io.WriteString(h, data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sha1 算法：返回结果的16进制字符串
func Sha1Str(data []byte) string {
	return hex.EncodeToString(Sha1(data))
}

func SHA256(data string) string {
	sha256s := sha256.New()
	sha256s.Write([]byte(data))
	return hex.EncodeToString(sha256s.Sum([]byte("")))
}

// hmac_sha1 算法：返回结果的原始buffer
func HmacSha1(data []byte, key []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// hmac_sha1 算法：返回结果的16进制字符串
func HmacSha1Str(data []byte, key []byte) string {
	return hex.EncodeToString(HmacSha1(data, key))
}

// hmac_sha256 算法：返回结果的原始buffer
func HmacSha256(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// hmac_sha256 算法：返回结果的16进制字符串
func HmacSha256Str(data []byte, key []byte) string {
	return hex.EncodeToString(HmacSha256(data, key))
}

// hash_time33 算法
func HashTime33(str string) int {
	bStr := []byte(str)
	bLen := len(bStr)
	hash := 5381

	for i := 0; i < bLen; i++ {
		hash += ((hash << 5) & 0x7FFFFFF) + int(bStr[i]) //& 0x7FFFFFFF后保证其值为unsigned int范围
	}
	return hash & 0x7FFFFFF
}

// base64 编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// base64 解码
func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// rsa 加密
func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// rsa 解密
func RsaDecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func RsaSign(origData []byte, privateKey []byte) (string, error) {

	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.New(fmt.Sprintf("err =%+v", err))
	}

	h := sha256.New()
	h.Write([]byte(origData))
	digest := h.Sum(nil)
	s, err := rsa.SignPKCS1v15(nil, priv, crypto.SHA256, digest)
	if err != nil {
		return "", errors.New(fmt.Sprintf("err2 =%+v", err))
	}
	data := base64.StdEncoding.EncodeToString(s)
	return string(data), nil
}

func RsaVerifySign(origData []byte, publiKey []byte, sign []byte) error {

	block, _ := pem.Decode(publiKey)
	if block == nil {
		return errors.New("private key error!")
	}
	public, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return errors.New(fmt.Sprintf("err =%+v", err))
	}
	pub, ok := public.(*rsa.PublicKey)
	if !ok {
		return errors.New(fmt.Sprintf("err3 =%+v", ok))
	}
	h := sha256.New()
	h.Write([]byte(origData))
	digest := h.Sum(nil)

	unsign, err := Base64Decode(string(sign))
	if err != nil {
		return errors.New(fmt.Sprintf("err4 =%+v", err))
	}
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest, unsign)
	if err != nil {
		return errors.New(fmt.Sprintf("err2 =%+v", err))
	}
	return nil
}
