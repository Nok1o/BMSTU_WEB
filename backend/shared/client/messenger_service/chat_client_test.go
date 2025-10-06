//go:build unit
// +build unit

package messenger_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
	"quickflow/shared/proto/messenger_service/mocks"
)

func TestChatServiceClient_GetUserChats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()
	limit := 10

	t.Run("success", func(t *testing.T) {
		expectedChats := []*pb.Chat{
			{Id: uuid.New().String(), Name: "Chat 1"},
			{Id: uuid.New().String(), Name: "Chat 2"},
		}

		mockClient.EXPECT().GetUserChats(ctx, &pb.GetUserChatsRequest{
			UserId:    userID.String(),
			ChatsNum:  int32(limit),
			UpdatedAt: timestamppb.New(now),
		}).Return(&pb.GetUserChatsResponse{Chats: expectedChats}, nil)

		chats, err := client.GetUserChats(ctx, userID, limit, now)
		require.NoError(t, err)
		assert.Len(t, chats, 2)
		assert.Equal(t, expectedChats[0].Id, chats[0].ID.String())
		assert.Equal(t, expectedChats[1].Id, chats[1].ID.String())
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().GetUserChats(ctx, gomock.Any()).Return(nil, expectedErr)

		chats, err := client.GetUserChats(ctx, userID, limit, now)
		assert.Error(t, err)
		assert.Nil(t, chats)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_CreateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	chatInfo := models.ChatCreationInfo{
		Name:   "New Chat",
		Type:   models.ChatTypeGroup,
		Avatar: &models.File{Name: "avatar.jpg"},
	}

	t.Run("success", func(t *testing.T) {
		expectedChat := &pb.Chat{
			Id:   uuid.New().String(),
			Name: chatInfo.Name,
			Type: pb.ChatType_CHAT_TYPE_GROUP,
		}

		mockClient.EXPECT().CreateChat(ctx, &pb.CreateChatRequest{
			UserId: userID.String(),
			ChatInfo: &pb.ChatCreationInfo{
				Name:   chatInfo.Name,
				Type:   pb.ChatType(chatInfo.Type),
				Avatar: file_service.ModelFileToProto(chatInfo.Avatar),
			},
		}).Return(&pb.CreateChatResponse{Chat: expectedChat}, nil)

		chat, err := client.CreateChat(ctx, userID, chatInfo)
		require.NoError(t, err)
		assert.Equal(t, expectedChat.Id, chat.ID.String())
		assert.Equal(t, expectedChat.Name, chat.Name)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().CreateChat(ctx, gomock.Any()).Return(nil, expectedErr)

		chat, err := client.CreateChat(ctx, userID, chatInfo)
		assert.Error(t, err)
		assert.Nil(t, chat)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("nil avatar", func(t *testing.T) {
		info := models.ChatCreationInfo{
			Name: "No Avatar",
			Type: models.ChatTypePrivate,
		}

		expectedChat := &pb.Chat{
			Id:   uuid.New().String(),
			Name: info.Name,
			Type: pb.ChatType_CHAT_TYPE_PRIVATE,
		}

		mockClient.EXPECT().CreateChat(ctx, &pb.CreateChatRequest{
			UserId: userID.String(),
			ChatInfo: &pb.ChatCreationInfo{
				Name:   info.Name,
				Type:   pb.ChatType(info.Type),
				Avatar: nil,
			},
		}).Return(&pb.CreateChatResponse{Chat: expectedChat}, nil)

		chat, err := client.CreateChat(ctx, userID, info)
		require.NoError(t, err)
		assert.Equal(t, expectedChat.Id, chat.ID.String())
	})
}

func TestChatServiceClient_GetChatParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	participantIDs := []string{uuid.New().String(), uuid.New().String()}

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().GetChatParticipants(ctx, &pb.GetChatParticipantsRequest{
			ChatId: chatID.String(),
		}).Return(&pb.GetChatParticipantsResponse{ParticipantIds: participantIDs}, nil)

		ids, err := client.GetChatParticipants(ctx, chatID)
		require.NoError(t, err)
		require.Len(t, ids, 2)
		assert.Equal(t, participantIDs[0], ids[0].String())
		assert.Equal(t, participantIDs[1], ids[1].String())
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().GetChatParticipants(ctx, gomock.Any()).Return(nil, expectedErr)

		ids, err := client.GetChatParticipants(ctx, chatID)
		assert.Error(t, err)
		assert.Nil(t, ids)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("invalid uuid in response", func(t *testing.T) {
		invalidIDs := []string{"invalid-uuid"}
		mockClient.EXPECT().GetChatParticipants(ctx, gomock.Any()).Return(&pb.GetChatParticipantsResponse{
			ParticipantIds: invalidIDs,
		}, nil)

		ids, err := client.GetChatParticipants(ctx, chatID)
		assert.Error(t, err)
		assert.Nil(t, ids)
	})
}

func TestChatServiceClient_GetPrivateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	user1 := uuid.New()
	user2 := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedChat := &pb.Chat{
			Id:   uuid.New().String(),
			Name: "Private Chat",
			Type: pb.ChatType_CHAT_TYPE_PRIVATE,
		}

		mockClient.EXPECT().GetPrivateChat(ctx, &pb.GetPrivateChatRequest{
			User1Id: user1.String(),
			User2Id: user2.String(),
		}).Return(&pb.GetPrivateChatResponse{Chat: expectedChat}, nil)

		chat, err := client.GetPrivateChat(ctx, user1, user2)
		require.NoError(t, err)
		assert.Equal(t, expectedChat.Id, chat.ID.String())
		assert.Equal(t, expectedChat.Name, chat.Name)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().GetPrivateChat(ctx, gomock.Any()).Return(nil, expectedErr)

		chat, err := client.GetPrivateChat(ctx, user1, user2)
		assert.Error(t, err)
		assert.Nil(t, chat)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_DeleteChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().DeleteChat(ctx, &pb.DeleteChatRequest{
			ChatId: chatID.String(),
		}).Return(&pb.DeleteChatResponse{}, nil)

		err := client.DeleteChat(ctx, chatID)
		assert.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().DeleteChat(ctx, gomock.Any()).Return(nil, expectedErr)

		err := client.DeleteChat(ctx, chatID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_GetChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedChat := &pb.Chat{
			Id:   chatID.String(),
			Name: "Test Chat",
			Type: pb.ChatType_CHAT_TYPE_GROUP,
		}

		mockClient.EXPECT().GetChat(ctx, &pb.GetChatRequest{
			ChatId: chatID.String(),
		}).Return(&pb.GetChatResponse{Chat: expectedChat}, nil)

		chat, err := client.GetChat(ctx, chatID)
		require.NoError(t, err)
		assert.Equal(t, expectedChat.Id, chat.ID.String())
		assert.Equal(t, expectedChat.Name, chat.Name)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().GetChat(ctx, gomock.Any()).Return(nil, expectedErr)

		chat, err := client.GetChat(ctx, chatID)
		assert.Error(t, err)
		assert.Nil(t, chat)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_JoinChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().JoinChat(ctx, &pb.JoinChatRequest{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).Return(&pb.JoinChatResponse{}, nil)

		err := client.JoinChat(ctx, chatID, userID)
		assert.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().JoinChat(ctx, gomock.Any()).Return(nil, expectedErr)

		err := client.JoinChat(ctx, chatID, userID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_LeaveChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().LeaveChat(ctx, &pb.LeaveChatRequest{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).Return(&pb.LeaveChatResponse{}, nil)

		err := client.LeaveChat(ctx, chatID, userID)
		assert.NoError(t, err)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().LeaveChat(ctx, gomock.Any()).Return(nil, expectedErr)

		err := client.LeaveChat(ctx, chatID, userID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestChatServiceClient_GetNumUnreadChats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockChatServiceClient(ctrl)
	client := &ChatServiceClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedNum := 5
		mockClient.EXPECT().GetNumUnreadChats(ctx, &pb.GetNumUnreadChatsRequest{
			UserId: userID.String(),
		}).Return(&pb.GetNumUnreadChatsResponse{NumChats: int32(expectedNum)}, nil)

		num, err := client.GetNumUnreadChats(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedNum, num)
	})

	t.Run("error from server", func(t *testing.T) {
		expectedErr := errors.New("server error")
		mockClient.EXPECT().GetNumUnreadChats(ctx, gomock.Any()).Return(nil, expectedErr)

		num, err := client.GetNumUnreadChats(ctx, userID)
		assert.Error(t, err)
		assert.Equal(t, 0, num)
		assert.Equal(t, expectedErr, err)
	})
}
