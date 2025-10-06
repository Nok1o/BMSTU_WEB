//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"quickflow/config/test"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	community_errors "quickflow/community_service/internal/errors"
	"quickflow/shared/models"
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
				Name:        "Test Community",
				Description: "Test Community Description",
			},
			NickName:    "test-community",
			CreatedAt:   time.Now(),
			Avatar:      &models.File{URL: "https://avatars2.githubusercontent.com/u/8640"},
			Cover:       &models.File{URL: "https://cover.jpg.com/cover.jpg"},
			ContactInfo: nil,
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

func (b *CommunityBuilder) WithContactInfo(contactInfo models.ContactInfo) *CommunityBuilder {
	b.community.ContactInfo = &contactInfo
	return b
}

func (b *CommunityBuilder) Build() models.Community {
	return b.community
}

type CommunityMemberBuilder struct {
	member models.CommunityMember
}

func NewCommunityMemberBuilder() *CommunityMemberBuilder {
	return &CommunityMemberBuilder{
		member: models.CommunityMember{
			UserID:      uuid.New(),
			CommunityID: uuid.New(),
			Role:        models.CommunityRoleMember,
			JoinedAt:    time.Now(),
		},
	}
}

func (b *CommunityMemberBuilder) WithUserID(userID uuid.UUID) *CommunityMemberBuilder {
	b.member.UserID = userID
	return b
}

func (b *CommunityMemberBuilder) WithCommunityID(communityID uuid.UUID) *CommunityMemberBuilder {
	b.member.CommunityID = communityID
	return b
}

func (b *CommunityMemberBuilder) WithRole(role models.CommunityRole) *CommunityMemberBuilder {
	b.member.Role = role
	return b
}

func (b *CommunityMemberBuilder) Build() models.CommunityMember {
	return b.member
}

type ContactInfoBuilder struct {
	contactInfo models.ContactInfo
}

func NewContactInfoBuilder() *ContactInfoBuilder {
	return &ContactInfoBuilder{
		contactInfo: models.ContactInfo{
			City:  "Test City",
			Email: "test@example.com",
			Phone: "+1234567890",
		},
	}
}

func (b *ContactInfoBuilder) WithCity(city string) *ContactInfoBuilder {
	b.contactInfo.City = city
	return b
}

func (b *ContactInfoBuilder) WithEmail(email string) *ContactInfoBuilder {
	b.contactInfo.Email = email
	return b
}

func (b *ContactInfoBuilder) WithPhone(phone string) *ContactInfoBuilder {
	b.contactInfo.Phone = phone
	return b
}

func (b *ContactInfoBuilder) Build() models.ContactInfo {
	return b.contactInfo
}

type CommunityRepositoryIntegrationSuite struct {
	suite.Suite
	db         *sql.DB
	repo       *SqlCommunityRepository
	testData   TestData
	cleanupIDs []uuid.UUID
}

type TestData struct {
	userID          uuid.UUID
	communityID     uuid.UUID
	adminUserID     uuid.UUID
	moderatorUserID uuid.UUID
}

func TestCommunityRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(CommunityRepositoryIntegrationSuite))
}

func (s *CommunityRepositoryIntegrationSuite) BeforeAll(t provider.T) {
	t.Feature("Community Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов репозитория сообществ")

	connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
	require.NotEmpty(t, connString, "Connection string must not be empty")

	var err error
	s.db, err = sql.Open("pgx", connString)
	require.NoError(t, err, "Failed to connect to test database")

	err = s.db.Ping()
	require.NoError(t, err, "Failed to ping database")

	s.repo = NewSqlCommunityRepository(s.db)
	s.cleanupIDs = make([]uuid.UUID, 0)
}

func (s *CommunityRepositoryIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	s.cleanupTestData(t)
	if s.db != nil {
		s.db.Close()
	}
}

func (s *CommunityRepositoryIntegrationSuite) BeforeEach(t provider.T) {
	s.cleanupTestData(t)
	s.testData = s.setupTestData(t)
	t.Epic("Integration")
}

func (s *CommunityRepositoryIntegrationSuite) setupTestData(t provider.T) TestData {
	// Создание тестовых пользователей
	userID := uuid.New()
	adminUserID := uuid.New()
	moderatorUserID := uuid.New()

	users := []struct {
		id       uuid.UUID
		username string
	}{
		{userID, "testuser"},
		{adminUserID, "adminuser"},
		{moderatorUserID, "moderatoruser"},
	}

	for _, user := range users {
		_, err := s.db.Exec(
			`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'hash', 'salt')`,
			user.id, user.username,
		)
		require.NoError(t, err, "Failed to create test user: %s", user.username)
		s.cleanupIDs = append(s.cleanupIDs, user.id)
	}

	// Создание тестового сообщества
	communityID := uuid.New()
	_, err := s.db.Exec(`
		INSERT INTO community (id, owner_id, name, description, created_at, avatar_url, cover_url, nickname) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, communityID, userID, "Test Community", "Test Description", time.Now(),
		"https://example.com/avatar.jpg", "https://example.com/cover.jpg", "test-community")
	require.NoError(t, err, "Failed to create test community")
	s.cleanupIDs = append(s.cleanupIDs, communityID)

	// Добавление пользователей в сообщество с разными ролями
	members := []struct {
		userID uuid.UUID
		role   models.CommunityRole
	}{
		{userID, models.CommunityRoleOwner},
		{adminUserID, models.CommunityRoleAdmin},
		{moderatorUserID, models.CommunityRoleMember},
	}

	for _, member := range members {
		_, err := s.db.Exec(`
			INSERT INTO community_user (user_id, community_id, role, joined_at)
			VALUES ($1, $2, $3, $4)
		`, member.userID, communityID, string(member.role), time.Now())
		require.NoError(t, err, "Failed to add user to community")
	}

	return TestData{
		userID:          userID,
		communityID:     communityID,
		adminUserID:     adminUserID,
		moderatorUserID: moderatorUserID,
	}
}

func (s *CommunityRepositoryIntegrationSuite) cleanupTestData(t provider.T) {
	if len(s.cleanupIDs) == 0 {
		return
	}

	// Удаление в правильном порядке для избежания foreign key violations
	_, err := s.db.Exec(`
		DELETE FROM community_user WHERE community_id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up community_user:", err)
	}

	_, err = s.db.Exec(`
		DELETE FROM community WHERE id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up community:", err)
	}

	_, err = s.db.Exec(`
		DELETE FROM "user" WHERE id = ANY($1)
	`, s.cleanupIDs)
	if err != nil {
		t.Error("Failed to clean up user:", err)
	}
}

func (s *CommunityRepositoryIntegrationSuite) TestCreateCommunity_Success(t provider.T) {
	t.Tags("create", "community", "success")
	t.Description("Тестирование успешного создания сообщества")

	// Arrange
	community := NewCommunityBuilder().
		WithOwnerID(s.testData.userID).
		WithName("New Test Community").
		WithNickname("new-test-community").
		Build()

	// Act
	err := s.repo.CreateCommunity(context.Background(), community)

	// Assert
	require.NoError(t, err)

	// Verify
	createdCommunity, err := s.repo.GetCommunityById(context.Background(), community.ID)
	require.NoError(t, err)
	assert.Equal(t, community.ID, createdCommunity.ID)
	assert.Equal(t, community.BasicInfo.Name, createdCommunity.BasicInfo.Name)
	assert.Equal(t, community.NickName, createdCommunity.NickName)

	// Check that owner is added as admin
	isMember, role, err := s.repo.IsCommunityMember(context.Background(), community.OwnerID, community.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, models.CommunityRoleOwner, *role)

	// Cleanup
	s.cleanupIDs = append(s.cleanupIDs, community.ID)
}

func (s *CommunityRepositoryIntegrationSuite) TestGetCommunityById_Success(t provider.T) {
	t.Tags("read", "community", "success")
	t.Description("Тестирование успешного получения сообщества по ID")

	// Act
	community, err := s.repo.GetCommunityById(context.Background(), s.testData.communityID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, s.testData.communityID, community.ID)
	assert.Equal(t, "Test Community", community.BasicInfo.Name)
	assert.Equal(t, "test-community", community.NickName)
}

func (s *CommunityRepositoryIntegrationSuite) TestGetCommunityById_NotFound(t provider.T) {
	t.Tags("read", "community", "error")
	t.Description("Тестирование получения несуществующего сообщества")

	// Arrange
	nonExistentID := uuid.New()

	// Act
	community, err := s.repo.GetCommunityById(context.Background(), nonExistentID)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, community_errors.ErrNotFound))
	assert.Equal(t, uuid.Nil, community.ID)
}

func (s *CommunityRepositoryIntegrationSuite) TestIsCommunityMember_Success(t provider.T) {
	t.Tags("membership", "success")
	t.Description("Тестирование проверки членства в сообществе")

	testCases := []struct {
		name         string
		userID       uuid.UUID
		expected     bool
		expectedRole models.CommunityRole
	}{
		{
			name:         "Owner should be member",
			userID:       s.testData.userID,
			expected:     true,
			expectedRole: models.CommunityRoleOwner,
		},
		{
			name:         "Admin should be member",
			userID:       s.testData.adminUserID,
			expected:     true,
			expectedRole: models.CommunityRoleAdmin,
		},
		{
			name:         "Non-member should not be member",
			userID:       uuid.New(),
			expected:     false,
			expectedRole: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")

			isMember, role, err := s.repo.IsCommunityMember(
				context.Background(),
				tc.userID,
				s.testData.communityID,
			)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expected, isMember)
			if tc.expected {
				assert.Equal(t, tc.expectedRole, *role)
			} else {
				assert.Nil(t, role)
			}
		})
	}
}

func (s *CommunityRepositoryIntegrationSuite) TestJoinAndLeaveCommunity_Success(t provider.T) {
	t.Tags("membership", "success")
	t.Description("Тестирование вступления и выхода из сообщества")

	// Arrange
	newUserID := uuid.New()
	_, err := s.db.Exec(
		`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, 'joinuser', 'hash', 'salt')`,
		newUserID,
	)
	require.NoError(t, err)
	defer s.db.Exec(`DELETE FROM "user" WHERE id = $1`, newUserID)

	member := NewCommunityMemberBuilder().
		WithUserID(newUserID).
		WithCommunityID(s.testData.communityID).
		WithRole(models.CommunityRoleMember).
		Build()

	// Act & Assert - Join
	err = s.repo.JoinCommunity(context.Background(), member)
	require.NoError(t, err)

	// Verify joined
	isMember, role, err := s.repo.IsCommunityMember(context.Background(), newUserID, s.testData.communityID)
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, models.CommunityRoleMember, *role)

	// Act & Assert - Leave
	err = s.repo.LeaveCommunity(context.Background(), newUserID, s.testData.communityID)
	require.NoError(t, err)

	// Verify left
	isMember, _, err = s.repo.IsCommunityMember(context.Background(), newUserID, s.testData.communityID)
	require.NoError(t, err)
	assert.False(t, isMember)
}

func (s *CommunityRepositoryIntegrationSuite) TestUpdateCommunityTextInfo_Success(t provider.T) {
	t.Tags("update", "community", "success")
	t.Description("Тестирование обновления текстовой информации сообщества")

	// Arrange
	newName := "Updated Community Name"
	newDescription := "Updated Community Description"
	newNickname := "updated-community"

	community := NewCommunityBuilder().
		WithID(s.testData.communityID).
		WithName(newName).
		WithNickname(newNickname).
		Build()
	community.BasicInfo.Description = newDescription

	// Act
	err := s.repo.UpdateCommunityTextInfo(context.Background(), community)

	// Assert
	require.NoError(t, err)

	// Verify
	updatedCommunity, err := s.repo.GetCommunityById(context.Background(), s.testData.communityID)
	require.NoError(t, err)
	assert.Equal(t, newName, updatedCommunity.BasicInfo.Name)
	assert.Equal(t, newDescription, updatedCommunity.BasicInfo.Description)
	assert.Equal(t, newNickname, updatedCommunity.NickName)
}

func (s *CommunityRepositoryIntegrationSuite) TestSearchSimilarCommunities_Success(t provider.T) {
	t.Tags("search", "community", "success")
	t.Description("Тестирование поиска похожих сообществ")

	// Arrange - Create additional communities for search
	similarCommunities := []models.Community{
		NewCommunityBuilder().
			WithOwnerID(s.testData.userID).
			WithName("Tech Community").
			WithNickname("tech-community").
			Build(),
		NewCommunityBuilder().
			WithOwnerID(s.testData.userID).
			WithName("Technology Enthusiasts").
			WithNickname("tech-enthusiasts").
			Build(),
	}

	for _, community := range similarCommunities {
		err := s.repo.CreateCommunity(context.Background(), community)
		require.NoError(t, err)
		s.cleanupIDs = append(s.cleanupIDs, community.ID)
	}

	// Act
	results, err := s.repo.SearchSimilarCommunities(context.Background(), "tech", 10)

	// Assert
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should find at least 2 similar communities")

	// Verify results contain expected communities
	foundNames := make(map[string]bool)
	for _, result := range results {
		foundNames[result.BasicInfo.Name] = true
	}

	assert.True(t, foundNames["Tech Community"], "Should find Tech Community")
	assert.True(t, foundNames["Technology Enthusiasts"], "Should find Technology Enthusiasts")
}

func (s *CommunityRepositoryIntegrationSuite) TestChangeUserRole_Success(t provider.T) {
	t.Tags("update", "role", "success")
	t.Description("Тестирование изменения роли пользователя в сообществе")

	// Arrange
	newRole := models.CommunityRoleAdmin

	// Act
	err := s.repo.ChangeUserRole(
		context.Background(),
		s.testData.moderatorUserID,
		s.testData.communityID,
		newRole,
	)

	// Assert
	require.NoError(t, err)

	// Verify
	isMember, role, err := s.repo.IsCommunityMember(
		context.Background(),
		s.testData.moderatorUserID,
		s.testData.communityID,
	)
	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, newRole, *role)
}

func (s *CommunityRepositoryIntegrationSuite) TestGetControlledCommunities_Success(t provider.T) {
	t.Tags("read", "controlled", "success")
	t.Description("Тестирование получения управляемых сообществ")

	// Act
	communities, err := s.repo.GetControlledCommunities(
		context.Background(),
		s.testData.userID, // owner should have controlled communities
		10,
		time.Now().Add(time.Hour),
	)

	// Assert
	require.NoError(t, err)
	assert.Greater(t, len(communities), 0, "Owner should have controlled communities")

	// Verify the community is in the results
	found := false
	for _, community := range communities {
		if community.ID == s.testData.communityID {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find the test community in controlled communities")
}
