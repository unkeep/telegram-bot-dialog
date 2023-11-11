package tgbotdlg

import (
	"context"
)

type Interceptor interface {
	HandleChatUpdate(ctx context.Context, name string, state any, upd ChatUpdate, next func(context.Context, ChatUpdate) (*Switch, error)) (*Switch, error)
	HandleForeignEvent(ctx context.Context, name string, state any, chatID int64, event any, next func(context.Context, int64, any) (*Switch, error)) (*Switch, error)
	OnStart(ctx context.Context, name string, state any, chatID int64, next func(context.Context, int64) error) error
	OnFinish(ctx context.Context, name string, state any, chatID int64, next func(context.Context, int64) error) error
}

type dialogWithInterceptors struct {
	original     Dialog
	interceptors []Interceptor
}

func (d *dialogWithInterceptors) Name() string {
	return d.original.Name()
}

func (d *dialogWithInterceptors) HandleChatUpdate(ctx context.Context, upd ChatUpdate) (*Switch, error) {
	name := d.original.Name()
	state := stateFromCtx(ctx)
	next := d.original.HandleChatUpdate

	for i := len(d.interceptors); i >= 0; i-- {
		interceptor := d.interceptors[i]

		next = func(ctx context.Context, upd ChatUpdate) (*Switch, error) {
			return interceptor.HandleChatUpdate(ctx, name, state, upd, next)
		}
	}

	return next(ctx, upd)
}

func (d *dialogWithInterceptors) HandleForeignEvent(ctx context.Context, chatID int64, event any) (*Switch, error) {
	name := d.original.Name()
	state := stateFromCtx(ctx)
	next := d.original.HandleForeignEvent

	for i := len(d.interceptors); i >= 0; i-- {
		interceptor := d.interceptors[i]

		next = func(ctx context.Context, chatID int64, event any) (*Switch, error) {
			return interceptor.HandleForeignEvent(ctx, name, state, chatID, event, next)
		}
	}

	return next(ctx, chatID, event)
}

func (d *dialogWithInterceptors) OnStart(ctx context.Context, chatID int64) error {
	name := d.original.Name()
	state := stateFromCtx(ctx)
	next := d.original.OnStart

	for i := len(d.interceptors); i >= 0; i-- {
		interceptor := d.interceptors[i]

		next = func(ctx context.Context, chatID int64) error {
			return interceptor.OnStart(ctx, name, state, chatID, next)
		}
	}

	return next(ctx, chatID)
}

func (d *dialogWithInterceptors) OnFinish(ctx context.Context, chatID int64) error {
	name := d.original.Name()
	state := stateFromCtx(ctx)
	next := d.original.OnFinish

	for i := len(d.interceptors); i >= 0; i-- {
		interceptor := d.interceptors[i]

		next = func(ctx context.Context, chatID int64) error {
			return interceptor.OnFinish(ctx, name, state, chatID, next)
		}
	}

	return next(ctx, chatID)
}

func (d *dialogWithInterceptors) newState() any {
	return d.original.newState()
}
