package im

import (
	"foxdice/utils"
)

type EventType = utils.EventType

type ChannelType int

const (
	TextChannel ChannelType = iota
	DirectChannel
	CategoryChannel
	VoiceChannel
)

type Channel struct {
	Id       string
	Type     ChannelType
	Name     string
	ParentId string
}

type Guild struct {
	Id     string
	Name   string
	Avatar string
}

type GuildMember struct {
	User     *User
	Nick     string
	Avatar   string
	JoinedAt int
}

type GuildRole struct {
	Id   string
	Name string
}

type User struct {
	Id     string
	Name   string
	Nick   string
	Avatar string
	IsBot  bool
}

type Message struct {
	Id        string
	Content   string
	CreatedAt int64
	UpdatedAt int64
}

type Status int

const (
	Offline Status = iota
	OnLine
	Connect
	DisConnect
	ReConnect
)

type Argv struct {
	Name      string
	Arguments []any
	Option    any
}

type Button struct {
	Id string
}

type StdEvent struct {
	// 事件类型
	Type EventType
	// 消息元素
	Elements []*Element
	// Endpoint
	Endpoint *Endpoint
}

type Event struct {
	StdEvent
	Id        int64
	Platform  string
	SelfId    string
	Timestamp int64
	Argv      Argv
	Button    Button
	Login     *LoginInfo
	Channel   Channel
	Guild     Guild
	Role      GuildRole
	Member    GuildMember
	Message   Message
	// 消息发生者、操作者
	User User
	// 消息接受者、目标用户
	TargetUser User
}

func (e *Event) Append(el *Element) {
	e.Elements = append(e.Elements, el)
}
