package core

import (
	"foxdice/endpoints/im"
	"foxdice/utils/syncx"
	"sync"
)

type Group struct {
	Guild          im.Guild
	Channel        im.Channel
	Extensions     []*Extension
	EndpointIdList []string

	sessions *sessions
}

func (g *Group) BindEndpoint(ep *im.Endpoint) {
	g.EndpointIdList = append(g.EndpointIdList, ep.Id)
}

func (g *Group) setExtEnable(name string, b bool) {
	for _, extension := range g.Extensions {
		if extension.Name == name {
			extension.Enable = b
		}
	}
}

func (g *Group) EnableExt(name string) {
	g.setExtEnable(name, true)
}

func (g *Group) DisableExt(name string) {
	g.setExtEnable(name, false)
}

type User struct {
	im.User
}

type sessionStore interface {
	Load(id string) map[string]any
	Store(id string, m map[string]any)
}

// 缓存数据
// 协调上下文之间的交互（如 prompt）
type sessions struct {
	mu sync.Mutex

	dialogues syncx.RWMap[string, chan string] // 多协程写入单协程读取
	caches    map[string]*syncx.Map[string, any]
	groups    map[string]*Group
	users     map[string]*User

	store   sessionStore
	manager *Manager
}

func (s *sessions) tryContinuousDialog(id, text string) bool {
	// TODO: 有可能要求的输入是一张图片
	if text == "" {
		return false
	}
	if ch, ok := s.dialogues.Load(id); ok {
		ch <- text
		s.dialogues.Delete(id)
		return true
	}
	return false
}
