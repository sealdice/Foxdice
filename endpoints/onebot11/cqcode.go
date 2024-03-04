package onebot11

import (
	"regexp"
	"strings"
)

// NewCQCode 解析为 CQCode 结构体 并反转义参数
func NewCQCode(code string) *CQCode {
	code = code[4 : len(code)-1]
	attr := strings.Split(code, ",")
	cq := &CQCode{Type: attr[0]}
	for i := 1; i < len(attr); i++ {
		kv := strings.SplitN(attr[i], "=", 1)
		cq.Args[kv[0]] = UnescapeValue(kv[1])
	}
	return cq
}

type CQCode struct {
	Type      string
	Args      map[string]string
	Overwrite string
}

func (c *CQCode) Compile() string {
	if c.Overwrite != "" {
		return c.Overwrite
	}
	sb := strings.Builder{}
	sb.WriteString("[CQ:")
	sb.WriteString(c.Type)
	for k, v := range c.Args {
		sb.WriteString(",")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(EscapeValue(v))
	}
	sb.WriteString("]")
	return sb.String()
}

var CQCodeRegexp = regexp.MustCompile(`\[CQ:([^],]+)(,[^]]+)?]`)

// ExecuteCQCode 提取 CQ码 并使用 cf 处理（纯文本部分转为 [CQ:text,text=pure_text]）
func ExecuteCQCode(dst string, cf func(cq *CQCode)) {
	res := CQCodeRegexp.FindAllStringIndex(dst, -1)
	i := 0
	for _, its := range res {
		a := its[0]
		b := its[1]
		if i < a {
			cf(&CQCode{Type: Text, Args: map[string]string{Text: EscapeValue(dst[i:a])}})
		}
		code := dst[a:b]
		cf(NewCQCode(code))
		i = b
	}
	if len(dst) != i {
		cf(&CQCode{Type: Text, Args: map[string]string{Text: EscapeText(dst[i:])}})
	}
}

func CQCodeString2Array(text string) []byte {
	return []byte("")
}

// 转义/反转义函数来自：https://github.com/Mrs4s/go-cqhttp/blob/master/internal/msg/element.go

// EscapeText 将字符串raw中部分字符转义
//
//   - & -> &amp;
//   - [ -> &#91;
//   - ] -> &#93;
func EscapeText(s string) string {
	count := strings.Count(s, "&")
	count += strings.Count(s, "[")
	count += strings.Count(s, "]")
	if count == 0 {
		return s
	}

	// Apply replacements to buffer.
	var b strings.Builder
	b.Grow(len(s) + count*4)
	start := 0
	for i := 0; i < count; i++ {
		j := start
		for index, r := range s[start:] {
			if r == '&' || r == '[' || r == ']' {
				j += index
				break
			}
		}
		b.WriteString(s[start:j])
		switch s[j] {
		case '&':
			b.WriteString("&amp;")
		case '[':
			b.WriteString("&#91;")
		case ']':
			b.WriteString("&#93;")
		}
		start = j + 1
	}
	b.WriteString(s[start:])
	return b.String()
}

// EscapeValue 将字符串value中部分字符转义
//
//   - , -> &#44;
//   - & -> &amp;
//   - [ -> &#91;
//   - ] -> &#93;
func EscapeValue(value string) string {
	ret := EscapeText(value)
	return strings.ReplaceAll(ret, ",", "&#44;")
}

// UnescapeText 将字符串content中部分字符反转义
//
//   - &amp; -> &
//   - &#91; -> [
//   - &#93; -> ]
func UnescapeText(content string) string {
	ret := content
	ret = strings.ReplaceAll(ret, "&#91;", "[")
	ret = strings.ReplaceAll(ret, "&#93;", "]")
	ret = strings.ReplaceAll(ret, "&amp;", "&")
	return ret
}

// UnescapeValue 将字符串content中部分字符反转义
//
//   - &#44; -> ,
//   - &amp; -> &
//   - &#91; -> [
//   - &#93; -> ]
func UnescapeValue(content string) string {
	ret := strings.ReplaceAll(content, "&#44;", ",")
	return UnescapeText(ret)
}
