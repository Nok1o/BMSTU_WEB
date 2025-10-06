//go:build unit
// +build unit

package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
	"quickflow/user_service/internal/delivery/grpc/mocks"
	user_errors "quickflow/user_service/internal/errors"
)

func TestProfileServiceServer_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.CreateProfileRequest
		mockSetup   func()
		want        *pb.CreateProfileResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.CreateProfileRequest{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			mockSetup: func() {
				mockUC.EXPECT().CreateProfile(ctx, gomock.Any()).Return(shared_models.Profile{
					UserId:   userID,
					Username: "testuser",
					BasicInfo: &shared_models.BasicInfo{
						Name:    "John",
						Surname: "Doe",
					},
					LastSeen: now,
				}, nil)
			},
			want: &pb.CreateProfileResponse{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			wantErr: false,
		},
		{
			name: "nil profile",
			req: &pb.CreateProfileRequest{
				Profile: nil,
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidProfileInfo,
		},
		{
			name: "invalid profile data",
			req: &pb.CreateProfileRequest{
				Profile: &pb.Profile{
					Id: "invalid-uuid",
				},
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidProfileInfo,
		},
		{
			name: "use case error",
			req: &pb.CreateProfileRequest{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
				},
			},
			mockSetup: func() {
				mockUC.EXPECT().CreateProfile(ctx, gomock.Any()).Return(shared_models.Profile{}, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.CreateProfile(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Profile.Id, got.Profile.Id)
			assert.Equal(t, tt.want.Profile.Username, got.Profile.Username)
			assert.Equal(t, tt.want.Profile.LastSeen, got.Profile.LastSeen)
		})
	}
}

func TestProfileServiceServer_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.UpdateProfileRequest
		mockSetup   func()
		want        *pb.UpdateProfileResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.UpdateProfileRequest{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "updateduser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "Updated",
						Lastname:  "Name",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			mockSetup: func() {
				mockUC.EXPECT().UpdateProfile(ctx, gomock.Any()).Return(&shared_models.Profile{
					UserId:   userID,
					Username: "updateduser",
					BasicInfo: &shared_models.BasicInfo{
						Name:    "Updated",
						Surname: "Name",
					},
					LastSeen: now,
				}, nil)
			},
			want: &pb.UpdateProfileResponse{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "updateduser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "Updated",
						Lastname:  "Name",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			wantErr: false,
		},
		{
			name: "nil profile",
			req: &pb.UpdateProfileRequest{
				Profile: nil,
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidProfileInfo,
		},
		{
			name: "use case error",
			req: &pb.UpdateProfileRequest{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "updateduser",
				},
			},
			mockSetup: func() {
				mockUC.EXPECT().UpdateProfile(ctx, gomock.Any()).Return(nil, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.UpdateProfile(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Profile.Id, got.Profile.Id)
			assert.Equal(t, tt.want.Profile.Username, got.Profile.Username)
			assert.Equal(t, tt.want.Profile.LastSeen, got.Profile.LastSeen)
		})
	}
}

func TestProfileServiceServer_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetProfileRequest
		mockSetup   func()
		want        *pb.GetProfileResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.GetProfileRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().GetProfile(ctx, userID).Return(shared_models.Profile{
					UserId:   userID,
					Username: "testuser",
					BasicInfo: &shared_models.BasicInfo{
						Name:    "John",
						Surname: "Doe",
					},
					LastSeen: now,
				}, nil)
			},
			want: &pb.GetProfileResponse{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetProfileRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidUserId,
		},
		{
			name: "use case error",
			req: &pb.GetProfileRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().GetProfile(ctx, userID).Return(shared_models.Profile{}, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.GetProfile(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Profile.Id, got.Profile.Id)
			assert.Equal(t, tt.want.Profile.Username, got.Profile.Username)
			assert.Equal(t, tt.want.Profile.LastSeen, got.Profile.LastSeen)
		})
	}
}

func TestProfileServiceServer_GetProfileByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetProfileByUsernameRequest
		mockSetup   func()
		want        *pb.GetProfileByUsernameResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.GetProfileByUsernameRequest{
				Username: "testuser",
			},
			mockSetup: func() {
				mockUC.EXPECT().GetProfileByUsername(ctx, "testuser").Return(shared_models.Profile{
					UserId:   userID,
					Username: "testuser",
					BasicInfo: &shared_models.BasicInfo{
						Name:    "John",
						Surname: "Doe",
					},
					LastSeen: now,
				}, nil)
			},
			want: &pb.GetProfileByUsernameResponse{
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "John",
						Lastname:  "Doe",
					},
					LastSeen: timestamppb.New(now),
				},
			},
			wantErr: false,
		},
		{
			name: "empty username",
			req: &pb.GetProfileByUsernameRequest{
				Username: "",
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrUserValidation,
		},
		{
			name: "use case error",
			req: &pb.GetProfileByUsernameRequest{
				Username: "testuser",
			},
			mockSetup: func() {
				mockUC.EXPECT().GetProfileByUsername(ctx, "testuser").Return(shared_models.Profile{}, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.GetProfileByUsername(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Profile.Id, got.Profile.Id)
			assert.Equal(t, tt.want.Profile.Username, got.Profile.Username)
			assert.Equal(t, tt.want.Profile.LastSeen, got.Profile.LastSeen)
		})
	}
}

func TestProfileServiceServer_UpdateLastSeen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.UpdateLastSeenRequest
		mockSetup   func()
		want        *pb.UpdateLastSeenResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.UpdateLastSeenRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().UpdateLastSeen(ctx, userID).Return(nil)
			},
			want: &pb.UpdateLastSeenResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.UpdateLastSeenRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidUserId,
		},
		{
			name: "use case error",
			req: &pb.UpdateLastSeenRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().UpdateLastSeen(ctx, userID).Return(assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.UpdateLastSeen(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProfileServiceServer_GetPublicUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetPublicUserInfoRequest
		mockSetup   func()
		want        *pb.GetPublicUserInfoResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.GetPublicUserInfoRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().GetPublicUserInfo(ctx, userID).Return(shared_models.PublicUserInfo{
					Id:        userID,
					Username:  "testuser",
					Firstname: "John",
					Lastname:  "Doe",
					AvatarURL: "http://example.com/avatar.jpg",
					LastSeen:  now,
				}, nil)
			},
			want: &pb.GetPublicUserInfoResponse{
				UserInfo: &pb.PublicUserInfo{
					Id:        userID.String(),
					Username:  "testuser",
					Firstname: "John",
					Lastname:  "Doe",
					AvatarUrl: "http://example.com/avatar.jpg",
					LastSeen:  timestamppb.New(now),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetPublicUserInfoRequest{
				UserId: "invalid-uuid",
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidUserId,
		},
		{
			name: "use case error",
			req: &pb.GetPublicUserInfoRequest{
				UserId: userID.String(),
			},
			mockSetup: func() {
				mockUC.EXPECT().GetPublicUserInfo(ctx, userID).Return(shared_models.PublicUserInfo{}, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.GetPublicUserInfo(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProfileServiceServer_GetPublicUsersInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUseCase(ctrl)
	server := NewProfileServiceServer(mockUC)

	ctx := context.Background()
	userID1 := uuid.New()
	userID2 := uuid.New()
	now := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetPublicUsersInfoRequest
		mockSetup   func()
		want        *pb.GetPublicUsersInfoResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			req: &pb.GetPublicUsersInfoRequest{
				UserIds: []string{userID1.String(), userID2.String()},
			},
			mockSetup: func() {
				mockUC.EXPECT().GetPublicUsersInfo(ctx, []uuid.UUID{userID1, userID2}).Return([]shared_models.PublicUserInfo{
					{
						Id:        userID1,
						Username:  "user1",
						Firstname: "John",
						Lastname:  "Doe",
						AvatarURL: "http://example.com/avatar1.jpg",
						LastSeen:  now,
					},
					{
						Id:        userID2,
						Username:  "user2",
						Firstname: "Jane",
						Lastname:  "Smith",
						AvatarURL: "http://example.com/avatar2.jpg",
						LastSeen:  now,
					},
				}, nil)
			},
			want: &pb.GetPublicUsersInfoResponse{
				UsersInfo: []*pb.PublicUserInfo{
					{
						Id:        userID1.String(),
						Username:  "user1",
						Firstname: "John",
						Lastname:  "Doe",
						AvatarUrl: "http://example.com/avatar1.jpg",
						LastSeen:  timestamppb.New(now),
					},
					{
						Id:        userID2.String(),
						Username:  "user2",
						Firstname: "Jane",
						Lastname:  "Smith",
						AvatarUrl: "http://example.com/avatar2.jpg",
						LastSeen:  timestamppb.New(now),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty user ids",
			req: &pb.GetPublicUsersInfoRequest{
				UserIds: []string{},
			},
			mockSetup: func() {},
			want: &pb.GetPublicUsersInfoResponse{
				UsersInfo: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			req: &pb.GetPublicUsersInfoRequest{
				UserIds: []string{"invalid-uuid"},
			},
			mockSetup:   func() {},
			want:        nil,
			wantErr:     true,
			expectedErr: user_errors.ErrInvalidUserId,
		},
		{
			name: "use case error",
			req: &pb.GetPublicUsersInfoRequest{
				UserIds: []string{userID1.String()},
			},
			mockSetup: func() {
				mockUC.EXPECT().GetPublicUsersInfo(ctx, []uuid.UUID{userID1}).Return(nil, assert.AnError)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := server.GetPublicUsersInfo(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
