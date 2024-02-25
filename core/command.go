package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Command interface {
	Action(ctx Context)
	Help(ctx Context) string
	Name() string
	Ext() *Extension
	// Try 输入是以指令名或者指令别名为前缀的字符串
	Try(text string) (ok bool, argsText string)
	// Parse 解析参数 并使用 Context.Set 将值写入上下文缓存
	Parse(ctx Context, argsText string)
}

const (
	IntParamType = "int"
	OptParamType = "opt"
	AllParamType = "all"
	IDParamType  = "id"
	AtParamType  = "at"
)

var (
	ErrOptParam = errors.New("未知的选项")
	ErrIntParam = errors.New("无法转换为整数")
)

type Param struct {
	IsOmit  bool
	Min     int64
	Max     int64
	Key     string
	Type    string
	Show    string
	Opts    []string
	Default any
}

type command struct {
	name string
	act  func(ctx Context)
	help string
	subs map[string]Command

	aliases []string
	params  []Param
	ext     *Extension
	root    *command

	onlyGroup bool // 指令只在群中可用
}

func (c *command) Name() string {
	if c.root != nil {
		return c.root.name + "/" + c.name
	}
	return c.name
}

func (c *command) Action(ctx Context) {
	c.act(ctx)
}

func (c *command) Help(ctx Context) string {
	return ""
}

func (c *command) Ext() *Extension {
	return c.ext
}

func (c *command) Try(text string) (bool, string) {
	if after, ok := strings.CutPrefix(text, c.name); ok {
		return ok, after
	}
	for _, alias := range c.aliases {
		if after, ok := strings.CutPrefix(text, alias); ok {
			return ok, after
		}
	}
	return false, ""
}

func (c *command) Parse(ctx Context, text string) {
	r := ctx.Unsafe()
	if c.onlyGroup && r.Private() {
		ctx.JoinError(fmt.Errorf("指令 %s 只在群组中可用", c.name))
		return
	}

	if r.err != nil {
		ctx.Logger().Error(r.err)
		r.err = nil // 感觉不太好 要不暂存一下？
	}

	if c.subs != nil {
		for _, sub := range c.subs {
			if ok, after := sub.Try(text); ok {
				sub.Parse(ctx, after)
				return
			}
		}
	}

	inputs := strings.Split(text, " ")
	li := len(inputs)   // 输入参数的数量
	lp := len(c.params) // 定义参数的数量

	for i, p := range c.params {
		// 定义的参数比输入的多 后面都是可忽略的参数？
		if i >= li {
			if p.IsOmit {
				ctx.Set(p.Key, p.Default)
			} else {
				ctx.JoinError(fmt.Errorf("缺少必须参数 <%s>", p.Key))
				return
			}
			continue
		}

		input := inputs[i]

		if input == "help" && i == 0 {
			ctx.Set("help", true) // 所有指令的第一个子命令都是帮助信息
			return
		}

		switch p.Type {
		case OptParamType:
			if slices.Contains(p.Opts, input) {
				ctx.Set(p.Key, input)
			} else {
				ctx.JoinError(ErrOptParam, fmt.Errorf("选项只有 %v", p.Opts))
			}
		case AllParamType:
			ctx.Set(p.Key, text)
			return
		case IntParamType:
			parseInt, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				ctx.JoinError(ErrIntParam, err)
			}
			if parseInt < p.Min || parseInt > p.Max {
				ctx.JoinError(fmt.Errorf("超出参数许可范围 %d ~ %d", p.Min, p.Max))
			}
			ctx.Set(p.Key, parseInt)
		default:
			// 想自行解析就随便写点占个位置
			ctx.Set(p.Key, input)
		}

		if r.err != nil && i+1 <= lp {
			if c.params[i+1].IsOmit {
				r.err = nil // 对于连续可省略的参数 解析错误时使用默认值 不传出错误
			}
		}
	}

	if li > lp {
		polymeric := inputs[lp:]
		for _, s := range polymeric {
			if after, ok := strings.CutPrefix(s, "--"); ok {
				key, val, _ := strings.Cut(after, "=")
				if key == "h" {
					ctx.Set("help", true)
				}
				ctx.Set(key, val)
			}
		}
	}
}

func (c *command) Sub(name string, alias ...string) *command {
	if c.subs == nil {
		c.subs = make(map[string]Command)
	}
	cmd := new(command)
	cmd.SetAlias(alias...)
	cmd.root = c
	c.subs[name] = cmd
	return cmd
}

func (c *command) SetAlias(s ...string) {
	c.aliases = append(c.aliases, s...)
}

func (c *command) SetParam(ps []Param) {
	c.params = ps
}

func (c *command) SetOnlyGroup() {
	c.onlyGroup = true
}

func (c *command) SetAct(act func(ctx Context)) {
	c.act = act
}

type Cmd = command

// 本来是想包装成链式调用的 但是太不灵活了 fmt 之后的格式也很难看

func NewCmd(group string, name string, aliases ...string) *Cmd {
	// return _m.Commander.NewCmd(group, name, aliases...)
	return &Cmd{}
}

// BuildParams
/*
快速生成 Param 一般格式 key:type?default

例子：

全选（不以空格区分参数） all

	r exp:all?1d100
	r 1d6 + 20 // exp =  "1d6 + 20"
	r // exp = "1d100"

选项和数字 opt<o1,o2,o3> int<min,max>

	name country:opt<zh,jp,en>?zh num:int<1,10>?5 // 定义随机名字指令 参数地区、生成数
	name // country = zh , num = 5
	name en 10 // country = en , num = 10

不提供默认值时为必须参数

	st val:all
	st xx10xx10 // val = "xx10xx10"
	st // 报错

*/
func BuildParams(params ...string) []Param {
	var r []Param
	for _, s := range params {
		p := Param{}
		if i := strings.Index(s, "?"); i > -1 {
			p.IsOmit = true
			p.Default = s[i+1:]
			s = s[:i]
		}
		if i := strings.Index(s, ":"); i > -1 {
			p.Type = s[:i+1]
			s = s[:i]
		}
		p.Key = s

		if strings.HasPrefix(p.Type, OptParamType) {
			opt := p.Type[len(OptParamType)+1 : len(p.Type)-1]
			ls := strings.Split(opt, ",")
			p.Opts = ls
			p.Type = OptParamType
		}

		if strings.HasPrefix(p.Type, IntParamType) {
			opt := p.Type[len(IntParamType)+1 : len(p.Type)-1]
			ls := strings.Split(opt, ",")
			pMin, err := strconv.ParseInt(ls[0], 10, 64)
			if err != nil {
				panic(err)
			}
			pMax, err := strconv.ParseInt(ls[1], 10, 64)
			if err != nil {
				panic(err)
			}
			p.Min = pMin
			p.Max = pMax
			p.Type = IntParamType
		}

		r = append(r, p)
	}
	return r
}

func (c *command) BuildParam(s ...string) {
	c.params = append(c.params, BuildParams(s...)...)
}

func (c *command) AddCaptions(tags []string, m map[string]*Caption) {
	prefix := c.ext.Name + ":"
	prefix = strings.ToUpper(prefix)
	for s, caption := range m {
		_m.DefaultBot.captions[prefix+s] = caption
	}
}

func (c *command) AppendNumParam(key string, def, min, max int64) *Cmd {
	c.params = append(c.params, Param{
		Type:    IntParamType,
		Key:     key,
		Default: def,
		Max:     max,
		Min:     min,
	})
	return c
}

func (c *command) AppendAllParam(key string, def string) *Cmd {
	c.params = append(c.params, Param{
		Type:    AllParamType,
		Key:     key,
		Default: def,
	})
	return c
}

func (c *command) AppendOptParam(key, def string, opt ...string) *Cmd {
	c.params = append(c.params, Param{
		Type:    OptParamType,
		Key:     key,
		Default: def,
		Opts:    opt,
	})
	return c
}
