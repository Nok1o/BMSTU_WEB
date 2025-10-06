//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/friends_service/internal/delivery/grpc/mocks"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/friends_service"
)

func TestGetFriendsInfo(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.GetFriendsInfoRequest
		mockSetup    func(*mocks.MockFriendsUseCase)
		expectedResp *pb.GetFriendsInfoResponse
		expectedErr  bool
	}{
		{
			name: "successful fetch",
			req: &pb.GetFriendsInfoRequest{
				UserId:  "user1",
				Limit:   "10",
				Offset:  "0",
				ReqType: "friends",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().GetFriendsInfo(gomock.Any(), "user1", "10", "0", "friends").
					Return([]models.FriendInfo{
						{
							Id:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
							Username:  "testuser",
							Firstname: "Test",
							Lastname:  "User",
						},
					}, 1, nil)
			},
			expectedResp: &pb.GetFriendsInfoResponse{
				Friends: []*pb.GetFriendInfo{
					{
						Id:        "123e4567-e89b-12d3-a456-426614174000",
						Username:  "testuser",
						Firstname: "Test",
						Lastname:  "User",
					},
				},
				TotalCount: 1,
			},
			expectedErr: false,
		},
		{
			name: "usecase error",
			req: &pb.GetFriendsInfoRequest{
				UserId: "user1",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().GetFriendsInfo(gomock.Any(), "user1", "", "", "").
					Return(nil, 0, errors.New("some error"))
			},
			expectedResp: &pb.GetFriendsInfoResponse{},
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			resp, err := server.GetFriendsInfo(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestSendFriendRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.FriendRequest
		mockSetup   func(*mocks.MockFriendsUseCase)
		expectedErr error
	}{
		{
			name: "successful request",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().IsExistsFriendRequest(gomock.Any(), "user1", "user2").Return(false, nil)
				m.EXPECT().SendFriendRequest(gomock.Any(), "user1", "user2").Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "request already exists",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().IsExistsFriendRequest(gomock.Any(), "user1", "user2").Return(true, nil)
			},
			expectedErr: errors.New("friend request already exists"),
		},
		{
			name: "check exists error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().IsExistsFriendRequest(gomock.Any(), "user1", "user2").Return(false, errors.New("check error"))
			},
			expectedErr: errors.New("check error"),
		},
		{
			name: "send request error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().IsExistsFriendRequest(gomock.Any(), "user1", "user2").Return(false, nil)
				m.EXPECT().SendFriendRequest(gomock.Any(), "user1", "user2").Return(errors.New("send error"))
			},
			expectedErr: errors.New("send error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			_, err := server.SendFriendRequest(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.FriendRequest
		mockSetup   func(*mocks.MockFriendsUseCase)
		expectedErr bool
	}{
		{
			name: "successful accept",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().AcceptFriendRequest(gomock.Any(), "user1", "user2").Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "accept error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().AcceptFriendRequest(gomock.Any(), "user1", "user2").Return(errors.New("accept error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			_, err := server.AcceptFriendRequest(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnfollow(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.FriendRequest
		mockSetup   func(*mocks.MockFriendsUseCase)
		expectedErr bool
	}{
		{
			name: "successful unfollow",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().Unfollow(gomock.Any(), "user1", "user2").Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "unfollow error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().Unfollow(gomock.Any(), "user1", "user2").Return(errors.New("unfollow error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			_, err := server.Unfollow(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteFriend(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.FriendRequest
		mockSetup   func(*mocks.MockFriendsUseCase)
		expectedErr bool
	}{
		{
			name: "successful delete",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().DeleteFriend(gomock.Any(), "user1", "user2").Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "delete error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().DeleteFriend(gomock.Any(), "user1", "user2").Return(errors.New("delete error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			_, err := server.DeleteFriend(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserRelation(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.FriendRequest
		mockSetup    func(*mocks.MockFriendsUseCase)
		expectedResp *pb.RelationResponse
		expectedErr  bool
	}{
		{
			name: "successful get relation",
			req: &pb.FriendRequest{
				UserId:     "123e4567-e89b-12d3-a456-426614174000",
				ReceiverId: "123e4567-e89b-12d3-a456-426614174001",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				user1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				user2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
				m.EXPECT().GetUserRelation(gomock.Any(), user1, user2).
					Return(models.RelationFriend, nil)
			},
			expectedResp: &pb.RelationResponse{
				Relation: string(models.RelationFriend),
			},
			expectedErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.FriendRequest{
				UserId:     "invalid",
				ReceiverId: "123e4567-e89b-12d3-a456-426614174001",
			},
			mockSetup:    func(m *mocks.MockFriendsUseCase) {},
			expectedResp: &pb.RelationResponse{},
			expectedErr:  false, // В текущей реализации ошибка парсинга не возвращается как ошибка
		},
		{
			name: "usecase error",
			req: &pb.FriendRequest{
				UserId:     "123e4567-e89b-12d3-a456-426614174000",
				ReceiverId: "123e4567-e89b-12d3-a456-426614174001",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				user1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				user2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
				m.EXPECT().GetUserRelation(gomock.Any(), user1, user2).
					Return(models.UserRelation(""), errors.New("relation error"))
			},
			expectedResp: &pb.RelationResponse{},
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			resp, err := server.GetUserRelation(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestMarkReadFriendRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.FriendRequest
		mockSetup   func(*mocks.MockFriendsUseCase)
		expectedErr bool
	}{
		{
			name: "successful mark read",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().MarkRead(gomock.Any(), "user1", "user2").Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "mark read error",
			req: &pb.FriendRequest{
				UserId:     "user1",
				ReceiverId: "user2",
			},
			mockSetup: func(m *mocks.MockFriendsUseCase) {
				m.EXPECT().MarkRead(gomock.Any(), "user1", "user2").Return(errors.New("mark read error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUseCase := mocks.NewMockFriendsUseCase(ctrl)
			tt.mockSetup(mockUseCase)

			server := NewFriendsServiceServer(mockUseCase)
			_, err := server.MarkReadFriendRequest(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
