//go:build integration
// +build integration

package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"quickflow/config/test"
	"quickflow/post_service/internal/repository/postgres"
	"quickflow/post_service/utils/validation"
	"quickflow/shared/models"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"quickflow/post_service/internal/usecase"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type PostUseCaseTestSuite struct {
	suite.Suite
	db          *sql.DB
	postRepo    *postgres.PostgresPostRepository
	postUseCase *usecase.PostUseCase
	testUser1   uuid.UUID
	testUser2   uuid.UUID
	testPost1   uuid.UUID
	testPost2   uuid.UUID
}

func (s *PostUseCaseTestSuite) BeforeAll(t provider.T) {
	t.WithNewStep("Setup test environment", func(ctx provider.StepCtx) {
		// Setup PostgreSQL
		connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
		require.NotEmpty(t, connString, "Connection string must not be empty")

		ctx.WithNewAttachment("connection_string", allure.Text, []byte(connString))

		var err error
		s.db, err = sql.Open("pgx", connString)
		if err != nil {
			log.Fatalf("Failed to connect to test database: %v", err)
		}

		err = s.db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		// Generate test data
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testPost1 = uuid.New()
		s.testPost2 = uuid.New()

		// Initialize repositories and use case
		s.postRepo = postgres.NewPostgresPostRepository(s.db)

		// Mock file service for testing
		mockFileService := &MockFileService{}
		validator := validation.NewPostValidator()

		s.postUseCase = usecase.NewPostUseCase(
			s.postRepo,
			mockFileService,
			validator,
		)
	})
}

func (s *PostUseCaseTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostUseCaseTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupTestData()
		s.insertTestData()
	})
	t.Epic("Integration")
}
func (s *PostUseCaseTestSuite) insertTestData() error {
	// Insert test users
	_, err := s.db.Exec(`
		INSERT INTO "user" (id, username, psw_hash, salt)
		VALUES
			($1, 'user1', 'hash1', 'salt1'),
			($2, 'user2', 'hash2', 'salt2')
	`, s.testUser1, s.testUser2)
	if err != nil {
		return fmt.Errorf("insert users: %w", err)
	}

	// Insert contact info
	var contactInfoID1, contactInfoID2 int
	err = s.db.QueryRow(`
		INSERT INTO contact_info (city, phone_number, email)
		VALUES ('Moscow', '+79161234567', 'test1@example.com')
		RETURNING id
	`).Scan(&contactInfoID1)
	if err != nil {
		return fmt.Errorf("insert contact_info 1: %w", err)
	}

	err = s.db.QueryRow(`
		INSERT INTO contact_info (city, phone_number, email)
		VALUES ('St. Petersburg', '+79169876543', 'test2@example.com')
		RETURNING id
	`).Scan(&contactInfoID2)
	if err != nil {
		return fmt.Errorf("insert contact_info 2: %w", err)
	}

	// Insert school
	var schoolID int
	err = s.db.QueryRow(`
		INSERT INTO school (city, name)
		VALUES ('Moscow', 'Test School')
		RETURNING id
	`).Scan(&schoolID)
	if err != nil {
		return fmt.Errorf("insert school: %w", err)
	}

	// Insert profiles
	_, err = s.db.Exec(`
		INSERT INTO profile (id, bio, firstname, lastname, sex, birth_date, contact_info_id, school_id, last_seen)
		VALUES
			($1, 'Test bio 1', 'John', 'Doe', 1, '1990-01-01', $2, $3, NOW()),
			($4, 'Test bio 2', 'Jane', 'Smith', 2, '1992-02-02', $5, $3, NOW())
	`, s.testUser1, contactInfoID1, schoolID, s.testUser2, contactInfoID2)
	if err != nil {
		return fmt.Errorf("insert profiles: %w", err)
	}

	// Insert university
	var universityID int
	err = s.db.QueryRow(`
		INSERT INTO university (name, city)
		VALUES ('Test University', 'Moscow')
		RETURNING id
	`).Scan(&universityID)
	if err != nil {
		return fmt.Errorf("insert university: %w", err)
	}

	// Insert faculty
	var facultyID int
	err = s.db.QueryRow(`
		INSERT INTO faculty (university_id, name)
		VALUES ($1, 'Computer Science')
		RETURNING id
	`, universityID).Scan(&facultyID)
	if err != nil {
		return fmt.Errorf("insert faculty: %w", err)
	}

	// Insert education
	_, err = s.db.Exec(`
		INSERT INTO education (profile_id, faculty_id, graduation_year)
		VALUES
			($1, $2, 2023),
			($3, $2, 2024)
	`, s.testUser1, facultyID, s.testUser2)
	if err != nil {
		return fmt.Errorf("insert education: %w", err)
	}

	// Insert friendship
	_, err = s.db.Exec(`
		INSERT INTO friendship (user1_id, user2_id, status, is_read)
		VALUES ($1, $2, 'friends', true)
	`, s.testUser1, s.testUser2)
	if err != nil {
		return fmt.Errorf("insert friendship: %w", err)
	}

	return nil
}

func (s *PostUseCaseTestSuite) insertTestPost() error {
	// Insert test post
	_, err := s.db.Exec(`
		INSERT INTO post (id, creator_id, creator_type, text, created_at, updated_at)
		VALUES ($1, $2, 'user', 'Test post content', NOW(), NOW())
	`, s.testPost1, s.testUser1)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}

	// Insert another test post for user2
	_, err = s.db.Exec(`
		INSERT INTO post (id, creator_id, creator_type, text, created_at, updated_at)
		VALUES ($1, $2, 'user', 'Another test post', NOW(), NOW())
	`, s.testPost2, s.testUser2)
	if err != nil {
		return fmt.Errorf("insert post2: %w", err)
	}

	return nil
}

func (s *PostUseCaseTestSuite) insertTestPostWithLike() error {
	// Insert test post
	err := s.insertTestPost()
	if err != nil {
		return err
	}

	// Insert like for the post
	_, err = s.db.Exec(`
		INSERT INTO like_post (user_id, post_id)
		VALUES ($1, $2)
	`, s.testUser1, s.testPost1)
	if err != nil {
		return fmt.Errorf("insert like: %w", err)
	}

	return nil
}

func (s *PostUseCaseTestSuite) cleanupTestData() {
	// Clean up in reverse order to respect foreign key constraints
	tables := []string{
		"like_comment", "like_post", "comment_file", "comment",
		"post_file", "repost", "post", "education", "profile",
		"friendship", "user_follow", "community_user", "community",
		"chat_user", "message_file", "message", "chat", "feedback",
		"files", "sticker", "sticker_pack", "contact_info", "school",
		"faculty", "university", `"user"`,
	}

	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

func (s *PostUseCaseTestSuite) cleanupPostData() {
	// Clean up post-related data
	tables := []string{
		"like_comment", "like_post", "comment_file", "comment",
		"post_file", "repost", "post",
	}

	for _, table := range tables {
		s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}
func (s *PostUseCaseTestSuite) TestAddPost(t provider.T) {
	t.WithNewStep("Test AddPost successfully", func(ctx provider.StepCtx) {
		// Arrange
		post := models.Post{
			CreatorId:   s.testUser1,
			CreatorType: models.PostUser,
			Desc:        "Test post content", // Используем Text вместо Content
			Files:       []*models.File{},
		}

		// Act
		result, err := s.postUseCase.AddPost(context.Background(), post)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEqual(t, uuid.Nil, result.Id)
		require.Equal(t, post.Desc, result.Desc)
		require.Equal(t, post.CreatorId, result.CreatorId)
	})
}

func (s *PostUseCaseTestSuite) TestUpdatePost(t provider.T) {
	t.WithNewStep("Test UpdatePost successfully", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPost()
		require.NoError(t, err)

		update := models.PostUpdate{
			Id:    s.testPost1,
			Desc:  "Updated content", // Используем Text вместо Content
			Files: []*models.File{},
		}

		// Act
		result, err := s.postUseCase.UpdatePost(context.Background(), update, s.testUser1)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Updated content", result.Desc)
	})
}

func (s *PostUseCaseTestSuite) TestGetPostWithLike(t provider.T) {
	t.WithNewStep("Test GetPost returns post with like status", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPostWithLike()
		require.NoError(t, err)

		// Act
		post, err := s.postUseCase.GetPost(context.Background(), s.testPost1, s.testUser1)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, post)
		require.Equal(t, s.testPost1, post.Id)
		require.True(t, post.IsLiked) // Должен быть лайкнут
	})
}

func (s *PostUseCaseTestSuite) TestDeletePost(t provider.T) {
	t.WithNewStep("Test DeletePost successfully", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPost()
		require.NoError(t, err)

		// Act
		err = s.postUseCase.DeletePost(context.Background(), s.testUser1, s.testPost1)

		// Assert
		require.NoError(t, err)

		// Verify post is deleted
		_, err = s.postRepo.GetPost(context.Background(), s.testPost1)
		require.Error(t, err)
	})
}

func (s *PostUseCaseTestSuite) TestLikePost(t provider.T) {
	t.WithNewStep("Test LikePost successfully", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPost()
		require.NoError(t, err)

		// Act
		err = s.postUseCase.LikePost(context.Background(), s.testPost1, s.testUser1)

		// Assert
		require.NoError(t, err)

		// Verify like exists
		liked, err := s.postRepo.CheckIfPostLiked(context.Background(), s.testPost1, s.testUser1)
		require.NoError(t, err)
		require.True(t, liked)
	})
}

func (s *PostUseCaseTestSuite) TestUnlikePost(t provider.T) {
	t.WithNewStep("Test UnlikePost successfully", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPost()
		require.NoError(t, err)

		// First like the post
		err = s.postUseCase.LikePost(context.Background(), s.testPost1, s.testUser1)
		require.NoError(t, err)

		// Act
		err = s.postUseCase.UnlikePost(context.Background(), s.testPost1, s.testUser1)

		// Assert
		require.NoError(t, err)

		// Verify like is removed
		liked, err := s.postRepo.CheckIfPostLiked(context.Background(), s.testPost1, s.testUser1)
		require.NoError(t, err)
		require.False(t, liked)
	})
}

func (s *PostUseCaseTestSuite) TestGetPost(t provider.T) {
	t.WithNewStep("Test GetPost returns post with like status", func(ctx provider.StepCtx) {
		// Arrange
		err := s.insertTestPost()
		require.NoError(t, err)

		// Act
		post, err := s.postUseCase.GetPost(context.Background(), s.testPost1, s.testUser1)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, post)
		require.Equal(t, s.testPost1, post.Id)
		require.False(t, post.IsLiked) // Should not be liked initially
	})
}

func (s *PostUseCaseTestSuite) TestAddPostValidation(t provider.T) {
	t.WithNewStep("Test AddPost with invalid data", func(ctx provider.StepCtx) {
		// Arrange
		invalidPost := models.Post{
			CreatorId:   uuid.Nil, // Invalid user ID
			CreatorType: models.PostUser,
			Desc:        "",
		}

		// Act
		result, err := s.postUseCase.AddPost(context.Background(), invalidPost)

		// Assert
		require.Error(t, err)
		require.Nil(t, result)
	})
}

func (s *PostUseCaseTestSuite) TestFetchFeedInvalidParams(t provider.T) {
	t.WithNewStep("Test FetchFeed with invalid parameters", func(ctx provider.StepCtx) {
		// Act & Assert - Invalid numPosts
		_, err := s.postUseCase.FetchFeed(context.Background(), s.testUser1, -1, time.Now())
		require.Error(t, err)

		// Act & Assert - Invalid timestamp
		_, err = s.postUseCase.FetchFeed(context.Background(), s.testUser1, 10, time.Time{})
		require.Error(t, err)
	})
}

// MockFileService1 для тестирования
type MockFileService struct{}

func (m *MockFileService) UploadFile(ctx context.Context, file *models.File) (string, error) {
	return "mock_url", nil
}

func (m *MockFileService) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
	urls := make([]string, len(files))
	for i := range files {
		urls[i] = fmt.Sprintf("mock_url_%d", i)
	}
	return urls, nil
}

func (m *MockFileService) DeleteFile(ctx context.Context, filename string) error {
	return nil
}

func TestPostUseCaseInt(t *testing.T) {
	suite.RunSuite(t, new(PostUseCaseTestSuite))
}
