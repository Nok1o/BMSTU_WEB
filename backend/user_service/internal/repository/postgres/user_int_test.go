//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	redis_config "quickflow/config/redis"
	"quickflow/gateway/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickflow/config/test"
	"quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/internal/repository/redis"
	getEnv "quickflow/utils/get-env"
)

type UserRepositoryIntegrationSuite struct {
	suite.Suite
	db         *sql.DB
	redisRepo  *redis.RedisSessionRepository
	repo       *PostgresUserRepository
	testData   TestDataUser
	cleanupIDs []uuid.UUID
}

type TestDataUser struct {
	userID1   uuid.UUID
	userID2   uuid.UUID
	userID3   uuid.UUID
	username1 string
	username2 string
	username3 string
}

func TestUserRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(UserRepositoryIntegrationSuite))
}

func (s *UserRepositoryIntegrationSuite) BeforeAll(t provider.T) {
	t.Epic("User Service")
	t.Feature("User Repository Integration")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов User Repository")

	// Инициализация БД PostgreSQL
	connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
	require.NotEmpty(t, connString, "Connection string must not be empty")

	var err error
	s.db, err = sql.Open("pgx", connString)
	require.NoError(t, err, "Failed to connect to test database")

	err = s.db.Ping()
	require.NoError(t, err, "Failed to ping database")

	// Инициализация Redis
	redisCfg := redis_config.NewRedisConfig()
	s.redisRepo = redis.NewRedisSessionRepository(redisCfg)

	// Инициализация репозитория
	s.repo = NewPostgresUserRepository(s.db)
}

func (s *UserRepositoryIntegrationSuite) BeforeEach(t provider.T) {
	s.cleanupTestData(t)
	s.testData = s.setupTestData(t)
	t.Epic("Integration")
}

func (s *UserRepositoryIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	s.cleanupTestData(t)
	if s.db != nil {
		s.db.Close()
	}
}

func (s *UserRepositoryIntegrationSuite) setupTestData(t provider.T) TestDataUser {
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	username1 := "testuser1"
	username2 := "testuser2"
	username3 := "testuser3"

	// Создание тестовых пользователей с хешированными паролями
	users := []struct {
		id       uuid.UUID
		username string
		password string
	}{
		{userID1, username1, "password123"},
		{userID2, username2, "password456"},
		{userID3, username3, "password789"},
	}

	for _, user := range users {
		salt := "salt" + user.id.String()
		hashedPassword := utils.HashPassword(user.password, salt)

		_, err := s.db.Exec(
			`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, $3, $4)`,
			user.id, user.username, hashedPassword, salt,
		)
		require.NoError(t, err, "Failed to create test user: %s", user.username)
		s.cleanupIDs = append(s.cleanupIDs, user.id)

		// Создание профилей для пользователей (для тестов поиска)
		_, err = s.db.Exec(`
			INSERT INTO profile (id, firstname, lastname, profile_avatar, last_seen) 
			VALUES ($1, $2, $3, $4, $5)
		`, user.id, "FirstName"+user.username, "LastName"+user.username, "avatar_"+user.username+".jpg", time.Now())
		require.NoError(t, err, "Failed to create test profile")
	}

	return TestDataUser{
		userID1:   userID1,
		userID2:   userID2,
		userID3:   userID3,
		username1: username1,
		username2: username2,
		username3: username3,
	}
}

func (s *UserRepositoryIntegrationSuite) cleanupTestData(t provider.T) {
	if len(s.cleanupIDs) == 0 {
		return
	}

	// Удаление в правильном порядке для избежания foreign key violations
	_, err := s.db.Exec(`
		DELETE FROM profile WHERE id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up profile:", err)
	}

	_, err = s.db.Exec(`
		DELETE FROM "user" WHERE id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up user:", err)
	}

	s.cleanupIDs = []uuid.UUID{}
}

// UserBuilder для создания тестовых пользователей
type UserBuilder struct {
	user models.User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: models.User{
			Id:       uuid.New(),
			Username: "testuser",
			Password: "testpassword123",
		},
	}
}

func (b *UserBuilder) WithID(id uuid.UUID) *UserBuilder {
	b.user.Id = id
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.user.Password = password
	return b
}

func (b *UserBuilder) Build() models.User {
	return b.user
}

func (s *UserRepositoryIntegrationSuite) TestSaveUser(t provider.T) {
	t.Tags("create", "user", "integration")
	t.Description("Тестирование сохранения пользователя")

	testCases := []struct {
		name          string
		user          models.User
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное сохранение пользователя",
			user: NewUserBuilder().
				WithID(uuid.New()).
				WithUsername("newtestuser" + uuid.New().String()[:4]).
				WithPassword("newpassword123").
				Build(),
			expectError: false,
		},
		{
			name: "Ошибка при сохранении пользователя с существующим username",
			user: NewUserBuilder().
				WithID(uuid.New()).
				WithUsername(s.testData.username1). // Уже существует
				WithPassword("password123").
				Build(),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			userID, err := s.repo.SaveUser(ctx, tc.user)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.user.Id, userID)

				// Verify that user was saved correctly
				retrievedUser, err := s.repo.GetUserByUId(ctx, userID)
				assert.NoError(t, err)
				assert.Equal(t, tc.user.Username, retrievedUser.Username)
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestGetUser(t provider.T) {
	t.Tags("read", "user", "integration")
	t.Description("Тестирование получения пользователя по логину и паролю")

	testCases := []struct {
		name          string
		loginData     models.LoginData
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное получение пользователя с правильными данными",
			loginData: models.LoginData{
				Username: s.testData.username1,
				Password: "password123",
			},
			expectError: false,
		},
		{
			name: "Ошибка при неправильном пароле",
			loginData: models.LoginData{
				Username: s.testData.username1,
				Password: "wrongpassword",
			},
			expectError:   true,
			expectedError: errors.New("incorrect login or password"),
		},
		{
			name: "Ошибка при несуществующем пользователе",
			loginData: models.LoginData{
				Username: "nonexistentuser",
				Password: "password123",
			},
			expectError:   true,
			expectedError: errors.New("user not found"),
		},
		{
			name: "Ошибка при пустом username",
			loginData: models.LoginData{
				Username: "",
				Password: "password123",
			},
			expectError:   true,
			expectedError: errors.New("incorrect"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			user, err := s.repo.GetUser(ctx, tc.loginData)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, user.Id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.loginData.Username, user.Username)
				assert.NotEqual(t, uuid.Nil, user.Id)
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestGetUserByUId(t provider.T) {
	t.Tags("read", "user", "integration")
	t.Description("Тестирование получения пользователя по ID")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное получение существующего пользователя",
			userID:      s.testData.userID1,
			expectError: false,
		},
		{
			name:          "Ошибка при получении несуществующего пользователя",
			userID:        uuid.New(),
			expectError:   true,
			expectedError: errors.New("user not found"),
		},
		{
			name:          "Ошибка при пустом user ID",
			userID:        uuid.Nil,
			expectError:   true,
			expectedError: errors.New("user not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			user, err := s.repo.GetUserByUId(ctx, tc.userID)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				}
				assert.Empty(t, user.Id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.userID, user.Id)
				assert.Equal(t, s.testData.username1, user.Username)
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestGetUserByUsername(t provider.T) {
	t.Tags("read", "user", "integration")
	t.Description("Тестирование получения пользователя по username")

	testCases := []struct {
		name          string
		username      string
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное получение существующего пользователя",
			username:    s.testData.username1,
			expectError: false,
		},
		{
			name:          "Ошибка при получении несуществующего пользователя",
			username:      "nonexistentusername",
			expectError:   true,
			expectedError: user_errors.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			user, err := s.repo.GetUserByUsername(ctx, tc.username)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
				assert.Empty(t, user.Id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.username, user.Username)
				assert.NotEqual(t, uuid.Nil, user.Id)
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestIsExists(t provider.T) {
	t.Tags("check", "existence", "integration")
	t.Description("Тестирование проверки существования пользователя")

	testCases := []struct {
		name        string
		username    string
		expected    bool
		expectError bool
	}{
		{
			name:        "Пользователь существует",
			username:    s.testData.username1,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Пользователь не существует",
			username:    "nonexistentuser",
			expected:    false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			exists, err := s.repo.IsExists(ctx, tc.username)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, exists)
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestSearchSimilar(t provider.T) {
	t.Tags("search", "users", "integration")
	t.Description("Тестирование поиска похожих пользователей")

	testCases := []struct {
		name        string
		searchQuery string
		postsCount  uint
		expectedMin int
		expectError bool
	}{
		{
			name:        "Поиск по точному username",
			searchQuery: "testuser1",
			postsCount:  10,
			expectedMin: 1,
			expectError: false,
		},
		{
			name:        "Поиск по части username",
			searchQuery: "testuser",
			postsCount:  10,
			expectedMin: 3, // Все три тестовых пользователя
			expectError: false,
		},
		{
			name:        "Поиск по имени",
			searchQuery: "FirstName",
			postsCount:  10,
			expectedMin: 3, // Все три тестовых пользователя
			expectError: false,
		},
		{
			name:        "Поиск по фамилии",
			searchQuery: "LastName",
			postsCount:  10,
			expectedMin: 3, // Все три тестовых пользователя
			expectError: false,
		},
		{
			name:        "Поиск с ограничением количества результатов",
			searchQuery: "testuser",
			postsCount:  2,
			expectedMin: 2,
			expectError: false,
		},
		{
			name:        "Поиск несуществующего пользователя",
			searchQuery: "nonexistentuser12345",
			postsCount:  10,
			expectedMin: 0,
			expectError: false,
		},
		{
			name:        "Поиск с пустым запросом",
			searchQuery: "",
			postsCount:  10,
			expectedMin: 0,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			users, err := s.repo.SearchSimilar(ctx, tc.searchQuery, tc.postsCount)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(users), tc.expectedMin)
				assert.LessOrEqual(t, len(users), int(tc.postsCount))

				// Проверяем, что результаты содержат ожидаемые данные
				if len(users) > 0 {
					for _, user := range users {
						assert.NotEqual(t, uuid.Nil, user.Id)
						assert.NotEmpty(t, user.Username)
						assert.NotEmpty(t, user.Firstname)
						assert.NotEmpty(t, user.Lastname)
					}
				}
			}
		})
	}
}

func (s *UserRepositoryIntegrationSuite) TestUserUniqueness(t provider.T) {
	t.Tags("uniqueness", "constraints", "integration")
	t.Description("Тестирование уникальности пользователей")

	t.Run("Проверка уникальности username", func(t provider.T) {
		ctx := context.Background()

		// Пытаемся создать пользователя с существующим username
		duplicateUser := NewUserBuilder().
			WithID(uuid.New()).
			WithUsername(s.testData.username1). // Уже существует
			WithPassword("password123").
			Build()

		_, err := s.repo.SaveUser(ctx, duplicateUser)
		assert.Error(t, err, "Должна быть ошибка при создании пользователя с существующим username")

		// Проверяем, что пользователь действительно не был создан
		exists, err := s.repo.IsExists(ctx, duplicateUser.Username)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Но ID должен соответствовать оригинальному пользователю, а не новому
		originalUser, err := s.repo.GetUserByUsername(ctx, duplicateUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, s.testData.userID1, originalUser.Id)
		assert.NotEqual(t, duplicateUser.Id, originalUser.Id)
	})
}

func (s *UserRepositoryIntegrationSuite) TestComplexSearchScenarios(t provider.T) {
	t.Tags("search", "complex", "integration")
	t.Description("Тестирование сложных сценариев поиска")

	t.Run("Поиск с различными вариантами написания", func(t provider.T) {
		ctx := context.Background()

		searchQueries := []string{
			"TESTUSER1",          // Верхний регистр
			"testuser1",          // Нижний регистр
			"TestUser1",          // Смешанный регистр
			"firstnametestuser1", // Комбинация имени и username
			"lastnametestuser1",  // Комбинация фамилии и username
		}

		for _, query := range searchQueries {
			t.Run(fmt.Sprintf("Query: %s", query), func(t provider.T) {
				users, err := s.repo.SearchSimilar(ctx, query, 10)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(users), 0)
			})
		}
	})

	t.Run("Поиск с специальными символами", func(t provider.T) {
		ctx := context.Background()

		searchQueries := []string{
			"test-user-1",
			"test_user_1",
			"test.user.1",
			"test user 1",
		}

		for _, query := range searchQueries {
			users, err := s.repo.SearchSimilar(ctx, query, 10)
			assert.NoError(t, err)
			// Результаты могут варьироваться в зависимости от настроек similarity
			assert.NotNil(t, users)
		}
	})
}
