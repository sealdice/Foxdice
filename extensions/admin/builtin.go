package admin

import (
	"errors"
	"fmt"

	"foxdice/core"
	"foxdice/extensions/trpg"
	"foxdice/utils"
	"foxdice/utils/str"

	"gorm.io/gorm"
)

const builtinExtname = "admin"

type Redirect struct {
	gorm.Model
	Skey         string `gorm:"index"`
	FormId       string `gorm:"index"`
	FormPlatform string
	ToId         string
	ToPlatform   string
}

func Build(m *core.Manager, db *gorm.DB) {
	botCmd := core.NewCmd(builtinExtname, "bot")
	botCmd.SetAct(func(ctx core.Context) {
		ctx.Reply("bots")
		// ctx.Reply(fmt.Sprintf("软件名 %s 版本 %s", utils.AppName, utils.Version))
	})

	botCmd.Sub("off").SetAct(func(ctx core.Context) {
		mctx := ctx.(*trpg.BotContext)
		mctx.SetBotOff()
		ctx.Reply("bots 被关闭")
	})

	botCmd.Sub("on").SetAct(func(ctx core.Context) {
		mctx := ctx.(*trpg.BotContext)
		mctx.SetBotOn()
		ctx.Reply("bots 被打开")
	})

	// 重定向 ID
	err := db.AutoMigrate(Redirect{})
	if err != nil {
		m.Logger.Error("重定向功能无法使用:", err)
	} else {
		reCmd := core.NewCmd(builtinExtname, "re")

		// re *hello_QQ/123/123 -r
		reCmd.SetOnlyGroup()
		reCmd.BuildParam("skey!")
		reCmd.SetAct(func(ctx core.Context) {
			m := ctx.Unsafe()
			r := Redirect{}
			r.Skey = ctx.String("skey")

			err := db.Take(r).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.Reply("未能找到密钥对应的群组")
				}
				return
			}

			r.FormId = m.Event.Channel.Id
			r.FormPlatform = m.Event.Platform
			if ctx.Bool("r") {
				r.Skey = ""
			}

			if err := db.Save(r).Error; err != nil {
				ctx.Reply("内部异常")
				return
			}

			ctx.Reply(fmt.Sprintf("%s 将被重定向到 %s", r.FormId, r.ToId))
		})

		// re skey
		reCmd.Sub("skey").SetAct(func(ctx core.Context) {
			m := ctx.Unsafe()
			skey := "*" + str.NanoId(6) + m.Gid()

			r := Redirect{}
			r.Skey = skey
			r.ToId = m.Event.Channel.Id
			r.ToPlatform = m.Event.Platform

			if err := db.Save(r).Error; err != nil {
				ctx.Reply("内部异常")
				return
			}

			ctx.Reply(fmt.Sprintf("已更新当前频道/群 %s 的密钥\n可使用在其他群组使用指令 re %s 使 ID 重定向到此处",
				m.Event.Channel.Id, skey))
		})

		core.OnBefore(utils.MessageEvent, func(ctx core.Context) {
			e := ctx.Unsafe()
			if e.Private() {
				return
			}
			r := Redirect{}
			r.FormId = e.Event.Channel.Id
			r.FormPlatform = e.Event.Platform
			db.First(r)
			if r.ToId != "" && r.ToPlatform != "" {
				e.Event.Channel.Id = r.ToId
				e.Event.Platform = r.ToPlatform
				e.Logger().Infof("[redirect] %s:%s => %s:%s", r.FormPlatform, r.FormId, r.ToPlatform, r.ToId)
			}
			e.SetBlock()
			e.Manager.Hub.Source <- e.Event
		})

	}

	// 模板替换
	core.OnBefore(utils.SendMessageEvent, func(ctx core.Context) {
		// seal code
		// cq code
	})

	// 日志记录
	core.On(utils.MessageEvent, func(ctx core.Context) {
		event := ctx.IMEvent()
		ctx.Logger().Infof("[%s:%s] << %s", event.Platform, event.Channel.Id, event.Message.Content)
	})

	core.OnBefore(utils.SendMessageEvent, func(ctx core.Context) {
		event := ctx.IMEvent()
		ctx.Logger().Infof("[%s:%s] >> %s", event.Platform, event.Channel.Id, event.Message.Content)
	})
}
