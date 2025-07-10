package strings

import(
	"text/template"
)

// 对输出文本中的特殊字符(",',',&,<,>)转义，防止XSS注入
func HtmlEncode(str string) string {
	s := template.HTMLEscapeString(str)
	return s
}

// 对输出JavaScript代码中特殊字符转义，防止XSS注入
func JSEncode(str string) string {
	s := template.JSEscapeString(str)
	return s
}

