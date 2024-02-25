package trpg

import (
	"foxdice/core"
	"strings"
)

type Bot struct {
	IsOff bool
}

func (b Bot) SetOn(id string) {

}

type BotContext struct {
	core.EventContext

	Bot Bot
}

func (c *BotContext) LoadGame() *Game {
	game := new(Game)
	c.Range(func(k string, v any) {
		if strings.HasPrefix(k, "_") {
			return
		}
	})
	return game
}

func (c *BotContext) SetBotOff() {

}

func (c *BotContext) SetBotOn() {

}

func NewCmd(ext string, name string, alias ...string) *Command {
	cmd := new(Command)
	cmd.Cmd = *core.NewCmd(ext, name, alias...)
	return cmd
}

type act func(ctx *BotContext, game *Game)

type Command struct {
	core.Cmd
	act act
}

func (c *Command) Action(ctx core.Context) {
	mctx := ctx.(*BotContext)
	game := mctx.LoadGame()
	c.act(mctx, game)
}
func (c *Command) Sub(n string, p ...string) *Command {
	return c
}

func (c *Command) SetAct(act act) *core.Cmd {
	c.act = act
	return &c.Cmd
}
