package onebot11

import (
	"context"

	"foxdice/endpoints/im"
	"foxdice/utils/websocket"
)

func (a *Adapter) websocket(ctx context.Context, ch chan<- *im.Event) {
	socket := websocket.New(a.ConnectURL, a.Endpoint.ILogger)
	socket.OnDisconnected = func(err error) {
		a.Endpoint.Info("OneBot 服务的连接被关闭")
	}
	socket.OnConnected = func(socket websocket.Socket) {
		implementationAdapter(a)
	}
	socket.OnBinaryMessage = func(msg []byte) {
		e, obe := a.parseEvent(msg)
		if obe.Status != "" {
			if ch2, ok := a.echo.Load(obe.Echo); ok {
				ch2 <- obe.Data
			}
			return
		}
		if e != nil {
			ch <- e
		}
	}
	a.socket = socket
	a.socket.Connect()
	for {
		select {
		case <-ctx.Done():
			a.Close()
			return
		}
	}
}
