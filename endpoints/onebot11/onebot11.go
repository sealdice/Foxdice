package onebot11

import (
	"context"
	"net/http"

	"foxdice/endpoints/im"
	"foxdice/utils/syncx"
	"foxdice/utils/websocket"
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
	socket          websocket.Socket
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

func implementationAdapter(a *Adapter) {
	a.Endpoint.Platform = im.QQ // ob11 目前有 QQ 平台以外的实现吗
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
