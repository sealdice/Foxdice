package core

import (
	"context"
	"errors"
	"foxdice/utils"
	"sync"
)

func NewManager(config utils.IConfig, logger utils.ILogger) *Manager {
	m := &Manager{
		Config: config,
		Logger: logger,
		DefaultBot: &Bot{
			Id:             1,
			Enable:         true,
			EndpointIdList: nil,
			captions:       make(map[string]*Caption),
			captionHooks:   make(map[string]func(ctx Context) string),
			Name:           "点头狐🦊",
		},
	}
	return m
}

type Manager struct {
	Eventbus   eventbus
	Sessions   *sessions
	Hub        *Hub
	Extensions []*Extension

	botMutex   sync.RWMutex
	BotList    []*Bot
	DefaultBot *Bot

	Logger utils.ILogger
	Config utils.IConfig

	serveCtx          context.Context
	loopCtx           context.Context
	loopCtxCancelFunc context.CancelCauseFunc
	maxGoroutine      uint
	limit             chan struct{}
	pool              sync.Pool
	wg                sync.WaitGroup
	mu                sync.Mutex
}

func (m *Manager) emit(et EventType, ctx Context) {
	m.Eventbus.emit(et, ctx)
}

func (m *Manager) emitBefore(et EventType, ctx Context) {
	m.emit(utils.Before+et, ctx)
}

func (m *Manager) emitAfter(et EventType, ctx Context) {
	m.emit(utils.After+et, ctx)
}

func (m *Manager) emitAll(et EventType, ctx Context) {
	m.emitBefore(et, ctx)
	m.emit(et, ctx)
	m.emitAfter(et, ctx)
}

func (m *Manager) EmitCustom(name string, ctx Context) {
	m.emit(EventType("custom/"+name), ctx)
}

func (m *Manager) On(et EventType, h HandlerFun) {
	m.Eventbus.on(et, h)
}

func (m *Manager) OnBefore(et EventType, h HandlerFun) {
	m.On(utils.Before+et, h)
}

func (m *Manager) OnAfter(et EventType, h HandlerFun) {
	m.On(utils.After+et, h)
}

func (m *Manager) TopOn(et EventType, h TopHandler) {
	m.Eventbus.topOn(et, h)
}

func (m *Manager) Serve(ctx context.Context) {
	m.serveCtx = ctx
	go m.Hub.Serve(ctx)
	<-m.Hub.startCompletionSignal
	m.DefaultBot.EndpointIdList = nil
	for _, e := range _m.Hub.Endpoints {
		m.DefaultBot.EndpointIdList = append(m.DefaultBot.EndpointIdList, e.Id)
	}
	m.ResetLoop()
}

func (m *Manager) StartEventLoop(stdCtx context.Context) {
	m.maxGoroutine = 5 * 1000
	m.limit = make(chan struct{}, m.maxGoroutine)
	for event := range m.Hub.Publish {
		select {
		case <-stdCtx.Done():
			return
		default:
		}
		ctx := m.pool.Get().(*EventContext)
		ctx.Manager = m
		ctx.Event = event
		ctx.Bot = m.DefaultBot
		ctx.reset(m.Sessions)
		m.limit <- struct{}{}
		m.wg.Add(1)
		go func(ctx *EventContext) {
			defer func() {
				<-m.limit
				m.wg.Done()
				if err := recover(); err != nil {
					m.Logger.Error(err)
					m.pool.Put(ctx)
					return // 发生错误就不保存了
				}
				if !ctx.beenCopied {
					ctx.Save()
				}
				m.pool.Put(ctx)
			}()
			m.emit("loop/event", ctx) // 存在一种需求 不要任何过滤
			if m.Sessions.tryContinuousDialog(ctx.Gid(), ctx.Text()) {
				m.Logger.Infof("此消息被判定为 prompt 的响应, 因此不做处理：[%s << %s]", ctx.Gid(), ctx.Text())
				return
			}
			// 端点自己发生变化、产生事件就没必要在上下文添加 bot 了
			if ctx.Event.Endpoint.SelfId == ctx.Event.User.Id {
				m.emit(utils.Endpoint+ctx.Event.Type, ctx)
				return
			}
			m.emitBefore(BotAttach, ctx)
			if ctx.block {
				return
			}
			for _, b := range m.BotList {
				ctx.Bot = b
				m.emitAfter(BotAttach, ctx)
				if ctx.block {
					return
				}
				switch ctx.Event.Type {
				case utils.MessageEvent:
					m.emitAll(utils.MessageEvent, ctx)
				default:
					m.emit(ctx.Event.Type, ctx)
				}
			}
		}(ctx)
	}
}

func (m *Manager) StopEventLoop() {
	m.loopCtxCancelFunc(nil)
	m.wg.Wait()
}

func (m *Manager) ResetLoop() {
	if !errors.Is(m.loopCtx.Err(), context.Canceled) {
		m.StopEventLoop()
	}
	m.loopCtx, m.loopCtxCancelFunc = context.WithCancelCause(m.serveCtx)
	m.StartEventLoop(m.loopCtx)
}
