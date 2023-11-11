package tgbotdlg

import "context"

type stateCtxKey struct{}

func ctxWithState(ctx context.Context, state any) context.Context {
	return context.WithValue(ctx, stateCtxKey{}, state)
}

func stateFromCtx(ctx context.Context) any {
	return ctx.Value(stateCtxKey{})
}
