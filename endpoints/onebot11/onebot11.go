package onebot11

import (
	"context"
	"foxdice/endpoints/im"
	"foxdice/utils/syncx"
	"net/http"

	"github.com/sacOO7/gowebsocket"
)

const (
	WsMode        = "ws"
	HttpMode      = "http"
	WsReverseMode = "ws-reverse"
)

func New(ep *im.Endpoint) {
	adapter := new(Adapter)
	_ = ep.Unmarshal("config", adapter)
	adapter.Endpoint = ep
}

type Adapter struct {
	im.EmptyAdapter
	socket          gowebsocket.Socket
	httpServer      http.Server
	httpClient      http.Client
	echo            syncx.RWMap[string, chan *RetData]
	Implementation  versionInfo
	ReverseAddr     string
	ConnectURL      string
	AccessToken     string
	Mode            string
	UseArrayMessage bool
}

func (a *Adapter) Serve(ctx context.Context, ch chan<- *im.Event) {
	switch a.Mode {
	case WsMode:
		a.websocket(ctx, ch)
	case HttpMode:
		a.http(ctx, ch)
	case WsReverseMode:
		a.websocketReverse(ctx, ch)
	}
}

func (a *Adapter) Close() {
	switch a.Mode {
	case WsMode:
		a.socket.Close()
	case HttpMode:
	case WsReverseMode:
	}
}
