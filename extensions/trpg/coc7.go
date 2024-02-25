package trpg

import (
	"foxdice/core"
)

// 克苏鲁的呼唤

func BuildCOC7(m *core.Manager) {
	/* roll check
	e.g.
		rc str
		rc[p/b] str[int]
	e.e.
		rc[p/b]str[50] => rc[p/b] xx[int]
		rcstr => rc str
		rcpow => rc pow
	*/
	rc := NewCmd("coc7", "rc")
	rc.AddCaptions(nil, map[string]*core.Caption{})
	rc.BuildParam("x:any")
	rc.SetAct(func(ctx *BotContext, game *Game) {
		ctx.ReplyCaption("RC_ONE")
	})

	st := NewCmd("coc7", "st")
	st.SetAct(func(ctx *BotContext, game *Game) {

	}).
		Try("")
	st.SetAct(func(ctx *BotContext, game *Game) {})
	stRm := st.Sub("rm")
	stRm.SetAlias("del")
	stRm.SetAct(func(ctx *BotContext, game *Game) {})
}
