package discord

import (
	"context"
	"foxdice/endpoints/im"
	"net/http"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

type Adapter struct {
	im.EmptyAdapter
	Session  *discordgo.Session
	Token    string
	ProxyURL string
}

func Attach(ep *im.Endpoint) {
	a := &Adapter{
		EmptyAdapter: im.EmptyAdapter{Endpoint: ep},
	}
	a.Token = ep.String(ep.Id + ".token")
	a.ProxyURL = ep.String(ep.Id + ".proxy_url")
	a.Endpoint = ep
	ep.Adapter = a
}

func (a *Adapter) Serve(ctx context.Context, ch chan<- *im.Event) {
	ep := a.Endpoint
	dg, err := discordgo.New("Bot " + a.Token)
	if err != nil {
		ep.Errorf("创建DiscordSession时出错:%s", err.Error())
	}

	if a.ProxyURL != "" {
		u, e := url.Parse(a.ProxyURL)
		if e != nil {
			ep.Errorf("代理地址解析错误%s", e.Error())

		}
		dg.Client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
		dg.Dialer = &websocket.Dialer{HandshakeTimeout: 45 * time.Second}
		dg.Dialer.Proxy = http.ProxyURL(u)
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot || m.Author.System {
			return
		}
		e := a.NewEvent()
		e.Guild.Id = m.GuildID
		e.Channel.Id = m.ChannelID
		e.Message.Id = m.Message.ID
		e.Message.Content = m.Message.Content
		e.Message.CreatedAt = m.Message.Timestamp.Unix()
		ch <- e
	})
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {})

	dg.Identify.Intents = discordgo.IntentsAll
	a.Session = dg
	err = dg.Open()

	if err != nil {
		ep.Errorf("与Discord服务进行连接时出错:%s", err.Error())
	}
	_ = a.Session.UpdateGameStatus(0, "FoxDice")
	ep.Info("Discord bots 上线")
}

func (a *Adapter) Close() {
	err := a.Session.Close()
	if err != nil {
		a.Endpoint.Error(err)
	}
}
