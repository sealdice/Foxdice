package core

import (
	"foxdice/utils"
	"sync"
)

type EventType = utils.EventType

type eventbus struct {
	mu         sync.RWMutex // 不确定 再想想真的存在并发读写的场景吗
	listener   map[EventType][]Handler
	topHandler map[EventType]TopHandler
}

type HandlerFun = func(ctx Context)

type Handler struct {
	enable bool
	fun    HandlerFun
}

type TopHandler = func(ctx Context) Context

func (b *eventbus) emit(et EventType, ctx Context) {
	b.mu.RLock()
	if h, ok := b.topHandler[et]; ok {
		h(ctx)
	}
	for _, handler := range b.listener[et] {
		if handler.enable {
			if ctx.Block() {
				b.mu.RUnlock()
				return
			}
			handler.fun(ctx)
		}
	}
	b.mu.RUnlock()
}

func (b *eventbus) topOn(et EventType, h TopHandler) {
	b.mu.Lock()
	b.topHandler[et] = h
	b.mu.Unlock()
}

func (b *eventbus) on(et EventType, h HandlerFun) int {
	b.mu.Lock()
	b.listener[et] = append(b.listener[et], Handler{enable: true, fun: h})
	b.mu.Unlock()
	return len(b.listener) - 1
}

func (b *eventbus) DisableHandler(et EventType, index int) {
	b.mu.Lock()
	if len(b.listener) > index && index >= 0 {
		b.listener[et][index].enable = false
	}
	b.mu.Unlock()
}

func (b *eventbus) EnableHandler(et EventType, index int) {
	b.mu.Lock()
	if len(b.listener) > index && index > 0 {
		b.listener[et][index].enable = false
	}
	b.mu.Unlock()
}
