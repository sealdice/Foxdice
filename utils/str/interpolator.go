package str

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// StringInterpolator 静态模板的字符串插值 不允许嵌套结构
type StringInterpolator struct {
	Template string
	texts    []string
	keys     []string
	mu       sync.Mutex
	sb       strings.Builder
	done     uint32
}

func (s *StringInterpolator) Reset(str, left, right string) error {
	s.Template = str
	li := 0
	ri := 0
	for {
		li = strings.Index(str, left)
		if li == -1 {
			if str != "" {
				s.texts = append(s.texts, str)
			}
			break
		}
		s.texts = append(s.texts, str[:li])
		str = str[li+1:]
		ri = strings.Index(str, right)
		if ri == -1 {
			return fmt.Errorf("缺少对应的右标记: %s <", str[:li])
		}
		nli := strings.Index(str, left)
		if nli < ri && nli != -1 {
			return fmt.Errorf("不允许嵌套: %s <", str[:nli])
		}
		s.keys = append(s.keys, str[:ri])
		str = str[ri+1:]
	}
	return nil
}

// Try 是否无需执行就能获得结果
func (s *StringInterpolator) Try() (bool, string) {
	if len(s.keys) == 0 {
		return true, s.Template
	}
	return false, ""
}

// Once 对于同一个实例 不要同时调用 Once 和 Execute
func (s *StringInterpolator) Once(f func(name string) string) string {
	if atomic.LoadUint32(&s.done) == 0 {
		defer atomic.StoreUint32(&s.done, 1) // 要不要 keys = nil ？
		return s.Execute(f)
	}
	return s.sb.String()
}

func (s *StringInterpolator) Execute(f func(name string) string) string {
	if ok, v := s.Try(); ok {
		return v
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_l := s.sb.Len()
	s.sb.Reset()
	// sb.Reset 是内部直接赋值成 nil 来重置值
	// 并发下如果不提前 lock 就有可能出现这边刚重置为 nil 那边就拿来用了然后 panic
	// 教训：以后干脆无脑 lock defter unlock（划掉）不要假设一个方法并发安全
	if _l > 0 {
		s.sb.Grow(_l)
	} else {
		s.sb.Grow(len(s.Template))
	}
	kl, tl := len(s.keys), len(s.texts)
	for i := 0; i < tl; i++ {
		s.sb.WriteString(s.texts[i])
		if i < kl {
			s.sb.WriteString(f(s.keys[i]))
		}
	}
	return s.sb.String()
}
