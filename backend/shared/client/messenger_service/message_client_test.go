//go:build unit
// +build unit

package messenger_service

import (
	"context"
	"errors"
	"quickflow/shared/proto/messenger_service/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func TestMessageServiceClient_SendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	msgID := uuid.New()
	userID := uuid.New()
	chatID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		inputMsg    models.Message
		inputUserID uuid.UUID
		expected    *models.Message
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().SendMessage(ctx, gomock.Any()).Return(&pb.SendMessageResponse{
					Message: &pb.Message{
						Id:        msgID.String(),
						SenderId:  userID.String(),
						ChatId:    chatID.String(),
						Text:      "test",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
				}, nil)
			},
			inputMsg: models.Message{
				ID:        msgID,
				SenderID:  userID,
				ChatID:    chatID,
				Text:      "test",
				CreatedAt: now,
				UpdatedAt: now,
			},
			inputUserID: userID,
			expected: &models.Message{
				ID:        msgID,
				SenderID:  userID,
				ChatID:    chatID,
				Text:      "test",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().SendMessage(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			inputMsg: models.Message{
				ID: msgID,
			},
			inputUserID: userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := client.SendMessage(ctx, &tt.inputMsg, tt.inputUserID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Text, result.Text)
			}
		})
	}
}

func TestMessageServiceClient_GetMessagesForChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name         string
		setup        func()
		chatID       uuid.UUID
		num          int
		updatedAt    time.Time
		userID       uuid.UUID
		expectedMsgs []*models.Message
		expectError  bool
	}{
		{
			name: "success with messages",
			setup: func() {
				mockClient.EXPECT().GetMessagesForChat(ctx, gomock.Any()).Return(&pb.GetMessagesForChatResponse{
					Messages: []*pb.Message{
						{
							Id:        uuid.New().String(),
							Text:      "msg1",
							CreatedAt: timestamppb.New(now),
							SenderId:  userID.String(),
							ChatId:    chatID.String(),
						},
						{
							Id:        uuid.New().String(),
							Text:      "msg2",
							CreatedAt: timestamppb.New(now.Add(time.Hour)),
							SenderId:  userID.String(),
							ChatId:    chatID.String(),
						},
					},
				}, nil)
			},
			chatID:    chatID,
			num:       10,
			updatedAt: now,
			userID:    userID,
			expectedMsgs: []*models.Message{
				{
					ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"), // will be overwritten
					Text:      "msg1",
					CreatedAt: now,
				},
				{
					ID:        uuid.MustParse("00000000-0000-0000-0000-000000000002"), // will be overwritten
					Text:      "msg2",
					CreatedAt: now.Add(time.Hour),
				},
			},
		},
		{
			name: "empty response",
			setup: func() {
				mockClient.EXPECT().GetMessagesForChat(ctx, gomock.Any()).Return(&pb.GetMessagesForChatResponse{
					Messages: []*pb.Message{},
				}, nil)
			},
			chatID:       chatID,
			num:          10,
			updatedAt:    now,
			userID:       userID,
			expectedMsgs: []*models.Message{},
		},
		{
			name: "grpc error",
			setup: func() {
				mockClient.EXPECT().GetMessagesForChat(ctx, gomock.Any()).Return(nil, errors.New("grpc error"))
			},
			chatID:      chatID,
			num:         10,
			updatedAt:   now,
			userID:      userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			msgs, err := client.GetMessagesForChat(ctx, tt.chatID, tt.num, tt.updatedAt, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, msgs)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expectedMsgs), len(msgs))

				for i, expected := range tt.expectedMsgs {
					assert.Equal(t, expected.Text, msgs[i].Text)
					// ID will be different since we generate new UUIDs in the test
					assert.NotEqual(t, uuid.Nil, msgs[i].ID)
				}
			}
		})
	}
}

func TestMessageServiceClient_GetMessageById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	msgID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		inputID     uuid.UUID
		expected    *models.Message
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().GetMessageById(ctx, &pb.GetMessageByIdRequest{
					MessageId: msgID.String(),
				}).Return(&pb.GetMessageByIdResponse{
					Message: &pb.Message{
						SenderId:  msgID.String(),
						ChatId:    msgID.String(),
						Id:        msgID.String(),
						Text:      "test",
						CreatedAt: timestamppb.New(now),
					},
				}, nil)
			},
			inputID: msgID,
			expected: &models.Message{
				ChatID:    msgID,
				SenderID:  msgID,
				ID:        msgID,
				Text:      "test",
				CreatedAt: now,
			},
		},
		{
			name: "not found",
			setup: func() {
				mockClient.EXPECT().GetMessageById(ctx, gomock.Any()).Return(nil, errors.New("not found"))
			},
			inputID:     msgID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := client.GetMessageById(ctx, tt.inputID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMessageServiceClient_DeleteMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	msgID := uuid.New()

	tests := []struct {
		name        string
		setup       func()
		inputID     uuid.UUID
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().DeleteMessage(ctx, &pb.DeleteMessageRequest{
					MessageId: msgID.String(),
				}).Return(&pb.DeleteMessageResponse{}, nil)
			},
			inputID: msgID,
		},
		{
			name: "error",
			setup: func() {
				mockClient.EXPECT().DeleteMessage(ctx, gomock.Any()).Return(nil, errors.New("delete error"))
			},
			inputID:     msgID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := client.DeleteMessage(ctx, tt.inputID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMessageServiceClient_UpdateLastReadTs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()
	authID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		chatID      uuid.UUID
		userID      uuid.UUID
		ts          time.Time
		authID      uuid.UUID
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().UpdateLastReadTs(ctx, &pb.UpdateLastReadTsRequest{
					ChatId:            chatID.String(),
					UserId:            userID.String(),
					LastReadTimestamp: timestamppb.New(now),
					UserAuthId:        authID.String(),
				}).Return(&pb.UpdateLastReadTsResponse{}, nil)
			},
			chatID: chatID,
			userID: userID,
			ts:     now,
			authID: authID,
		},
		{
			name: "error",
			setup: func() {
				mockClient.EXPECT().UpdateLastReadTs(ctx, gomock.Any()).Return(nil, errors.New("update error"))
			},
			chatID:      chatID,
			userID:      userID,
			ts:          now,
			authID:      authID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := client.UpdateLastReadTs(ctx, tt.chatID, tt.userID, tt.ts, tt.authID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMessageServiceClient_GetLastReadTs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		setup       func()
		chatID      uuid.UUID
		userID      uuid.UUID
		expected    time.Time
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().GetLastReadTs(ctx, &pb.GetLastReadTsRequest{
					ChatId: chatID.String(),
					UserId: userID.String(),
				}).Return(&pb.GetLastReadTsResponse{
					LastReadTs: timestamppb.New(now),
				}, nil)
			},
			chatID:   chatID,
			userID:   userID,
			expected: now,
		},
		{
			name: "error",
			setup: func() {
				mockClient.EXPECT().GetLastReadTs(ctx, gomock.Any()).Return(nil, errors.New("get error"))
			},
			chatID:      chatID,
			userID:      userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := client.GetLastReadTs(ctx, tt.chatID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.UTC(), result.UTC())
			}
		})
	}
}

func TestMessageServiceClient_GetNumUnreadMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMessageServiceClient(ctrl)
	client := &MessageServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		setup       func()
		chatID      uuid.UUID
		userID      uuid.UUID
		expected    int
		expectError bool
	}{
		{
			name: "success",
			setup: func() {
				mockClient.EXPECT().GetNumUnreadMessages(ctx, &pb.GetNumUnreadMessagesRequest{
					ChatId: chatID.String(),
					UserId: userID.String(),
				}).Return(&pb.GetNumUnreadMessagesResponse{
					NumMessages: 5,
				}, nil)
			},
			chatID:   chatID,
			userID:   userID,
			expected: 5,
		},
		{
			name: "error",
			setup: func() {
				mockClient.EXPECT().GetNumUnreadMessages(ctx, gomock.Any()).Return(nil, errors.New("get error"))
			},
			chatID:      chatID,
			userID:      userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := client.GetNumUnreadMessages(ctx, tt.chatID, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, 0, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Mock for grpc.ClientConnInterface since we can't mock the real one
type MockClientConnInterface struct {
	grpc.ClientConnInterface
}

func NewMockClientConnInterface(ctrl *gomock.Controller) *MockClientConnInterface {
	return &MockClientConnInterface{}
}
