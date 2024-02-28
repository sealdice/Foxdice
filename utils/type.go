package utils

type EventType string

const (
	Before   = "before/"
	After    = "after/"
	Private  = "private/"
	Endpoint = "endpoint/"
)

const (
	MessageEvent     EventType = "message"
	GroupAddEvent    EventType = "GroupAdd"
	GroupQuitEvent   EventType = "GroupQuit"
	SendMessageEvent EventType = "SendMessage"
	CommandEvent     EventType = "CommandEvent"
)

type ILogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
}

type IConfig interface {
	Int(path string) int
	String(path string) string
	Bool(path string) bool
	Unmarshal(path string, o any) error
	Exists(path string) bool
	Set(path string, a any) error
	Sub(path string) IConfig
}
