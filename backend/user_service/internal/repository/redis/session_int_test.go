//go:build integration
// +build integration

package redis_test

import (
	"context"
	"github.com/redis/go-redis/v9"
	redis2 "quickflow/config/redis"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickflow/shared/models"
	redisRepo "quickflow/user_service/internal/repository/redis"
)

type RedisSessionRepositoryTestSuite struct {
	suite.Suite
	repo         *redisRepo.RedisSessionRepository
	redisCfg     *redis2.RedisConfig
	ctx          context.Context
	testUsers    []uuid.UUID
	testSessions []uuid.UUID
}

func (s *RedisSessionRepositoryTestSuite) BeforeAll(t provider.T) {
	t.Epic("Redis Session Repository")
	t.Feature("Integration Tests")

	t.WithNewStep("Setup test suite", func(ctx provider.StepCtx) {
		s.ctx = context.Background()
		s.redisCfg = redis2.NewRedisConfig()
		s.repo = redisRepo.NewRedisSessionRepository(s.redisCfg)
		s.testUsers = []uuid.UUID{
			uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		}
		s.testSessions = []uuid.UUID{
			uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
			uuid.MustParse("223e4567-e89b-12d3-a456-426614174001"),
		}
	})
}

func (s *RedisSessionRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Tear down test suite", func(ctx provider.StepCtx) {
		s.repo.Close()
	})
}

func (s *RedisSessionRepositoryTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup Redis before test", func(ctx provider.StepCtx) {
		// Очищаем тестовые данные перед каждым тестом
		client := redis.NewClient(&redis.Options{
			Addr:     s.redisCfg.GetURL(),
			Password: s.redisCfg.GetPass(),
		})
		defer client.Close()

		for _, sessionID := range s.testSessions {
			client.Del(s.ctx, sessionID.String())
		}
	})
}

func TestRedisSessionRepository(t *testing.T) {
	suite.RunSuite(t, new(RedisSessionRepositoryTestSuite))
}

func (s *RedisSessionRepositoryTestSuite) TestSaveSession(t provider.T) {
	t.Description("Test saving session to Redis with valid data")
	t.Tags("integration", "redis", "save")

	t.WithNewStep("Save session to Redis", func(ctx provider.StepCtx) {
		// Arrange
		userID := s.testUsers[0]
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		// Act
		err := s.repo.SaveSession(s.ctx, userID, session)

		// Assert
		require.NoError(t, err)
	})

	t.WithNewStep("Verify session was saved", func(ctx provider.StepCtx) {
		// Verify that session was actually saved
		exists, err := s.repo.IsExists(s.ctx, s.testSessions[0])
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestSaveSession_Expired(t provider.T) {
	t.Description("Test saving session with expired TTL")
	t.Tags("integration", "redis", "expiration")

	t.WithNewStep("Save session with negative TTL", func(ctx provider.StepCtx) {
		// Arrange
		userID := s.testUsers[0]
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(-10 * time.Minute), // Прошедшее время
		}

		// Act
		err := s.repo.SaveSession(s.ctx, userID, session)

		// Assert
		require.NoError(t, err)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestLookupUserSession(t provider.T) {
	t.Description("Test retrieving user ID by session ID")
	t.Tags("integration", "redis", "lookup")

	t.WithNewStep("Save session first", func(ctx provider.StepCtx) {
		// Arrange
		userID := s.testUsers[0]
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		err := s.repo.SaveSession(s.ctx, userID, session)
		require.NoError(t, err)
	})

	t.WithNewStep("Lookup user session", func(ctx provider.StepCtx) {
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		// Act
		retrievedUserID, err := s.repo.LookupUserSession(s.ctx, session)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, s.testUsers[0], retrievedUserID)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestLookupUserSession_NotFound(t provider.T) {
	t.Description("Test retrieving non-existent session")
	t.Tags("integration", "redis", "lookup", "error")

	t.WithNewStep("Attempt to lookup non-existent session", func(ctx provider.StepCtx) {
		// Arrange
		nonExistentSession := models.Session{
			SessionId:  uuid.MustParse("333e4567-e89b-12d3-a456-426614174000"),
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		// Act
		userID, err := s.repo.LookupUserSession(s.ctx, nonExistentSession)

		// Assert
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, userID)
		assert.Contains(t, err.Error(), "unable to get userId")
	})
}

func (s *RedisSessionRepositoryTestSuite) TestIsExists(t provider.T) {
	t.Description("Test checking if session exists")
	t.Tags("integration", "redis", "exists")

	t.WithNewStep("Save session first", func(ctx provider.StepCtx) {
		// Arrange
		userID := s.testUsers[0]
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		err := s.repo.SaveSession(s.ctx, userID, session)
		require.NoError(t, err)
	})

	t.WithNewStep("Check if session exists", func(ctx provider.StepCtx) {
		// Act
		exists, err := s.repo.IsExists(s.ctx, s.testSessions[0])

		// Assert
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestIsExists_NotFound(t provider.T) {
	t.Description("Test checking if non-existent session exists")
	t.Tags("integration", "redis", "exists", "error")

	t.WithNewStep("Check non-existent session", func(ctx provider.StepCtx) {
		// Arrange
		nonExistentSession := uuid.MustParse("333e4567-e89b-12d3-a456-426614174000")

		// Act
		exists, err := s.repo.IsExists(s.ctx, nonExistentSession)

		// Assert
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestDeleteSession(t provider.T) {
	t.Description("Test deleting existing session")
	t.Tags("integration", "redis", "delete")

	t.WithNewStep("Save session first", func(ctx provider.StepCtx) {
		// Arrange
		userID := s.testUsers[0]
		session := models.Session{
			SessionId:  s.testSessions[0],
			ExpireDate: time.Now().Add(10 * time.Minute),
		}

		err := s.repo.SaveSession(s.ctx, userID, session)
		require.NoError(t, err)
	})

	t.WithNewStep("Verify session exists before deletion", func(ctx provider.StepCtx) {
		exists, err := s.repo.IsExists(s.ctx, s.testSessions[0])
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.WithNewStep("Delete session", func(ctx provider.StepCtx) {
		// Act
		err := s.repo.DeleteSession(s.ctx, s.testSessions[0].String())
		require.NoError(t, err)
	})

	t.WithNewStep("Verify session was deleted", func(ctx provider.StepCtx) {
		// Verify session was deleted
		exists, err := s.repo.IsExists(s.ctx, s.testSessions[0])
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func (s *RedisSessionRepositoryTestSuite) TestDeleteSession_NonExistent(t provider.T) {
	t.Description("Test deleting non-existent session")
	t.Tags("integration", "redis", "delete")

	t.WithNewStep("Delete non-existent session", func(ctx provider.StepCtx) {
		// Arrange
		nonExistentSession := "333e4567-e89b-12d3-a456-426614174000"

		// Act
		err := s.repo.DeleteSession(s.ctx, nonExistentSession)

		// Assert
		require.NoError(t, err)
	})
}
