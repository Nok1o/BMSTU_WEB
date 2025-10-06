//go:build unit
// +build unit

package community_service

import (
	"context"
	pb "quickflow/shared/proto/community_service"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	mocks "quickflow/shared/proto/community_service/mocks"
)

func TestClient_CreateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	now := time.Now()
	ownerID := uuid.New()
	community := &models.Community{
		NickName: "test-community",
		OwnerID:  ownerID,
		BasicInfo: &models.BasicCommunityInfo{
			Name:        "Test Community",
			Description: "Test Description",
			AvatarUrl:   "http://example.com/avatar.jpg",
			CoverUrl:    "http://example.com/cover.jpg",
		},
		Avatar: &models.File{URL: uuid.New().String()},
		Cover:  &models.File{URL: uuid.New().String()},
	}

	expectedProtoCommunity := &pb.Community{
		Id:          uuid.New().String(),
		OwnerId:     ownerID.String(),
		Name:        "Test Community",
		Description: "Test Description",
		AvatarUrl:   "http://example.com/avatar.jpg",
		CoverUrl:    "http://example.com/cover.jpg",
		Nickname:    "test-community",
		CreatedAt:   timestamppb.New(now),
	}

	tests := []struct {
		name          string
		setupMock     func()
		input         *models.Community
		expected      *models.Community
		expectedError bool
	}{
		{
			name: "successful creation",
			setupMock: func() {
				mockClient.EXPECT().CreateCommunity(gomock.Any(), &pb.CreateCommunityRequest{
					Name:        community.BasicInfo.Name,
					Nickname:    community.NickName,
					Description: community.BasicInfo.Description,
					AvatarUrl:   community.BasicInfo.AvatarUrl,
					CoverUrl:    community.BasicInfo.CoverUrl,
					Avatar:      file_service.ModelFileToProto(community.Avatar),
					Cover:       file_service.ModelFileToProto(community.Cover),
					OwnerId:     community.OwnerID.String(),
				}).Return(&pb.CreateCommunityResponse{
					Community: expectedProtoCommunity,
				}, nil)
			},
			input: community,
			expected: &models.Community{
				ID:       uuid.MustParse(expectedProtoCommunity.Id),
				OwnerID:  ownerID,
				NickName: "test-community",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "Test Description",
					AvatarUrl:   "http://example.com/avatar.jpg",
					CoverUrl:    "http://example.com/cover.jpg",
				},
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().CreateCommunity(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			input:         community,
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := client.CreateCommunity(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.OwnerID, result.OwnerID)
			assert.Equal(t, tt.expected.NickName, result.NickName)
			assert.Equal(t, tt.expected.BasicInfo.Name, result.BasicInfo.Name)
		})
	}
}

func TestClient_GetCommunityById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	id := uuid.New()
	protoCommunity := &pb.Community{
		Id:        id.String(),
		OwnerId:   id.String(),
		Name:      "Test Community",
		Nickname:  "test-comm",
		AvatarUrl: "http://example.com/avatar.jpg",
	}

	tests := []struct {
		name          string
		setupMock     func()
		input         uuid.UUID
		expected      *models.Community
		expectedError bool
	}{
		{
			name: "successful get",
			setupMock: func() {
				mockClient.EXPECT().GetCommunityById(gomock.Any(), &pb.GetCommunityByIdRequest{
					CommunityId: id.String(),
				}).Return(&pb.GetCommunityByIdResponse{
					Community: protoCommunity,
				}, nil)
			},
			input: id,
			expected: &models.Community{
				ID:       id,
				NickName: "test-comm",
				BasicInfo: &models.BasicCommunityInfo{
					Name:      "Test Community",
					AvatarUrl: "http://example.com/avatar.jpg",
				},
			},
			expectedError: false,
		},
		{
			name: "not found",
			setupMock: func() {
				mockClient.EXPECT().GetCommunityById(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			input:         id,
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := client.GetCommunityById(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.NickName, result.NickName)
			assert.Equal(t, tt.expected.BasicInfo.Name, result.BasicInfo.Name)
		})
	}
}

func TestClient_IsCommunityMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	userID := uuid.New()
	communityID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func()
		userID        uuid.UUID
		communityID   uuid.UUID
		expectedIs    bool
		expectedRole  *models.CommunityRole
		expectedError bool
	}{
		{
			name: "is member with role",
			setupMock: func() {
				mockClient.EXPECT().IsCommunityMember(gomock.Any(), &pb.IsCommunityMemberRequest{
					UserId:      userID.String(),
					CommunityId: communityID.String(),
				}).Return(&pb.IsCommunityMemberResponse{
					IsMember: true,
					Role:     pb.CommunityRole_COMMUNITY_ROLE_ADMIN,
				}, nil)
			},
			userID:        userID,
			communityID:   communityID,
			expectedIs:    true,
			expectedRole:  func() *models.CommunityRole { r := models.CommunityRoleAdmin; return &r }(),
			expectedError: false,
		},
		{
			name: "is not member",
			setupMock: func() {
				mockClient.EXPECT().IsCommunityMember(gomock.Any(), gomock.Any()).
					Return(&pb.IsCommunityMemberResponse{
						IsMember: false,
						Role:     -1,
					}, nil)
			},
			userID:        userID,
			communityID:   communityID,
			expectedIs:    false,
			expectedRole:  nil,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().IsCommunityMember(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			userID:        userID,
			communityID:   communityID,
			expectedIs:    false,
			expectedRole:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			is, role, err := client.IsCommunityMember(context.Background(), tt.userID, tt.communityID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedIs, is)
			assert.Equal(t, tt.expectedRole, role)
		})
	}
}

func TestClient_GetCommunityMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	communityID := uuid.New()
	userID := uuid.New()
	ts := time.Now()

	protoMembers := []*pb.CommunityMember{
		{
			UserId:      userID.String(),
			CommunityId: communityID.String(),
			Role:        pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
			JoinedAt:    timestamppb.New(ts),
		},
	}

	tests := []struct {
		name          string
		setupMock     func()
		communityID   uuid.UUID
		count         int
		ts            time.Time
		expected      []*models.CommunityMember
		expectedError bool
	}{
		{
			name: "successful get members",
			setupMock: func() {
				mockClient.EXPECT().GetCommunityMembers(gomock.Any(), &pb.GetCommunityMembersRequest{
					CommunityId: communityID.String(),
					Count:       10,
					Ts:          timestamppb.New(ts),
				}).Return(&pb.GetCommunityMembersResponse{
					Members: protoMembers,
				}, nil)
			},
			communityID: communityID,
			count:       10,
			ts:          ts,
			expected: []*models.CommunityMember{
				{
					UserID:      userID,
					CommunityID: communityID,
					Role:        models.CommunityRoleMember,
					JoinedAt:    ts,
				},
			},
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().GetCommunityMembers(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			communityID:   communityID,
			count:         10,
			ts:            ts,
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := client.GetCommunityMembers(context.Background(), tt.communityID, tt.count, tt.ts)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))
			if len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].UserID, result[0].UserID)
				assert.Equal(t, tt.expected[0].Role, result[0].Role)
			}
		})
	}
}

func TestClient_DeleteCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	communityID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func()
		communityID   uuid.UUID
		userID        uuid.UUID
		expectedError bool
	}{
		{
			name: "successful delete",
			setupMock: func() {
				mockClient.EXPECT().DeleteCommunity(gomock.Any(), &pb.DeleteCommunityRequest{
					CommunityId: communityID.String(),
					UserId:      userID.String(),
				}).Return(&pb.DeleteCommunityResponse{}, nil)
			},
			communityID:   communityID,
			userID:        userID,
			expectedError: false,
		},
		{
			name: "error from server",
			setupMock: func() {
				mockClient.EXPECT().DeleteCommunity(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			communityID:   communityID,
			userID:        userID,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := client.DeleteCommunity(context.Background(), tt.communityID, tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestClient_UpdateCommunity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockCommunityServiceClient(ctrl)
	client := &Client{client: mockClient}

	communityID := uuid.New()
	userID := uuid.New()
	community := &models.Community{
		ID:       communityID,
		NickName: "updated-comm",
		BasicInfo: &models.BasicCommunityInfo{
			Name:        "Updated Community",
			Description: "Updated Description",
		},
	}

	updatedProto := &pb.Community{
		Id:          communityID.String(),
		OwnerId:     userID.String(),
		Name:        "Updated Community",
		Nickname:    "updated-comm",
		Description: "Updated Description",
	}

	tests := []struct {
		name          string
		setupMock     func()
		community     *models.Community
		userID        uuid.UUID
		expected      *models.Community
		expectedError bool
	}{
		{
			name: "successful update",
			setupMock: func() {
				mockClient.EXPECT().UpdateCommunity(gomock.Any(), &pb.UpdateCommunityRequest{
					Id:          communityID.String(),
					UserId:      userID.String(),
					Name:        community.BasicInfo.Name,
					Nickname:    community.NickName,
					Description: community.BasicInfo.Description,
				}).Return(&pb.UpdateCommunityResponse{
					Community: updatedProto,
				}, nil)
			},
			community: community,
			userID:    userID,
			expected: &models.Community{
				ID:       communityID,
				NickName: "updated-comm",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Updated Community",
					Description: "Updated Description",
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := client.UpdateCommunity(context.Background(), tt.community, tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.NickName, result.NickName)
			assert.Equal(t, tt.expected.BasicInfo.Name, result.BasicInfo.Name)
		})
	}
}
