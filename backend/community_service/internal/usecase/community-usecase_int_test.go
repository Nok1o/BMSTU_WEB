//go:build integration
// +build integration

package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"quickflow/community_service/config"
	"quickflow/community_service/internal/repository/postgres"
	"quickflow/community_service/utils/validation"
	addr "quickflow/config/micro-addr"
	"quickflow/config/test"
	fileService "quickflow/shared/client/file_service"
	"quickflow/shared/interceptors"
	"quickflow/shared/models"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	community_errors "quickflow/community_service/internal/errors"
)

type CommunityBuilder struct {
	community models.Community
}

func NewCommunityBuilder() *CommunityBuilder {
	return &CommunityBuilder{
		community: models.Community{
			ID:      uuid.New(),
			OwnerID: uuid.New(),
			BasicInfo: &models.BasicCommunityInfo{
				Name:        generateRandomName("Community"),
				Description: generateRandomDescription(),
			},
			NickName:  generateRandomNickname(),
			CreatedAt: time.Now(),
		},
	}
}

func (b *CommunityBuilder) WithID(id uuid.UUID) *CommunityBuilder {
	b.community.ID = id
	return b
}

func (b *CommunityBuilder) WithOwnerID(ownerID uuid.UUID) *CommunityBuilder {
	b.community.OwnerID = ownerID
	return b
}

func (b *CommunityBuilder) WithName(name string) *CommunityBuilder {
	b.community.BasicInfo.Name = name
	return b
}

func (b *CommunityBuilder) WithNickname(nickname string) *CommunityBuilder {
	b.community.NickName = nickname
	return b
}

func (b *CommunityBuilder) WithDescription(description string) *CommunityBuilder {
	b.community.BasicInfo.Description = description
	return b
}

func (b *CommunityBuilder) WithAvatar(file *models.File) *CommunityBuilder {
	b.community.Avatar = file
	return b
}

func (b *CommunityBuilder) WithCover(file *models.File) *CommunityBuilder {
	b.community.Cover = file
	return b
}

func (b *CommunityBuilder) Build() models.Community {
	return b.community
}

type FileBuilder struct {
	file models.File
}

func NewFileBuilder() *FileBuilder {
	return &FileBuilder{
		file: models.File{
			Name:     generateRandomFilename(),
			Reader:   bytes.NewReader(generateRandomData()),
			Size:     int64(rand.Intn(1000) + 100),
			MimeType: "image/jpeg",
		},
	}
}

func (b *FileBuilder) WithFilename(filename string) *FileBuilder {
	b.file.Name = filename
	return b
}

func (b *FileBuilder) WithData(data []byte) *FileBuilder {
	b.file.Reader = bytes.NewReader(data)
	return b
}

func (b *FileBuilder) Build() *models.File {
	return &b.file
}

// Генерация случайных данных
func generateRandomName(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), rand.Intn(10000))
}

func generateRandomDescription() string {
	descriptions := []string{
		"Test community description for integration testing",
		"Random community for parallel test execution",
		"Temporary community created during integration tests",
		"Community with randomized data for isolation",
	}
	return descriptions[rand.Intn(len(descriptions))]
}

func generateRandomNickname() string {
	return fmt.Sprintf("test-community-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
}

func generateRandomUsername() string {
	return fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(10000))
}

func generateRandomFilename() string {
	extensions := []string{"jpg", "png", "gif", "jpeg"}
	return fmt.Sprintf("test-%d.%s", time.Now().UnixNano(), extensions[rand.Intn(len(extensions))])
}

func generateRandomData() []byte {
	size := rand.Intn(100) + 50
	data := make([]byte, size)
	rand.Read(data)
	return data
}

type CommunityUseCaseIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	useCase     *CommunityUseCase
	fileService FileService
	validator   *validation.CommunityValidator
	testData    TestData
	cleanupData CleanupData
}

type TestData struct {
	userID       uuid.UUID
	adminUserID  uuid.UUID
	memberUserID uuid.UUID
	communityID  uuid.UUID
	communityID2 uuid.UUID
	sessionID    string // Уникальный идентификатор сессии теста
}

type CleanupData struct {
	userIDs      []uuid.UUID
	communityIDs []uuid.UUID
}

func TestCommunityUseCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Инициализация генератора случайных чисел для каждого теста
	rand.Seed(time.Now().UnixNano())
	suite.RunSuite(t, new(CommunityUseCaseIntegrationSuite))
}

func (s *CommunityUseCaseIntegrationSuite) BeforeAll(t provider.T) {
	t.Epic("Community Service")
	t.Feature("Community UseCase Integration")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов UseCase")

	// Инициализация БД
	connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
	require.NotEmpty(t, connString, "Connection string must not be empty")

	var err error
	s.db, err = sql.Open("pgx", connString)
	require.NoError(t, err, "Failed to connect to test database")

	err = s.db.Ping()
	require.NoError(t, err, "Failed to ping database")

	grpcConnFileService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFileServiceAddrEnv, addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnFileService.Close()
	s.fileService = fileService.NewFileClient(grpcConnFileService)

	// Инициализация валидатора
	cfg := config.CommunityConfig{
		CommunityNameMaxLength:        100,
		CommunityNameMinLength:        3,
		CommunityDescriptionMaxLength: 500,
		CommunityAvatarMaxSize:        10 * 1024 * 1024, // 10MB
	}
	s.validator = validation.NewCommunityValidator(cfg)

	// Инициализация репозитория
	repo := postgres.NewSqlCommunityRepository(s.db)

	// Инициализация use case
	s.useCase = NewCommunityUseCase(repo, s.fileService, s.validator)
}

func (s *CommunityUseCaseIntegrationSuite) BeforeEach(t provider.T) {
	// Очистка данных предыдущего теста
	s.cleanupTestData(t)

	// Инициализация новых случайных данных для текущего теста
	s.testData = s.setupTestData(t)
	s.cleanupData = CleanupData{}
	t.Epic("Integration")
}

func (s *CommunityUseCaseIntegrationSuite) AfterEach(t provider.T) {
	// Очистка только данных, созданных в текущем тесте
	s.cleanupTestData(t)
}

func (s *CommunityUseCaseIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	if s.db != nil {
		s.db.Close()
	}
}

func (s *CommunityUseCaseIntegrationSuite) setupTestData(t provider.T) TestData {
	// Генерация уникального идентификатора сессии для изоляции тестов
	sessionID := fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

	// Создание тестовых пользователей со случайными данными
	userID := uuid.New()
	adminUserID := uuid.New()
	memberUserID := uuid.New()

	users := []struct {
		id       uuid.UUID
		username string
	}{
		{userID, generateRandomUsername()},
		{adminUserID, generateRandomUsername()},
		{memberUserID, generateRandomUsername()},
	}

	for _, user := range users {
		_, err := s.db.Exec(
			`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'hash', 'salt')`,
			user.id, user.username,
		)
		require.NoError(t, err, "Failed to create test user: %s", user.username)
		s.cleanupData.userIDs = append(s.cleanupData.userIDs, user.id)
	}

	// Создание тестовых сообществ со случайными данными
	communityID := uuid.New()
	communityID2 := uuid.New()

	communities := []struct {
		id          uuid.UUID
		ownerID     uuid.UUID
		name        string
		description string
		nickname    string
	}{
		{
			communityID,
			userID,
			generateRandomName("Community"),
			generateRandomDescription(),
			generateRandomNickname(),
		},
		{
			communityID2,
			userID,
			generateRandomName("Another"),
			generateRandomDescription(),
			generateRandomNickname(),
		},
	}

	for _, community := range communities {
		_, err := s.db.Exec(`
			INSERT INTO community (id, owner_id, name, description, created_at, nickname) 
			VALUES ($1, $2, $3, $4, $5, $6)
		`, community.id, community.ownerID, community.name, community.description, time.Now(), community.nickname)
		require.NoError(t, err, "Failed to create test community")
		s.cleanupData.communityIDs = append(s.cleanupData.communityIDs, community.id)
	}

	// Добавление пользователей в сообщества
	members := []struct {
		userID      uuid.UUID
		communityID uuid.UUID
		role        models.CommunityRole
	}{
		{userID, communityID, models.CommunityRoleOwner},
		{adminUserID, communityID, models.CommunityRoleAdmin},
		{memberUserID, communityID, models.CommunityRoleMember},
		{userID, communityID2, models.CommunityRoleOwner},
	}

	for _, member := range members {
		_, err := s.db.Exec(`
			INSERT INTO community_user (user_id, community_id, role, joined_at)
			VALUES ($1, $2, $3, $4)
		`, member.userID, member.communityID, string(member.role), time.Now())
		require.NoError(t, err, "Failed to add user to community")
	}

	return TestData{
		userID:       userID,
		adminUserID:  adminUserID,
		memberUserID: memberUserID,
		communityID:  communityID,
		communityID2: communityID2,
		sessionID:    sessionID,
	}
}

func (s *CommunityUseCaseIntegrationSuite) cleanupTestData(t provider.T) {
	if len(s.cleanupData.communityIDs) == 0 && len(s.cleanupData.userIDs) == 0 {
		return
	}

	// Удаление в правильном порядке для избежания foreign key violations
	if len(s.cleanupData.communityIDs) > 0 {
		// Удаляем связи пользователей с сообществами
		_, err := s.db.Exec(`
			DELETE FROM community_user WHERE community_id = ANY($1)
		`, s.cleanupData.communityIDs)
		if err != nil {
			t.Logf("Failed to clean up community_user: %v", err)
		}

		// Удаляем сообщества
		_, err = s.db.Exec(`
			DELETE FROM community WHERE id = ANY($1)
		`, s.cleanupData.communityIDs)
		if err != nil {
			t.Logf("Failed to clean up community: %v", err)
		}
	}

	if len(s.cleanupData.userIDs) > 0 {
		// Удаляем пользователей (если они не используются в других тестах)
		_, err := s.db.Exec(`
			DELETE FROM "user" WHERE id = ANY($1)
		`, s.cleanupData.userIDs)
		if err != nil {
			t.Logf("Failed to clean up user: %v", err)
		}
	}

	// Очищаем списки ID после удаления
	s.cleanupData = CleanupData{}
}

func (s *CommunityUseCaseIntegrationSuite) TestCreateCommunity(t provider.T) {
	t.Tags("create", "community", "integration")
	t.Description("Тестирование создания сообщества")

	testCases := []struct {
		name          string
		community     models.Community
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное создание сообщества",
			community: NewCommunityBuilder().
				WithOwnerID(s.testData.userID).
				WithName(generateRandomName("NewCommunity")).
				WithNickname(generateRandomNickname()).
				Build(),
			expectError: false,
		},
		{
			name: "Ошибка при создании сообщества с пустым именем",
			community: func() models.Community {
				comm := NewCommunityBuilder().
					WithOwnerID(s.testData.userID).
					WithNickname(generateRandomNickname()).
					Build()
				comm.BasicInfo.Name = "" // Пустое имя
				return comm
			}(),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			result, err := s.useCase.CreateCommunity(ctx, tc.community)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.community.BasicInfo.Name, result.BasicInfo.Name)

				// Добавляем созданное сообщество в список для очистки
				if result != nil && result.ID != uuid.Nil {
					s.cleanupData.communityIDs = append(s.cleanupData.communityIDs, result.ID)
				}
			}
		})
	}
}

func (s *CommunityUseCaseIntegrationSuite) TestGetCommunityMembers(t provider.T) {
	t.Tags("read", "members", "integration")
	t.Description("Тестирование получения участников сообщества")

	testCases := []struct {
		name          string
		communityID   uuid.UUID
		numMembers    int
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное получение участников",
			communityID: s.testData.communityID,
			numMembers:  10,
			expectError: false,
		},
		{
			name:          "Ошибка при пустом ID сообщества",
			communityID:   uuid.Nil,
			numMembers:    10,
			expectError:   true,
			expectedError: fmt.Errorf("community ID cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			members, err := s.useCase.GetCommunityMembers(ctx, tc.communityID, tc.numMembers, time.Now())

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, members)

				if tc.communityID == s.testData.communityID {
					assert.Len(t, members, 3) // 3 участника в тестовых данных
				} else {
					assert.Len(t, members, 0) // Пустой список для несуществующего сообщества
				}
			}
		})
	}
}

func (s *CommunityUseCaseIntegrationSuite) TestGetUserCommunities(t provider.T) {
	t.Tags("read", "user", "integration")
	t.Description("Тестирование получения сообществ пользователя")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		count         int
		expectedCount int
		expectError   bool
		expectedError error
	}{
		{
			name:          "Успешное получение сообществ пользователя",
			userID:        s.testData.userID,
			count:         10,
			expectedCount: 2, // Пользователь состоит в 2 сообществах
			expectError:   false,
		},
		{
			name:          "Ошибка при пустом user ID",
			userID:        uuid.Nil,
			count:         10,
			expectedCount: 0,
			expectError:   true,
			expectedError: fmt.Errorf("user ID cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			communities, err := s.useCase.GetUserCommunities(ctx, tc.userID, tc.count, time.Now())

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, communities)
				assert.Len(t, communities, tc.expectedCount)
			}
		})
	}
}

func (s *CommunityUseCaseIntegrationSuite) TestChangeUserRole(t provider.T) {
	t.Tags("update", "role", "integration")
	t.Description("Тестирование изменения роли пользователя")

	testCases := []struct {
		name          string
		targetUserID  uuid.UUID
		requesterID   uuid.UUID
		newRole       models.CommunityRole
		expectError   bool
		expectedError error
	}{
		{
			name:         "Владелец может изменить роль на админа",
			targetUserID: s.testData.memberUserID,
			requesterID:  s.testData.userID, // Владелец
			newRole:      models.CommunityRoleAdmin,
			expectError:  false,
		},
		{
			name:         "Админ может изменить роль на модератора",
			targetUserID: s.testData.memberUserID,
			requesterID:  s.testData.adminUserID, // Админ
			newRole:      models.CommunityRoleMember,
			expectError:  false,
		},
		{
			name:          "Обычный участник не может изменять роли",
			targetUserID:  s.testData.adminUserID,
			requesterID:   s.testData.memberUserID, // Обычный участник
			newRole:       models.CommunityRoleMember,
			expectError:   true,
			expectedError: community_errors.ErrForbidden,
		},
		{
			name:          "Не участник не может изменять роли",
			targetUserID:  s.testData.memberUserID,
			requesterID:   uuid.New(), // Не участник
			newRole:       models.CommunityRoleAdmin,
			expectError:   true,
			expectedError: community_errors.ErrNotParticipant,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			err := s.useCase.ChangeUserRole(ctx, tc.targetUserID, s.testData.communityID, tc.newRole, tc.requesterID)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify
				isMember, role, err := s.useCase.IsCommunityMember(ctx, tc.targetUserID, s.testData.communityID)
				assert.NoError(t, err)
				assert.True(t, isMember)
				assert.Equal(t, tc.newRole, *role)
			}
		})
	}
}

func (s *CommunityUseCaseIntegrationSuite) TestGetControlledCommunities(t provider.T) {
	t.Tags("read", "controlled", "integration")
	t.Description("Тестирование получения управляемых сообществ")

	testCases := []struct {
		name          string
		userID        uuid.UUID
		count         int
		expectedCount int
		expectError   bool
		expectedError error
	}{
		{
			name:          "Успешное получение управляемых сообществ владельца",
			userID:        s.testData.userID,
			count:         10,
			expectedCount: 2, // Владелец управляет 2 сообществами
			expectError:   false,
		},
		{
			name:          "Успешное получение управляемых сообществ админа",
			userID:        s.testData.adminUserID,
			count:         10,
			expectedCount: 1, // Админ управляет 1 сообществом
			expectError:   false,
		},
		{
			name:          "Ошибка при пустом user ID",
			userID:        uuid.Nil,
			count:         10,
			expectedCount: 0,
			expectError:   true,
			expectedError: fmt.Errorf("user ID cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			communities, err := s.useCase.GetControlledCommunities(ctx, tc.userID, tc.count, time.Now())

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, communities)
				assert.Len(t, communities, tc.expectedCount)
			}
		})
	}
}

// Вспомогательная функция для создания изолированных тестовых данных в отдельных тестах
func (s *CommunityUseCaseIntegrationSuite) createIsolatedTestCommunity(t provider.T, ownerID uuid.UUID) uuid.UUID {
	communityID := uuid.New()
	communityName := generateRandomName("IsolatedCommunity")
	communityNickname := generateRandomNickname()

	_, err := s.db.Exec(`
		INSERT INTO community (id, owner_id, name, description, created_at, nickname) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, communityID, ownerID, communityName, generateRandomDescription(), time.Now(), communityNickname)
	require.NoError(t, err, "Failed to create isolated test community")

	// Добавляем в список для очистки
	s.cleanupData.communityIDs = append(s.cleanupData.communityIDs, communityID)

	return communityID
}

// Вспомогательная функция для создания изолированного тестового пользователя
func (s *CommunityUseCaseIntegrationSuite) createIsolatedTestUser(t provider.T) uuid.UUID {
	userID := uuid.New()
	username := generateRandomUsername()

	_, err := s.db.Exec(
		`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'hash', 'salt')`,
		userID, username,
	)
	require.NoError(t, err, "Failed to create isolated test user")

	// Добавляем в список для очистки
	s.cleanupData.userIDs = append(s.cleanupData.userIDs, userID)

	return userID
}
