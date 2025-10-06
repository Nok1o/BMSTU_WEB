//go:build unit
// +build unit

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/require"

	messengerErrors "quickflow/messenger_service/internal/errors"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/shared/models"
)

func TestMessageRepository_GetMessagesForChatOlder(t *testing.T) {
	ctx := context.Background()
	chatID := uuid.New()
	senderID := uuid.New()
	ts := time.Now()

	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
		wantCount int
	}{
		{
			name: "success with one message and file",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(
					[]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"},
				).AddRow(uuid.New(), chatID, senderID, "hello", ts, ts)

				mock.ExpectQuery(`SELECT id, chat_id`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 10).
					WillReturnRows(rows)

				fileRows := sqlmock.NewRows([]string{"file_url", "file_type", "filename"}).
					AddRow("url1", "image", "file.png")
				mock.ExpectQuery(`SELECT mf.file_url`).WillReturnRows(fileRows)
			},
			wantCount: 1,
		},
		{
			name: "db error on query",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, chat_id`).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "scan error on message",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow("bad-id")
				mock.ExpectQuery(`SELECT id, chat_id`).WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "scan error on file",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(
					[]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"},
				).AddRow(uuid.New(), chatID, senderID, "hello", ts, ts)

				mock.ExpectQuery(`SELECT id, chat_id`).WillReturnRows(rows)

				fileRows := sqlmock.NewRows([]string{"file_url", "file_type"}).
					AddRow("url1", "image")
				mock.ExpectQuery(`SELECT mf.file_url`).WillReturnRows(fileRows)
			},
			wantErr: true,
		},
	}

	runner.Run(t, "GetMessagesForChatOlder", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetMessagesForChatOlder")
		t.Severity(allure.CRITICAL)
		t.Description("Tests fetching older messages with and without files, handling errors")

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t provider.T) {
				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresMessageRepository(db)

				tt.mockSetup(mock)
				msgs, err := repo.GetMessagesForChatOlder(ctx, chatID, 10, ts)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Len(t, msgs, tt.wantCount)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestMessageRepository_SaveMessage(t *testing.T) {
	ctx := context.Background()
	chatID := uuid.New()
	senderID := uuid.New()
	ts := time.Now()

	tests := []struct {
		name    string
		message models.Message
		mock    func(sqlmock.Sqlmock, models.Message)
		wantErr bool
	}{
		{
			name: "success without files",
			message: models.Message{
				ID:        uuid.New(),
				ChatID:    chatID,
				SenderID:  senderID,
				Text:      "hello",
				CreatedAt: ts,
				UpdatedAt: ts,
			},
			mock: func(mock sqlmock.Sqlmock, m models.Message) {
				mock.ExpectExec(`INSERT INTO message`).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`update chat set updated_at`).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "success with files",
			message: models.Message{
				ID:        uuid.New(),
				ChatID:    chatID,
				SenderID:  senderID,
				Text:      "hi",
				CreatedAt: ts,
				UpdatedAt: ts,
				Attachments: []*models.File{
					{URL: "file1", DisplayType: "image"},
				},
			},
			mock: func(mock sqlmock.Sqlmock, m models.Message) {
				mock.ExpectExec(`INSERT INTO message`).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO message_file`).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`update chat set updated_at`).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "error on insert message",
			message: models.Message{
				ID:        uuid.New(),
				ChatID:    chatID,
				SenderID:  senderID,
				Text:      "bad",
				CreatedAt: ts,
				UpdatedAt: ts,
			},
			mock: func(mock sqlmock.Sqlmock, m models.Message) {
				mock.ExpectExec(`INSERT INTO message`).WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "error on insert file",
			message: models.Message{
				ID:        uuid.New(),
				ChatID:    chatID,
				SenderID:  senderID,
				Text:      "hi",
				CreatedAt: ts,
				UpdatedAt: ts,
				Attachments: []*models.File{
					{URL: "file1", DisplayType: "image"},
				},
			},
			mock: func(mock sqlmock.Sqlmock, m models.Message) {
				mock.ExpectExec(`INSERT INTO message`).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`INSERT INTO message_file`).WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "error on update chat",
			message: models.Message{
				ID:        uuid.New(),
				ChatID:    chatID,
				SenderID:  senderID,
				Text:      "hi",
				CreatedAt: ts,
				UpdatedAt: ts,
			},
			mock: func(mock sqlmock.Sqlmock, m models.Message) {
				mock.ExpectExec(`INSERT INTO message`).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`update chat set updated_at`).WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	runner.Run(t, "SaveMessage", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("SaveMessage")
		t.Severity(allure.BLOCKER)
		t.Description("Tests saving messages with and without files, covering errors")

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t provider.T) {
				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresMessageRepository(db)

				tt.mock(mock, tt.message)
				err := repo.SaveMessage(ctx, tt.message)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestMessageRepository_DeleteMessage(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM message`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "db error",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM message`).WithArgs(id).WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	runner.Run(t, "DeleteMessage", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("DeleteMessage")
		t.Severity(allure.CRITICAL)
		t.Description("Tests deleting messages successfully and handling DB errors")

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t provider.T) {
				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresMessageRepository(db)

				tt.mock(mock)
				err := repo.DeleteMessage(ctx, id)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestMessageRepository_UpdateLastReadTs(t *testing.T) {
	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()
	ts := time.Now()

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "success",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update chat_user`).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name: "not found",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update chat_user`).WillReturnError(sql.ErrNoRows)
			},
			wantErr: messengerErrors.ErrNotFound,
		},
		{
			name: "db error",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update chat_user`).WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	runner.Run(t, "UpdateLastReadTs", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UpdateLastReadTs")
		t.Severity(allure.CRITICAL)
		t.Description("Tests updating last read timestamp in chat_user table")

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t provider.T) {
				db, mock, _ := sqlmock.New()
				defer db.Close()
				repo := postgres.NewPostgresMessageRepository(db)

				tt.mock(mock)
				err := repo.UpdateLastReadTs(ctx, ts, chatID, userID)
				if tt.wantErr == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}
