# telegram-bot-dialog
Simple library for managing Telegram bot stateful dialogs

## Example

First, ensure the library is installed and up to date by running `go get -u github.com/unkeep/telegram-bot-dialog`.

This is a very simple bot that asks username and password in a dialog manner

```go
package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dialog "github.com/unkeep/telegram-bot-dialog"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	var dialogsStorage dialog.Storage // TODO: provide dialogs storage

	dialogsDispatcher := dialog.NewDispatcher(dialogsStorage)
	dialogsDispatcher.RegisterDialogs(
		func() dialog.Dialog {
			return &rootDialog{bot: bot}
		},
		func() dialog.Dialog {
			return &enterUserNameDialog{bot: bot}
		},
		func() dialog.Dialog {
			return &enterPasswordDialog{bot: bot}
		},
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if err := dialogsDispatcher.HandleUpdate(context.Background(), &update); err != nil {
			log.Panic(err)
		}
	}
}

type rootDialog struct {
	bot *tgbotapi.BotAPI
}

func (d *rootDialog) Name() string {
	return "root"
}

func (d *rootDialog) OnMessage(ctx context.Context, updateID int, msg *tgbotapi.Message) (dialog.Dialog, error) {
	if msg.Text == "/start" {
		if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Hi! I'm a bot")); err != nil {
			return nil, err
		}
		return d, nil // stay on the same dialog
	}

	if msg.Text == "/signup" {
		if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Enter a username")); err != nil {
			return nil, err
		}
		return &enterUserNameDialog{bot: d.bot}, nil
	}

	if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Sorry, I don't understand you")); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *rootDialog) OnCallbackQuery(ctx context.Context, updateID int, q *tgbotapi.CallbackQuery) (dialog.Dialog, error) {
	return d, nil
}

type enterUserNameDialog struct {
	bot *tgbotapi.BotAPI
}

func (d *enterUserNameDialog) Name() string {
	return "enter_username"
}

func (d *enterUserNameDialog) OnMessage(ctx context.Context, updateID int, msg *tgbotapi.Message) (dialog.Dialog, error) {
	username := msg.Text
	return &enterPasswordDialog{
		bot:      d.bot,
		Username: username,
	}, nil
}

func (d *enterUserNameDialog) OnCallbackQuery(ctx context.Context, updateID int, q *tgbotapi.CallbackQuery) (dialog.Dialog, error) {
	return d, nil
}

type enterPasswordDialog struct {
	bot      *tgbotapi.BotAPI
	Username string `json:"username,omitempty"`
}

func (d *enterPasswordDialog) Name() string {
	return "enter_password"
}

func (d *enterPasswordDialog) OnMessage(ctx context.Context, updateID int, msg *tgbotapi.Message) (dialog.Dialog, error) {
	password := msg.Text
	login(d.Username, password)

	return &rootDialog{bot: d.bot}, nil
}

func (d *enterPasswordDialog) OnCallbackQuery(ctx context.Context, updateID int, q *tgbotapi.CallbackQuery) (dialog.Dialog, error) {
	return d, nil
}

func login(username string, password string) {}

```
