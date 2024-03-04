package websocket

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"foxdice/utils"

	"github.com/gorilla/websocket"
)

type Socket struct {
	Conn              *websocket.Conn
	WebsocketDialer   *websocket.Dialer
	Url               string
	ConnectionOptions ConnectionOptions
	RequestHeader     http.Header
	OnConnected       func(socket Socket)
	OnTextMessage     func(message string)
	OnBinaryMessage   func(data []byte)
	OnConnectError    func(err error)
	OnDisconnected    func(err error)
	OnPingReceived    func(data string)
	OnPongReceived    func(data string)
	IsConnected       bool
	Timeout           time.Duration
	sendMu            *sync.Mutex
	receiveMu         *sync.Mutex
	logger            utils.ILogger
}

type ConnectionOptions struct {
	UseCompression bool
	UseSSL         bool
	Proxy          func(*http.Request) (*url.URL, error)
	Subprotocols   []string
}

func New(url string, logger utils.ILogger) Socket {
	return Socket{
		Url:           url,
		RequestHeader: http.Header{},
		ConnectionOptions: ConnectionOptions{
			UseCompression: false,
			UseSSL:         true,
		},
		WebsocketDialer: &websocket.Dialer{},
		Timeout:         0,
		sendMu:          &sync.Mutex{},
		receiveMu:       &sync.Mutex{},
		logger:          logger,
	}
}

func (s *Socket) setConnectionOptions() {
	s.WebsocketDialer.EnableCompression = s.ConnectionOptions.UseCompression
	s.WebsocketDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: s.ConnectionOptions.UseSSL}
	s.WebsocketDialer.Proxy = s.ConnectionOptions.Proxy
	s.WebsocketDialer.Subprotocols = s.ConnectionOptions.Subprotocols
}

func (s *Socket) Connect() {
	var err error
	var resp *http.Response
	s.setConnectionOptions()

	s.Conn, resp, err = s.WebsocketDialer.Dial(s.Url, s.RequestHeader)

	if err != nil {
		s.logger.Error("Error while connecting to server ", err)
		if resp != nil {
			s.logger.Error("HTTP Response %d status: %s", resp.StatusCode, resp.Status)
		}
		s.IsConnected = false
		if s.OnConnectError != nil {
			s.OnConnectError(err)
		}
		return
	}

	s.logger.Info("Connected to server")

	if s.OnConnected != nil {
		s.IsConnected = true
		s.OnConnected(*s)
	}

	defaultPingHandler := s.Conn.PingHandler()
	s.Conn.SetPingHandler(func(appData string) error {
		s.logger.Debug("Received PING from server")
		if s.OnPingReceived != nil {
			s.OnPingReceived(appData)
		}
		return defaultPingHandler(appData)
	})

	defaultPongHandler := s.Conn.PongHandler()
	s.Conn.SetPongHandler(func(appData string) error {
		s.logger.Debug("Received PONG from server")
		if s.OnPongReceived != nil {
			s.OnPongReceived(appData)
		}
		return defaultPongHandler(appData)
	})

	defaultCloseHandler := s.Conn.CloseHandler()
	s.Conn.SetCloseHandler(func(code int, text string) error {
		result := defaultCloseHandler(code, text)
		s.logger.Warn("Disconnected from server ", result)
		if s.OnDisconnected != nil {
			s.IsConnected = false
			s.OnDisconnected(errors.New(text))
		}
		return result
	})

	go func() {
		for {
			s.receiveMu.Lock()
			if s.Timeout != 0 {
				s.Conn.SetReadDeadline(time.Now().Add(s.Timeout))
			}
			messageType, message, err := s.Conn.ReadMessage()
			s.receiveMu.Unlock()
			if err != nil {
				s.logger.Error("read:", err)
				if s.OnDisconnected != nil {
					s.IsConnected = false
					s.OnDisconnected(err)
				}
				return
			}
			s.logger.Debug("recv: %s", message)

			switch messageType {
			case websocket.TextMessage:
				if s.OnTextMessage != nil {
					s.OnTextMessage(string(message))
				}
			case websocket.BinaryMessage:
				if s.OnBinaryMessage != nil {
					s.OnBinaryMessage(message)
				}
			}
		}
	}()
}

func (s *Socket) SendText(message string) {
	err := s.send(websocket.TextMessage, []byte(message))
	if err != nil {
		s.logger.Error("ws.send:", err)
		return
	}
}

func (s *Socket) SendBinary(data []byte) {
	err := s.send(websocket.BinaryMessage, data)
	if err != nil {
		s.logger.Error("ws.send:", err)
		return
	}
}

func (s *Socket) send(messageType int, data []byte) error {
	s.sendMu.Lock()
	err := s.Conn.WriteMessage(messageType, data)
	s.sendMu.Unlock()
	return err
}

func (s *Socket) Close() {
	err := s.send(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		s.logger.Error("write close:", err)
	}
	err = s.Conn.Close()
	if err != nil {
		s.logger.Error(err)
	}
	if s.OnDisconnected != nil {
		s.IsConnected = false
		s.OnDisconnected(err)
	}
}
