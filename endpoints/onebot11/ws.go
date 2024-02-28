package onebot11

import (
	"context"

	"foxdice/endpoints/im"

	"github.com/sacOO7/gowebsocket"
)

func (a *Adapter) websocket(ctx context.Context, ch chan<- *im.Event) {
	socket := gowebsocket.New(a.ConnectURL)
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		a.Endpoint.Debug("[ping] " + data)
	}
	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		a.Endpoint.Debug("[pong] " + data)
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		a.Endpoint.Info("OneBot 服务的连接被关闭")
	}
	socket.OnConnected = func(socket gowebsocket.Socket) {
		ret := a.getVersionInfo()
		a.Implementation.AppName = ret.AppName
	}
	socket.OnBinaryMessage = func(msg []byte, socket gowebsocket.Socket) {
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
