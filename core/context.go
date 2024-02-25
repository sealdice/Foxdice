package core

import (
	"errors"
	"fmt"
	"foxdice/endpoints/im"
	"foxdice/utils"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type Context interface {
	// Copy 因为 Context(EventContext) 不是并发安全的
	// 而且使用 sync.Pool 复用了对象
	// 传入其他 goroutine 时要 Copy 一份, 并手动 Save 防止数据竞争
	// (但是 真的有这种需求吗？)
	Copy() Context
	Unsafe() *EventContext // 直接访问 EventContext 的字段 无法保证其并发安全
	// Block 是否中断
	Block() bool
	// SetBlock 不需要执行下一个 Handler 时使用
	SetBlock()

	// 缓存

	// Set 写入缓存 map 其生命周期与 Context 一起结束
	Set(key string, val any)
	// Get 从缓存（或者 Store 中）获取值
	Get(key string) any
	// Bool key 存在 和 值为真 返回真；其他情况返回假
	Bool(s string) bool
	// Int 断言为 int 失败为零值
	Int(s string) int
	// String 断言为 string 失败为零值
	String(s string) string
	// Range 遍历处理缓存值
	Range(f func(k string, v any))

	// Store 写入存储 map
	Store(key string, val any)
	Save()

	// 错误处理

	// Errs 返回错误链
	Errs() error
	JoinError(err ...error)

	// 回复

	Reply(text string)        // 发送到当前频道
	SilenceReply(text string) // 同 Reply 但是不触发事件
	ReplyPrivate(text string) // 同 Reply 但是发送目标是当前用户的私信频道
	ReplyCaption(key string)

	// 交互

	// Prompt 使用 SilenceReply 发送提 tip 消息并获取当前上下文 deadline 秒内的回应消息
	//
	// 特别的是，起反馈作用的消息也不会触发事件
	Prompt(tip string, deadline int64) string
	// ConfirmCustom 是基于 Dialogue 的封装 反馈消息的文本值(会强制转为英文大写) 是 res 中的一个时时返回 true
	ConfirmCustom(tip string, deadline int64, res ...string) bool
	// Confirm 等于 ConfirmCustom(tip, 5, "YES", "Y", ".", "。")
	Confirm(tip string) bool

	IMEvent() *im.Event
	Config() utils.IConfig
	Logger() utils.ILogger
	// BeenFirstAt 被第一个 At
	BeenFirstAt() bool
	// BeenAt 被 At
	BeenAt() bool
	// Private 是否是私信环境
	Private() bool
	// Text 纯文本元素的拼接
	Text() string
	// Gid 跨平台的唯一会话ID
	Gid() string
	// Uid 跨平台的唯一用户ID
	Uid() string
}

type EventContext struct {
	Manager  *Manager
	Bot      *Bot
	Sessions *sessions
	Group    *Group
	User     *User
	Event    *im.Event

	uid         string
	gid         string
	text        string
	private     bool
	beenAt      bool
	firstBeenAt bool
	block       bool // true 已经什么都不需要做了
	beenCopied  bool
	cmp         map[string]any
	smp         map[string]any
	err         error
	cmd         Command
}

func (ctx *EventContext) Unsafe() *EventContext {
	return ctx
}

func (ctx *EventContext) Copy() Context {
	ctx.beenCopied = true
	c := &EventContext{
		Manager: ctx.Manager,
		Event:   ctx.Event,
		smp:     map[string]any{},
		cmp:     map[string]any{},
		err:     nil,
	}
	for s, a := range ctx.cmp {
		c.cmp[s] = a
	}
	for s, a := range ctx.smp {
		c.smp[s] = a
	}
	return c
}

func (ctx *EventContext) reset(s *sessions) {
	e := ctx.Event
	ctx.uid = fmt.Sprintf("%s:%s", e.Platform, e.User.Id)
	if e.Guild.Id == "" {
		ctx.private = true
		e.Guild.Id = "U"
	}
	if e.Channel.Id == "" {
		if ctx.private {
			e.Channel.Id = e.User.Id
		} else {
			e.Channel.Id = e.Guild.Id
			e.Guild.Id = "G"
		}
	}
	ctx.gid = fmt.Sprintf("%s:%s:%s", e.Platform, e.Guild.Id, e.Channel.Id)

	// --- data

	ctx.Sessions = s
	var smp map[string]any
	if s.store == nil {

	} else {
		smp = s.store.Load(ctx.Gid())
	}
	ctx.cmp = make(map[string]any, len(smp))
	for k, v := range smp {
		ctx.cmp[k] = v
	}
	ctx.smp = smp
	g, ok := s.groups[ctx.Gid()]
	if !ok {
		g = &Group{
			Guild:          ctx.Event.Guild,
			Channel:        ctx.Event.Channel,
			EndpointIdList: []string{ctx.Event.Endpoint.Id},
			sessions:       s,
		}
		s.manager.emit("xx", ctx)
		s.groups[ctx.Gid()] = g
	}
	ctx.Group = g
	u, ok := s.users[ctx.Gid()]
	if !ok {
		u = &User{
			User: ctx.Event.User,
		}
		s.manager.emit("xx", ctx)
		s.users[ctx.Uid()] = u
	}
	ctx.User = u
}

func (ctx *EventContext) BeenFirstAt() bool {
	return ctx.firstBeenAt
}

func (ctx *EventContext) BeenAt() bool {
	return ctx.beenAt
}

func (ctx *EventContext) Private() bool {
	return ctx.private
}

func (ctx *EventContext) Text() string {
	return ctx.text
}

func (ctx *EventContext) Gid() string {
	return ctx.gid
}

func (ctx *EventContext) Uid() string {
	return ctx.uid
}
func (ctx *EventContext) Block() bool {
	return ctx.block
}

func (ctx *EventContext) SetBlock() {
	ctx.block = true
}

func (ctx *EventContext) Store(k string, v any) {
	ctx.smp[k] = v
}

func (ctx *EventContext) Save() {
	if len(ctx.smp) > 0 {
		// ctx.Manager.Sessions.store.Store(ctx.Gid, ctx.smp)
	}
}

func (ctx *EventContext) Set(key string, val any) {
	if _, ok := ctx.cmp[key]; ok {
		ctx.Logger().Warn("重复写入", key)
	}
	ctx.cmp[key] = val
}

func (ctx *EventContext) Get(s string) any {
	r, ok := ctx.cmp[s]
	if ok {
		return r
	}
	return nil
}

func (ctx *EventContext) Range(f func(k string, v any)) {
	for s, a := range ctx.cmp {
		f(s, a)
	}
}

func (ctx *EventContext) Bool(s string) bool {
	if val, ok := ctx.cmp[s]; ok {
		if b, ok2 := val.(bool); ok2 {
			return b
		}
		return true
	}
	return false
}

func (ctx *EventContext) Int(s string) int {
	r, ok := ctx.Get(s).(int)
	if ok {
		return r
	}
	return 0
}

func (ctx *EventContext) String(s string) string {
	r, ok := ctx.Get(s).(string)
	if ok {
		return r
	}
	return ""
}

func (ctx *EventContext) reply(text string) {
	e := ctx.Event.CloneId()

	ctx.Event.Endpoint.Adapter.Send(e)
}

func (ctx *EventContext) SilenceReply(text string) {
	ctx.reply(text)
}

func (ctx *EventContext) Reply(text string) {
	ctx.Manager.emit(utils.SendMessageEvent, ctx)
	ctx.reply(text)
}

func (ctx *EventContext) ReplyPrivate(text string) {

}

func (ctx *EventContext) getCaption(key string) (string, bool) {
	ctx.Manager.botMutex.RLock()
	defer ctx.Manager.botMutex.RUnlock()
	if v, ok := ctx.Bot.getCaption(ctx, key); ok {
		return v, ok
	}
	return ctx.Manager.DefaultBot.getCaption(ctx, key)
}

func (ctx *EventContext) ReplyCaption(key string) {
	if ctx.cmd != nil {
		if v, ok := ctx.getCaption(ctx.cmd.Ext().Name + ":" + key); ok {
			ctx.Reply(v)
			return
		}
	}
	if v, ok := ctx.getCaption(key); ok {
		ctx.Reply(v)
		return
	}
	ctx.Reply("找不到文案：" + key)
}

func (ctx *EventContext) Prompt(tip string, second int64) string {
	ctx.SilenceReply(tip)
	ch := make(chan string)
	id := ctx.Gid()
	ctx.Manager.Sessions.dialogues.Store(id, ch)
	select {
	case r := <-ch:
		return r
	case <-time.After(time.Duration(second) * time.Second):
		ctx.Manager.Sessions.dialogues.Delete(id)
		return ""
	}
}

func (ctx *EventContext) Confirm(tip string) bool {
	return ctx.ConfirmCustom(tip, 5, "YES", "Y", ".", "。")
}

func (ctx *EventContext) ConfirmCustom(tip string, t int64, res ...string) bool {
	r := ctx.Prompt(tip, t)
	r = strings.ToUpper(r)
	if slices.Contains(res, r) {
		return true
	}
	return false
}

func (ctx *EventContext) Errs() error {
	return ctx.err
}

func (ctx *EventContext) JoinError(err ...error) {
	ctx.err = errors.Join(err...)
}

func (ctx *EventContext) Logger() utils.ILogger {
	return ctx.Manager.Logger
}

func (ctx *EventContext) Config() utils.IConfig {
	return ctx.Manager.Config
}

func (ctx *EventContext) IMEvent() *im.Event {
	return ctx.Event
}
