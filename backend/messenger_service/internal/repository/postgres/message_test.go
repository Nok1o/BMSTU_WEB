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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	messenger_service "quickflow/messenger_service/internal/errors"
	postgres "quickflow/messenger_service/internal/repository/postgres"
	pgmodels "quickflow/messenger_service/internal/repository/postgres-models"
	"quickflow/shared/models"
)

type MessageBuilder struct {
	message models.Message
}

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		message: models.Message{
			ID:        uuid.New(),
			ChatID:    uuid.New(),
			SenderID:  uuid.New(),
			Text:      "Test message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *MessageBuilder) WithID(id uuid.UUID) *MessageBuilder {
	b.message.ID = id
	return b
}

func (b *MessageBuilder) WithChatID(chatID uuid.UUID) *MessageBuilder {
	b.message.ChatID = chatID
	return b
}

func (b *MessageBuilder) WithSenderID(senderID uuid.UUID) *MessageBuilder {
	b.message.SenderID = senderID
	return b
}

func (b *MessageBuilder) WithText(text string) *MessageBuilder {
	b.message.Text = text
	return b
}

func (b *MessageBuilder) WithCreatedAt(createdAt time.Time) *MessageBuilder {
	b.message.CreatedAt = createdAt
	return b
}

func (b *MessageBuilder) WithUpdatedAt(updatedAt time.Time) *MessageBuilder {
	b.message.UpdatedAt = updatedAt
	return b
}

func (b *MessageBuilder) WithAttachment(file models.File) *MessageBuilder {
	b.message.Attachments = append(b.message.Attachments, &file)
	return b
}

func (b *MessageBuilder) Build() models.Message {
	return b.message
}

type FileBuilder struct {
	file models.File
}

func NewFileBuilder() *FileBuilder {
	return &FileBuilder{
		file: models.File{
			URL:         "https://example.com/file.txt",
			DisplayType: "text/plain",
			Name:        "file.txt",
		},
	}
}

func (b *FileBuilder) WithURL(url string) *FileBuilder {
	b.file.URL = url
	return b
}

func (b *FileBuilder) WithDisplayType(displayType string) *FileBuilder {
	b.file.DisplayType = models.DisplayType(displayType)
	return b
}

func (b *FileBuilder) WithName(name string) *FileBuilder {
	b.file.Name = name
	return b
}

func (b *FileBuilder) Build() models.File {
	return b.file
}

func TestGetMessagesForChatOlder(t *testing.T) {
	runner.Run(t, "Get Messages For Chat Older Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Epic("Unit")
		t.Feature("Get Messages For Chat Older")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving messages older than specified timestamp for a chat")

		tests := []struct {
			name          string
			chatID        uuid.UUID
			numMessages   int
			timestamp     time.Time
			mockSetup     func(mock sqlmock.Sqlmock, chatID uuid.UUID, timestamp time.Time, numMessages int)
			expectedCount int
			expectError   bool
		}{
			{
				name:        "positive - messages found",
				chatID:      uuid.New(),
				numMessages: 10,
				timestamp:   time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID, timestamp time.Time, numMessages int) {

					messageID := uuid.New()
					rows := sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"}).
						AddRow(messageID, chatID, uuid.New(), "Hello", timestamp, timestamp)

					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.Timestamptz{Time: timestamp, Valid: true}, numMessages).
						WillReturnRows(rows)

					fileRows := sqlmock.NewRows([]string{"file_url", "file_type", "filename"}).
						AddRow("file1.txt", "text/plain", "file1.txt")

					mock.ExpectQuery(`SELECT mf.file_url, mf.file_type, f.filename`).
						WithArgs(messageID).
						WillReturnRows(fileRows)
				},
				expectedCount: 1,
				expectError:   false,
			},
			{
				name:        "negative - database error on messages query",
				chatID:      uuid.New(),
				numMessages: 10,
				timestamp:   time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID, timestamp time.Time, numMessages int) {
					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.Timestamptz{Time: timestamp, Valid: true}, numMessages).
						WillReturnError(errors.New("database error"))
				},
				expectedCount: 0,
				expectError:   true,
			},
			{
				name:        "positive - no messages found",
				chatID:      uuid.New(),
				numMessages: 10,
				timestamp:   time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID, timestamp time.Time, numMessages int) {
					rows := sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"})
					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.Timestamptz{Time: timestamp, Valid: true}, numMessages).
						WillReturnRows(rows)
				},
				expectedCount: 0,
				expectError:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.chatID, tt.timestamp, tt.numMessages)

				messages, err := repo.GetMessagesForChatOlder(context.Background(), tt.chatID, tt.numMessages, tt.timestamp)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, messages, tt.expectedCount)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestSaveMessage(t *testing.T) {
	runner.Run(t, "Save Message Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Save Message")
		t.Severity(allure.CRITICAL)
		t.Description("Test saving messages to the database")

		tests := []struct {
			name        string
			message     models.Message
			mockSetup   func(mock sqlmock.Sqlmock, message models.Message)
			expectError bool
		}{
			{
				name: "positive - message saved successfully",
				message: NewMessageBuilder().
					WithAttachment(NewFileBuilder().Build()).
					Build(),
				mockSetup: func(mock sqlmock.Sqlmock, message models.Message) {
					pgMessage := pgmodels.FromMessage(message)

					mock.ExpectExec(`INSERT INTO message`).
						WithArgs(
							pgMessage.ID, pgMessage.ChatID, pgMessage.SenderID,
							pgMessage.Text, pgMessage.CreatedAt, pgMessage.UpdatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))

					for _, file := range pgMessage.Attachments {
						mock.ExpectExec(`INSERT INTO message_file`).
							WithArgs(pgMessage.ID, file.URL, file.DisplayType).
							WillReturnResult(sqlmock.NewResult(1, 1))
					}

					mock.ExpectExec(`update chat set updated_at`).
						WithArgs(pgMessage.UpdatedAt, pgMessage.ChatID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				expectError: false,
			},
			{
				name:    "negative - database error on message insert",
				message: NewMessageBuilder().Build(),
				mockSetup: func(mock sqlmock.Sqlmock, message models.Message) {
					pgMessage := pgmodels.FromMessage(message)

					mock.ExpectExec(`INSERT INTO message`).
						WithArgs(
							pgMessage.ID, pgMessage.ChatID, pgMessage.SenderID,
							pgMessage.Text, pgMessage.CreatedAt, pgMessage.UpdatedAt,
						).
						WillReturnError(errors.New("database error"))
				},
				expectError: true,
			},
			{
				name: "negative - database error on file insert",
				message: NewMessageBuilder().
					WithAttachment(NewFileBuilder().Build()).
					Build(),
				mockSetup: func(mock sqlmock.Sqlmock, message models.Message) {
					pgMessage := pgmodels.FromMessage(message)

					mock.ExpectExec(`INSERT INTO message`).
						WithArgs(
							pgMessage.ID, pgMessage.ChatID, pgMessage.SenderID,
							pgMessage.Text, pgMessage.CreatedAt, pgMessage.UpdatedAt,
						).
						WillReturnResult(sqlmock.NewResult(1, 1))

					mock.ExpectExec(`INSERT INTO message_file`).
						WithArgs(pgMessage.ID, pgMessage.Attachments[0].URL, pgMessage.Attachments[0].DisplayType).
						WillReturnError(errors.New("database error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.message)

				err = repo.SaveMessage(context.Background(), tt.message)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestDeleteMessage(t *testing.T) {
	runner.Run(t, "Delete Message Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Delete Message")
		t.Severity(allure.NORMAL)
		t.Description("Test deleting messages from the database")

		tests := []struct {
			name        string
			messageID   uuid.UUID
			mockSetup   func(mock sqlmock.Sqlmock, messageID uuid.UUID)
			expectError bool
		}{
			{
				name:      "positive - message deleted successfully",
				messageID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, messageID uuid.UUID) {
					mock.ExpectExec(`DELETE FROM message WHERE id = \$1`).
						WithArgs(messageID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				expectError: false,
			},
			{
				name:      "negative - database error",
				messageID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, messageID uuid.UUID) {
					mock.ExpectExec(`DELETE FROM message WHERE id = \$1`).
						WithArgs(messageID).
						WillReturnError(errors.New("database error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.messageID)

				err = repo.DeleteMessage(context.Background(), tt.messageID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestUpdateLastReadTs(t *testing.T) {
	runner.Run(t, "Update Last Read Timestamp Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Update Last Read Timestamp")
		t.Severity(allure.NORMAL)
		t.Description("Test updating last read timestamp for users in chats")

		tests := []struct {
			name        string
			chatID      uuid.UUID
			userID      uuid.UUID
			timestamp   time.Time
			mockSetup   func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID, timestamp time.Time)
			expectError bool
		}{
			{
				name:      "positive - timestamp updated successfully",
				chatID:    uuid.New(),
				userID:    uuid.New(),
				timestamp: time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID, timestamp time.Time) {
					mock.ExpectExec(`update chat_user`).
						WithArgs(chatID, userID, pgtype.Timestamptz{Time: timestamp, Valid: true}).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				expectError: false,
			},
			{
				name:      "negative - not found error",
				chatID:    uuid.New(),
				userID:    uuid.New(),
				timestamp: time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID, timestamp time.Time) {
					mock.ExpectExec(`update chat_user`).
						WithArgs(chatID, userID, pgtype.Timestamptz{Time: timestamp, Valid: true}).
						WillReturnError(sql.ErrNoRows)
				},
				expectError: true,
			},
			{
				name:      "negative - database error",
				chatID:    uuid.New(),
				userID:    uuid.New(),
				timestamp: time.Now(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID, timestamp time.Time) {
					mock.ExpectExec(`update chat_user`).
						WithArgs(chatID, userID, pgtype.Timestamptz{Time: timestamp, Valid: true}).
						WillReturnError(errors.New("database error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.chatID, tt.userID, tt.timestamp)

				err = repo.UpdateLastReadTs(context.Background(), tt.timestamp, tt.chatID, tt.userID)

				if tt.expectError {
					assert.Error(t, err)
					if errors.Is(err, sql.ErrNoRows) {
						assert.ErrorIs(t, err, messenger_service.ErrNotFound)
					}
				} else {
					assert.NoError(t, err)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestGetLastReadTs(t *testing.T) {
	runner.Run(t, "Get Last Read Timestamp Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Last Read Timestamp")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving last read timestamp for users in chats")

		tests := []struct {
			name           string
			chatID         uuid.UUID
			userID         uuid.UUID
			mockSetup      func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID)
			expectedResult *time.Time
			expectError    bool
		}{
			{
				name:   "positive - timestamp found",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					expectedTime := time.Now()
					mock.ExpectQuery(`select max\(last_read\)`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.UUID{Bytes: userID, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(expectedTime))
				},
				expectedResult: func() *time.Time { t := time.Now(); return &t }(),
				expectError:    false,
			},
			{
				name:   "positive - no timestamp found",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					mock.ExpectQuery(`select max\(last_read\)`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.UUID{Bytes: userID, Valid: true}).
						WillReturnError(sql.ErrNoRows)
				},
				expectedResult: nil,
				expectError:    false,
			},
			{
				name:   "negative - database error",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					mock.ExpectQuery(`select max\(last_read\)`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}, pgtype.UUID{Bytes: userID, Valid: true}).
						WillReturnError(errors.New("database error"))
				},
				expectedResult: nil,
				expectError:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.chatID, tt.userID)

				result, err := repo.GetLastReadTs(context.Background(), tt.chatID, tt.userID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					if tt.expectedResult != nil {
						assert.NotNil(t, result)
					} else {
						assert.Nil(t, result)
					}
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestGetLastChatMessage(t *testing.T) {
	runner.Run(t, "Get Last Chat Message Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Last Chat Message")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving the last message in a chat")

		tests := []struct {
			name           string
			chatID         uuid.UUID
			mockSetup      func(mock sqlmock.Sqlmock, chatID uuid.UUID)
			expectedResult *models.Message
			expectError    bool
		}{
			{
				name:   "positive - last message found",
				chatID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID) {
					messageID := uuid.New()
					timestamp := time.Now()

					mock.ExpectQuery(`with otv as`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"}).
							AddRow(messageID, chatID, uuid.New(), "Last message", timestamp, timestamp))

					mock.ExpectQuery(`SELECT mf.file_url, mf.file_type, f.filename`).
						WithArgs(messageID).
						WillReturnRows(sqlmock.NewRows([]string{"file_url", "file_type", "filename"}))
				},
				expectedResult: &models.Message{},
				expectError:    false,
			},
			{
				name:   "positive - no message found",
				chatID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID) {
					mock.ExpectQuery(`with otv as`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnError(sql.ErrNoRows)
				},
				expectedResult: nil,
				expectError:    false,
			},
			{
				name:   "negative - database error",
				chatID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID uuid.UUID) {
					mock.ExpectQuery(`with otv as`).
						WithArgs(pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnError(errors.New("database error"))
				},
				expectedResult: nil,
				expectError:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.chatID)

				result, err := repo.GetLastChatMessage(context.Background(), tt.chatID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					if tt.expectedResult != nil {
						assert.NotNil(t, result)
					} else {
						assert.Nil(t, result)
					}
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestGetMessageById(t *testing.T) {
	runner.Run(t, "Get Message By ID Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Message By ID")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving messages by their ID")

		tests := []struct {
			name           string
			messageID      uuid.UUID
			mockSetup      func(mock sqlmock.Sqlmock, messageID uuid.UUID)
			expectError    bool
			expectNotFound bool
		}{
			{
				name:      "positive - message found",
				messageID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, messageID uuid.UUID) {

					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at FROM message WHERE id = \$1`).
						WithArgs(messageID).
						WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "sender_id", "text", "created_at", "updated_at"}).
							AddRow(messageID, uuid.New(), uuid.New(), "Test message", time.Now(), time.Now()))

					mock.ExpectQuery(`SELECT mf.file_url, mf.file_type, f.filename`).
						WithArgs(messageID).
						WillReturnRows(sqlmock.NewRows([]string{"file_url", "file_type", "filename"}))
				},
				expectError:    false,
				expectNotFound: false,
			},
			{
				name:      "negative - message not found",
				messageID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, messageID uuid.UUID) {
					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at FROM message WHERE id = \$1`).
						WithArgs(messageID).
						WillReturnError(sql.ErrNoRows)
				},
				expectError:    true,
				expectNotFound: true,
			},
			{
				name:      "negative - database error",
				messageID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, messageID uuid.UUID) {
					mock.ExpectQuery(`SELECT id, chat_id, sender_id, text, created_at, updated_at FROM message WHERE id = \$1`).
						WithArgs(messageID).
						WillReturnError(errors.New("database error"))
				},
				expectError:    true,
				expectNotFound: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.messageID)

				result, err := repo.GetMessageById(context.Background(), tt.messageID)

				if tt.expectError {
					assert.Error(t, err)
					if tt.expectNotFound {
						assert.ErrorIs(t, err, messenger_service.ErrNotFound)
					}
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestGetNumUnreadMessages(t *testing.T) {
	runner.Run(t, "Get Number of Unread Messages Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Number of Unread Messages")
		t.Severity(allure.NORMAL)
		t.Description("Test counting unread messages for a user in a chat")

		tests := []struct {
			name           string
			chatID         uuid.UUID
			userID         uuid.UUID
			mockSetup      func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID)
			expectedResult int
			expectError    bool
		}{
			{
				name:   "positive - unread messages found",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					mock.ExpectQuery(`select count\(\*\)`).
						WithArgs(pgtype.UUID{Bytes: userID, Valid: true}, pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				},
				expectedResult: 5,
				expectError:    false,
			},
			{
				name:   "positive - no unread messages",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					mock.ExpectQuery(`select count\(\*\)`).
						WithArgs(pgtype.UUID{Bytes: userID, Valid: true}, pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnError(sql.ErrNoRows)
				},
				expectedResult: 0,
				expectError:    false,
			},
			{
				name:   "negative - database error",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mock sqlmock.Sqlmock, chatID, userID uuid.UUID) {
					mock.ExpectQuery(`select count\(\*\)`).
						WithArgs(pgtype.UUID{Bytes: userID, Valid: true}, pgtype.UUID{Bytes: chatID, Valid: true}).
						WillReturnError(errors.New("database error"))
				},
				expectedResult: 0,
				expectError:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()

				repo := postgres.NewPostgresMessageRepository(db)
				tt.mockSetup(mock, tt.chatID, tt.userID)

				result, err := repo.GetNumUnreadMessages(context.Background(), tt.chatID, tt.userID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedResult, result)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestClose(t *testing.T) {
	runner.Run(t, "Close Repository Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Close Repository")
		t.Severity(allure.MINOR)
		t.Description("Test closing the database connection")

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		repo := postgres.NewPostgresMessageRepository(db)

		mock.ExpectClose()

		repo.Close()

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
