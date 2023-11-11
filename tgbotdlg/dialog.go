package tgbotdlg

import (
	"context"
)

// Dialog is a dialog interface
type Dialog interface {
	// Name is a unique dialog name. Used to match a dialog stored in DB with corresponding Dialog instance
	Name() string

	// HandleChatUpdate is called on current dialog when an update is received from a chat
	HandleChatUpdate(ctx context.Context, upd ChatUpdate) (*Switch, error)

	// HandleForeignEvent is called on current dialog when a foreign event is received to the dispatcher
	HandleForeignEvent(ctx context.Context, chatID int64, event any) (*Switch, error)

	// OnStart is called when this dialog is started
	// Note. If this func returned an error, dispatcher does not make this dialog a current dialog
	OnStart(ctx context.Context, chatID int64) error

	// OnFinish is called when this dialog is about to be finished.
	// Note. If this func returned an error, a current active dialog remains active
	OnFinish(ctx context.Context, chatID int64) error

	newState() any
}

// StatelessDialog should be embedded to a dialog which don't have a initialState
type StatelessDialog = StatefulDialog[struct{}]

// StatefulDialog should be embedded to a dialog which has its state
type StatefulDialog[State any] struct {
	noopImpl
}

func (d StatefulDialog[State]) dialog(State) {}

func (d StatefulDialog[State]) newState() any {
	return new(State)
}

// State returns a non-nil pointer to a mutable initialState of the current dialog.
func (d StatefulDialog[State]) State(ctx context.Context) *State {
	return stateFromCtx(ctx).(*State)
}

type dialog[State any] interface {
	Dialog
	dialog(State)
}

type noopImpl struct{}

func (noopImpl) OnStart(context.Context, int64) error {
	return nil
}

func (noopImpl) OnFinish(context.Context, int64) error {
	return nil
}

func (noopImpl) HandleForeignEvent(context.Context, int64, any) (*Switch, error) {
	return nil, nil
}
