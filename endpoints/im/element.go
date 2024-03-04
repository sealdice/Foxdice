package im

import "encoding/xml"

type ElementType string // TODO 改成字符

const (
	TextElement      ElementType = ""        // 纯文本
	AtElement        ElementType = "at"      // 提及用户
	SharpElement     ElementType = "sharp"   // 提及频道
	LinkElement      ElementType = "a"       // 链接
	ImgElement       ElementType = "img"     // 图片
	AudioElement     ElementType = "audio"   // 语音
	VideoElement     ElementType = "video"   // 视频
	FileElement      ElementType = "file"    // 文件
	StrongElement    ElementType = "b"       // 粗体
	EmElement        ElementType = "i"       // 斜体
	InsElement       ElementType = "u"       // 下划线
	DelElement       ElementType = "s"       // 删除线
	SplElement       ElementType = "spl"     // 剧透
	CodeElement      ElementType = "code"    // 代码
	SupElement       ElementType = "sup"     // 上标
	SubElement       ElementType = "sub"     // 下标
	BrElement        ElementType = "br"      // 换行
	ParagraphElement ElementType = "p"       // 段落
	MessageElement   ElementType = "message" // 消息
	QuoteElement     ElementType = "quote"   // 引用
	AuthorElement    ElementType = "author"  // 作者
)

type Attr struct {
	key   string
	value string
}

type Element struct {
	Type ElementType `xml:"name"`
	xml.Name
	data []Attr
}

func (el *Element) Get(key string) (string, bool) {
	for i := 0; i < len(el.data); i++ {
		if el.data[i].key == key {
			return el.data[i].value, true
		}
	}
	return "", false
}

func (el *Element) Set(key string, val string) {
	el.data = append(el.data, Attr{key: key, value: val})
}
