package core

type Extension struct {
	Index    int
	Enable   bool
	Name     string
	commands []Command
	m        *Manager
}

func (e *Extension) OnNotCommand(handler HandlerFun) {
	On(NotCommand, handler)
}

func (e *Extension) NewCommand(name string) Command {
	cmd := &command{name: name, ext: e}
	e.commands = append(e.commands, cmd)
	return cmd
}

func (e *Extension) RegConfig(m map[string]*Caption) {

}
