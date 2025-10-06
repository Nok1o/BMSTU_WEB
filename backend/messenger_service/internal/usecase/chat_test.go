//go:build unit
// +build unit

package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/messenger_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestNewChatUseCase(t *testing.T) {
	runner.Run(t, "New Chat UseCase Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Chat UseCase Constructor")
		t.Severity(allure.CRITICAL)
		t.Description("Test creating new Chat UseCase instance")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockChatRepo := mocks.NewMockChatRepository(ctrl)
		mockFileRepo := mocks.NewMockFileService(ctrl)
		mockProfileRepo := mocks.NewMockProfileService(ctrl)
		mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
		mockValidator := mocks.NewMockChatValidator(ctrl)

		service := NewChatUseCase(
			mockChatRepo,
			mockFileRepo,
			mockProfileRepo,
			mockMessageRepo,
			mockValidator,
		)

		assert.NotNil(t, service)
		assert.Equal(t, mockChatRepo, service.chatRepo)
		assert.Equal(t, mockFileRepo, service.fileRepo)
		assert.Equal(t, mockProfileRepo, service.profileRepo)
		assert.Equal(t, mockMessageRepo, service.messageRepo)
		assert.Equal(t, mockValidator, service.validator)
	})
}

func TestCreateChat(t *testing.T) {
	runner.Run(t, "Create Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Create Chat")
		t.Severity(allure.CRITICAL)
		t.Description("Test creating different types of chats with validation")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatInfo    models.ChatCreationInfo
			mockSetup   func(*mocks.MockChatRepository, *mocks.MockFileService, *mocks.MockChatValidator)
			expectError bool
			errorType   error
		}{
			{
				name: "positive - private chat",
				chatInfo: models.ChatCreationInfo{
					Type: models.ChatTypePrivate,
				},
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockFileRepo *mocks.MockFileService, mockValidator *mocks.MockChatValidator) {
					mockValidator.EXPECT().
						ValidateChatCreationInfo(gomock.Any()).
						Return(nil)
					mockChatRepo.EXPECT().
						CreateChat(gomock.Any(), gomock.Any()).
						Do(func(_ context.Context, chat models.Chat) {
							assert.Equal(t, models.ChatTypePrivate, chat.Type)
							assert.NotEqual(t, uuid.Nil, chat.ID)
						}).
						Return(nil)
				},
				expectError: false,
			},
			{
				name: "positive - group chat with avatar",
				chatInfo: models.ChatCreationInfo{
					Type:   models.ChatTypeGroup,
					Name:   "Test Group",
					Avatar: &models.File{},
				},
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockFileRepo *mocks.MockFileService, mockValidator *mocks.MockChatValidator) {
					mockValidator.EXPECT().
						ValidateChatCreationInfo(gomock.Any()).
						Return(nil)
					mockFileRepo.EXPECT().
						UploadFile(gomock.Any(), gomock.Any()).
						Return("avatar_url", nil)
					mockChatRepo.EXPECT().
						CreateChat(gomock.Any(), gomock.Any()).
						Do(func(_ context.Context, chat models.Chat) {
							assert.Equal(t, models.ChatTypeGroup, chat.Type)
							assert.Equal(t, "Test Group", chat.Name)
							assert.Equal(t, "avatar_url", chat.AvatarURL)
						}).
						Return(nil)
				},
				expectError: false,
			},
			{
				name:     "negative - validation error",
				chatInfo: models.ChatCreationInfo{},
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockFileRepo *mocks.MockFileService, mockValidator *mocks.MockChatValidator) {
					mockValidator.EXPECT().
						ValidateChatCreationInfo(gomock.Any()).
						Return(errors.New("validation error"))
				},
				expectError: true,
				errorType:   messenger_errors.ErrInvalidChatCreationInfo,
			},
			{
				name: "negative - upload error",
				chatInfo: models.ChatCreationInfo{
					Type:   models.ChatTypeGroup,
					Avatar: &models.File{},
				},
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockFileRepo *mocks.MockFileService, mockValidator *mocks.MockChatValidator) {
					mockValidator.EXPECT().
						ValidateChatCreationInfo(gomock.Any()).
						Return(nil)
					mockFileRepo.EXPECT().
						UploadFile(gomock.Any(), gomock.Any()).
						Return("", errors.New("upload error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				mockFileRepo := mocks.NewMockFileService(ctrl)
				mockValidator := mocks.NewMockChatValidator(ctrl)

				service := NewChatUseCase(
					mockChatRepo,
					mockFileRepo, nil, nil,
					mockValidator,
				)

				tt.mockSetup(mockChatRepo, mockFileRepo, mockValidator)

				result, err := service.CreateChat(context.Background(), tt.chatInfo)

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorType != nil {
						assert.Equal(t, tt.errorType, err)
					}
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})
}

func TestGetUserChats(t *testing.T) {
	runner.Run(t, "Get User Chats Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get User Chats")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving user chats with participant information and last messages")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		uid := uuid.New()

		tests := []struct {
			name        string
			userID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository, *mocks.MockProfileService, *mocks.MockMessageRepository)
			expectError bool
		}{
			{
				name:   "positive - successful retrieval",
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockProfileRepo *mocks.MockProfileService, mockMessageRepo *mocks.MockMessageRepository) {
					chatID := uuid.New()
					otherUserID := uuid.New()

					chats := []models.Chat{{ID: chatID, Type: models.ChatTypePrivate}}
					mockChatRepo.EXPECT().GetUserChats(gomock.Any(), gomock.Any()).Return(chats, nil)
					mockChatRepo.EXPECT().GetChatParticipants(gomock.Any(), chatID).Return([]uuid.UUID{uid, otherUserID}, nil).AnyTimes()
					mockProfileRepo.EXPECT().GetPublicUsersInfo(gomock.Any(), gomock.Any()).Return([]models.PublicUserInfo{
						{Id: uid, Firstname: "Me", Lastname: "User"},
						{Id: otherUserID, Firstname: "Other", Lastname: "User", AvatarURL: "avatar_url"},
					}, nil)
					mockMessageRepo.EXPECT().GetLastChatMessage(gomock.Any(), chatID).Return(&models.Message{Text: "last message"}, nil)
				},
				expectError: false,
			},
			{
				name:   "negative - get user chats error",
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockProfileRepo *mocks.MockProfileService, mockMessageRepo *mocks.MockMessageRepository) {
					mockChatRepo.EXPECT().GetUserChats(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
				},
				expectError: true,
			},
			{
				name:   "negative - get participants error",
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository, mockProfileRepo *mocks.MockProfileService, mockMessageRepo *mocks.MockMessageRepository) {
					chats := []models.Chat{{ID: uuid.New(), Type: models.ChatTypePrivate}}
					mockChatRepo.EXPECT().GetUserChats(gomock.Any(), gomock.Any()).Return(chats, nil)
					mockChatRepo.EXPECT().GetChatParticipants(gomock.Any(), gomock.Any()).Return(nil, errors.New("participants error")).AnyTimes()
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				mockProfileRepo := mocks.NewMockProfileService(ctrl)
				mockMessageRepo := mocks.NewMockMessageRepository(ctrl)

				service := NewChatUseCase(
					mockChatRepo,
					nil, mockProfileRepo, mockMessageRepo,
					nil,
				)

				tt.mockSetup(mockChatRepo, mockProfileRepo, mockMessageRepo)

				result, err := service.GetUserChats(context.Background(), tt.userID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})
}

func TestGetChat(t *testing.T) {
	runner.Run(t, "Get Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Chat")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving chat by ID")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
		}{
			{
				name:   "positive - chat found",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetChat(gomock.Any(), gomock.Any()).Return(models.Chat{ID: uuid.New()}, nil)
				},
				expectError: false,
			},
			{
				name:   "negative - chat not found",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetChat(gomock.Any(), gomock.Any()).Return(models.Chat{}, errors.New("not found"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				result, err := service.GetChat(context.Background(), tt.chatID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})
}

func TestDeleteChat(t *testing.T) {
	runner.Run(t, "Delete Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Delete Chat")
		t.Severity(allure.CRITICAL)
		t.Description("Test deleting chats with existence validation")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
			errorType   error
		}{
			{
				name:   "positive - successful deletion",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().DeleteChat(gomock.Any(), gomock.Any()).Return(nil)
				},
				expectError: false,
			},
			{
				name:   "negative - chat not found",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
				},
				expectError: true,
				errorType:   messenger_errors.ErrNotFound,
			},
			{
				name:   "negative - exists check error",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, errors.New("check error"))
				},
				expectError: true,
			},
			{
				name:   "negative - deletion error",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().DeleteChat(gomock.Any(), gomock.Any()).Return(errors.New("deletion error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				err := service.DeleteChat(context.Background(), tt.chatID)

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorType != nil {
						assert.Equal(t, tt.errorType, err)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestJoinChat(t *testing.T) {
	runner.Run(t, "Join Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Join Chat")
		t.Severity(allure.NORMAL)
		t.Description("Test joining chats with participation validation")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatID      uuid.UUID
			userID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
			errorType   error
		}{
			{
				name:   "positive - successful join",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().IsParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
					mockChatRepo.EXPECT().JoinChat(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				},
				expectError: false,
			},
			{
				name:   "negative - chat not found",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
				},
				expectError: true,
				errorType:   messenger_errors.ErrNotFound,
			},
			{
				name:   "negative - already in chat",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().IsParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
				},
				expectError: true,
				errorType:   messenger_errors.ErrAlreadyInChat,
			},
			{
				name:   "negative - exists check error",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, errors.New("check error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				err := service.JoinChat(context.Background(), tt.chatID, tt.userID)

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorType != nil {
						assert.Equal(t, tt.errorType, err)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestLeaveChat(t *testing.T) {
	runner.Run(t, "Leave Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Leave Chat")
		t.Severity(allure.NORMAL)
		t.Description("Test leaving chats with participation validation")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatID      uuid.UUID
			userID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
			errorType   error
		}{
			{
				name:   "positive - successful leave",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().IsParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().LeaveChat(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				},
				expectError: false,
			},
			{
				name:   "negative - chat not found",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil)
				},
				expectError: true,
				errorType:   messenger_errors.ErrNotFound,
			},
			{
				name:   "negative - not a participant",
				chatID: uuid.New(),
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil)
					mockChatRepo.EXPECT().IsParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
				},
				expectError: true,
				errorType:   messenger_errors.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				err := service.LeaveChat(context.Background(), tt.chatID, tt.userID)

				if tt.expectError {
					assert.Error(t, err)
					if tt.errorType != nil {
						assert.Equal(t, tt.errorType, err)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestGetChatParticipants(t *testing.T) {
	runner.Run(t, "Get Chat Participants Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Chat Participants")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving chat participants")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			chatID      uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
		}{
			{
				name:   "positive - participants found",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					expected := []uuid.UUID{uuid.New(), uuid.New()}
					mockChatRepo.EXPECT().GetChatParticipants(gomock.Any(), gomock.Any()).Return(expected, nil)
				},
				expectError: false,
			},
			{
				name:   "negative - error getting participants",
				chatID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetChatParticipants(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				result, err := service.GetChatParticipants(context.Background(), tt.chatID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})
}

func TestGetPrivateChat(t *testing.T) {
	runner.Run(t, "Get Private Chat Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Private Chat")
		t.Severity(allure.NORMAL)
		t.Description("Test retrieving private chat between two users")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name        string
			user1       uuid.UUID
			user2       uuid.UUID
			mockSetup   func(*mocks.MockChatRepository)
			expectError bool
		}{
			{
				name:  "positive - private chat found",
				user1: uuid.New(),
				user2: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					expected := models.Chat{ID: uuid.New()}
					mockChatRepo.EXPECT().GetPrivateChat(gomock.Any(), gomock.Any(), gomock.Any()).Return(expected, nil)
				},
				expectError: false,
			},
			{
				name:  "negative - private chat not found",
				user1: uuid.New(),
				user2: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetPrivateChat(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.Chat{}, errors.New("not found"))
				},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				result, err := service.GetPrivateChat(context.Background(), tt.user1, tt.user2)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})
}

func TestGetNumUnreadChats(t *testing.T) {
	runner.Run(t, "Get Number of Unread Chats Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Get Number of Unread Chats")
		t.Severity(allure.NORMAL)
		t.Description("Test counting unread chats for a user")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tests := []struct {
			name          string
			userID        uuid.UUID
			mockSetup     func(*mocks.MockChatRepository)
			expectedCount int
			expectError   bool
		}{
			{
				name:   "positive - unread chats found",
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetNumUnreadChats(gomock.Any(), gomock.Any()).Return(5, nil)
				},
				expectedCount: 5,
				expectError:   false,
			},
			{
				name:   "negative - error getting unread chats",
				userID: uuid.New(),
				mockSetup: func(mockChatRepo *mocks.MockChatRepository) {
					mockChatRepo.EXPECT().GetNumUnreadChats(gomock.Any(), gomock.Any()).Return(0, errors.New("database error"))
				},
				expectedCount: 0,
				expectError:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")

				t.Description(tt.name)

				mockChatRepo := mocks.NewMockChatRepository(ctrl)
				service := NewChatUseCase(mockChatRepo, nil, nil, nil, nil)

				tt.mockSetup(mockChatRepo)

				result, err := service.GetNumUnreadChats(context.Background(), tt.userID)

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedCount, result)
				}
			})
		}
	})
}

func TestGetUserChats_ConcurrentProcessing(t *testing.T) {
	runner.Run(t, "Get User Chats Concurrent Processing Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Concurrent Processing")
		t.Severity(allure.NORMAL)
		t.Description("Test concurrent processing of multiple user chats")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockChatRepo := mocks.NewMockChatRepository(ctrl)
		mockProfileRepo := mocks.NewMockProfileService(ctrl)
		mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
		service := NewChatUseCase(
			mockChatRepo,
			nil, mockProfileRepo, mockMessageRepo,
			nil,
		)

		ctx := context.Background()
		userID := uuid.New()

		// Create multiple private chats
		chats := make([]models.Chat, 10)
		for i := 0; i < 10; i++ {
			chats[i] = models.Chat{
				ID:   uuid.New(),
				Type: models.ChatTypePrivate,
			}
		}

		// Mock expectations
		mockChatRepo.EXPECT().
			GetUserChats(ctx, userID).
			Return(chats, nil)

		// Each chat will have its own participants
		for _, chat := range chats {
			otherUser := uuid.New()
			mockChatRepo.EXPECT().
				GetChatParticipants(gomock.Any(), chat.ID).
				Return([]uuid.UUID{userID, otherUser}, nil).AnyTimes()

			mockProfileRepo.EXPECT().
				GetPublicUsersInfo(gomock.Any(), []uuid.UUID{userID, otherUser}).
				Return([]models.PublicUserInfo{
					{Id: userID},
					{Id: otherUser, Firstname: "User", Lastname: "Other"},
				}, nil)

			mockMessageRepo.EXPECT().
				GetLastChatMessage(gomock.Any(), chat.ID).
				Return(&models.Message{}, nil)
		}

		// Execute and verify concurrent processing
		result, err := service.GetUserChats(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, result, 10)
	})
}
