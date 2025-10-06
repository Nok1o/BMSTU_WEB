//go:build integration
// +build integration

package usecase_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"

	"quickflow/config/test"
	post_errors "quickflow/post_service/internal/errors"
	"quickflow/post_service/internal/repository/postgres"
	"quickflow/post_service/internal/usecase"
	shared_models "quickflow/shared/models"
	getEnv "quickflow/utils/get-env"
)

// MockFileService1 для тестирования
type MockFileService1 struct {
	uploadedFiles map[string]string
	deletedFiles  []string
}

func NewMockFileService() *MockFileService1 {
	return &MockFileService1{
		uploadedFiles: make(map[string]string),
		deletedFiles:  make([]string, 0),
	}
}

func (m *MockFileService1) UploadFile(ctx context.Context, file *shared_models.File) (string, error) {
	filename := fmt.Sprintf("https://storage.example.com/%s", file.Name)
	m.uploadedFiles[file.Name] = filename
	return filename, nil
}

func (m *MockFileService1) UploadManyFiles(ctx context.Context, files []*shared_models.File) ([]string, error) {
	res := make([]string, 0)
	for _, file := range files {
		link, err := m.UploadFile(ctx, file)
		if err != nil {
			return nil, err
		}
		res = append(res, link)
	}
	return res, nil
}

func (m *MockFileService1) DeleteFile(ctx context.Context, filename string) error {
	m.deletedFiles = append(m.deletedFiles, filename)
	return nil
}

// MockValidator для тестирования
type MockValidator struct{}

func (m *MockValidator) ValidateFeedParams(numPosts int, timestamp time.Time) error {
	if numPosts <= 0 || numPosts > 1000 {
		return fmt.Errorf("invalid number of posts")
	}
	if timestamp.After(time.Now()) {
		return fmt.Errorf("invalid timestamp")
	}
	return nil
}

type CommentUseCaseTestSuite struct {
	suite.Suite
	db             *sql.DB
	commentRepo    *postgres.PostgresCommentRepository
	fileService    *MockFileService1
	validator      *MockValidator
	commentUseCase *usecase.CommentUseCase
	testUser1      uuid.UUID
	testUser2      uuid.UUID
	testPost1      uuid.UUID
	testComment1   uuid.UUID
	testComment2   uuid.UUID
}

func (s *CommentUseCaseTestSuite) BeforeAll(t provider.T) {
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

		s.cleanupTestData()
		// Generate test IDs
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testPost1 = uuid.New()
		s.testComment1 = uuid.New()
		s.testComment2 = uuid.New()

		// Initialize repositories
		s.commentRepo = postgres.NewPostgresCommentRepository(s.db)
		s.fileService = NewMockFileService()
		s.validator = &MockValidator{}

		// Initialize comment use case
		s.commentUseCase = usecase.NewCommentUseCase(s.commentRepo, s.fileService, s.validator)
	})
}

func (s *CommentUseCaseTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *CommentUseCaseTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Setup test data for each test", func(ctx provider.StepCtx) {
		// Clean up first
		s.cleanupTestData()

		// Then insert fresh test data
		err := s.insertTestData()
		if err != nil {
			log.Fatalf("Failed to insert test data: %v", err)
		}

		// Reset mock file service
		s.fileService = NewMockFileService()
		s.commentUseCase = usecase.NewCommentUseCase(s.commentRepo, s.fileService, s.validator)
	})
	t.Epic("Integration")
}

func (s *CommentUseCaseTestSuite) insertTestData() error {
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

	// Insert test post
	_, err = s.db.Exec(`
		INSERT INTO post (id, creator_id, Text, created_at, updated_at, creator_type)
		VALUES ($1, $2, 'Test post Text', NOW(), NOW(), 'user')
	`, s.testPost1, s.testUser1)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}

	// Insert test comments
	_, err = s.db.Exec(`
		INSERT INTO comment (id, post_id, user_id, Text, created_at, updated_at)
		VALUES 
			($1, $2, $3, 'First comment', NOW(), NOW()),
			($4, $2, $5, 'Second comment', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour')
	`, s.testComment1, s.testPost1, s.testUser1, s.testComment2, s.testUser2)
	if err != nil {
		return fmt.Errorf("insert comments: %w", err)
	}

	return nil
}

func (s *CommentUseCaseTestSuite) cleanupTestData() {
	tables := []string{"like_comment", "comment", "post", `"user"`}

	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("Warning: failed to clean up table %s: %v", table, err)
		}
	}
}

func (s *CommentUseCaseTestSuite) TestAddComment(t provider.T) {
	t.WithNewStep("Test AddComment successfully", func(ctx provider.StepCtx) {
		// Arrange
		comment := shared_models.Comment{
			PostId: s.testPost1,
			UserId: s.testUser1,
			Text:   "New test comment",
		}

		// Act
		createdComment, err := s.commentUseCase.AddComment(context.Background(), comment)

		// Assert
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, createdComment.Id)
		require.Equal(t, s.testPost1, createdComment.PostId)
		require.Equal(t, s.testUser1, createdComment.UserId)
		require.Equal(t, "New test comment", createdComment.Text)
	})
}

func (s *CommentUseCaseTestSuite) TestUpdateComment_NotOwner(t provider.T) {
	t.WithNewStep("Test UpdateComment with non-owner user", func(ctx provider.StepCtx) {
		// Arrange
		commentUpdate := shared_models.CommentUpdate{
			Id:   s.testComment1,
			Text: "Updated comment Text",
		}

		// Act - user2 tries to update user1's comment
		_, err := s.commentUseCase.UpdateComment(context.Background(), commentUpdate, s.testUser2)

		// Assert
		require.Error(t, err)
		require.True(t, errors.Is(err, post_errors.ErrDoesNotBelongToUser))
	})
}

func (s *CommentUseCaseTestSuite) TestLikeComment(t provider.T) {
	t.WithNewStep("Test LikeComment successfully", func(ctx provider.StepCtx) {
		// Act
		err := s.commentUseCase.LikeComment(context.Background(), s.testComment1, s.testUser1)

		// Assert
		require.NoError(t, err)

		// Verify like was added
		liked, err := s.commentRepo.CheckIfCommentLiked(context.Background(), s.testComment1, s.testUser1)
		require.NoError(t, err)
		require.True(t, liked)
	})
}

func (s *CommentUseCaseTestSuite) TestLikeComment_AlreadyLiked(t provider.T) {
	t.WithNewStep("Test LikeComment when already liked (idempotency)", func(ctx provider.StepCtx) {
		// Arrange - like first
		err := s.commentUseCase.LikeComment(context.Background(), s.testComment1, s.testUser1)
		require.NoError(t, err)

		// Act - like again
		err = s.commentUseCase.LikeComment(context.Background(), s.testComment1, s.testUser1)

		// Assert - should be idempotent
		require.NoError(t, err)
	})
}

func (s *CommentUseCaseTestSuite) TestUnlikeComment(t provider.T) {
	t.WithNewStep("Test UnlikeComment successfully", func(ctx provider.StepCtx) {
		// Arrange - like first
		err := s.commentUseCase.LikeComment(context.Background(), s.testComment1, s.testUser1)
		require.NoError(t, err)

		// Act
		err = s.commentUseCase.UnlikeComment(context.Background(), s.testComment1, s.testUser1)

		// Assert
		require.NoError(t, err)

		// Verify like was removed
		liked, err := s.commentRepo.CheckIfCommentLiked(context.Background(), s.testComment1, s.testUser1)
		require.NoError(t, err)
		require.False(t, liked)
	})
}

func (s *CommentUseCaseTestSuite) TestUnlikeComment_NotLiked(t provider.T) {
	t.WithNewStep("Test UnlikeComment when not liked (idempotency)", func(ctx provider.StepCtx) {
		// Act
		err := s.commentUseCase.UnlikeComment(context.Background(), s.testComment1, s.testUser1)

		// Assert - should be idempotent
		require.NoError(t, err)
	})
}

func (s *CommentUseCaseTestSuite) TestGetComment(t provider.T) {
	t.WithNewStep("Test GetComment successfully", func(ctx provider.StepCtx) {
		// Act
		comment, err := s.commentUseCase.GetComment(context.Background(), s.testComment1, s.testUser1)

		// Assert
		require.NoError(t, err)
		require.Equal(t, s.testComment1, comment.Id)
		require.Equal(t, "First comment", comment.Text)
		require.Equal(t, s.testUser1, comment.UserId)
	})
}

// ErrorValidator для тестирования ошибок валидации
type ErrorValidator struct{}

func (e *ErrorValidator) ValidateFeedParams(numPosts int, timestamp time.Time) error {
	if numPosts <= 0 {
		return fmt.Errorf("invalid number of posts")
	}
	return nil
}

func TestCommentUseCase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(CommentUseCaseTestSuite))
}
