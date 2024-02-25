package core

import (
	"foxdice/utils/str"
)

func NewCaption(text string, note string, categories []string, example []string) (*Caption, error) {
	c := &Caption{
		Note:       note,
		Categories: categories,
		Example:    example,
	}
	err := c.Reset(text, "{", "}")
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Caption struct {
	str.StringInterpolator
	Note       string   // 备注
	Categories []string // 分类
	Example    []string
}

type captionHook = func(ctx Context) string

type Bot struct {
	Id             int
	Enable         bool
	Name           string
	EndpointIdList []string
	captions       map[string]*Caption
	captionHooks   map[string]captionHook
	m              *Manager
}

// 记得加锁
func (b *Bot) getCaption(ctx Context, key string) (string, bool) {
	caption, ok := b.captions[key]
	if !ok {
		return "", ok
	}
	return caption.Execute(func(name string) string {
		if name == key {
			return "异常:自我调用的文案:" + key
		}
		if f, ok := b.captionHooks[name]; ok {
			return f(ctx)
		}
		if v, ok := b.getCaption(ctx, name); ok {
			return v
		}
		return "{" + name + "}"
	}), true
}

func (b *Bot) RegCaptionHooks(m map[string]captionHook) {
	b.m.botMutex.Lock()
	b.captionHooks = m
	b.m.botMutex.Unlock()
}

func (b *Bot) ResetCaptions(m map[string]*Caption) {
	b.m.botMutex.Lock()
	b.captions = m
	b.m.botMutex.Unlock()
}
