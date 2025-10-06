//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"quickflow/config/test"
	"quickflow/post_service/internal/repository/postgres"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	post_errors "quickflow/post_service/internal/errors"
	"quickflow/shared/models"
)

type PostgresCommentRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.PostgresCommentRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
	testPost1  uuid.UUID
	testPost2  uuid.UUID
}

func (s *PostgresCommentRepositoryTestSuite) BeforeAll(t provider.T) {
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

		// Generate test users and posts
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testPost1 = uuid.New()
		s.testPost2 = uuid.New()

		s.repository = postgres.NewPostgresCommentRepository(s.db)
	})
}

func (s *PostgresCommentRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostgresCommentRepositoryTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		err := s.cleanupTestData()
		require.NoError(t, err)
		err = s.insertTestData()
		require.NoError(t, err)
	})
	t.Epic("Integration")
}

func (s *PostgresCommentRepositoryTestSuite) TestAddComment(t provider.T) {
	t.WithNewStep("Test add comment", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		comment := models.Comment{
			Id:        commentId,
			PostId:    s.testPost1,
			UserId:    s.testUser1,
			Text:      "Test comment",
			CreatedAt: now,
			UpdatedAt: now,
			LikeCount: 0,
		}

		err := s.repository.AddComment(context.Background(), comment)
		require.NoError(t, err)

		// Verify comment was created
		retrievedComment, err := s.repository.GetComment(context.Background(), commentId)
		require.NoError(t, err)
		require.Equal(t, commentId, retrievedComment.Id)
		require.Equal(t, s.testPost1, retrievedComment.PostId)
		require.Equal(t, s.testUser1, retrievedComment.UserId)
		require.Equal(t, "Test comment", retrievedComment.Text)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestAddComment_WithFiles(t provider.T) {
	t.WithNewStep("Test add comment with files", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		comment := models.Comment{
			Id:        commentId,
			PostId:    s.testPost1,
			UserId:    s.testUser1,
			Text:      "Test comment with files",
			CreatedAt: now,
			UpdatedAt: now,
			LikeCount: 0,
			Images: []*models.File{
				{
					URL:         "https://example.com/file1.jpg",
					Name:        "file1.jpg",
					DisplayType: "image",
				},
				{
					URL:         "https://example.com/file2.png",
					Name:        "file2.png",
					DisplayType: "image",
				},
			},
		}

		err := s.repository.AddComment(context.Background(), comment)
		require.NoError(t, err)

		// Verify comment was created with files
		retrievedComment, err := s.repository.GetComment(context.Background(), commentId)
		require.NoError(t, err)
		require.Equal(t, commentId, retrievedComment.Id)
		require.Len(t, retrievedComment.Images, 2)
		require.Equal(t, "https://example.com/file1.jpg", retrievedComment.Images[0].URL)
		require.Equal(t, "https://example.com/file2.png", retrievedComment.Images[1].URL)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetComment_Exists(t provider.T) {
	t.WithNewStep("Test get existing comment", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		// Create test comment
		err := s.createTestComment(commentId, s.testPost1, s.testUser1, "Test comment", now)
		require.NoError(t, err)

		comment, err := s.repository.GetComment(context.Background(), commentId)
		require.NoError(t, err)
		require.Equal(t, commentId, comment.Id)
		require.Equal(t, s.testPost1, comment.PostId)
		require.Equal(t, s.testUser1, comment.UserId)
		require.Equal(t, "Test comment", comment.Text)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetComment_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent comment", func(ctx provider.StepCtx) {
		nonExistentId := uuid.New()

		_, err := s.repository.GetComment(context.Background(), nonExistentId)
		require.Error(t, err)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetCommentsForPost_WithComments(t provider.T) {
	t.WithNewStep("Test get comments for post", func(ctx provider.StepCtx) {
		commentId1 := uuid.New()
		commentId2 := uuid.New()
		now := time.Now()

		// Create test comments
		err := s.createTestComment(commentId1, s.testPost1, s.testUser1, "Comment 1", now)
		require.NoError(t, err)
		err = s.createTestComment(commentId2, s.testPost1, s.testUser2, "Comment 2", now.Add(time.Minute))
		require.NoError(t, err)

		comments, err := s.repository.GetCommentsForPost(context.Background(), s.testPost1, 10, now.Add(-time.Hour))
		require.NoError(t, err)
		require.Len(t, comments, 2)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetCommentsForPost_NoComments(t provider.T) {
	t.WithNewStep("Test get comments for post with no comments", func(ctx provider.StepCtx) {
		comments, err := s.repository.GetCommentsForPost(context.Background(), s.testPost2, 10, time.Now())
		require.NoError(t, err)
		require.Empty(t, comments)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetCommentsForPost_WithTimestampFilter(t provider.T) {
	t.WithNewStep("Test get comments with timestamp filter", func(ctx provider.StepCtx) {
		commentId1 := uuid.New()
		commentId2 := uuid.New()
		now := time.Now()

		// Create test comments at different times
		err := s.createTestComment(commentId1, s.testPost1, s.testUser1, "Comment 1", now.Add(-time.Hour))
		require.NoError(t, err)
		err = s.createTestComment(commentId2, s.testPost1, s.testUser2, "Comment 2", now)
		require.NoError(t, err)

		// Get only comments after 30 minutes ago
		comments, err := s.repository.GetCommentsForPost(context.Background(), s.testPost1, 10, now.Add(-30*time.Minute))
		require.NoError(t, err)
		require.Len(t, comments, 1)
		require.Equal(t, commentId2, comments[0].Id)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestDeleteComment(t provider.T) {
	t.WithNewStep("Test delete comment", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		// Create test comment
		err := s.createTestComment(commentId, s.testPost1, s.testUser1, "Test comment", now)
		require.NoError(t, err)

		// Verify comment exists
		_, err = s.repository.GetComment(context.Background(), commentId)
		require.NoError(t, err)

		// Delete comment
		err = s.repository.DeleteComment(context.Background(), commentId)
		require.NoError(t, err)

		// Verify comment no longer exists
		_, err = s.repository.GetComment(context.Background(), commentId)
		require.Error(t, err)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestUpdateComment(t provider.T) {
	t.WithNewStep("Test update comment", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		// Create test comment
		err := s.createTestComment(commentId, s.testPost1, s.testUser1, "Original text", now)
		require.NoError(t, err)

		// Update comment
		update := models.CommentUpdate{
			Id:   commentId,
			Text: "Updated text",
			Files: []*models.File{
				{
					URL:         "https://example.com/updated.jpg",
					Name:        "updated.jpg",
					DisplayType: "image",
				},
			},
		}

		err = s.repository.UpdateComment(context.Background(), update)
		require.NoError(t, err)

		// Verify comment was updated
		updatedComment, err := s.repository.GetComment(context.Background(), commentId)
		require.NoError(t, err)
		require.Equal(t, "Updated text", updatedComment.Text)
		require.Len(t, updatedComment.Images, 1)
		require.Equal(t, "https://example.com/updated.jpg", updatedComment.Images[0].URL)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestLikeAndUnlikeComment(t provider.T) {
	t.WithNewStep("Test like and unlike comment", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		// Create test comment
		err := s.createTestComment(commentId, s.testPost1, s.testUser1, "Test comment", now)
		require.NoError(t, err)

		// Like comment
		err = s.repository.LikeComment(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)

		// Verify comment is liked
		isLiked, err := s.repository.CheckIfCommentLiked(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)
		require.True(t, isLiked)

		// Unlike comment
		err = s.repository.UnlikeComment(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)

		// Verify comment is no longer liked
		isLiked, err = s.repository.CheckIfCommentLiked(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)
		require.False(t, isLiked)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestCheckIfCommentLiked(t provider.T) {
	t.WithNewStep("Test check if comment is liked", func(ctx provider.StepCtx) {
		commentId := uuid.New()
		now := time.Now()

		// Create test comment
		err := s.createTestComment(commentId, s.testPost1, s.testUser1, "Test comment", now)
		require.NoError(t, err)

		// Check if not liked initially
		isLiked, err := s.repository.CheckIfCommentLiked(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)
		require.False(t, isLiked)

		// Like comment
		err = s.addCommentLike(commentId, s.testUser2)
		require.NoError(t, err)

		// Check if liked after like
		isLiked, err = s.repository.CheckIfCommentLiked(context.Background(), commentId, s.testUser2)
		require.NoError(t, err)
		require.True(t, isLiked)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetLastPostComment(t provider.T) {
	t.WithNewStep("Test get last post comment", func(ctx provider.StepCtx) {
		commentId1 := uuid.New()
		commentId2 := uuid.New()
		now := time.Now()

		// Create test comments with different like counts
		err := s.createTestComment(commentId1, s.testPost1, s.testUser1, "Comment with 5 likes", now)
		require.NoError(t, err)
		err = s.setCommentLikeCount(commentId1, 5)
		require.NoError(t, err)

		err = s.createTestComment(commentId2, s.testPost1, s.testUser2, "Comment with 10 likes", now.Add(time.Minute))
		require.NoError(t, err)
		err = s.setCommentLikeCount(commentId2, 10)
		require.NoError(t, err)

		// Get last comment (should be the one with most likes)
		lastComment, err := s.repository.GetLastPostComment(context.Background(), s.testPost1)
		require.NoError(t, err)
		require.Equal(t, commentId2, lastComment.Id)
		require.Equal(t, "Comment with 10 likes", lastComment.Text)
	})
}

func (s *PostgresCommentRepositoryTestSuite) TestGetLastPostComment_NoComments(t provider.T) {
	t.WithNewStep("Test get last post comment when no comments", func(ctx provider.StepCtx) {
		_, err := s.repository.GetLastPostComment(context.Background(), s.testPost2)
		require.Error(t, err)
		require.True(t, errors.Is(err, post_errors.ErrNotFound))
	})
}

// Helper methods
func (s *PostgresCommentRepositoryTestSuite) insertTestData() error {
	// Insert test users
	users := []struct {
		id       string
		username string
	}{
		{s.testUser1.String(), "testuser1"},
		{s.testUser2.String(), "testuser2"},
	}

	for _, user := range users {
		_, err := s.db.Exec(`
			INSERT INTO "user" (id, username, psw_hash, salt) 
			VALUES ($1, $2, 'hashedpassword', 'randomsalt')
		`, user.id, user.username)
		if err != nil {
			return err
		}

		_, err = s.db.Exec(`
			INSERT INTO profile (id, firstname, lastname) 
			VALUES ($1, $2, $3)
		`, user.id, "Test", "User")
		if err != nil {
			return err
		}
	}

	// Insert test posts
	posts := []struct {
		id string
	}{
		{s.testPost1.String()},
		{s.testPost2.String()},
	}

	for _, post := range posts {
		_, err := s.db.Exec(`
			INSERT INTO post (id, creator_id, text, created_at, updated_at, creator_type)
			VALUES ($1, $2, $3, NOW(), NOW(), $4)
		`, post.id, s.testUser1.String(), "Test post content", models.PostUser)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresCommentRepositoryTestSuite) createTestComment(commentId, postId, userId uuid.UUID, text string, createdAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO comment (id, post_id, user_id, text, created_at, updated_at, like_count)
		VALUES ($1, $2, $3, $4, $5, $5, 0)
	`, commentId, postId, userId, text, createdAt)
	return err
}

func (s *PostgresCommentRepositoryTestSuite) addCommentFile(commentId uuid.UUID, fileURL, fileType string) error {
	// First ensure file exists in files table
	_, err := s.db.Exec(`
		INSERT INTO files (file_url, filename, user_id, uploaded_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (file_url) DO NOTHING
	`, fileURL, "testfile", s.testUser1)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO comment_file (comment_id, file_url, file_type)
		VALUES ($1, $2, $3)
	`, commentId, fileURL, fileType)
	return err
}

func (s *PostgresCommentRepositoryTestSuite) addCommentLike(commentId, userId uuid.UUID) error {
	_, err := s.db.Exec(`
		INSERT INTO like_comment (user_id, comment_id)
		VALUES ($1, $2)
	`, userId, commentId)
	return err
}

func (s *PostgresCommentRepositoryTestSuite) setCommentLikeCount(commentId uuid.UUID, count int) error {
	_, err := s.db.Exec(`
		UPDATE comment SET like_count = $1 WHERE id = $2
	`, count, commentId)
	return err
}

func (s *PostgresCommentRepositoryTestSuite) cleanupTestData() error {
	_, err := s.db.Exec(`TRUNCATE TABLE 
    like_comment, comment_file, comment, post, profile, "user", files 
	RESTART IDENTITY CASCADE
    `)
	if err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}

	return nil
}

func TestPostgresCommentRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresCommentRepositoryTestSuite))
}
