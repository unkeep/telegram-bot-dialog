package tgbotdlg

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_RegisterDialogs(t *testing.T) {
	d := Dispatcher{}

	t.Run("when all dialogs are unique", func(t *testing.T) {
		newDialog1 := func() Dialog {
			return &dialogMock{name: "dialog1"}
		}

		newDialog2 := func() Dialog {
			return &dialogMock{name: "dialog2"}
		}

		d.RegisterDialogs(newDialog1, newDialog2)

		assert.Len(t, d.register, 2)
		assert.Equal(t, reflect.ValueOf(d.register["dialog1"]), reflect.ValueOf(newDialog1))
		assert.Equal(t, reflect.ValueOf(d.register["dialog2"]), reflect.ValueOf(newDialog2))
	})

	t.Run("when not all dialogs are unique", func(t *testing.T) {
		newDialog1 := func() Dialog {
			return &dialogMock{name: "dialog1"}
		}

		newDialog2 := func() Dialog {
			return &dialogMock{name: "dialog2"}
		}

		assert.Panics(t, func() { d.RegisterDialogs(newDialog1, newDialog2, newDialog1) })
	})
}

func TestDispatcher_HandleUpdate(t *testing.T) {
	storage := &storageMock{}

	dispatcher := &Dispatcher{
		storage: storage,
	}

	ctx := context.TODO()

	t.Run("update is not a message or button click", func(t *testing.T) {
		var upd *tgbotapi.Update
		assert.NoError(t, faker.FakeData(&upd))
		upd.Message = nil
		upd.CallbackQuery = nil

		// it does nothing returning nil err
		err := dispatcher.HandleUpdate(ctx, upd)
		assert.NoError(t, err)
	})

	var chatID, userID int64
	_ = faker.FakeData(&chatID)
	_ = faker.FakeData(&userID)

	for name, upd := range map[string]*tgbotapi.Update{
		"update is a message from user": {
			Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}},
		},
		"update is a button click": {
			CallbackQuery: &tgbotapi.CallbackQuery{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, From: &tgbotapi.User{ID: userID}},
		},
	} {
		upd := upd
		t.Run(name, func(t *testing.T) {
			storageGetCalls := 0
			storage.GetDialogFunc = func(ctx context.Context, gotChatID, gotUserID int64) (*Data, error) {
				storageGetCalls++
				assert.Equal(t, chatID, gotChatID)
				assert.Equal(t, userID, gotUserID)
				return nil, fmt.Errorf("test")
			}

			_ = dispatcher.HandleUpdate(ctx, upd)
			assert.Equal(t, 1, storageGetCalls)

			t.Run("storage fails", func(t *testing.T) {
				storage.GetDialogFunc = func(_ context.Context, _, _ int64) (*Data, error) {
					return nil, fmt.Errorf("test")
				}

				err := dispatcher.HandleUpdate(ctx, upd)
				assert.Error(t, err)
			})

			var storageErrNotFound error
			var storageData *Data
			for name, setup := range map[string]func(t *testing.T){
				"storage returns ErrNotFound": func(t *testing.T) {
					storageErrNotFound = ErrNotFound
					storageData = nil
				},
				"storage returns dialog data with state": func(t *testing.T) {
					storageErrNotFound = nil
					storageData = &Data{
						Name:  "dlg1",
						State: json.RawMessage(`{"state":"state1"}`),
					}
				},
				"storage returns dialog data with empty state": func(t *testing.T) {
					storageErrNotFound = nil
					storageData = &Data{
						Name:  "dlg1",
						State: json.RawMessage(`{}`),
					}
				},
				"storage returns dialog data with nil state": func(t *testing.T) {
					storageErrNotFound = nil
					storageData = &Data{
						Name:  "dlg1",
						State: nil,
					}
				},
			} {
				setup := setup
				t.Run(name, func(t *testing.T) {
					setup(t)
					storage.GetDialogFunc = func(_ context.Context, _, _ int64) (*Data, error) {
						return storageData, storageErrNotFound
					}

					t.Run("dialog is not registered", func(t *testing.T) {
						dispatcher.register = nil

						err := dispatcher.HandleUpdate(ctx, upd)
						assert.Error(t, err)
					})

					t.Run("dialogs are registered", func(t *testing.T) {
						rootDlgMock := &dialogMock{name: "root"}
						dlg1Mock := &dialogMock{name: "dlg1"}

						var handlerDlg *dialogMock
						if storageErrNotFound != nil {
							handlerDlg = rootDlgMock
						} else {
							handlerDlg = dlg1Mock
						}

						dispatcher.RegisterDialogs(
							func() Dialog { return rootDlgMock },
							func() Dialog { return dlg1Mock },
						)

						var handleUpdateCalls int
						handlerDlg.HandleUpdateFunc = func(ctx context.Context, dlgUpd Update) (Dialog, error) {
							handleUpdateCalls++
							expectedUpd := Update{
								ID:            upd.UpdateID,
								Message:       upd.Message,
								CallbackQuery: upd.CallbackQuery,
							}
							assert.Equal(t, expectedUpd, dlgUpd)
							return nil, fmt.Errorf("test")
						}

						_ = dispatcher.HandleUpdate(ctx, upd)

						assert.Equal(t, 1, handleUpdateCalls)

						if storageData != nil && storageData.State != nil && string(storageData.State) != "{}" {
							assert.Equal(t, "state1", handlerDlg.State)
						}

						t.Run("dialog handler fails", func(t *testing.T) {
							handlerDlg.HandleUpdateFunc = func(ctx context.Context, dlgUpd Update) (Dialog, error) {
								return nil, fmt.Errorf("HandleUpdateFuncErr")
							}

							err := dispatcher.HandleUpdate(ctx, upd)
							assert.Error(t, err)
						})

						for name, newDlg := range map[string]*dialogMock{
							"dialog handler returns a new dialog":                    {name: "newDlg", State: "newState"},
							"dialog handler returns same dialog ptr":                 handlerDlg,
							"dialog handler returns same dialog with the same state": {name: handlerDlg.name, State: handlerDlg.State},
							"dialog handler returns same dialog with new state":      {name: handlerDlg.name, State: "newState"},
						} {
							newDlg := newDlg
							t.Run(name, func(t *testing.T) {
								handlerDlg.HandleUpdateFunc = func(ctx context.Context, dlgUpd Update) (Dialog, error) {
									return newDlg, nil
								}

								var expectedStorageSaveCalls int
								var storageSaveCalls int

								shouldSaveNewDialog := handlerDlg.name != newDlg.name || handlerDlg.State != newDlg.State
								if shouldSaveNewDialog {
									expectedStorageSaveCalls = 1
									storage.SaveDialogFunc = func(_ context.Context, gotChatID, gotUserID int64, gotData *Data) error {
										storageSaveCalls++
										assert.Equal(t, chatID, gotChatID)
										assert.Equal(t, userID, gotUserID)
										assert.Equal(t, &Data{
											Name:  newDlg.name,
											State: json.RawMessage(`{"state":"newState"}`),
										}, gotData)

										return fmt.Errorf("test")
									}
								}

								err := dispatcher.HandleUpdate(ctx, upd)
								assert.Equal(t, expectedStorageSaveCalls, storageSaveCalls)

								if !shouldSaveNewDialog {
									assert.NoError(t, err)
								} else {
									t.Run("storage Save fails", func(t *testing.T) {
										storage.SaveDialogFunc = func(_ context.Context, _, _ int64, _ *Data) error {
											return fmt.Errorf("test")
										}

										err := dispatcher.HandleUpdate(ctx, upd)
										assert.Error(t, err)
									})

									t.Run("storage Save succeeds", func(t *testing.T) {
										storage.SaveDialogFunc = func(_ context.Context, _, _ int64, _ *Data) error {
											return nil
										}

										err := dispatcher.HandleUpdate(ctx, upd)
										assert.NoError(t, err)
									})
								}
							})
						}
					})
				})
			}
		})
	}
}

type dialogMock struct {
	name             string
	HandleUpdateFunc func(ctx context.Context, upd Update) (Dialog, error) `json:"-"`

	State string `json:"state,omitempty"`
}

func (d *dialogMock) Name() string {
	return d.name
}

func (d *dialogMock) HandleUpdate(ctx context.Context, upd Update) (Dialog, error) {
	return d.HandleUpdateFunc(ctx, upd)
}

type storageMock struct {
	GetDialogFunc  func(ctx context.Context, chatID, userID int64) (*Data, error)
	SaveDialogFunc func(ctx context.Context, chatID, userID int64, data *Data) error
}

func (s *storageMock) GetDialog(ctx context.Context, chatID, userID int64) (*Data, error) {
	return s.GetDialogFunc(ctx, chatID, userID)
}

func (s *storageMock) SaveDialog(ctx context.Context, chatID, userID int64, data *Data) error {
	return s.SaveDialogFunc(ctx, chatID, userID, data)
}
