package dialog

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Dialog interface {
	Name() string
	OnMessage(ctx context.Context, updateID int, m *tgbotapi.Message) (Dialog, error)
	OnCallbackQuery(ctx context.Context, updateID int, q *tgbotapi.CallbackQuery) (Dialog, error)
}
