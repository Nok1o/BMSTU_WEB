//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"quickflow/config/test"
	"quickflow/shared/models"

	redis_config "quickflow/config/redis"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickflow/user_service/internal/repository/redis"
	getEnv "quickflow/utils/get-env"
)

type ProfileRepositoryIntegrationSuite struct {
	suite.Suite
	db         *sql.DB
	redisRepo  *redis.RedisSessionRepository
	repo       *PostgresProfileRepository
	testData   TestData
	cleanupIDs []uuid.UUID
}

type TestData struct {
	userID1    uuid.UUID
	userID2    uuid.UUID
	userID3    uuid.UUID
	profileID1 uuid.UUID
	profileID2 uuid.UUID
}

func TestProfileRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(ProfileRepositoryIntegrationSuite))
}

func (s *ProfileRepositoryIntegrationSuite) BeforeAll(t provider.T) {
	t.Epic("User Service")
	t.Feature("Profile Repository Integration")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов Profile Repository")

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
	s.repo = NewPostgresProfileRepository(s.db)
}

func (s *ProfileRepositoryIntegrationSuite) BeforeEach(t provider.T) {
	s.cleanupTestData(t)
	s.testData = s.setupTestData(t)
	t.Epic("Integration")
}

func (s *ProfileRepositoryIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	s.cleanupTestData(t)
	if s.db != nil {
		s.db.Close()
	}
}

func (s *ProfileRepositoryIntegrationSuite) setupTestData(t provider.T) TestData {
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	// Создание тестовых пользователей
	users := []struct {
		id       uuid.UUID
		username string
	}{
		{userID1, "testuser1"},
		{userID2, "testuser2"},
		{userID3, "testuser3"},
	}

	for _, user := range users {
		_, err := s.db.Exec(
			`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'hash', 'salt')`,
			user.id, user.username,
		)
		require.NoError(t, err, "Failed to create test user: %s", user.username)
		s.cleanupIDs = append(s.cleanupIDs, user.id)
	}

	// Создание тестовых профилей
	profiles := []struct {
		userID    uuid.UUID
		bio       string
		firstname string
		lastname  string
		sex       int
		birthDate time.Time
	}{
		{
			userID:    userID1,
			bio:       "Test bio 1",
			firstname: "John",
			lastname:  "Doe",
			sex:       0,
			birthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			userID:    userID2,
			bio:       "Test bio 2",
			firstname: "Jane",
			lastname:  "Smith",
			sex:       1,
			birthDate: time.Date(1995, 5, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, profile := range profiles {
		_, err := s.db.Exec(`
			INSERT INTO profile (id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date, last_seen) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, profile.userID, profile.bio, "", "", profile.firstname, profile.lastname, profile.sex, profile.birthDate, time.Now())
		require.NoError(t, err, "Failed to create test profile")
	}

	return TestData{
		userID1:    userID1,
		userID2:    userID2,
		userID3:    userID3,
		profileID1: userID1,
		profileID2: userID2,
	}
}

func (s *ProfileRepositoryIntegrationSuite) cleanupTestData(t provider.T) {
	if len(s.cleanupIDs) == 0 {
		return
	}

	// Удаление в правильном порядке для избежания foreign key violations
	_, err := s.db.Exec(`
		DELETE FROM education WHERE profile_id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up education:", err)
	}

	_, err = s.db.Exec(`
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

// ProfileBuilder для создания тестовых профилей
type ProfileBuilder struct {
	profile models.Profile
}

func NewProfileBuilder(userID uuid.UUID) *ProfileBuilder {
	return &ProfileBuilder{
		profile: models.Profile{
			UserId: userID,
			BasicInfo: &models.BasicInfo{
				Name:        "Test",
				Surname:     "User",
				Bio:         "Test bio",
				Sex:         0,
				DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func (b *ProfileBuilder) WithBasicInfo(name, surname, bio string, sex int, birthDate time.Time) *ProfileBuilder {
	b.profile.BasicInfo = &models.BasicInfo{
		Name:        name,
		Surname:     surname,
		Bio:         bio,
		Sex:         0,
		DateOfBirth: birthDate,
	}
	return b
}

func (b *ProfileBuilder) WithContactInfo(city, email, phone string) *ProfileBuilder {
	b.profile.ContactInfo = &models.ContactInfo{
		City:  city,
		Email: email,
		Phone: phone,
	}
	return b
}

func (b *ProfileBuilder) WithSchoolEducation(city, school string) *ProfileBuilder {
	b.profile.SchoolEducation = &models.SchoolEducation{
		City:   city,
		School: school,
	}
	return b
}

func (b *ProfileBuilder) WithUniversityEducation(university, city, faculty string, graduationYear int) *ProfileBuilder {
	b.profile.UniversityEducation = &models.UniversityEducation{
		University:     university,
		City:           city,
		Faculty:        faculty,
		GraduationYear: graduationYear,
	}
	return b
}

func (b *ProfileBuilder) WithUsername(username string) *ProfileBuilder {
	b.profile.Username = username
	return b
}

func (b *ProfileBuilder) Build() models.Profile {
	return b.profile
}

func (s *ProfileRepositoryIntegrationSuite) TestSaveProfile(t provider.T) {
	t.Tags("create", "profile", "integration")
	t.Description("Тестирование сохранения профиля")

	testCases := []struct {
		name          string
		profile       models.Profile
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное сохранение профиля с базовой информацией",
			profile: NewProfileBuilder(s.testData.userID3).
				WithBasicInfo("Alice", "Johnson", "New user bio", 1, time.Date(1992, 3, 20, 0, 0, 0, 0, time.UTC)).
				Build(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			err := s.repo.SaveProfile(ctx, tc.profile)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify that profile was saved correctly
				retrievedProfile, err := s.repo.GetProfile(ctx, tc.profile.UserId)
				assert.NoError(t, err)
				assert.Equal(t, tc.profile.BasicInfo.Name, retrievedProfile.BasicInfo.Name)
				assert.Equal(t, tc.profile.BasicInfo.Surname, retrievedProfile.BasicInfo.Surname)
				assert.Equal(t, tc.profile.BasicInfo.Bio, retrievedProfile.BasicInfo.Bio)
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestGetProfile(t provider.T) {
	t.Tags("read", "profile", "integration")
	t.Description("Тестирование получения профиля")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное получение существующего профиля",
			userID:      s.testData.userID1,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			profile, err := s.repo.GetProfile(ctx, tc.userID)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
				assert.Empty(t, profile.UserId)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.userID, profile.UserId)
				assert.Equal(t, "John", profile.BasicInfo.Name)
				assert.Equal(t, "Doe", profile.BasicInfo.Surname)
				assert.Equal(t, "Test bio 1", profile.BasicInfo.Bio)
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestUpdateProfileTextInfo(t provider.T) {
	t.Tags("update", "profile", "integration")
	t.Description("Тестирование обновления текстовой информации профиля")

	testCases := []struct {
		name          string
		profile       models.Profile
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное обновление базовой информации",
			profile: NewProfileBuilder(s.testData.userID1).
				WithBasicInfo("JohnUpdated", "DoeUpdated", "Updated bio", 0, time.Date(1991, 2, 2, 0, 0, 0, 0, time.UTC)).
				WithUsername("john_updated").
				Build(),
			expectError: false,
		},
		{
			name: "Успешное обновление с добавлением контактной информации",
			profile: NewProfileBuilder(s.testData.userID1).
				WithBasicInfo("John", "Doe", "Bio with contacts", 0, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)).
				WithContactInfo("Moscow", "john@example.com", "+79160000000").
				Build(),
			expectError: false,
		},
		{
			name: "Успешное обновление с добавлением образования",
			profile: NewProfileBuilder(s.testData.userID1).
				WithBasicInfo("John", "Doe", "Bio with education", 0, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)).
				WithSchoolEducation("SPb", "School 456").
				WithUniversityEducation("HSE", "Moscow", "Economics", 2012).
				Build(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			err := s.repo.UpdateProfileTextInfo(ctx, tc.profile)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify that profile was updated correctly
				retrievedProfile, err := s.repo.GetProfile(ctx, tc.profile.UserId)
				assert.NoError(t, err)

				if tc.profile.BasicInfo != nil {
					assert.Equal(t, tc.profile.BasicInfo.Name, retrievedProfile.BasicInfo.Name)
					assert.Equal(t, tc.profile.BasicInfo.Surname, retrievedProfile.BasicInfo.Surname)
					assert.Equal(t, tc.profile.BasicInfo.Bio, retrievedProfile.BasicInfo.Bio)
				}

				if tc.profile.Username != "" {
					// Verify username update
					publicInfo, err := s.repo.GetPublicUserInfo(ctx, tc.profile.UserId)
					assert.NoError(t, err)
					assert.Equal(t, tc.profile.Username, publicInfo.Username)
				}
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestUpdateProfileAvatar(t provider.T) {
	t.Tags("update", "avatar", "integration")
	t.Description("Тестирование обновления аватара профиля")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		avatarURL     string
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное обновление аватара",
			userID:      s.testData.userID1,
			avatarURL:   "https://example.com/avatar1.jpg",
			expectError: false,
		},
		{
			name:        "Успешное обновление аватара другим URL",
			userID:      s.testData.userID1,
			avatarURL:   "https://example.com/avatar2.png",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			err := s.repo.UpdateProfileAvatar(ctx, tc.userID, tc.avatarURL)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify that avatar was updated correctly
				profile, err := s.repo.GetProfile(ctx, tc.userID)
				assert.NoError(t, err)
				assert.Equal(t, tc.avatarURL, profile.BasicInfo.AvatarUrl)
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestUpdateProfileCover(t provider.T) {
	t.Tags("update", "cover", "integration")
	t.Description("Тестирование обновления обложки профиля")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		coverURL      string
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное обновление обложки",
			userID:      s.testData.userID1,
			coverURL:    "https://example.com/cover1.jpg",
			expectError: false,
		},
		{
			name:        "Успешное обновление обложки другим URL",
			userID:      s.testData.userID1,
			coverURL:    "https://example.com/cover2.png",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			err := s.repo.UpdateProfileCover(ctx, tc.userID, tc.coverURL)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify that cover was updated correctly
				profile, err := s.repo.GetProfile(ctx, tc.userID)
				assert.NoError(t, err)
				assert.Equal(t, tc.coverURL, profile.BasicInfo.BackgroundUrl)
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestGetPublicUserInfo(t provider.T) {
	t.Tags("read", "public", "integration")
	t.Description("Тестирование получения публичной информации о пользователе")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное получение публичной информации",
			userID:      s.testData.userID1,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			publicInfo, err := s.repo.GetPublicUserInfo(ctx, tc.userID)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
				assert.Empty(t, publicInfo.Id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.userID, publicInfo.Id)
				assert.Equal(t, "John", publicInfo.Firstname)
				assert.Equal(t, "Doe", publicInfo.Lastname)
				assert.Equal(t, "testuser1", publicInfo.Username)
			}
		})
	}
}

func (s *ProfileRepositoryIntegrationSuite) TestGetPublicUsersInfo(t provider.T) {
	t.Tags("read", "public", "batch", "integration")
	t.Description("Тестирование получения публичной информации о нескольких пользователях")

	testCases := []struct {
		name          string
		userIDs       []uuid.UUID
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Успешное получение информации о нескольких пользователях",
			userIDs:       []uuid.UUID{s.testData.userID1, s.testData.userID2},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "Успешное получение информации об одном пользователе",
			userIDs:       []uuid.UUID{s.testData.userID1},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Пустой результат для несуществующих пользователей",
			userIDs:       []uuid.UUID{uuid.New(), uuid.New()},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Пустой список пользователей",
			userIDs:       []uuid.UUID{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Смешанный список существующих и несуществующих пользователей",
			userIDs:       []uuid.UUID{s.testData.userID1, uuid.New(), s.testData.userID2},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			ctx := context.Background()

			// Act
			publicInfos, err := s.repo.GetPublicUsersInfo(ctx, tc.userIDs)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, publicInfos, tc.expectedCount)

				// Verify that returned users are from the requested list
				for _, info := range publicInfos {
					assert.Contains(t, tc.userIDs, info.Id)
				}
			}
		})
	}
}
