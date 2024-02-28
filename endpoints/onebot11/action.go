package onebot11

import "foxdice/utils/str"

func (a *Adapter) BuildAction() {

}

type versionInfo struct {
	AppName         string
	AppVersion      string
	ProtocolVersion string
}

func (a *Adapter) getVersionInfo() *RetData {
	switch a.Mode {
	case WsMode, WsReverseMode:
		ch := make(chan *RetData)
		a.echo.Store(str.UUID(), ch)
		return <-ch
	case HttpMode:
		return nil
	}
	return nil
}
