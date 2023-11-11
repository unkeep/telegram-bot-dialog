package tgbotdlg

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// ChatUpdate is a shortened version of tgbotapi.Update. It contains only chat-specific updates. See other kind of updates
// in OffChatUpdate
type ChatUpdate struct {
	// UpdateID is the update's unique identifier.
	// Update identifiers start from a certain positive number and increase
	// sequentially.
	// This ID becomes especially handy if you're using Webhooks,
	// since it allows you to ignore repeated updates or to restore
	// the correct update sequence, should they get out of order.
	// If there are no new updates for at least a week, then identifier
	// of the next update will be chosen randomly instead of sequentially.
	UpdateID int
	// Message new incoming message of any kind — text, photo, sticker, etc.
	//
	// optional
	Message *tgbotapi.Message
	// EditedMessage new version of a message that is known to the bot and was
	// edited
	//
	// optional
	EditedMessage *tgbotapi.Message
	// ChannelPost new version of a message that is known to the bot and was
	// edited
	//
	// optional
	ChannelPost *tgbotapi.Message
	// EditedChannelPost new incoming channel post of any kind — text, photo,
	// sticker, etc.
	//
	// optional
	EditedChannelPost *tgbotapi.Message
	// CallbackQuery new incoming callback query
	//
	// optional
	CallbackQuery *tgbotapi.CallbackQuery
	// MyChatMember is the bot's chat member status was updated in a chat. For
	// private chats, this update is received only when the bot is blocked or
	// unblocked by the user.
	//
	// optional
	MyChatMember *tgbotapi.ChatMemberUpdated
	// ChatMember is a chat member's status was updated in a chat. The bot must
	// be an administrator in the chat and must explicitly specify "chat_member"
	// in the list of allowed_updates to receive these updates.
	//
	// optional
	ChatMember *tgbotapi.ChatMemberUpdated
	// ChatJoinRequest is a request to join the chat has been sent. The bot must
	// have the can_invite_users administrator right in the chat to receive
	// these updates.
	//
	// optional
	ChatJoinRequest *tgbotapi.ChatJoinRequest `json:"chat_join_request"`
}

type OffChatUpdate struct {
	// UpdateID is the update's unique identifier.
	// Update identifiers start from a certain positive number and increase
	// sequentially.
	// This ID becomes especially handy if you're using Webhooks,
	// since it allows you to ignore repeated updates or to restore
	// the correct update sequence, should they get out of order.
	// If there are no new updates for at least a week, then identifier
	// of the next update will be chosen randomly instead of sequentially.
	UpdateID int `json:"update_id"`

	InlineQuery *tgbotapi.InlineQuery `json:"inline_query,omitempty"`
	// ChosenInlineResult is the result of an inline query
	// that was chosen by a user and sent to their chat partner.
	// Please see our documentation on the feedback collecting
	// for details on how to enable these updates for your bot.
	//
	// optional
	ChosenInlineResult *tgbotapi.ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	// ShippingQuery new incoming shipping query. Only for invoices with
	// flexible price
	//
	// optional
	ShippingQuery *tgbotapi.ShippingQuery `json:"shipping_query,omitempty"`
	// PreCheckoutQuery new incoming pre-checkout query. Contains full
	// information about checkout
	//
	// optional
	PreCheckoutQuery *tgbotapi.PreCheckoutQuery `json:"pre_checkout_query,omitempty"`
	// Pool new poll state. Bots receive only updates about stopped polls and
	// polls, which are sent by the bot
	//
	// optional
	Poll *tgbotapi.Poll `json:"poll,omitempty"`
	// PollAnswer user changed their answer in a non-anonymous poll. Bots
	// receive new votes only in polls that were sent by the bot itself.
	//
	// optional
	PollAnswer *tgbotapi.PollAnswer `json:"poll_answer,omitempty"`
}

type botUpdate tgbotapi.Update

func (u botUpdate) chatID() (int64, bool) {
	switch {
	case u.Message != nil:
		return u.Message.Chat.ID, true
	case u.EditedMessage != nil:
		return u.EditedMessage.Chat.ID, true
	case u.ChannelPost != nil:
		return u.ChannelPost.Chat.ID, true
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost.Chat.ID, true
	case u.CallbackQuery != nil && u.CallbackQuery.Message != nil:
		return u.CallbackQuery.Message.Chat.ID, true
	case u.MyChatMember != nil:
		return u.MyChatMember.Chat.ID, true
	case u.ChatMember != nil:
		return u.ChatMember.Chat.ID, true
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.Chat.ID, true
	default:
		return 0, false
	}
}

func (u botUpdate) toChatUpdate() ChatUpdate {
	return ChatUpdate{
		UpdateID:          u.UpdateID,
		Message:           u.Message,
		EditedMessage:     u.EditedMessage,
		ChannelPost:       u.ChannelPost,
		EditedChannelPost: u.EditedChannelPost,
		CallbackQuery:     u.CallbackQuery,
		MyChatMember:      u.MyChatMember,
		ChatMember:        u.ChatMember,
		ChatJoinRequest:   u.ChatJoinRequest,
	}
}

func (u botUpdate) toOffChatUpdate() OffChatUpdate {
	return OffChatUpdate{
		UpdateID:           u.UpdateID,
		InlineQuery:        u.InlineQuery,
		ChosenInlineResult: u.ChosenInlineResult,
		ShippingQuery:      u.ShippingQuery,
		PreCheckoutQuery:   u.PreCheckoutQuery,
		Poll:               u.Poll,
		PollAnswer:         u.PollAnswer,
	}
}
