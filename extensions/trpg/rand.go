package trpg

func BuildRandIt() {
	draw := NewCmd("draw", "draw", "deck")
	draw.SetAct(func(ctx *BotContext, game *Game) {

	})
}
