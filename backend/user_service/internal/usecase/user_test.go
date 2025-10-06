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
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shared_models "quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/usecase/mocks"
)

type UserUseCaseSuite struct {
	suite.Suite
}

func TestUserUseCase(t *testing.T) {
	suite.RunSuite(t, new(UserUseCaseSuite))
}

func (s *UserUseCaseSuite) TestUserUseCase_CreateUser(t provider.T) {
	t.Epic("Unit")
	t.Feature("User UseCase")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование создания пользователя")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()
	username := "testuser"
	password := "validPass123"
	user := shared_models.User{
		Username: username,
		Password: password,
	}
	profile := shared_models.Profile{
		BasicInfo: &shared_models.BasicInfo{
			Name:    "Test",
			Surname: "User",
		},
	}

	t.Run("success", func(t provider.T) {

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
		mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

		uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

		mockUserRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil).AnyTimes()
		mockUserRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(userID, nil).AnyTimes()
		mockSessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil).AnyTimes()
		mockSessionRepo.EXPECT().SaveSession(ctx, userID, gomock.Any()).Return(nil).AnyTimes()
		mockProfileRepo.EXPECT().SaveProfile(ctx, gomock.Any()).Return(nil).AnyTimes()

		id, session, err := uc.CreateUser(ctx, user, profile)
		require.NoError(t, err)
		assert.Equal(t, userID, id)
		assert.NotEqual(t, uuid.Nil, session.SessionId)
	})

	t.Run("validation error - profile", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
		mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

		uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

		invalidProfile := shared_models.Profile{
			BasicInfo: &shared_models.BasicInfo{
				Name:    "", // empty name
				Surname: "User",
			},
		}

		_, _, err := uc.CreateUser(ctx, user, invalidProfile)
		assert.Error(t, err)
		assert.ErrorIs(t, err, user_errors.ErrProfileValidation)
	})

	t.Run("error saving user", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
		mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

		uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

		expectedErr := errors.New("save error")
		mockUserRepo.EXPECT().IsExists(ctx, username).Return(false, nil).AnyTimes()
		mockUserRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(uuid.Nil, expectedErr).AnyTimes()

		_, _, _ = uc.CreateUser(ctx, user, profile)
	})

	t.Run("error saving profile", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
		mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

		uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

		expectedErr := errors.New("profile save error")
		mockUserRepo.EXPECT().IsExists(ctx, username).Return(false, nil)
		mockUserRepo.EXPECT().SaveUser(ctx, gomock.Any()).Return(userID, nil)
		mockProfileRepo.EXPECT().SaveProfile(ctx, gomock.Any()).Return(expectedErr)

		_, _, _ = uc.CreateUser(ctx, user, profile)
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_AuthUser(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование аутентификации пользователя")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uuid.New()
	loginData := shared_models.LoginData{
		Username: "testuser",
		Password: "password",
	}
	user := shared_models.User{
		Id:       userID,
		Username: loginData.Username,
		Password: loginData.Password,
	}

	t.Run("success", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo.EXPECT().GetUser(ctx, loginData).Return(user, nil)
		mockSessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil)
		mockSessionRepo.EXPECT().SaveSession(ctx, userID, gomock.Any()).Return(nil)

		session, err := uc.AuthUser(ctx, loginData)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.SessionId)
	})

	t.Run("user not found", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().GetUser(ctx, loginData).Return(shared_models.User{}, expectedErr)

		_, err := uc.AuthUser(ctx, loginData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("error saving session", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("save error")
		mockUserRepo.EXPECT().GetUser(ctx, loginData).Return(user, nil)
		mockSessionRepo.EXPECT().IsExists(ctx, gomock.Any()).Return(false, nil).AnyTimes()
		mockSessionRepo.EXPECT().SaveSession(ctx, userID, gomock.Any()).Return(expectedErr)

		_, err := uc.AuthUser(ctx, loginData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_LookupUserSession(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование поиска пользователя по сессии")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uuid.New()
	session := shared_models.CreateSession()
	user := shared_models.User{
		Id: userID,
	}

	t.Run("success", func(t provider.T) {
		t.Epic("Unit")

		mockSessionRepo.EXPECT().LookupUserSession(ctx, session).Return(userID, nil)
		mockUserRepo.EXPECT().GetUserByUId(ctx, userID).Return(user, nil)

		result, err := uc.LookupUserSession(ctx, session)
		require.NoError(t, err)
		assert.Equal(t, userID, result.Id)
	})

	t.Run("session not found", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("session not found")
		mockSessionRepo.EXPECT().LookupUserSession(ctx, session).Return(uuid.Nil, expectedErr)

		_, err := uc.LookupUserSession(ctx, session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("user not found", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("user not found")
		mockSessionRepo.EXPECT().LookupUserSession(ctx, session).Return(userID, nil)
		mockUserRepo.EXPECT().GetUserByUId(ctx, userID).Return(shared_models.User{}, expectedErr)

		_, err := uc.LookupUserSession(ctx, session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_GetUserByUsername(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование получения пользователя по имени пользователя")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	username := "testuser"
	user := shared_models.User{
		Username: username,
	}

	t.Run("success", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)

		result, err := uc.GetUserByUsername(ctx, username)
		require.NoError(t, err)
		assert.Equal(t, username, result.Username)
	})

	t.Run("user not found", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().GetUserByUsername(ctx, username).Return(shared_models.User{}, expectedErr)

		_, err := uc.GetUserByUsername(ctx, username)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_GetUserById(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование получения пользователя по ID")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uuid.New()
	user := shared_models.User{
		Id: userID,
	}

	t.Run("success", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo.EXPECT().GetUserByUId(ctx, userID).Return(user, nil)

		result, err := uc.GetUserById(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, userID, result.Id)
	})

	t.Run("user not found", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().GetUserByUId(ctx, userID).Return(shared_models.User{}, expectedErr)

		_, err := uc.GetUserById(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_DeleteUserSession(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование удаления сессии пользователя")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	sessionID := "session123"

	t.Run("success", func(t provider.T) {

		t.Epic("Unit")
		mockSessionRepo.EXPECT().DeleteSession(ctx, sessionID).Return(nil)

		err := uc.DeleteUserSession(ctx, sessionID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("delete error")
		mockSessionRepo.EXPECT().DeleteSession(ctx, sessionID).Return(expectedErr)

		err := uc.DeleteUserSession(ctx, sessionID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func (s *UserUseCaseSuite) TestUserUseCase_SearchSimilarUser(t provider.T) {
	t.Epic("User Service")
	t.Feature("User UseCase")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование поиска похожих пользователей")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockProfileRepo := mocks.NewMockProfileRepository(ctrl)

	uc := NewUserUseCase(mockUserRepo, mockSessionRepo, mockProfileRepo)

	ctx := context.Background()
	searchTerm := "test"
	postsCount := uint(10)
	users := []shared_models.PublicUserInfo{
		{Username: "testuser1"},
		{Username: "testuser2"},
	}

	t.Run("success", func(t provider.T) {
		t.Epic("Unit")

		mockUserRepo.EXPECT().SearchSimilar(ctx, searchTerm, postsCount).Return(users, nil)

		result, err := uc.SearchSimilarUser(ctx, searchTerm, postsCount)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, users[0].Username, result[0].Username)
	})

	t.Run("error", func(t provider.T) {
		t.Epic("Unit")

		expectedErr := errors.New("search error")
		mockUserRepo.EXPECT().SearchSimilar(ctx, searchTerm, postsCount).Return(nil, expectedErr)

		_, err := uc.SearchSimilarUser(ctx, searchTerm, postsCount)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
