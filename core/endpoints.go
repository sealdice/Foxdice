package core

import (
	"context"
	"errors"
	"foxdice/endpoints/cli"
	"foxdice/endpoints/discord"
	"foxdice/endpoints/im"
	"foxdice/utils"
	"foxdice/utils/str"

	"golang.org/x/exp/slices"
)

type Hub struct {
	Publish               chan *im.Event
	Source                chan *im.Event
	Endpoints             []*im.Endpoint
	startCompletionSignal chan struct{}
	serveCtx              context.Context
	Config                utils.IConfig
	Logger                utils.ILogger
}

func (h *Hub) Serve(ctx context.Context) {
	h.serveCtx = ctx
	var list []im.LoginInfo
	err := h.Config.Unmarshal("list", &list)
	if err != nil {
		h.Logger.Error(err)
	}

	cliEp := &im.Endpoint{}
	cli.New(cliEp)
	go cliEp.Adapter.Serve(ctx, h.Source)

	for _, b := range list {
		ep := new(im.Endpoint)
		ep.LoginInfo = b
		ep.IConfig = h.Config
		ep.ILogger = h.Logger
		switch ep.Platform {
		case im.Discord:
			discord.Attach(ep)
		}
		h.Endpoints = append(h.Endpoints, ep)
		go func() {
			defer utils.ErrorPrint()
			ep.Adapter.Serve(ctx, h.Source)
		}()
	}

	h.startCompletionSignal <- struct{}{}

	for msg := range h.Source {
		h.Publish <- msg
	}
}

func (h *Hub) Add(platform string) *im.Endpoint {
	ep := new(im.Endpoint)
	ep.Id = str.UUID()
	ep.Platform = platform
	h.Endpoints = append(h.Endpoints, ep)
	go func() {
		defer utils.ErrorPrint()
		ep.Adapter.Serve(h.serveCtx, h.Source)
	}()
	return ep
}

func (h *Hub) Del(id string) error {
	i := slices.IndexFunc(h.Endpoints, func(endpoint *im.Endpoint) bool {
		return endpoint.Id == id
	})
	ep := h.Endpoints[i]
	ep.Adapter.Close()
	h.Endpoints = slices.Delete(h.Endpoints, i, i+1)
	return nil
}

func (h *Hub) ById(id string) (*im.Endpoint, error) {
	for _, ep := range h.Endpoints {
		if ep.Id == id {
			return ep, nil
		}
	}
	return nil, errors.New("not found")
}
