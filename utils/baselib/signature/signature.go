package signature

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aiden2048/pkg/utils/baselib/crypto"
)

/*
  Sig生成规范 <<腾讯开放平台第三方应用签名参数sig的说明>>
  http://wiki.open.qq.com/wiki/%E8%85%BE%E8%AE%AF%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0%E7%AC%AC%E4%B8%89%E6%96%B9%E5%BA%94%E7%94%A8%E7%AD%BE%E5%90%8D%E5%8F%82%E6%95%B0sig%E7%9A%84%E8%AF%B4%E6%98%8E
*/

// 生成sig串
func GenerateSig(method string, urlPath string, params map[string]string, sigKey string) (sig string) {

	// ======= 1.拼接原始数据和加密 key =======
	// (1.1) 参数排序
	keys := make([]string, len(params))
	i := 0
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// (1.2) 参数拼接字符串
	strParams := ""
	for _, k := range keys {
		strParams += k + "=" + params[k] + "&"
	}
	if len(strParams) > 0 {
		// 将整个字符串最后的 "&" 去掉
		strParams = strParams[0 : len(strParams)-1]
	}

	source := strings.ToUpper(method) + "&" + urlEncodeForSig(urlPath) + "&" + urlEncodeForSig(strParams)
	secret := sigKey + "&"

	// ======= 2.hmac_sha1 处理数据 =====
	byteSig := crypto.HmacSha1([]byte(source), []byte(secret))

	// ======= 3.将密文数据明文化 =======
	// 利用 base64 算法，而不是将 buf 转换成 16 进制
	strSig := crypto.Base64Encode(byteSig)
	return strSig
}

func GenerateSig2(method string, urlPath string, params map[string]interface{}, secret string) (sig string) {
	// 参数排序
	keys := make([]string, len(params))
	i := 0
	for k, _ := range params {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	// 参数拼接字符串
	strParams := ""
	for _, k := range keys {
		strParams += k + "=" + fmt.Sprintf("%v", params[k]) + "&"
	}
	if len(strParams) > 0 {
		strParams = strParams[0 : len(strParams)-1]
	}
	source := strings.ToUpper(method) + "&" + urlEncodeForSig(urlPath) + "&" + urlEncodeForSig(strParams)
	byteSig := crypto.HmacSha1([]byte(source), []byte(secret+"&"))
	strSig := crypto.Base64Encode(byteSig)
	return strSig
}

// 计算sig专用的url encode (修复了urlEncodeForSigReg占用CPU过高的问题)
func urlEncodeForSig(url string) string {
	encodingUrl := make([]byte, 0)
	HexDigits := []byte("0123456789ABCDEF")
	for i := 0; i < len(url); i++ {
		c := url[i]
		// 除字母数字.-_以外全部转码
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '.' || c == '-' || c == '_' {
			encodingUrl = append(encodingUrl, c)
		} else {
			encodingUrl = append(encodingUrl, '%', HexDigits[(c>>4)&0x0f], HexDigits[c&0x0f])
		}
	}
	return string(encodingUrl)
}

// 计算sig专用的url encode
// 使用正则匹配要转换的ASCII, 当sig参数比较多时, 请求量大的时候, cpu占用率过高
func urlEncodeForSigReg(url string) string {
	encodingUrl := ""
	reg, _ := regexp.Compile("[a-zA-Z0-9._-]")
	for i := 0; i < len(url); i++ {
		b := []byte(url[i : i+1])
		if reg.Match(b) {
			encodingUrl += fmt.Sprintf("%s", b)
		} else {
			encodingUrl += fmt.Sprintf("%%%X", b)
		}
	}
	return encodingUrl
}
