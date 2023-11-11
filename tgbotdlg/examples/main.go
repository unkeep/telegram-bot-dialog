package main

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/unkeep/telegram-bot-dialog/tgbotdlg"
)

type rootDialog struct {
	tgbotdlg.StatelessDialog
	bot *tgbotapi.BotAPI
}

func (d *rootDialog) Name() string {
	return "root"
}

func (d *rootDialog) HandleChatUpdate(_ context.Context, upd tgbotdlg.ChatUpdate) (*tgbotdlg.Switch, error) {
	msg := upd.Message
	if msg == nil {
		return nil, nil
	}

	switch msg.Text {
	case "/start":
		_, _ = d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Hi! I'm a signup bot"))
	case "/signup":
		return tgbotdlg.SwitchTo[*signupDialog](), nil
	default:
		_, _ = d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Sorry, I do not know this command"))
	}

	return nil, nil
}

type signupDialogState struct {
	Step     int    `json:"step"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type signupDialog struct {
	tgbotdlg.StatefulDialog[signupDialogState]
	bot *tgbotapi.BotAPI
}

func (d *signupDialog) Name() string {
	return "signup"
}

func (d *signupDialog) OnStart(ctx context.Context, chatID int64) error {
	return d.show(chatID, d.State(ctx))
}

func (d *signupDialog) HandleChatUpdate(ctx context.Context, upd tgbotdlg.ChatUpdate) (*tgbotdlg.Switch, error) {
	msg := upd.Message
	if msg == nil {
		return nil, nil
	}

	state := d.State(ctx)

	switch state.Step {
	case 0: // username
		state.UserName = msg.Text
		state.Step++
		return nil, d.show(msg.Chat.ID, state)
	case 1: // password
		state.Password = msg.Text
		state.Step++
		return nil, d.show(msg.Chat.ID, state)
	case 2: // password confirmation
		passConfirmation := msg.Text
		if signup(state.UserName, state.Password, passConfirmation) {
			_, _ = d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Success!"))
		} else {
			_, _ = d.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Fail!"))
		}
		return tgbotdlg.SwitchTo[*rootDialog](), nil
	default:
		return nil, fmt.Errorf("unexpected signup dialog step")
	}
}

func (d *signupDialog) show(chatID int64, state *signupDialogState) error {
	switch state.Step {
	case 0: // username
		_, err := d.bot.Send(tgbotapi.NewMessage(chatID, "Enter username"))
		return err
	case 1: // password
		_, err := d.bot.Send(tgbotapi.NewMessage(chatID, "Enter a password"))
		return err
	case 2: // password confirmation
		_, err := d.bot.Send(tgbotapi.NewMessage(chatID, "Confirm the password"))
		return err
	default:
		return fmt.Errorf("unexpected signup dialog step")
	}

	return nil
}

func signup(username string, password string, passwordConfirmation string) bool {
	return len(username) > 5 && len(password) > 5 && password == passwordConfirmation
}

func main() {
	bot, err := tgbotapi.NewBotAPI("MyAwesomeBotToken")
	if err != nil {
		log.Panic(err)
	}

	var dialogsStorage tgbotdlg.Storage

	dialogsDispatcher := tgbotdlg.NewDispatcher(
		dialogsStorage,
		&rootDialog{bot: bot},
		tgbotdlg.WithDialogs(&signupDialog{bot: bot}),
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if err := dialogsDispatcher.HandleBotUpdate(context.Background(), update); err != nil {
			log.Panic(err)
		}
	}
}
