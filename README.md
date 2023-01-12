# Simple library for managing Telegram bot stateful dialogs

[![Go Reference](https://pkg.go.dev/badge/github.com/unkeep/telegram-bot-dialog.svg)](https://pkg.go.dev/github.com/unkeep/telegram-bot-dialog)
[![Build Status](https://app.travis-ci.com/unkeep/telegram-bot-dialog.svg?branch=main)](https://app.travis-ci.com/unkeep/telegram-bot-dialog)
[![Coverage Status](https://coveralls.io/repos/github/unkeep/telegram-bot-dialog/badge.svg?branch=main)](https://coveralls.io/github/unkeep/telegram-bot-dialog?branch=main)

## Example

First, ensure the library is installed and up to date by running `go get -u github.com/unkeep/telegram-bot-dialog`.

This is a very simple bot that asks username and password in a dialog manner

```go
package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/unkeep/telegram-bot-dialog/tgbotdlg"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	var dialogsStorage tgbotdlg.Storage // TODO: provide dialogs storage

	dialogsDispatcher := tgbotdlg.NewDispatcher(dialogsStorage)
	dialogsDispatcher.RegisterDialogs(
		func() tgbotdlg.Dialog {
			return &rootDialog{bot: bot}
		},
		func() tgbotdlg.Dialog {
			return &enterUserNameDialog{bot: bot}
		},
		func() tgbotdlg.Dialog {
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
	IsLoggedIn bool `json:"is_logged_in,omitempty"`
}

func (d *rootDialog) Name() string {
	return "root"
}

func (d *rootDialog) HandleUpdate(ctx context.Context, upd tgbotdlg.Update) (tgbotdlg.Dialog, error) {
	msg := upd.Message
	if msg == nil {
		return d, nil
	}
	
	if msg.Text == "/start" {
		if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Hi! I'm a bot")); err != nil {
			return nil, err
		}
		return d, nil // stay on the same dialog
	}

	if msg.Text == "/login" {
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

type enterUserNameDialog struct {
	bot *tgbotapi.BotAPI
}

func (d *enterUserNameDialog) Name() string {
	return "enter_username"
}

func (d *enterUserNameDialog) HandleUpdate(ctx context.Context, upd tgbotdlg.Update) (tgbotdlg.Dialog, error) {
	msg := upd.Message
	if msg == nil {
		return d, nil
	}
	
	username := msg.Text
	
	if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Enter a password")); err != nil {
		return nil, err
	}
	
	return &enterPasswordDialog{
		bot:      d.bot,
		Username: username,
	}, nil
}

type enterPasswordDialog struct {
	bot      *tgbotapi.BotAPI
	Username string `json:"username,omitempty"`
}

func (d *enterPasswordDialog) Name() string {
	return "enter_password"
}

func (d *enterPasswordDialog) HandleUpdate(ctx context.Context, upd tgbotdlg.Update) (tgbotdlg.Dialog, error) {
	msg := upd.Message
	if msg == nil {
		return d, nil
	}
	
	password := msg.Text
	login(d.Username, password)

	if _, err := d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You have successfully logged in")); err != nil {
		return nil, err
	}

	return &rootDialog{bot: d.bot, IsLoggedIn: true}, nil
}

func login(username string, password string) {}

```
