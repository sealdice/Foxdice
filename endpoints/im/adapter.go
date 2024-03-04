package im

import (
	"context"
	"time"
)

// Adapter 最基本适配器实现 收发消息
type Adapter interface {
	Serve(ctx context.Context, ch chan<- *Event)
	Send(msg *Event)
	Close()
}

// EmptyAdapter 一个默认的实现 提供了平台不提供功能的默认实现
// 平台适配器应该嵌入 EmptyAdapter
type EmptyAdapter struct {
	Endpoint *Endpoint
}

// NewEvent 平台适配器实现应使用此方法创建事件
func (a *EmptyAdapter) NewEvent() *Event {
	return &Event{
		StdEvent:   StdEvent{},
		Platform:   a.Endpoint.Platform,
		SelfId:     a.Endpoint.SelfId,
		Timestamp:  time.Now().Unix(),
		Argv:       Argv{},
		Button:     Button{},
		Login:      &a.Endpoint.LoginInfo,
		Channel:    Channel{},
		Guild:      Guild{},
		Role:       GuildRole{},
		Member:     GuildMember{},
		Message:    Message{},
		User:       User{},
		TargetUser: User{},
	}
}

func (a *EmptyAdapter) Serve(ctx context.Context, ch chan<- *Event) {
}

func (a *EmptyAdapter) Close() {
}

func (a *EmptyAdapter) Send(msg *Event) {

}

func (a *EmptyAdapter) GetChannel(id string) {}

func (a *EmptyAdapter) CreateChannel(id string) {}

func (a *EmptyAdapter) UpdateChannel(id string) {}

func (a *EmptyAdapter) DeleteChannel(id string) {}

func (a *EmptyAdapter) GetChannelList(guildId string) {}

func (a *EmptyAdapter) GetGuild(id string) {}

func (a *EmptyAdapter) GetGuildList() {}

func (a *EmptyAdapter) ApproveGuild(id string) {}

// func (a *EmptyAdapter) CreateGuild(id string) {}
//
// func (a *EmptyAdapter) UpdateGuild(id string) {}
//
// func (a *EmptyAdapter) DeleteGuild(id string) {}
