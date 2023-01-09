package dialog

import (
	"context"
	"encoding/json"
	"fmt"
)

// ErrNotFound is the error which should be returned if a dialog is not found
var ErrNotFound = fmt.Errorf("dialog not found")

// Data is a dialog persistent data
type Data struct {
	Name  string
	State json.RawMessage
}

// Storage is a storage interface of current dialogs between bot and users
type Storage interface {
	// GetDialog returns current dialog data
	// ErrNotFound must be returned if there is no such dialog
	GetDialog(ctx context.Context, chatID, userID int64) (*Data, error)
	// SaveDialog saves curre
	SaveDialog(ctx context.Context, chatID, userID int64, data *Data) error
}
