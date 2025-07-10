package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

func RsaWithSHA256Base64(origData string, block []byte) (string, error) {
	blocks, _ := pem.Decode(block)
	if blocks == nil {
		//logs.LogDebug("blocks is nil:%+v", blocks)
		return "", errors.New("blocks is nil")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(blocks.Bytes)
	if err != nil {
		//logs.LogDebug("ParsePKCS8PrivateKey err:%+v", err)
		return "", err
	}
	h := crypto.Hash.New(crypto.SHA256)
	h.Write([]byte(origData))
	hashed := h.Sum(nil)

	// 进行rsa加密签名
	signedData, err := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed)
	if err != nil {
		//logs.LogDebug("SignPKCS1v15 err:%+v", err)
		return "", err
	}
	sign := base64.StdEncoding.EncodeToString(signedData)
	return sign, nil
}
