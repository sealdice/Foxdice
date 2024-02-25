package str

import (
	"testing"
)

func TestStringInterpolator_Reset(t *testing.T) {
	l := []string{
		"x1{x2}x3{x4}x5{x6}x7",
		"{x1}x2",
		"x1{x2}",
		"{x1}",
		"x1",
		"",
		//
		"}x1",
		"{x1}}",
		//
		"x1{x2}x3{x4}x5{x6x7",
		"xxx{x0{x1}}x2",
	}
	for i, txt := range l {
		s := StringInterpolator{}
		err := s.Reset(txt, "{", "}")
		switch i {
		case 0:
			if len(s.keys) != 3 {
				panic(txt)
			}
		case 1, 2, 3, 7:
			if len(s.keys) != 1 {
				panic(txt)
			}
		case 4, 5, 6:
			if len(s.keys) != 0 {
				panic(txt)
			}
		default:
			if err == nil {
				panic(txt)
			}
		}
	}
}

func TestStringInterpolator_Execute(t *testing.T) {
	s := StringInterpolator{}
	_ = s.Reset("x1{x2}x3{x4}x5{x6}x7", "{", "}")
	txt := "x1{Y}x3{Y}x5{Y}x7"
	// 没有并发的单测吗？
	for _ = range make([]struct{}, 100*100) {
		go func() {
			ret := s.Execute(func(name string) string {
				return "{Y}"
			})
			if ret != txt {
				panic(ret)
			}
		}()
	}
}
