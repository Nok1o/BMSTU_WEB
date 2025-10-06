//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"log"
	"quickflow/config/test"
	"quickflow/messenger_service/internal/repository/postgres"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"quickflow/shared/models"
)

type PostgresChatRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.ChatRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
	testUser3  uuid.UUID
}

func (s *PostgresChatRepositoryTestSuite) BeforeAll(t provider.T) {
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

		// Generate test users
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testUser3 = uuid.New()

		// Insert test users and profiles
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		s.repository = postgres.NewPostgresChatRepository(s.db)
	})
}

func (s *PostgresChatRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostgresChatRepositoryTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupChatData()
	})
	t.Epic("Integration")
}

func (s *PostgresChatRepositoryTestSuite) TestCreateChat_Private(t provider.T) {
	t.WithNewStep("Test create private chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		chat := models.Chat{
			ID:        chatID,
			Type:      models.ChatTypePrivate,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.repository.CreateChat(context.Background(), chat)
		require.NoError(t, err)

		// Verify chat was created
		exists, err := s.repository.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.True(t, exists)

		// Verify chat type is private
		retrievedChat, err := s.repository.GetChat(context.Background(), chatID)
		require.NoError(t, err)
		require.Equal(t, models.ChatTypePrivate, retrievedChat.Type)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestCreateChat_Group(t provider.T) {
	t.WithNewStep("Test create group chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		chat := models.Chat{
			ID:        chatID,
			Name:      "Test Group Chat",
			AvatarURL: "https://example.com/avatar.png",
			Type:      models.ChatTypeGroup,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.repository.CreateChat(context.Background(), chat)
		require.NoError(t, err)

		// Verify chat was created with correct data
		retrievedChat, err := s.repository.GetChat(context.Background(), chatID)
		require.NoError(t, err)
		require.Equal(t, chat.Name, retrievedChat.Name)
		require.Equal(t, chat.AvatarURL, retrievedChat.AvatarURL)
		require.Equal(t, chat.Type, retrievedChat.Type)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetChat_Exists(t provider.T) {
	t.WithNewStep("Test get existing chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create test chat
		err := s.createTestChat(chatID, "Test Chat", models.ChatTypeGroup, now)
		require.NoError(t, err)

		chat, err := s.repository.GetChat(context.Background(), chatID)
		require.NoError(t, err)
		require.Equal(t, chatID, chat.ID)
		require.Equal(t, "Test Chat", chat.Name)
		require.Equal(t, models.ChatTypeGroup, chat.Type)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetChat_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent chat", func(ctx provider.StepCtx) {
		nonExistentID := uuid.New()

		_, err := s.repository.GetChat(context.Background(), nonExistentID)
		require.Error(t, err)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetPrivateChat_Exists(t provider.T) {
	t.WithNewStep("Test get existing private chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create private chat and add users
		err := s.createPrivateChatWithUsers(chatID, s.testUser1, s.testUser2, now)
		require.NoError(t, err)

		chat, err := s.repository.GetPrivateChat(context.Background(), s.testUser1, s.testUser2)
		require.NoError(t, err)
		require.Equal(t, chatID, chat.ID)
		require.Equal(t, models.ChatTypePrivate, chat.Type)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetPrivateChat_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent private chat", func(ctx provider.StepCtx) {
		_, err := s.repository.GetPrivateChat(context.Background(), s.testUser1, s.testUser2)
		require.Error(t, err)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetUserChats_WithChats(t provider.T) {
	t.WithNewStep("Test get user chats", func(ctx provider.StepCtx) {
		chatID1 := uuid.New()
		chatID2 := uuid.New()
		now := time.Now()

		// Create chats and add user to them
		err := s.createTestChat(chatID1, "Chat 1", models.ChatTypeGroup, now)
		require.NoError(t, err)
		err = s.createTestChat(chatID2, "Chat 2", models.ChatTypeGroup, now)
		require.NoError(t, err)

		err = s.addUserToChat(chatID1, s.testUser1)
		require.NoError(t, err)
		err = s.addUserToChat(chatID2, s.testUser1)
		require.NoError(t, err)

		chats, err := s.repository.GetUserChats(context.Background(), s.testUser1)
		require.NoError(t, err)
		require.Len(t, chats, 2)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetUserChats_NoChats(t provider.T) {
	t.WithNewStep("Test get user chats when user has no chats", func(ctx provider.StepCtx) {
		chats, err := s.repository.GetUserChats(context.Background(), s.testUser1)
		require.NoError(t, err)
		require.Empty(t, chats)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestGetChatParticipants(t provider.T) {
	t.WithNewStep("Test get chat participants", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create chat and add participants
		err := s.createTestChat(chatID, "Group Chat", models.ChatTypeGroup, now)
		require.NoError(t, err)

		err = s.addUserToChat(chatID, s.testUser1)
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser2)
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser3)
		require.NoError(t, err)

		participants, err := s.repository.GetChatParticipants(context.Background(), chatID)
		require.NoError(t, err)
		require.Len(t, participants, 3)
		require.Contains(t, participants, s.testUser1)
		require.Contains(t, participants, s.testUser2)
		require.Contains(t, participants, s.testUser3)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestJoinAndLeaveChat(t provider.T) {
	t.WithNewStep("Test join and leave chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create chat
		err := s.createTestChat(chatID, "Test Chat", models.ChatTypeGroup, now)
		require.NoError(t, err)

		// Join chat
		err = s.repository.JoinChat(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)

		// Verify user is participant
		isParticipant, err := s.repository.IsParticipant(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)
		require.True(t, isParticipant)

		// Leave chat
		err = s.repository.LeaveChat(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)

		// Verify user is no longer participant
		isParticipant, err = s.repository.IsParticipant(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)
		require.False(t, isParticipant)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestExistsChat(t provider.T) {
	t.WithNewStep("Test check chat existence", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create chat
		err := s.createTestChat(chatID, "Test Chat", models.ChatTypeGroup, now)
		require.NoError(t, err)

		exists, err := s.repository.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.True(t, exists)

		// Check non-existent chat
		nonExistentID := uuid.New()
		exists, err = s.repository.Exists(context.Background(), nonExistentID)
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func (s *PostgresChatRepositoryTestSuite) TestDeleteChat(t provider.T) {
	t.WithNewStep("Test delete chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		now := time.Now()

		// Create chat
		err := s.createTestChat(chatID, "Test Chat", models.ChatTypeGroup, now)
		require.NoError(t, err)

		// Verify chat exists
		exists, err := s.repository.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.True(t, exists)

		// Delete chat
		err = s.repository.DeleteChat(context.Background(), chatID)
		require.NoError(t, err)

		// Verify chat no longer exists
		exists, err = s.repository.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.False(t, exists)
	})
}

// Helper methods
func (s *PostgresChatRepositoryTestSuite) insertTestUsers() error {
	// Insert test users with minimal required fields
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

func (s *PostgresChatRepositoryTestSuite) createTestChat(chatID uuid.UUID, name string, chatType models.ChatType, createdAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO chat (id, name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, chatID, name, chatType, createdAt, createdAt)
	return err
}

func (s *PostgresChatRepositoryTestSuite) createPrivateChatWithUsers(chatID uuid.UUID, user1, user2 uuid.UUID, createdAt time.Time) error {
	// Create private chat
	_, err := s.db.Exec(`
		INSERT INTO chat (id, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, chatID, models.ChatTypePrivate, createdAt, createdAt)
	if err != nil {
		return err
	}

	// Add users to chat
	_, err = s.db.Exec(`
		INSERT INTO chat_user (chat_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, chatID, user1, user2)
	return err
}

func (s *PostgresChatRepositoryTestSuite) addUserToChat(chatID, userID uuid.UUID) error {
	_, err := s.db.Exec(`
		INSERT INTO chat_user (chat_id, user_id)
		VALUES ($1, $2)
	`, chatID, userID)
	return err
}

func (s *PostgresChatRepositoryTestSuite) addUserToChatWithLastRead(chatID, userID uuid.UUID, lastRead time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO chat_user (chat_id, user_id, last_read)
		VALUES ($1, $2, $3)
	`, chatID, userID, lastRead)
	return err
}

func (s *PostgresChatRepositoryTestSuite) cleanupChatData() error {
	queries := []string{
		`DELETE FROM message_file`,
		`DELETE FROM message`,
		`DELETE FROM chat_user`,
		`DELETE FROM chat`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresChatRepositoryTestSuite) cleanupTestData() error {
	queries := []string{
		`DELETE FROM message_file`,
		`DELETE FROM message`,
		`DELETE FROM chat_user`,
		`DELETE FROM chat`,
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

func TestPostgresChatRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresChatRepositoryTestSuite))
}
