package tgbotdlg

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Update is a shorter version of tgbotapi.Update containing only supported update kinds for a dialog interaction
type Update struct {
	// ID is tgbotapi.UpdateID.
	ID int

	// Message new incoming message of any kind — text, photo, sticker, etc.
	// optional
	Message *tgbotapi.Message

	// CallbackQuery new incoming callback query
	// optional
	CallbackQuery *tgbotapi.CallbackQuery
}

// Dialog is a dialog between your bot and user.
// It's responsible for handling updates in certain point and may have a state
type Dialog interface {
	// Name is a unique dialog name. Used to match a dialog stored in DB with an appropriate Dialog instance
	Name() string

	// HandleUpdate called for a current dialog if an update received from chat
	// Should return a new Dialog instance or the same instance if user staying on the dialog
	HandleUpdate(ctx context.Context, upd Update) (Dialog, error)
}
