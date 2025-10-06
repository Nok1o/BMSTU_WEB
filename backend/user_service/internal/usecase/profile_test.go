//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/usecase"
	"quickflow/user_service/internal/usecase/mocks"
)

type ProfileServiceSuite struct {
	suite.Suite
}

func TestProfileService(t *testing.T) {
	suite.RunSuite(t, new(ProfileServiceSuite))
}

func (s *ProfileServiceSuite) TestProfileService_GetUserInfo(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование получения информации о пользователе")

	uuid_ := uuid.New()

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(*mocks.MockProfileRepository)
		expected    models.Profile
		expectedErr error
	}{
		{
			name:   "success",
			userId: uuid.New(),
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{UserId: uuid_}, nil)
			},
			expected:    models.Profile{UserId: uuid_},
			expectedErr: nil,
		},
		{
			name:   "not found",
			userId: uuid.New(),
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{}, user_errors.ErrNotFound)
			},
			expected:    models.Profile{},
			expectedErr: user_errors.ErrNotFound,
		},
		{
			name:   "repository error",
			userId: uuid.New(),
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{}, errors.New("db error"))
			},
			expected:    models.Profile{},
			expectedErr: errors.New("p.profileRepo.GetProfile: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockProfileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			result, err := service.GetUserInfo(context.Background(), tt.userId)

			assert.Equal(t, tt.expected.UserId, result.UserId)
			assert.Equal(t, tt.expected.Username, result.Username)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileServiceSuite) TestProfileService_GetUserInfoByUserName(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование получения информации о пользователе по имени пользователя")

	userId := uuid.New()

	tests := []struct {
		name        string
		username    string
		mockSetup   func(*mocks.MockUserRepository, *mocks.MockProfileRepository)
		expected    models.Profile
		expectedErr error
	}{
		{
			name:     "success",
			username: "testuser",
			mockSetup: func(mur *mocks.MockUserRepository, mpr *mocks.MockProfileRepository) {
				mur.EXPECT().GetUserByUsername(gomock.Any(), "testuser").
					Return(models.User{Id: userId, Username: "testuser"}, nil)
				mpr.EXPECT().GetProfile(gomock.Any(), userId).
					Return(models.Profile{UserId: userId}, nil)
			},
			expected:    models.Profile{UserId: userId, Username: "testuser"},
			expectedErr: nil,
		},
		{
			name:     "user not found",
			username: "unknown",
			mockSetup: func(mur *mocks.MockUserRepository, mpr *mocks.MockProfileRepository) {
				mur.EXPECT().GetUserByUsername(gomock.Any(), "unknown").
					Return(models.User{}, user_errors.ErrNotFound)
			},
			expected:    models.Profile{},
			expectedErr: user_errors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockUserRepo, mockProfileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			result, err := service.GetUserInfoByUserName(context.Background(), tt.username)

			if tt.expected.UserId != uuid.Nil {
				assert.Equal(t, tt.expected.UserId, result.UserId)
			}
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileServiceSuite) TestProfileService_UpdateProfile(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование обновления профиля пользователя")

	validProfile := models.Profile{
		UserId: uuid.New(),
		BasicInfo: &models.BasicInfo{
			Name:    "Valid",
			Surname: "Name",
		},
	}

	invalidProfile := models.Profile{
		UserId: uuid.New(),
		BasicInfo: &models.BasicInfo{
			Name:    "", // Invalid name
			Surname: "",
		},
	}

	tests := []struct {
		name        string
		profile     models.Profile
		mockSetup   func(*mocks.MockProfileRepository, *mocks.MockUserRepository, *mocks.MockFileService)
		expectedErr error
	}{
		{
			name:    "success text only update",
			profile: validProfile,
			mockSetup: func(mpr *mocks.MockProfileRepository, mur *mocks.MockUserRepository, mfr *mocks.MockFileService) {
				mpr.EXPECT().GetProfile(gomock.Any(), validProfile.UserId).Return(models.Profile{UserId: validProfile.UserId}, nil)
				mpr.EXPECT().UpdateProfileTextInfo(gomock.Any(), validProfile).Return(nil)
				mpr.EXPECT().GetProfile(gomock.Any(), validProfile.UserId).Return(validProfile, nil)
			},
			expectedErr: nil,
		},
		{
			name: "success with avatar upload",
			profile: models.Profile{
				UserId: uuid.New(),
				BasicInfo: &models.BasicInfo{
					Name:    "Valid",
					Surname: "Name",
				},
				Avatar: &models.File{},
			},
			mockSetup: func(mpr *mocks.MockProfileRepository, mur *mocks.MockUserRepository, mfr *mocks.MockFileService) {
				mpr.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(models.Profile{UserId: uuid.New()}, nil)
				mpr.EXPECT().UpdateProfileTextInfo(gomock.Any(), gomock.Any()).Return(nil)
				mfr.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return("http://avatar.url", nil)
				mpr.EXPECT().UpdateProfileAvatar(gomock.Any(), gomock.Any(), "http://avatar.url").Return(nil)
				mpr.EXPECT().GetProfile(gomock.Any(), gomock.Any()).Return(models.Profile{UserId: uuid.New()}, nil)
			},
			expectedErr: nil,
		},
		{
			name:    "invalid profile info",
			profile: invalidProfile,
			mockSetup: func(mpr *mocks.MockProfileRepository, mur *mocks.MockUserRepository, mfr *mocks.MockFileService) {
				// No expectations as validation should fail first
			},
			expectedErr: user_errors.ErrInvalidProfileInfo,
		},
		{
			name: "username taken",
			profile: models.Profile{
				UserId:   uuid.New(),
				Username: "taken",
				BasicInfo: &models.BasicInfo{
					Name:    "Valid",
					Surname: "Name",
				},
			},
			mockSetup: func(mpr *mocks.MockProfileRepository, mur *mocks.MockUserRepository, mfr *mocks.MockFileService) {
				mur.EXPECT().GetUserByUsername(gomock.Any(), "taken").
					Return(models.User{Id: uuid.New()}, nil) // Different user with same username
			},
			expectedErr: user_errors.ErrAlreadyExists,
		},
		{
			name:    "error getting old profile",
			profile: validProfile,
			mockSetup: func(mpr *mocks.MockProfileRepository, mur *mocks.MockUserRepository, mfr *mocks.MockFileService) {
				mpr.EXPECT().GetProfile(gomock.Any(), validProfile.UserId).
					Return(models.Profile{}, errors.New("db error"))
			},
			expectedErr: errors.New("p.profileRepo.GetProfile: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockProfileRepo, mockUserRepo, mockFileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			_, err := service.UpdateProfile(context.Background(), tt.profile)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileServiceSuite) TestProfileService_GetPublicUserInfo(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Service")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование получения публичной информации о пользователе")

	uuid_ := uuid.New()

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(*mocks.MockProfileRepository)
		expected    models.PublicUserInfo
		expectedErr error
	}{
		{
			name:   "success",
			userId: uuid_,
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().GetPublicUserInfo(gomock.Any(), gomock.Any()).
					Return(models.PublicUserInfo{Id: uuid_}, nil)
			},
			expected:    models.PublicUserInfo{Id: uuid_},
			expectedErr: nil,
		},
		{
			name:   "invalid user id",
			userId: uuid.Nil,
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				// No expectations as validation should fail first
			},
			expected:    models.PublicUserInfo{},
			expectedErr: user_errors.ErrInvalidUserId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockProfileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			result, err := service.GetPublicUserInfo(context.Background(), tt.userId)

			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileServiceSuite) TestProfileService_CreateProfile(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование создания профиля пользователя")

	validProfile := models.Profile{
		UserId: uuid.New(),
		BasicInfo: &models.BasicInfo{
			Name:    "Valid",
			Surname: "Name",
		},
	}

	invalidProfile := models.Profile{
		UserId: uuid.New(),
		BasicInfo: &models.BasicInfo{
			Name:    "", // Invalid
			Surname: "",
		},
	}

	tests := []struct {
		name        string
		profile     models.Profile
		mockSetup   func(*mocks.MockProfileRepository)
		expected    models.Profile
		expectedErr error
	}{
		{
			name:    "success",
			profile: validProfile,
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().SaveProfile(gomock.Any(), validProfile).Return(nil)
			},
			expected:    validProfile,
			expectedErr: nil,
		},
		{
			name:        "invalid profile",
			profile:     invalidProfile,
			mockSetup:   func(mpr *mocks.MockProfileRepository) {},
			expected:    models.Profile{},
			expectedErr: user_errors.ErrInvalidProfileInfo,
		},
		{
			name:    "repository error",
			profile: validProfile,
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().SaveProfile(gomock.Any(), validProfile).Return(errors.New("db error"))
			},
			expected:    models.Profile{},
			expectedErr: errors.New("p.profileRepo.SaveProfile: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockProfileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			result, err := service.CreateProfile(context.Background(), tt.profile)

			assert.Equal(t, tt.expected, result)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileServiceSuite) TestProfileService_UpdateLastSeen(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Service")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование обновления времени последнего посещения пользователя")

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(*mocks.MockProfileRepository)
		expectedErr error
	}{
		{
			name:   "success",
			userId: uuid.New(),
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().UpdateLastSeen(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "repository error",
			userId: uuid.New(),
			mockSetup: func(mpr *mocks.MockProfileRepository) {
				mpr.EXPECT().UpdateLastSeen(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedErr: errors.New("a.userRepo.UpdateLastSeen: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := mocks.NewMockProfileRepository(ctrl)
			mockUserRepo := mocks.NewMockUserRepository(ctrl)
			mockFileRepo := mocks.NewMockFileService(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockProfileRepo)
			}

			service := usecase.NewProfileService(mockProfileRepo, mockUserRepo, mockFileRepo)
			err := service.UpdateLastSeen(context.Background(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
