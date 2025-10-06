//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/messenger_service/internal/usecase"
	"quickflow/messenger_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

type MessageBuilder struct {
	message models.Message
}

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		message: models.Message{
			ID:         uuid.New(),
			SenderID:   uuid.New(),
			ReceiverID: uuid.New(),
			ChatID:     uuid.New(),
			Text:       "Hello",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
}

func (b *MessageBuilder) WithChatID(chatID uuid.UUID) *MessageBuilder {
	b.message.ChatID = chatID
	return b
}

func (b *MessageBuilder) WithSenderID(senderID uuid.UUID) *MessageBuilder {
	b.message.SenderID = senderID
	return b
}

func (b *MessageBuilder) WithReceiverID(receiverID uuid.UUID) *MessageBuilder {
	b.message.ReceiverID = receiverID
	return b
}

func (b *MessageBuilder) WithText(text string) *MessageBuilder {
	b.message.Text = text
	return b
}

func (b *MessageBuilder) Build() models.Message {
	return b.message
}

type ChatBuilder struct {
	chat models.Chat
}

func NewChatBuilder() *ChatBuilder {
	return &ChatBuilder{
		chat: models.Chat{
			ID:        uuid.New(),
			Type:      models.ChatTypePrivate,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *ChatBuilder) WithID(id uuid.UUID) *ChatBuilder {
	b.chat.ID = id
	return b
}

func (b *ChatBuilder) WithType(t models.ChatType) *ChatBuilder {
	b.chat.Type = t
	return b
}

func (b *ChatBuilder) Build() models.Chat {
	return b.chat
}

// ==================== Табличные тесты ====================

type MessageServiceSuite struct {
	suite.Suite
}

func TestMessageService(t *testing.T) {
	suite.RunSuite(t, new(MessageServiceSuite))
}

func (s *MessageServiceSuite) TestMessageService(t provider.T) {
	t.Epic("Unit")
	t.Feature("Message Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование сервиса сообщений с различными сценариями")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	messageRepo := mocks.NewMockMessageRepository(ctrl)
	fileRepo := mocks.NewMockFileService(ctrl)
	chatRepo := mocks.NewMockChatRepository(ctrl)
	validator := mocks.NewMockMessageValidator(ctrl)
	service := usecase.NewMessageService(messageRepo, fileRepo, chatRepo, validator)

	ctx := context.Background()

	t.Run("SaveMessage", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Save Message")
		t.Severity(allure.BLOCKER)
		t.Description("Тестирование сохранения сообщений с различными сценариями")

		message := NewMessageBuilder().Build()

		tests := []struct {
			name          string
			modifyMessage func(m models.Message) models.Message
			setupMocks    func()
			wantError     bool
		}{
			{
				name: "Success",
				setupMocks: func() {
					validator.EXPECT().ValidateMessage(message).Return(nil)
					messageRepo.EXPECT().SaveMessage(ctx, message).Return(nil)
					messageRepo.EXPECT().GetMessageById(ctx, message.ID).Return(message, nil)
				},
				wantError: false,
			},
			{
				name: "ValidationError",
				setupMocks: func() {
					validator.EXPECT().ValidateMessage(message).Return(errors.New("validation error"))
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				if tt.setupMocks != nil {
					tt.setupMocks()
				}
				result, err := service.SaveMessage(ctx, message)
				if tt.wantError {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, message, *result)
				}
			})
		}
	})

	t.Run("GetMessagesForChatOlder", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Get Messages For Chat Older")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование получения сообщений чата с различными сценариями")

		chatID := uuid.New()
		userID := uuid.New()
		messages := []models.Message{NewMessageBuilder().WithChatID(chatID).Build()}

		tests := []struct {
			name          string
			numMessages   int
			isParticipant bool
			mockError     error
			wantError     bool
		}{
			{"Success", 5, true, nil, false},
			{"InvalidNumMessages", 0, true, nil, true},
			{"NotParticipant", 5, false, nil, true},
			{"RepoError", 5, true, errors.New("repo error"), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				if tt.numMessages > 0 {
					chatRepo.EXPECT().IsParticipant(ctx, chatID, userID).Return(tt.isParticipant, nil)
					if tt.isParticipant {
						messageRepo.EXPECT().GetMessagesForChatOlder(ctx, chatID, tt.numMessages, gomock.Any()).Return(messages, tt.mockError)
					}
				}
				res, err := service.GetMessagesForChatOlder(ctx, chatID, userID, tt.numMessages, time.Now())
				if tt.wantError {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, messages, res)
				}
			})
		}
	})

	t.Run("DeleteMessage", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Delete Message")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование удаления сообщений с различными сценариями")

		messageID := uuid.New()
		tests := []struct {
			name      string
			id        uuid.UUID
			mockError error
			wantError bool
		}{
			{"Success", messageID, nil, false},
			{"EmptyID", uuid.Nil, nil, true},
			{"RepoError", messageID, errors.New("repo error"), true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				if tt.id != uuid.Nil && tt.mockError == nil {
					messageRepo.EXPECT().DeleteMessage(ctx, tt.id).Return(nil)
				}
				if tt.id != uuid.Nil && tt.mockError != nil {
					messageRepo.EXPECT().DeleteMessage(ctx, tt.id).Return(tt.mockError)
				}
				err := service.DeleteMessage(ctx, tt.id)
				if tt.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("GetLastChatMessage", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Get Last Chat Message")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование получения последнего сообщения чата с различными сценариями")

		chatID := uuid.New()
		message := NewMessageBuilder().WithChatID(chatID).Build()
		tests := []struct {
			name      string
			chatID    uuid.UUID
			mockError error
			wantError bool
		}{
			{"Success", chatID, nil, false},
			{"EmptyChatID", uuid.Nil, nil, true},
			{"RepoError", chatID, errors.New("repo error"), true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				if tt.chatID != uuid.Nil && tt.mockError == nil {
					messageRepo.EXPECT().GetLastChatMessage(ctx, tt.chatID).Return(&message, nil)
				}
				if tt.chatID != uuid.Nil && tt.mockError != nil {
					messageRepo.EXPECT().GetLastChatMessage(ctx, tt.chatID).Return(nil, tt.mockError)
				}
				res, err := service.GetLastChatMessage(ctx, tt.chatID)
				if tt.wantError {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, &message, res)
				}
			})
		}
	})

	t.Run("GetMessageById", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Get Message By ID")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование получения сообщения по ID с различными сценариями")

		message := NewMessageBuilder().Build()
		tests := []struct {
			name      string
			id        uuid.UUID
			mockError error
			wantError bool
		}{
			{"Success", message.ID, nil, false},
			{"EmptyID", uuid.Nil, nil, true},
			{"RepoError", message.ID, errors.New("repo error"), true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				if tt.id != uuid.Nil && tt.mockError == nil {
					messageRepo.EXPECT().GetMessageById(ctx, tt.id).Return(message, nil)
				}
				if tt.id != uuid.Nil && tt.mockError != nil {
					messageRepo.EXPECT().GetMessageById(ctx, tt.id).Return(models.Message{}, tt.mockError)
				}
				res, err := service.GetMessageById(ctx, tt.id)
				if tt.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, message, res)
				}
			})
		}
	})
}
