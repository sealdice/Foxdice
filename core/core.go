package core

import (
	"context"
	"foxdice/utils"

	"golang.org/x/exp/slices"
)

var _m *Manager

func GetManager() *Manager {
	return _m
}

// Emit 触发自定义事件
func Emit(name string, ctx Context) {
	_m.EmitCustom(name, ctx)
}

func On(et EventType, h HandlerFun) {
	_m.On(et, h)
}

func OnBefore(et EventType, h HandlerFun) {
	_m.OnBefore(et, h)
}

func OnAfter(et EventType, h HandlerFun) {
	_m.OnAfter(et, h)
}

// TopOn 用于扩展上下文 每个事件只能存在一个
func TopOn(et EventType, h TopHandler) {
	_m.Eventbus.topOn(et, h)
}

const (
	BotAttach     EventType = "bot/attach"     // Bot 被添加到上下文
	NotCommand    EventType = "command/not"    // 没有匹配到指令
	CommandAction EventType = "command/action" // 指令执行
	GroupNew      EventType = "group/new"      //
)

func InitManager(config utils.IConfig, logger utils.ILogger) *Manager {
	_m = NewManager(config, logger)
	_m.DefaultBot.Id = 0
	_m.DefaultBot.Enable = true
	nc := func(text, note string) *Caption {
		c, err := NewCaption(text, note, nil, nil)
		if err != nil {
			panic(err)
		}
		return c
	}
	_m.DefaultBot.captions = map[string]*Caption{
		"NAME":          nc("点头狐🦊", "机器人名字"),
		"SELF_PRONOUNS": nc("狐狐", "机器人自称"),
		"DESC":          nc("{SELF_PRONOUNS}名字就是“{NAME}”，“🦊”也不能少！", "机器人自我介绍"),
	}
	return _m
}

// DefaultHandler 一系列默认的监听器：事件过滤、自动更新群组和用户的缓存、指令系统
func DefaultHandler() {
	On(GroupNew, func(ctx Context) {
		r := ctx.Unsafe()
		if !slices.Contains(r.Group.EndpointIdList, r.Event.Endpoint.Id) {
			r.Sessions.mu.Lock()
			r.Group.EndpointIdList = append(r.Group.EndpointIdList, r.Event.Endpoint.Id)
			r.Sessions.mu.Unlock()
		}
	})
	// 事件过滤
	On(BotAttach, func(ctx Context) {
		r := ctx.Unsafe()
		e := r.Event
		b := r.Bot
		g := r.Group
		if !b.Enable {
			_m.Logger.Debugf("虽然接收到事件, 但是 [Bot:%s] 已经全局禁用", b.Name)
			ctx.SetBlock()
			return
		}
		if !slices.Contains(b.EndpointIdList, e.Endpoint.Id) {
			_m.Logger.Debugf("虽然接收到事件, 但是 [Bot:%s] 没有绑定 [Endpoint:%s]", b.Name, e.Endpoint.Name)
			ctx.SetBlock()
			return
		}
		// 当前 Manager 的多个 EndPoint 监听了同一群组 但通常情况下只应该响应一次
		// TODO 可配置
		if len(g.EndpointIdList) > 1 {
			gep, _ := _m.Hub.ById(g.EndpointIdList[0])
			if gep != e.Endpoint {
				ctx.Logger().Debugf(
					"虽然接收到事件, 但是同一群组内存在内响应级别更高的 [Endpoint:%s], 因此 [Endpoint:%s] 静默",
					gep.Name, e.Endpoint.Name)
				ctx.SetBlock()
				return
			}
		}
	})

	// 指令系统
	On(utils.MessageEvent, func(ctx Context) {
		e := ctx.Unsafe()
		text := e.Text()
		for _, extension := range e.Group.Extensions {
			for _, c := range extension.commands {
				if ok, argsText := c.Try(text); ok {
					if !extension.Enable {
						e.Logger().Debugf("找到指令 但是扩展禁用")
						goto NEXT
					}
					c.Parse(ctx, argsText)
					e.cmd = c
					goto NEXT
				}
			}
		}
	NEXT:
		if e.cmd == nil {
			_m.emit(NotCommand, ctx)
			return
		}

		if ctx.Errs() != nil {
			ctx.Reply(ctx.Errs().Error() + "\n" + e.cmd.Help(ctx))
			return
		}

		if ctx.Bool("help") {
			ctx.Reply(e.cmd.Help(ctx))
			return
		}
		_m.emitBefore(CommandAction, ctx)
		if e.block {
			return
		}
		e.cmd.Action(ctx)
		_m.emitAfter(CommandAction, ctx)
	})
}

func Serve(ctx context.Context) {
	_m.Serve(ctx)
}
