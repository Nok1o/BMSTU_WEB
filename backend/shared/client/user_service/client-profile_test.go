//go:build unit
// +build unit

package userclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
	"quickflow/shared/proto/user_service/mocks"
)

func TestNewProfileClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConn := grpc.ClientConn{}
	client := NewProfileClient(&mockConn)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestCreateProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	profile := shared_models.Profile{
		UserId:   uuid.New(),
		Username: "testuser",
	}

	expectedRequest := &pb.CreateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	}
	expectedResponse := &pb.CreateProfileResponse{
		Profile: MapProfileToProfileDTO(&profile),
	}

	mockClient.EXPECT().
		CreateProfile(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.CreateProfile(ctx, profile)

	assert.NoError(t, err)
	assert.Equal(t, profile.UserId, result.UserId)
	assert.Equal(t, profile.Username, result.Username)
}

func TestCreateProfile_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	profile := shared_models.Profile{
		UserId:   uuid.New(),
		Username: "testuser",
	}

	expectedErr := errors.New("grpc error")

	mockClient.EXPECT().
		CreateProfile(ctx, gomock.Any()).
		Return(nil, expectedErr)

	_, err := client.CreateProfile(ctx, profile)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestUpdateProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	profile := shared_models.Profile{
		UserId:   uuid.New(),
		Username: "updateduser",
	}

	expectedRequest := &pb.UpdateProfileRequest{
		Profile: MapProfileToProfileDTO(&profile),
	}
	expectedResponse := &pb.UpdateProfileResponse{
		Profile: MapProfileToProfileDTO(&profile),
	}

	mockClient.EXPECT().
		UpdateProfile(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.UpdateProfile(ctx, profile)

	assert.NoError(t, err)
	assert.Equal(t, profile.UserId, result.UserId)
	assert.Equal(t, profile.Username, result.Username)
}

func TestUpdateProfile_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	profile := shared_models.Profile{
		UserId: uuid.New(),
	}

	expectedErr := errors.New("update error")

	mockClient.EXPECT().
		UpdateProfile(ctx, gomock.Any()).
		Return(nil, expectedErr)

	_, err := client.UpdateProfile(ctx, profile)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	profile := shared_models.Profile{
		UserId:   userID,
		Username: "testuser",
	}

	expectedRequest := &pb.GetProfileRequest{
		UserId: userID.String(),
	}
	expectedResponse := &pb.GetProfileResponse{
		Profile: MapProfileToProfileDTO(&profile),
	}

	mockClient.EXPECT().
		GetProfile(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.GetProfile(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, result.UserId)
	assert.Equal(t, "testuser", result.Username)
}

func TestGetProfile_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()

	expectedErr := errors.New("not found")

	mockClient.EXPECT().
		GetProfile(ctx, &pb.GetProfileRequest{UserId: userID.String()}).
		Return(nil, expectedErr)

	_, err := client.GetProfile(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetProfileByUsername_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	username := "testuser"
	profile := shared_models.Profile{
		UserId:   uuid.New(),
		Username: username,
	}

	expectedRequest := &pb.GetProfileByUsernameRequest{
		Username: username,
	}
	expectedResponse := &pb.GetProfileByUsernameResponse{
		Profile: MapProfileToProfileDTO(&profile),
	}

	mockClient.EXPECT().
		GetProfileByUsername(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.GetProfileByUsername(ctx, username)

	assert.NoError(t, err)
	assert.Equal(t, username, result.Username)
}

func TestGetProfileByUsername_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	username := "nonexistent"

	expectedErr := errors.New("not found")

	mockClient.EXPECT().
		GetProfileByUsername(ctx, &pb.GetProfileByUsernameRequest{Username: username}).
		Return(nil, expectedErr)

	_, err := client.GetProfileByUsername(ctx, username)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestUpdateLastSeen_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()

	expectedRequest := &pb.UpdateLastSeenRequest{
		UserId: userID.String(),
	}
	expectedResponse := &pb.UpdateLastSeenResponse{}

	mockClient.EXPECT().
		UpdateLastSeen(ctx, expectedRequest).
		Return(expectedResponse, nil)

	err := client.UpdateLastSeen(ctx, userID)

	assert.NoError(t, err)
}

func TestUpdateLastSeen_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()

	expectedErr := errors.New("update failed")

	mockClient.EXPECT().
		UpdateLastSeen(ctx, &pb.UpdateLastSeenRequest{UserId: userID.String()}).
		Return(nil, expectedErr)

	err := client.UpdateLastSeen(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetPublicUserInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()
	publicInfo := shared_models.PublicUserInfo{
		Id:        userID,
		Username:  "publicuser",
		Firstname: "John",
		Lastname:  "Doe",
		LastSeen:  time.Now(),
	}

	expectedRequest := &pb.GetPublicUserInfoRequest{
		UserId: userID.String(),
	}
	expectedResponse := &pb.GetPublicUserInfoResponse{
		UserInfo: MapPublicUserInfoToDTO(&publicInfo),
	}

	mockClient.EXPECT().
		GetPublicUserInfo(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.GetPublicUserInfo(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, result.Id)
	assert.Equal(t, "publicuser", result.Username)
	assert.Equal(t, "John", result.Firstname)
	assert.Equal(t, "Doe", result.Lastname)
}

func TestGetPublicUserInfo_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userID := uuid.New()

	expectedErr := errors.New("not found")

	mockClient.EXPECT().
		GetPublicUserInfo(ctx, &pb.GetPublicUserInfoRequest{UserId: userID.String()}).
		Return(nil, expectedErr)

	_, err := client.GetPublicUserInfo(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetPublicUsersInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	publicInfos := []shared_models.PublicUserInfo{
		{
			Id:       userIDs[0],
			Username: "user1",
		},
		{
			Id:       userIDs[1],
			Username: "user2",
		},
	}

	expectedRequest := &pb.GetPublicUsersInfoRequest{
		UserIds: []string{userIDs[0].String(), userIDs[1].String()},
	}
	expectedResponse := &pb.GetPublicUsersInfoResponse{
		UsersInfo: []*pb.PublicUserInfo{
			MapPublicUserInfoToDTO(&publicInfos[0]),
			MapPublicUserInfoToDTO(&publicInfos[1]),
		},
	}

	mockClient.EXPECT().
		GetPublicUsersInfo(ctx, expectedRequest).
		Return(expectedResponse, nil)

	result, err := client.GetPublicUsersInfo(ctx, userIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, userIDs[0], result[0].Id)
	assert.Equal(t, "user1", result[0].Username)
	assert.Equal(t, userIDs[1], result[1].Id)
	assert.Equal(t, "user2", result[1].Username)
}

func TestGetPublicUsersInfo_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockProfileServiceClient(ctrl)
	client := &ProfileClient{client: mockClient}

	ctx := context.Background()
	userIDs := []uuid.UUID{uuid.New()}

	expectedErr := errors.New("batch error")

	mockClient.EXPECT().
		GetPublicUsersInfo(ctx, &pb.GetPublicUsersInfoRequest{UserIds: []string{userIDs[0].String()}}).
		Return(nil, expectedErr)

	_, err := client.GetPublicUsersInfo(ctx, userIDs)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
