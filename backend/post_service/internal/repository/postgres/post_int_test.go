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

type PostgresPostRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.PostgresPostRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
	testUser3  uuid.UUID
	testPost1  uuid.UUID
	testPost2  uuid.UUID
	testPost3  uuid.UUID
}

func (s *PostgresPostRepositoryTestSuite) BeforeAll(t provider.T) {
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
		s.testUser3 = uuid.New()
		s.testPost1 = uuid.New()
		s.testPost2 = uuid.New()
		s.testPost3 = uuid.New()

		// Insert test data
		err = s.insertTestData()
		if err != nil {
			log.Fatalf("Failed to insert test data: %v", err)
		}

		s.repository = postgres.NewPostgresPostRepository(s.db)
	})
}

func (s *PostgresPostRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostgresPostRepositoryTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupTestData()
		s.insertTestData()
	})
	t.Epic("Integration")
}

func (s *PostgresPostRepositoryTestSuite) TestAddPost(t provider.T) {
	t.WithNewStep("Test add post", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		post := models.Post{
			Id:           postId,
			CreatorId:    s.testUser1,
			CreatorType:  models.PostUser,
			Desc:         "Test post content",
			CreatedAt:    now,
			UpdatedAt:    now,
			LikeCount:    0,
			RepostCount:  0,
			CommentCount: 0,
			IsRepost:     false,
		}

		err := s.repository.AddPost(context.Background(), post)
		require.NoError(t, err)

		// Verify post was created
		retrievedPost, err := s.repository.GetPost(context.Background(), postId)
		require.NoError(t, err)
		require.Equal(t, postId, retrievedPost.Id)
		require.Equal(t, s.testUser1, retrievedPost.CreatorId)
		require.Equal(t, "Test post content", retrievedPost.Desc)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestAddPost_WithFiles(t provider.T) {
	t.WithNewStep("Test add post with files", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		post := models.Post{
			Id:           postId,
			CreatorId:    s.testUser1,
			CreatorType:  models.PostUser,
			Desc:         "Test post with files",
			CreatedAt:    now,
			UpdatedAt:    now,
			LikeCount:    0,
			RepostCount:  0,
			CommentCount: 0,
			IsRepost:     false,
			Files: []*models.File{
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

		err := s.repository.AddPost(context.Background(), post)
		require.NoError(t, err)

		// Verify post was created with files
		retrievedPost, err := s.repository.GetPost(context.Background(), postId)
		require.NoError(t, err)
		require.Equal(t, postId, retrievedPost.Id)
		require.Len(t, retrievedPost.Files, 2)
		require.Equal(t, "https://example.com/file1.jpg", retrievedPost.Files[0].URL)
		require.Equal(t, "https://example.com/file2.png", retrievedPost.Files[1].URL)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestGetPost_Exists(t provider.T) {
	t.WithNewStep("Test get existing post", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		post, err := s.repository.GetPost(context.Background(), postId)
		require.NoError(t, err)
		require.Equal(t, postId, post.Id)
		require.Equal(t, s.testUser1, post.CreatorId)
		require.Equal(t, "Test post content", post.Desc)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestGetPost_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent post", func(ctx provider.StepCtx) {
		nonExistentId := uuid.New()

		_, err := s.repository.GetPost(context.Background(), nonExistentId)
		require.Error(t, err)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestDeletePost(t provider.T) {
	t.WithNewStep("Test delete post", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		// Verify post exists
		_, err = s.repository.GetPost(context.Background(), postId)
		require.NoError(t, err)

		// Delete post
		err = s.repository.DeletePost(context.Background(), postId)
		require.NoError(t, err)

		// Verify post no longer exists
		_, err = s.repository.GetPost(context.Background(), postId)
		require.Error(t, err)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestBelongsTo(t provider.T) {
	t.WithNewStep("Test belongs to", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		// Test belongs to owner
		belongs, err := s.repository.BelongsTo(context.Background(), s.testUser1, postId)
		require.NoError(t, err)
		require.True(t, belongs)

		// Test doesn't belong to other user
		belongs, err = s.repository.BelongsTo(context.Background(), s.testUser2, postId)
		require.NoError(t, err)
		require.False(t, belongs)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestGetUserPosts_NoPosts(t provider.T) {
	t.WithNewStep("Test get user posts with no posts", func(ctx provider.StepCtx) {
		posts, err := s.repository.GetUserPosts(context.Background(), s.testUser2, s.testUser1, 10, time.Now())
		require.NoError(t, err)
		require.Empty(t, posts)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestUpdatePost(t provider.T) {
	t.WithNewStep("Test update post", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Original text", now)
		require.NoError(t, err)

		// Update post
		update := models.PostUpdate{
			Id:   postId,
			Desc: "Updated text",
			Files: []*models.File{
				{
					URL:         "https://example.com/updated.jpg",
					Name:        "updated.jpg",
					DisplayType: "image",
				},
			},
		}

		err = s.repository.UpdatePost(context.Background(), update)
		require.NoError(t, err)

		// Verify post was updated
		updatedPost, err := s.repository.GetPost(context.Background(), postId)
		require.NoError(t, err)
		require.Equal(t, "Updated text", updatedPost.Desc)
		require.Len(t, updatedPost.Files, 1)
		require.Equal(t, "https://example.com/updated.jpg", updatedPost.Files[0].URL)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestLikeAndUnlikePost(t provider.T) {
	t.WithNewStep("Test like and unlike post", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		// Like post
		err = s.repository.LikePost(context.Background(), postId, s.testUser2)
		require.NoError(t, err)

		// Verify post is liked
		isLiked, err := s.repository.CheckIfPostLiked(context.Background(), postId, s.testUser2)
		require.NoError(t, err)
		require.True(t, isLiked)

		// Unlike post
		err = s.repository.UnlikePost(context.Background(), postId, s.testUser2)
		require.NoError(t, err)

		// Verify post is no longer liked
		isLiked, err = s.repository.CheckIfPostLiked(context.Background(), postId, s.testUser2)
		require.NoError(t, err)
		require.False(t, isLiked)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestCheckIfPostLiked(t provider.T) {
	t.WithNewStep("Test check if post is liked", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		// Check if not liked initially
		isLiked, err := s.repository.CheckIfPostLiked(context.Background(), postId, s.testUser2)
		require.NoError(t, err)
		require.False(t, isLiked)

		// Like post
		err = s.addPostLike(postId, s.testUser2)
		require.NoError(t, err)

		// Check if liked after like
		isLiked, err = s.repository.CheckIfPostLiked(context.Background(), postId, s.testUser2)
		require.NoError(t, err)
		require.True(t, isLiked)
	})
}

func (s *PostgresPostRepositoryTestSuite) TestLikePost_AlreadyLiked(t provider.T) {
	t.WithNewStep("Test like post when already liked", func(ctx provider.StepCtx) {
		postId := uuid.New()
		now := time.Now()

		// Create test post
		err := s.createTestPost(postId, s.testUser1, "Test post content", now)
		require.NoError(t, err)

		// Like post
		err = s.addPostLike(postId, s.testUser2)
		require.NoError(t, err)

		// Try to like again
		err = s.repository.LikePost(context.Background(), postId, s.testUser2)
		require.Error(t, err)
		require.True(t, errors.Is(err, post_errors.ErrAlreadyExists))
	})
}

// Helper methods
func (s *PostgresPostRepositoryTestSuite) insertTestData() error {
	// Insert test users
	users := []struct {
		id       string
		username string
	}{
		{s.testUser1.String(), "testuser1"},
		{s.testUser2.String(), "testuser2"},
		{s.testUser3.String(), "testuser3"},
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

	return nil
}

func (s *PostgresPostRepositoryTestSuite) createTestPost(postId, userId uuid.UUID, text string, createdAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO post (id, creator_id, creator_type, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost)
		VALUES ($1, $2, $3, $4, $5, $5, 0, 0, 0, false)
	`, postId, userId, models.PostUser, text, createdAt)
	return err
}

func (s *PostgresPostRepositoryTestSuite) addPostFile(postId uuid.UUID, fileURL, fileType string) error {
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
		INSERT INTO post_file (post_id, file_url, file_type)
		VALUES ($1, $2, $3)
	`, postId, fileURL, fileType)
	return err
}

func (s *PostgresPostRepositoryTestSuite) addPostLike(postId, userId uuid.UUID) error {
	_, err := s.db.Exec(`
		INSERT INTO like_post (user_id, post_id)
		VALUES ($1, $2)
	`, userId, postId)
	return err
}

func (s *PostgresPostRepositoryTestSuite) createFriendship(user1, user2 uuid.UUID, status models.UserRelation) error {
	_, err := s.db.Exec(`
		INSERT INTO friendship (user1_id, user2_id, status)
		VALUES ($1, $2, $3)
	`, user1, user2, status)
	return err
}

func (s *PostgresPostRepositoryTestSuite) cleanupPostData() error {
	queries := []string{
		`DELETE FROM like_post`,
		`DELETE FROM post_file`,
		`DELETE FROM post`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresPostRepositoryTestSuite) cleanupTestData() error {
	queries := []string{
		`DELETE FROM like_post`,
		`DELETE FROM post_file`,
		`DELETE FROM post`,
		`DELETE FROM friendship`,
		`DELETE FROM files`,
		`DELETE FROM profile`,
		`DELETE FROM "user"`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestPostgresPostRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresPostRepositoryTestSuite))
}
