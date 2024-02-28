package onebot11

import (
	"encoding/json"
	"foxdice/endpoints/im"
	"foxdice/utils"
	"regexp"
)

func (a *Adapter) parseMessage(s []byte, e *im.Event) bool {
	if !a.UseArrayMessage {
		msg := new(MessageCQCode)
		err := json.Unmarshal(s, msg)
		if err != nil {
			a.Endpoint.Error("解析 OneBot11 CQ 码类型的消息失败", err)
			a.Endpoint.Info("将尝试作为数组类型的消息解析")
			a.UseArrayMessage = true
			goto Array
		}
		longText := msg.Message
		re := regexp.MustCompile(`\[CQ:.+?]`)
		m := re.FindAllStringIndex(msg.Message, -1)
		newText := msg.Message
		var arr []any

		for i := len(m) - 1; i >= 0; i-- {
			p := m[i]
			cq := CQParse(longText[p[0]:p[1]])

			// 如果尾部有文本，将其拼入数组
			endText := newText[p[1]:]
			if len(endText) > 0 {
				i := OneBotV11ArrMsgItem[OneBotV11MsgItemTextType]{Type: "text", Data: OneBotV11MsgItemTextType{Text: endText}}
				arr = append(arr, i)
			}

			// 将 CQ 拼入数组
			switch cq.Type {
			case "image":
				i := OneBotV11ArrMsgItem[OneBotV11MsgItemImageType]{Type: "image", Data: OneBotV11MsgItemImageType{File: cq.Args["file"]}}
				arr = append(arr, i)
			case "record":
				i := OneBotV11ArrMsgItem[OneBotV11MsgItemRecordType]{Type: "record", Data: OneBotV11MsgItemRecordType{File: cq.Args["file"]}}
				arr = append(arr, i)
			case "at":
				// [CQ:at,qq=10001000]
				i := OneBotV11ArrMsgItem[OneBotV11MsgItemAtType]{Type: "at", Data: OneBotV11MsgItemAtType{QQ: cq.Args["qq"]}}
				arr = append(arr, i)
			default:
				data := make(map[string]interface{})
				for k, v := range cq.Args {
					data[k] = v
				}
				i := OneBotV11MsgItem{Type: cq.Type, Data: data}
				arr = append(arr, i)
			}

			newText = newText[:p[0]]
		}

		// 如果剩余有文本，将其拼入数组
		if len(newText) > 0 {
			i := OneBotV11ArrMsgItem[OneBotV11MsgItemTextType]{Type: "text", Data: OneBotV11MsgItemTextType{Text: newText}}
			arr = append(arr, i)
		}

	}
Array:
	msg := new(MessageArray)
	err := json.Unmarshal(s, msg)
	if err != nil {
		a.Endpoint.Error("解析 OneBot11 数组类型的消息失败", err)
		return false
	}

	for _, m := range msg.Message {
		switch m.Type {
		case "text":
			e.Elements = append(e.Elements, &im.Element{Type: im.TextElement, Data: m.Data["text"]})
		case "image":
			e.Elements = append(e.Elements, &im.Element{Type: im.ImgElement, Data: m.Data["file"]})
		case "face":
			e.Elements = append(e.Elements, &im.Element{Type: im.TextElement, Data: m.Data["id"]})
		case "record":
			e.Elements = append(e.Elements, &im.Element{Type: im.TextElement, Data: m.Data["file"]})
		case "at":
			e.Elements = append(e.Elements, &im.Element{Type: im.AtElement, Data: m.Data["qq"]})
		case "poke":
			e.Elements = append(e.Elements, &im.Element{Type: im.TextElement, Data: m.Data["poke"]})
		case "reply":
			e.Elements = append(e.Elements, &im.Element{Type: im.TextElement, Data: m.Data["id"]})
		}
	}
	return true
}

func (a *Adapter) parseEvent(msg []byte) (*im.Event, *Event) {
	obe := new(Event)
	event := a.NewEvent()
	err := json.Unmarshal(msg, &obe)
	if err != nil {
		a.Endpoint.Error("解析 OneBot11 事件失败", err)
		return nil, nil
	}
	switch obe.PostType {
	case "message":
		event.Type = utils.MessageEvent
		if ok := a.parseMessage(msg, event); ok == false {
			return nil, nil
		}
		switch obe.MessageType {
		case "private":
			switch obe.SubType {
			case "friend":
			case "group":
			case "other":
			}
		case "group":
			switch obe.SubType {
			case "normal":
			case "anonymous":
			case "notice":
			}
		}
	case "notice":
		switch obe.NoticeType {
		case "group_upload":
		case "group_admin":
			switch obe.SubType {
			case "set":
			case "unset":
			}
		case "group_decrease":
			switch obe.SubType {
			case "leave":
			case "kick":
			case "kick_me":
			}
		case "group_increase":
			switch obe.SubType {
			case "approve":
			case "invite":
			}
		case "group_ban":
			switch obe.SubType {
			case "ban":
			case "lift_ban":
			}
		case "friend_add":
		case "group_recall":
		case "friend_recall":
		case "notify":
			switch obe.SubType {
			case "poke":
			case "lucky_king":
			case "honor":
				switch obe.HonorType {
				case "talkative":
				case "performer":
				case "emotion":
				}
			}
		}
	case "request":
		switch obe.RequestType {
		case "friend":
		case "group":
			switch obe.SubType {
			case "add":
			case "invite":
			}
		}
	case "meta_event":
		switch obe.MetaEventType {
		case "heartbeat":
		case "lifecycle":
			switch obe.SubType {
			case "enable":
			case "disable":
			case "connect":
			}
		}
	}
	return event, obe
}

type Sender struct {
	Age      int32           `json:"age"`
	Card     string          `json:"card"`
	Nickname string          `json:"nickname"`
	Role     string          `json:"role"` // owner 群主
	UserID   json.RawMessage `json:"user_id"`
}

type OneBotUserInfo struct {
	// 个人信息
	Nickname string `json:"nickname"`
	UserID   string `json:"user_id"`

	// 群信息
	GroupID         string `json:"group_id"`          // 群号
	GroupCreateTime uint32 `json:"group_create_time"` // 群号
	MemberCount     int64  `json:"member_count"`
	GroupName       string `json:"group_name"`
	MaxMemberCount  int32  `json:"max_member_count"`
	Card            string `json:"card"`
}

type RetData struct {
	// 个人信息
	Nickname string          `json:"nickname"`
	UserID   json.RawMessage `json:"user_id"`

	// 群信息
	GroupID         json.RawMessage `json:"group_id"`          // 群号
	GroupCreateTime uint32          `json:"group_create_time"` // 群号
	MemberCount     int64           `json:"member_count"`
	GroupName       string          `json:"group_name"`
	MaxMemberCount  int32           `json:"max_member_count"`

	// 群成员信息
	Card    string `json:"card"`
	AppName string
}

type Event struct {
	MessageID     int64           `json:"message_id"`   // QQ信息此类型为int64，频道中为string
	MessageType   string          `json:"message_type"` // Group
	Sender        *Sender         `json:"sender"`       // 发送者
	RawMessage    string          `json:"raw_message"`
	Time          int64           `json:"time"` // 发送时间
	MetaEventType string          `json:"meta_event_type"`
	OperatorID    json.RawMessage `json:"operator_id"`  // 操作者帐号
	GroupID       json.RawMessage `json:"group_id"`     // 群号
	PostType      string          `json:"post_type"`    // 上报类型，如group、notice
	RequestType   string          `json:"request_type"` // 请求类型，如group
	SubType       string          `json:"sub_type"`     // 子类型，如add invite
	HonorType     string          `json:"honor_type"`
	Flag          string          `json:"flag"` // 请求 flag, 在调用处理请求的 API 时需要传入
	NoticeType    string          `json:"notice_type"`
	UserID        json.RawMessage `json:"user_id"`
	SelfID        json.RawMessage `json:"self_id"`
	Duration      int64           `json:"duration"`
	Comment       string          `json:"comment"`
	TargetID      json.RawMessage `json:"target_id"`

	Data    *RetData `json:"data"`
	RetCode int64    `json:"retcode"`
	Status  string   `json:"status"`
	Echo    string   `json:"echo"` // 声明类型而不是interface的原因是interface下数字不能正确转换

	Msg string `json:"msg"`
	// Status  interface{} `json:"status"`
	Wording string `json:"wording"`
}

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

type MessageArray struct {
	Event
	Message []*OneBotV11MsgItem `json:"message"` // 消息内容
}

type MessageCQCode struct {
	Event
	Message string `json:"message"` // 消息内容
}
