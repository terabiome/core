package workflow

import (
	"context"
	"uuid"
)

type Dispatch struct {
	UUID    uuid.UUID
	UOWUUID uuid.UUID
}

type dispatchContextKey struct{}

func WithDispatch(ctx context.Context, dispatch Dispatch) context.Context {
	if dispatch == (Dispatch{}) {
		return ctx
	}
	return context.WithValue(ctx, dispatchContextKey{}, dispatch)
}

func DispatchFromContext(ctx context.Context) (Dispatch, bool) {
	dispatch, ok := ctx.Value(dispatchContextKey{}).(Dispatch)
	return dispatch, ok
}
