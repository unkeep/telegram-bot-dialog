package dialog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewDispatcher(storage Storage) *Dispatcher {
	return &Dispatcher{storage: storage}
}

// Dispatcher routes users messages to an appropriate dialog
type Dispatcher struct {
	storage     Storage
	register    map[string]func() Dialog
	rootDlgName string
}

// RegisterDialogs registers dialogs which can receive users messages.
// Each func must return a pointer to a new dialog instance with a unique and constant Name
// The first registered dialog (root dialog) is the entry point for users messages if a conversation has not yet been
// initiated by the user.
func (e *Dispatcher) RegisterDialogs(newRootFunc func() Dialog, newDlgFuncs ...func() Dialog) {
	e.register = make(map[string]func() Dialog, len(newDlgFuncs)+1)

	rootDlg := newRootFunc()
	e.rootDlgName = rootDlg.Name()
	e.register[rootDlg.Name()] = newRootFunc

	for _, newDlgFunc := range newDlgFuncs {
		d := newDlgFunc()
		if _, exist := e.register[d.Name()]; exist {
			panic(fmt.Sprintf("dublicate dialog name found - %q", d.Name()))
		}
		e.register[d.Name()] = newDlgFunc
	}
}

// HandleUpdate handles an update from telegram chat.
// Note: it currently ignores other updates than users messages or inline buttons clicks
func (e *Dispatcher) HandleUpdate(ctx context.Context, upd *tgbotapi.Update) error {
	var chatID, userID int64

	var handler func(Dialog) (Dialog, error)

	switch {
	case upd.Message != nil:
		chatID, userID = upd.Message.Chat.ID, upd.Message.From.ID
		handler = func(dialog Dialog) (Dialog, error) {
			newDlg, err := dialog.OnMessage(ctx, upd.UpdateID, upd.Message)
			if err != nil {
				return nil, fmt.Errorf("OnMessage: %w", err)
			}
			return newDlg, nil
		}
	case upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil:
		chatID, userID = upd.CallbackQuery.Message.Chat.ID, upd.CallbackQuery.From.ID
		handler = func(dialog Dialog) (Dialog, error) {
			newDlg, err := dialog.OnCallbackQuery(ctx, upd.UpdateID, upd.CallbackQuery)
			if err != nil {
				return nil, fmt.Errorf("OnCallbackQuery: %w", err)
			}
			return newDlg, nil
		}
	default:
		return nil
	}

	data, err := e.storage.GetDialog(chatID, userID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("storage.GetDialog: %w", err)
		}
		data = &Data{Name: e.rootDlgName}
	}

	dlg, err := e.newDialog(data.Name)
	if err != nil {
		return fmt.Errorf("newDialog: %w", err)
	}

	if len(data.State) > 0 && string(data.State) != "{}" {
		if err := json.Unmarshal(data.State, dlg); err != nil {
			return fmt.Errorf("unmarshal dialog stateJSON: %w", err)
		}
	}

	newDlg, err := handler(dlg)
	if err != nil {
		return fmt.Errorf("dialog %s: %w", dlg.Name(), err)
	}

	newStateJSON, err := json.Marshal(newDlg)
	if err != nil {
		return fmt.Errorf("marshal new dialog stateJSON: %w", err)
	}

	newData := &Data{
		Name:  newDlg.Name(),
		State: newStateJSON,
	}

	if data != nil && newData.IsEquivalent(data) {
		// old and new dialogs are equivalent, no sense to update it in storage
		return nil
	}

	if err := e.storage.SaveDialog(chatID, userID, newData); err != nil {
		return fmt.Errorf("storage.SaveDialog: %w", err)
	}

	return nil
}

func (e *Dispatcher) newDialog(name string) (Dialog, error) {
	newFunc, ok := e.register[name]
	if !ok {
		return nil, fmt.Errorf("unregistered dialog %q", name)
	}

	return newFunc(), nil
}

func (d *Data) IsEquivalent(other *Data) bool {
	if d.Name != other.Name {
		return false
	}

	stateStr := string(d.State)
	if stateStr == "{}" {
		stateStr = ""
	}
	otherStateStr := string(other.State)
	if otherStateStr == "{}" {
		otherStateStr = ""
	}

	return stateStr == otherStateStr
}
