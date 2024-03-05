package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"foxdice/endpoints/im"
)

func New() (string, im.AdapterBuilder) {
	return im.Cli, func(ep *im.Endpoint) *im.Endpoint {
		ep.LoginInfo = im.LoginInfo{
			Status:   im.Offline,
			SelfId:   im.Cli,
			Name:     im.Cli,
			Platform: im.Cli,
		}
		ep.Adapter = &Adapter{
			EmptyAdapter: im.EmptyAdapter{Endpoint: ep},
			uid:          "1",
			cid:          "1",
			gid:          "1",
		}
		return ep
	}
}

type Adapter struct {
	im.EmptyAdapter
	uid string
	cid string
	gid string
	p   string
}

func (c *Adapter) Serve(ctx context.Context, ch chan<- *im.Event) {
	r := bufio.NewScanner(os.Stdin)
	for r.Scan() {
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
		}
		m := c.NewEvent()
		m.Guild.Id = c.gid
		m.User.Id = c.uid
		m.Channel.Id = c.cid
		m.Message.Content = r.Text()
		ch <- m
	}
}

func (c *Adapter) Send(msg *im.Event) {
	fmt.Println(msg.Message.Content)
}

var cmd = []string{"e", "uid", "gid", "sid", "p", "c"}

func (c *Adapter) Exec(event *im.Event) string {
	t := event.Message.Content
	switch t {
	case "s":
		fmt.Printf("UID=%s SID=%s GID=%s\n", event.User.Id, event.Channel.Id, event.Guild.Id)
	}
	for _, s := range cmd {
		if after, ok := strings.CutPrefix(t, s); ok {
			after = strings.TrimSpace(after)
			switch s {
			case "c":
				return after
			case "p":
				event.Platform = after
			case "e":
				fmt.Println("[ECHO]", after)
			case "uid":
				c.uid = after
			case "gid":
				c.gid = after
			case "sid":
				c.cid = after
			}
		}
	}

	fmt.Print("> ")
	return ""
}
