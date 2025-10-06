//go:build unit
// +build unit

package friends_service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
	"quickflow/shared/proto/friends_service/mocks"
)

func TestFriendsClient_GetFriendsInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	userID := "user123"
	limit := "10"
	offset := "0"
	reqType := "friends"

	expectedResponse := &pb.GetFriendsInfoResponse{
		Friends: []*pb.GetFriendInfo{
			{
				Id:        "123e4567-e89b-12d3-a456-426614174000",
				Username:  "user1",
				Firstname: "John",
			},
		},
		TotalCount: 1,
	}

	tests := []struct {
		name          string
		setupMock     func()
		userID        string
		limit         string
		offset        string
		reqType       string
		expected      []shared_models.FriendInfo
		expectedCount int
		expectedError bool
	}{
		{
			name: "successful get friends info",
			setupMock: func() {
				mockClient.EXPECT().GetFriendsInfo(gomock.Any(), &pb.GetFriendsInfoRequest{
					UserId:  userID,
					Limit:   limit,
					Offset:  offset,
					ReqType: reqType,
				}).Return(expectedResponse, nil)
			},
			userID:  userID,
			limit:   limit,
			offset:  offset,
			reqType: reqType,
			expected: []shared_models.FriendInfo{
				{
					Id:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Username:  "user1",
					Firstname: "John",
				},
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().GetFriendsInfo(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			userID:        userID,
			limit:         limit,
			offset:        offset,
			reqType:       reqType,
			expected:      nil,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, count, err := client.GetFriendsInfo(context.Background(), tt.userID, tt.limit, tt.offset, tt.reqType)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestFriendsClient_GetUserRelation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	user1 := uuid.New()
	user2 := uuid.New()

	tests := []struct {
		name          string
		setupMock     func()
		user1         uuid.UUID
		user2         uuid.UUID
		expected      shared_models.UserRelation
		expectedError bool
	}{
		{
			name: "friends relation",
			setupMock: func() {
				mockClient.EXPECT().GetUserRelation(gomock.Any(), &pb.FriendRequest{
					UserId:     user1.String(),
					ReceiverId: user2.String(),
				}).Return(&pb.RelationResponse{
					Relation: string(shared_models.RelationFriend),
				}, nil)
			},
			user1:         user1,
			user2:         user2,
			expected:      shared_models.RelationFriend,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().GetUserRelation(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			user1:         user1,
			user2:         user2,
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := client.GetUserRelation(context.Background(), tt.user1, tt.user2)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFriendsClient_SendFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	senderID := "sender123"
	receiverID := "receiver123"

	tests := []struct {
		name          string
		setupMock     func()
		senderID      string
		receiverID    string
		expectedError bool
	}{
		{
			name: "successful send friend request",
			setupMock: func() {
				mockClient.EXPECT().SendFriendRequest(gomock.Any(), &pb.FriendRequest{
					UserId:     senderID,
					ReceiverId: receiverID,
				}).Return(&emptypb.Empty{}, nil)
			},
			senderID:      senderID,
			receiverID:    receiverID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().SendFriendRequest(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			senderID:      senderID,
			receiverID:    receiverID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.SendFriendRequest(context.Background(), tt.senderID, tt.receiverID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestFriendsClient_AcceptFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	senderID := "sender123"
	receiverID := "receiver123"

	tests := []struct {
		name          string
		setupMock     func()
		senderID      string
		receiverID    string
		expectedError bool
	}{
		{
			name: "successful accept friend request",
			setupMock: func() {
				mockClient.EXPECT().AcceptFriendRequest(gomock.Any(), &pb.FriendRequest{
					UserId:     senderID,
					ReceiverId: receiverID,
				}).Return(&emptypb.Empty{}, nil)
			},
			senderID:      senderID,
			receiverID:    receiverID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().AcceptFriendRequest(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			senderID:      senderID,
			receiverID:    receiverID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.AcceptFriendRequest(context.Background(), tt.senderID, tt.receiverID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestFriendsClient_Unfollow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	userID := "user123"
	friendID := "friend123"

	tests := []struct {
		name          string
		setupMock     func()
		userID        string
		friendID      string
		expectedError bool
	}{
		{
			name: "successful unfollow",
			setupMock: func() {
				mockClient.EXPECT().Unfollow(gomock.Any(), &pb.FriendRequest{
					UserId:     userID,
					ReceiverId: friendID,
				}).Return(&emptypb.Empty{}, nil)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().Unfollow(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.Unfollow(context.Background(), tt.userID, tt.friendID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestFriendsClient_DeleteFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	userID := "user123"
	friendID := "friend123"

	tests := []struct {
		name          string
		setupMock     func()
		userID        string
		friendID      string
		expectedError bool
	}{
		{
			name: "successful delete friend",
			setupMock: func() {
				mockClient.EXPECT().DeleteFriend(gomock.Any(), &pb.FriendRequest{
					UserId:     userID,
					ReceiverId: friendID,
				}).Return(&emptypb.Empty{}, nil)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().DeleteFriend(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.DeleteFriend(context.Background(), tt.userID, tt.friendID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestFriendsClient_MarkRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockFriendsServiceClient(ctrl)
	client := &FriendsClient{client: mockClient}

	userID := "user123"
	friendID := "friend123"

	tests := []struct {
		name          string
		setupMock     func()
		userID        string
		friendID      string
		expectedError bool
	}{
		{
			name: "successful mark read",
			setupMock: func() {
				mockClient.EXPECT().MarkReadFriendRequest(gomock.Any(), &pb.FriendRequest{
					UserId:     userID,
					ReceiverId: friendID,
				}).Return(&emptypb.Empty{}, nil)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().MarkReadFriendRequest(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			userID:        userID,
			friendID:      friendID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.MarkRead(context.Background(), tt.userID, tt.friendID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
