package onebot11

import (
	"context"
	"foxdice/endpoints/im"
)

func (a *Adapter) websocketReverse(ctx context.Context, ch chan<- *im.Event) {
	a.Endpoint.Info("not impl")
}
