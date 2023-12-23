package proxy

import (
	"context"
	"time"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	types2 "github.com/acorn-io/brent/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/request"
)

type WatchRefresh struct {
	types2.Store
	asl accesscontrol.AccessSetLookup
}

func (w *WatchRefresh) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	user, ok := request.UserFrom(apiOp.Context())
	if !ok {
		return w.Store.Watch(apiOp, schema, wr)
	}

	as := w.asl.AccessFor(user)
	ctx, cancel := context.WithCancel(apiOp.Context())
	apiOp = apiOp.WithContext(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}

			newAs := w.asl.AccessFor(user)
			if as.ID != newAs.ID {
				// RBAC changed
				cancel()
				return
			}
		}
	}()

	return w.Store.Watch(apiOp, schema, wr)
}
