//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/community_service/internal/delivery/grpc/mocks"
	"quickflow/shared/models"
	"quickflow/shared/proto/community_service"
)

func TestIsCommunityMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()

	tests := []struct {
		name         string
		req          *community_service.IsCommunityMemberRequest
		expectedResp *community_service.IsCommunityMemberResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "IsCommunityMember - user is member",
			req: &community_service.IsCommunityMemberRequest{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
			},
			expectedResp: &community_service.IsCommunityMemberResponse{
				IsMember: true,
				Role:     community_service.CommunityRole_COMMUNITY_ROLE_ADMIN,
			},
			expectedErr: nil,
			mockSetup: func() {
				role := models.CommunityRoleAdmin
				mockCommunityUseCase.EXPECT().
					IsCommunityMember(gomock.Any(), userID, communityID).
					Return(true, &role, nil)
			},
		},
		{
			name: "IsCommunityMember - user is not member",
			req: &community_service.IsCommunityMemberRequest{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
			},
			expectedResp: &community_service.IsCommunityMemberResponse{
				IsMember: false,
				Role:     -1,
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					IsCommunityMember(gomock.Any(), userID, communityID).
					Return(false, nil, nil)
			},
		},
		{
			name: "IsCommunityMember - invalid user ID",
			req: &community_service.IsCommunityMemberRequest{
				UserId:      "invalid",
				CommunityId: communityID.String(),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid user ID"),
			mockSetup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.IsCommunityMember(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.IsMember, resp.IsMember)
				assert.Equal(t, tt.expectedResp.Role, resp.Role)
			}
		})
	}
}

func TestDeleteCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	communityID := uuid.New()

	tests := []struct {
		name        string
		req         *community_service.DeleteCommunityRequest
		expectedErr error
		mockSetup   func()
	}{
		{
			name: "DeleteCommunity - success",
			req: &community_service.DeleteCommunityRequest{
				CommunityId: communityID.String(),
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					DeleteCommunity(gomock.Any(), communityID).
					Return(nil)
			},
		},
		{
			name: "DeleteCommunity - invalid ID",
			req: &community_service.DeleteCommunityRequest{
				CommunityId: "invalid",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid community ID"),
			mockSetup:   func() {},
		},
		{
			name: "DeleteCommunity - not found",
			req: &community_service.DeleteCommunityRequest{
				CommunityId: communityID.String(),
			},
			expectedErr: status.Error(codes.NotFound, "community not found"),
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					DeleteCommunity(gomock.Any(), communityID).
					Return(errors.New("community not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.DeleteCommunity(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Nil(t, resp)
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NotNil(t, resp)
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()

	tests := []struct {
		name         string
		req          *community_service.UpdateCommunityRequest
		expectedResp *community_service.UpdateCommunityResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "UpdateCommunity - success",
			req: &community_service.UpdateCommunityRequest{
				Id:        communityID.String(),
				UserId:    userID.String(),
				Nickname:  "Updated Community",
				Name:      "Updated Name",
				Avatar:    nil,
				Cover:     nil,
				AvatarUrl: "http://example.com/avatar.jpg",
				CoverUrl:  "http://example.com/cover.jpg",
			},
			expectedResp: &community_service.UpdateCommunityResponse{
				Community: &community_service.Community{
					Id:        communityID.String(),
					OwnerId:   userID.String(),
					Nickname:  "Updated Community",
					Name:      "Updated Name",
					Avatar:    nil,
					Cover:     nil,
					AvatarUrl: "http://example.com/avatar.jpg",
					CoverUrl:  "http://example.com/cover.jpg",
				},
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					UpdateCommunity(gomock.Any(), gomock.Any(), userID).
					Return(&models.Community{
						ID:       communityID,
						OwnerID:  userID,
						NickName: "Updated Community",
						BasicInfo: &models.BasicCommunityInfo{
							Name:      "Updated Name",
							AvatarUrl: "http://example.com/avatar.jpg",
							CoverUrl:  "http://example.com/cover.jpg",
						},
					}, nil)
			},
		},
		{
			name: "UpdateCommunity - invalid user ID",
			req: &community_service.UpdateCommunityRequest{
				UserId: "invalid",
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid user ID"),
			mockSetup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.UpdateCommunity(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Community.Id, resp.Community.Id)
				assert.Equal(t, tt.expectedResp.Community.Nickname, resp.Community.Nickname)
				assert.Equal(t, tt.expectedResp.Community.Name, resp.Community.Name)
				assert.Equal(t, tt.expectedResp.Community.AvatarUrl, resp.Community.AvatarUrl)
				assert.Equal(t, tt.expectedResp.Community.CoverUrl, resp.Community.CoverUrl)
			}
		})
	}
}

func TestJoinCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	tests := []struct {
		name         string
		req          *community_service.JoinCommunityRequest
		expectedResp *community_service.JoinCommunityResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "JoinCommunity - success",
			req: &community_service.JoinCommunityRequest{
				NewMember: &community_service.CommunityMember{
					UserId:      uuid.New().String(),
					CommunityId: uuid.New().String(),
					Role:        community_service.CommunityRole_COMMUNITY_ROLE_MEMBER,
				},
			},
			expectedResp: &community_service.JoinCommunityResponse{
				Success: true,
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					JoinCommunity(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "JoinCommunity - invalid member data",
			req: &community_service.JoinCommunityRequest{
				NewMember: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "member data cannot be empty"),
			mockSetup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.JoinCommunity(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Success, resp.Success)
			}
		})
	}
}

func TestLeaveCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()

	tests := []struct {
		name         string
		req          *community_service.LeaveCommunityRequest
		expectedResp *community_service.LeaveCommunityResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "LeaveCommunity - success",
			req: &community_service.LeaveCommunityRequest{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
			},
			expectedResp: &community_service.LeaveCommunityResponse{
				Success: true,
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					LeaveCommunity(gomock.Any(), userID, communityID).
					Return(nil)
			},
		},
		{
			name: "LeaveCommunity - invalid user ID",
			req: &community_service.LeaveCommunityRequest{
				UserId:      "invalid",
				CommunityId: communityID.String(),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid user ID"),
			mockSetup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.LeaveCommunity(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Success, resp.Success)
			}
		})
	}
}

func TestGetUserCommunities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()
	ts := time.Now()

	tests := []struct {
		name         string
		req          *community_service.GetUserCommunitiesRequest
		expectedResp *community_service.GetUserCommunitiesResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "GetUserCommunities - success",
			req: &community_service.GetUserCommunitiesRequest{
				UserId: userID.String(),
				Count:  10,
				Ts:     timestamppb.New(ts),
			},
			expectedResp: &community_service.GetUserCommunitiesResponse{
				Communities: []*community_service.Community{
					{
						Id:       communityID.String(),
						OwnerId:  userID.String(),
						Nickname: "Test Community",
						Name:     "Test Community Name",
					},
				},
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					GetUserCommunities(gomock.Any(), userID, 10, gomock.Any()).
					Return([]models.Community{
						{
							ID:       communityID,
							OwnerID:  userID,
							NickName: "Test Community",
							BasicInfo: &models.BasicCommunityInfo{
								Name: "Test Community Name",
							},
						},
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.GetUserCommunities(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, len(tt.expectedResp.Communities), len(resp.Communities))
				if len(resp.Communities) > 0 {
					assert.Equal(t, tt.expectedResp.Communities[0].Id, resp.Communities[0].Id)
					assert.Equal(t, tt.expectedResp.Communities[0].Nickname, resp.Communities[0].Nickname)
				}
			}
		})
	}
}

func TestSearchSimilarCommunities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	communityID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name         string
		req          *community_service.SearchSimilarCommunitiesRequest
		expectedResp *community_service.SearchSimilarCommunitiesResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "SearchSimilarCommunities - success",
			req: &community_service.SearchSimilarCommunitiesRequest{
				Name:  "Test",
				Count: 5,
			},
			expectedResp: &community_service.SearchSimilarCommunitiesResponse{
				Communities: []*community_service.Community{
					{
						Id:       communityID.String(),
						OwnerId:  userID.String(),
						Nickname: "Test Community",
						Name:     "Test Community Name",
					},
				},
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					SearchSimilarCommunities(gomock.Any(), "Test", 5).
					Return([]models.Community{
						{
							ID:       communityID,
							OwnerID:  userID,
							NickName: "Test Community",
							BasicInfo: &models.BasicCommunityInfo{
								Name: "Test Community Name",
							},
						},
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.SearchSimilarCommunities(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, len(tt.expectedResp.Communities), len(resp.Communities))
				if len(resp.Communities) > 0 {
					assert.Equal(t, tt.expectedResp.Communities[0].Id, resp.Communities[0].Id)
					assert.Equal(t, tt.expectedResp.Communities[0].Nickname, resp.Communities[0].Nickname)
				}
			}
		})
	}
}

func TestChangeUserRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()
	requesterID := uuid.New()

	tests := []struct {
		name         string
		req          *community_service.ChangeUserRoleRequest
		expectedResp *community_service.ChangeUserRoleResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "ChangeUserRole - success",
			req: &community_service.ChangeUserRoleRequest{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
				Role:        community_service.CommunityRole_COMMUNITY_ROLE_ADMIN,
				RequesterId: requesterID.String(),
			},
			expectedResp: &community_service.ChangeUserRoleResponse{
				Success: true,
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					ChangeUserRole(gomock.Any(), userID, communityID, models.CommunityRoleAdmin, requesterID).
					Return(nil)
			},
		},
		{
			name: "ChangeUserRole - invalid user ID",
			req: &community_service.ChangeUserRoleRequest{
				UserId:      "invalid",
				CommunityId: communityID.String(),
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid user ID"),
			mockSetup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.ChangeUserRole(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Success, resp.Success)
			}
		})
	}
}

func TestGetControlledCommunities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCommunityUseCase := mocks.NewMockCommunityUseCase(ctrl)
	server := NewCommunityServiceServer(mockCommunityUseCase)

	userID := uuid.New()
	communityID := uuid.New()
	ts := time.Now()

	tests := []struct {
		name         string
		req          *community_service.GetControlledCommunitiesRequest
		expectedResp *community_service.GetControlledCommunitiesResponse
		expectedErr  error
		mockSetup    func()
	}{
		{
			name: "GetControlledCommunities - success",
			req: &community_service.GetControlledCommunitiesRequest{
				UserId: userID.String(),
				Count:  10,
				Ts:     timestamppb.New(ts),
			},
			expectedResp: &community_service.GetControlledCommunitiesResponse{
				Communities: []*community_service.Community{
					{
						Id:       communityID.String(),
						OwnerId:  userID.String(),
						Nickname: "Controlled Community",
						Name:     "Controlled Community Name",
					},
				},
			},
			expectedErr: nil,
			mockSetup: func() {
				mockCommunityUseCase.EXPECT().
					GetControlledCommunities(gomock.Any(), userID, 10, gomock.Any()).
					Return([]models.Community{
						{
							ID:       communityID,
							OwnerID:  userID,
							NickName: "Controlled Community",
							BasicInfo: &models.BasicCommunityInfo{
								Name: "Controlled Community Name",
							},
						},
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			resp, err := server.GetControlledCommunities(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedResp != nil {
				assert.Equal(t, len(tt.expectedResp.Communities), len(resp.Communities))
				if len(resp.Communities) > 0 {
					assert.Equal(t, tt.expectedResp.Communities[0].Id, resp.Communities[0].Id)
					assert.Equal(t, tt.expectedResp.Communities[0].Nickname, resp.Communities[0].Nickname)
				}
			}
		})
	}
}
