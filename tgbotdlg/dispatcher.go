package tgbotdlg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewDispatcher(storage Storage, rootDlg Dialog, opts ...optionFunc) *Dispatcher {
	builtOpts := buildOpts(opts)

	wrapDlg := func(original Dialog) Dialog {
		if len(builtOpts.interceptors) > 0 {
			return &dialogWithInterceptors{
				original:     original,
				interceptors: builtOpts.interceptors,
			}
		}
		return original
	}

	dialogs := make(map[string]Dialog, len(builtOpts.dialogs)+1)
	dialogs[rootDlg.Name()] = wrapDlg(rootDlg)

	for _, dlg := range builtOpts.dialogs {
		if _, exist := dialogs[dlg.Name()]; exist {
			panic(fmt.Sprintf("dublicate dialog name found - %q", dlg.Name()))
		}

		dialogs[dlg.Name()] = wrapDlg(dlg)
	}

	offChatUpdateHandler := builtOpts.offChatUpdateHandler
	if offChatUpdateHandler == nil {
		offChatUpdateHandler = func(context.Context, OffChatUpdate) error { return nil }
	}

	return &Dispatcher{
		storage:              storage,
		dialogs:              dialogs,
		rootDlg:              rootDlg,
		offChatUpdateHandler: offChatUpdateHandler,
	}
}

// Dispatcher routes users messages to an appropriate dialog
type Dispatcher struct {
	storage              Storage
	dialogs              map[string]Dialog
	rootDlg              Dialog
	offChatUpdateHandler func(context.Context, OffChatUpdate) error
}

// HandleBotUpdate handles an update of telegram bot
// in-chat update is handled by active dialog, off-chat update - by optionally provided handler
func (d *Dispatcher) HandleBotUpdate(ctx context.Context, upd tgbotapi.Update) error {
	botUpd := botUpdate(upd)

	chatID, ok := botUpd.chatID()
	if !ok {
		if err := d.offChatUpdateHandler(ctx, botUpd.toOffChatUpdate()); err != nil {
			return fmt.Errorf("handle off-chat update: %w", err)
		}
		return nil
	}

	handler := func(ctx context.Context, dlg Dialog) (*Switch, error) {
		s, err := dlg.HandleChatUpdate(ctx, botUpd.toChatUpdate())
		if err != nil {
			return nil, fmt.Errorf("HandleChatUpdate: %w", err)
		}
		return s, nil
	}

	return d.handleDialogEvent(ctx, chatID, handler)
}

func (d *Dispatcher) HandleForeignEvent(ctx context.Context, chatID int64, event any) error {
	handler := func(ctx context.Context, dlg Dialog) (*Switch, error) {
		s, err := dlg.HandleForeignEvent(ctx, chatID, event)
		if err != nil {
			return nil, fmt.Errorf("HandleForeignEvent: %w", err)
		}
		return s, nil
	}

	return d.handleDialogEvent(ctx, chatID, handler)
}

func (d *Dispatcher) ForceDialog(ctx context.Context, chatID int64, switchTo Switch) error {
	dlg, state, err := d.getCurrentDialogAndState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get current dialog: %w", err)
	}

	if err := d.switchDialog(ctx, chatID, dlg, state, switchTo); err != nil {
		return fmt.Errorf("force switch %q dialog to %q: %w", dlg.Name(), switchTo.name, err)
	}

	return nil
}

func (d *Dispatcher) handleDialogEvent(ctx context.Context, chatID int64, handler func(context.Context, Dialog) (*Switch, error)) error {
	dlg, state, err := d.getCurrentDialogAndState(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get current dialog: %w", err)
	}

	stateRaw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("dialog %q: marhsal state: %w", dlg.Name(), err)
	}

	ctx = ctxWithState(ctx, state)
	switchTo, err := handler(ctx, dlg)
	if err != nil {
		return fmt.Errorf("dialog %q: handle event: %w", dlg.Name(), err)
	}

	if switchTo != nil {
		if err := d.switchDialog(ctx, chatID, dlg, state, *switchTo); err != nil {
			return fmt.Errorf("switch %q dialog to %q: %w", dlg.Name(), switchTo.name, err)
		}
		return nil
	}

	newStateRaw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("dialog %q: marhsal state: %w", dlg.Name(), err)
	}

	if bytes.Equal(stateRaw, newStateRaw) {
		// nothing changed
		return nil
	}

	newData := Data{
		Name:  dlg.Name(),
		State: newStateRaw,
	}
	if err := d.storage.SaveDialog(ctx, chatID, newData); err != nil {
		return fmt.Errorf("storage.SaveDialog: %w", err)
	}

	return nil
}

func (d *Dispatcher) switchDialog(ctx context.Context, chatID int64, dlg Dialog, state any, switchTo Switch) error {
	nextDlg, ok := d.dialogs[switchTo.name]
	if !ok {
		return fmt.Errorf("dialog %q: unregistered", switchTo.name)
	}

	ctx = ctxWithState(ctx, state)
	if err := dlg.OnFinish(ctx, chatID); err != nil {
		return fmt.Errorf("dialog %q finish: %w", dlg.Name(), err)
	}

	nextState := switchTo.initialState
	ctx = ctxWithState(ctx, nextState)

	if err := nextDlg.OnStart(ctx, chatID); err != nil {
		return fmt.Errorf("dialog %q start: %w", switchTo.name, err)
	}

	nextStateRaw, err := json.Marshal(nextState)
	if err != nil {
		return fmt.Errorf("dialog %q: marhsal state: %w", nextDlg.Name(), err)
	}

	nextData := Data{
		Name:  nextDlg.Name(),
		State: nextStateRaw,
	}

	if err := d.storage.SaveDialog(ctx, chatID, nextData); err != nil {
		return fmt.Errorf("storage.SaveDialog: %w", err)
	}

	return nil
}

func (d *Dispatcher) getCurrentDialogAndState(ctx context.Context, chatID int64) (Dialog, any, error) {
	data, err := d.storage.GetDialog(ctx, chatID)
	if err != nil {
		return nil, nil, fmt.Errorf("storage.GetDialog: %w", err)
	}

	if data == nil {
		// absence of dialog data for a chat is considered as a root dialog is active atm.
		return d.rootDlg, d.rootDlg.newState(), nil
	}

	dlg, ok := d.dialogs[data.Name]
	if !ok {
		return nil, nil, fmt.Errorf("dialog %q: unregistered", data.Name)
	}

	state := dlg.newState()
	if err := json.Unmarshal(data.State, state); err != nil {
		return nil, nil, fmt.Errorf("dialog %q: unmarshal state: %w", dlg.Name(), err)
	}

	return dlg, state, nil
}
