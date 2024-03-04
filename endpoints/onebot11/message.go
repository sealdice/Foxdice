package onebot11

import (
	"foxdice/endpoints/im"

	"github.com/tidwall/gjson"
)

const (
	Text   = "text"
	Image  = "image"
	Face   = "face"
	Record = "record"
	At     = "at"
	Poke   = "poke"
	Reply  = "reply"
)

// https://github.com/botuniverse/onebot-11/blob/master/message/array.md

type OneBotV11MsgItemTextType struct {
	Text string `json:"text"`
}

type OneBotV11MsgItemImageType struct {
	File string `json:"file"`
}

type OneBotV11MsgItemFaceType struct {
	Id string `json:"id"`
}

type OneBotV11MsgItemRecordType struct {
	File string `json:"file"`
}

type OneBotV11MsgItemAtType struct {
	QQ string `json:"qq"`
}

type OneBotV11MsgItemPokeType struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

type OneBotV11MsgItemReplyType struct {
	Id string `json:"id"`
}

type OneBotV11MsgItem struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type OneBotV11ArrMsgItem[T any] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type MessageItem struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type MessageArray struct {
	Message []*OneBotV11MsgItem `json:"message"` // 消息内容
}

type MessageCQCode struct {
	Message string `json:"message"` // 消息内容
}

func (a *Adapter) parseMessage(bytes []byte, e *im.Event) bool {
	res := gjson.GetBytes(bytes, "message")
	// CQCode
	if res.Type == gjson.String {
		ExecuteCQCode(res.Str, func(cq *CQCode) {
			el := &im.Element{Type: im.TextElement}
			switch cq.Type {
			case Text:
				el.Set("text", cq.Args["text"])
			}
			e.Append(el)
		})
		return true
	}

	// JSON 形式
	convert := func(e gjson.Result) {
		var elem im.Element
		switch e.Get("type").Str {
		case Text:
			elem.Type = im.TextElement
			r := e.Get("data.text")
			if r.Type == gjson.String {
				elem.Set("text", r.Str)
			}
		}
	}

	if res.IsArray() {
		res.ForEach(func(_, e gjson.Result) bool {
			convert(e)
			return true
		})
		return true
	}
	if res.IsObject() {
		convert(res)
		return true
	}

	// 无法解析
	return false
}
