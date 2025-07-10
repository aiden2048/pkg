package baselib

import(
 	"golang.org/x/text/encoding/simplifiedchinese"
    "golang.org/x/text/transform"
)

// 编码转换
func GBKToUTF8(gbk string) string{
	// cd, _ := iconv.Open("utf-8", "gbk")
	// defer cd.Close()
	// return cd.ConvString(string(gbk))
    
	e := simplifiedchinese.GBK
	if utf8, _, err := transform.String(e.NewDecoder(), gbk); err != nil{
		return ""
	}else{
		return utf8
	}
}

// 编码转换
func UTF8ToGBK(utf8 string) string{
	// cd, _ := iconv.Open("gbk", "utf-8")
	// defer cd.Close()
	// return cd.ConvString(string(utf8))

	e := simplifiedchinese.GBK
	if gbk, _, err := transform.String(e.NewEncoder(), utf8); err != nil{
		return ""
	}else{
		return gbk
	}
}
