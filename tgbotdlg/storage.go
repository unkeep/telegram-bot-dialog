package tgbotdlg

import (
	"context"
	"encoding/json"
)

// Data is a dialog persistent data
type Data struct {
	Name  string
	State json.RawMessage
}

// Storage is a storage interface of current dialogs between bot and users
type Storage interface {
	// GetDialog returns current dialog data
	// nil must be returned if there is no such dialog
	GetDialog(ctx context.Context, chatID int64) (*Data, error)
	// SaveDialog saves dialog data
	SaveDialog(ctx context.Context, chatID int64, data Data) error
}
