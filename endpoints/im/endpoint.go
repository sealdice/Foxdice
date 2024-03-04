package im

import (
	"foxdice/utils"
	"foxdice/utils/str"
)

type LoginInfo struct {
	Name     string
	SelfId   string
	Platform string
	Status   Status
}

func NewEndpoint() *Endpoint {
	ep := &Endpoint{}
	ep.Id = str.UUID()
	return ep
}

type Endpoint struct {
	Id string
	LoginInfo
	Adapter Adapter
	utils.IConfig
	utils.ILogger
}

type Platform = string

const (
	Unknown  Platform = "UNKNOWN"
	Cli      Platform = "CLI"
	QQ       Platform = "QQ"
	KOOK     Platform = "KOOK"
	Discord  Platform = "DISCORD"
	Telegram Platform = "TELEGRAM"
)
