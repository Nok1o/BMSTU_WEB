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

type PostgresMessageRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.MessageRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
	testChat   uuid.UUID
}

func (s *PostgresMessageRepositoryTestSuite) BeforeAll(t provider.T) {
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

		// Generate test users and chat
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testChat = uuid.New()

		// Insert test users, profiles and chat
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		err = s.createTestChat()
		if err != nil {
			log.Fatalf("Failed to create test chat: %v", err)
		}

		s.repository = postgres.NewPostgresMessageRepository(s.db)
	})
}

func (s *PostgresMessageRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostgresMessageRepositoryTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupMessageData()
	})
	t.Epic("Integration")
}

func (s *PostgresMessageRepositoryTestSuite) TestSaveMessage(t provider.T) {
	t.WithNewStep("Test save message", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		message := models.Message{
			ID:        messageID,
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Hello, world!",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.repository.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Verify message was saved
		retrievedMessage, err := s.repository.GetMessageById(context.Background(), messageID)
		require.NoError(t, err)
		require.Equal(t, messageID, retrievedMessage.ID)
		require.Equal(t, s.testChat, retrievedMessage.ChatID)
		require.Equal(t, s.testUser1, retrievedMessage.SenderID)
		require.Equal(t, "Hello, world!", retrievedMessage.Text)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetMessagesForChatOlder(t provider.T) {
	t.WithNewStep("Test get messages for chat older than timestamp", func(ctx provider.StepCtx) {
		now := time.Now()

		// Create test messages
		message1 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Message 1",
			CreatedAt: now.Add(-5 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
		}

		message2 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser2,
			Text:      "Message 2",
			CreatedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now.Add(-3 * time.Minute),
		}

		message3 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Message 3",
			CreatedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}

		err := s.repository.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		err = s.repository.SaveMessage(context.Background(), message2)
		require.NoError(t, err)
		err = s.repository.SaveMessage(context.Background(), message3)
		require.NoError(t, err)

		// Get messages older than 2 minutes
		messages, err := s.repository.GetMessagesForChatOlder(
			context.Background(),
			s.testChat,
			10,
			now.Add(-2*time.Minute),
		)
		require.NoError(t, err)
		require.Len(t, messages, 2) // Should get message1 and message2

		// Verify messages are in correct order (oldest first)
		require.Equal(t, "Message 1", messages[0].Text)
		require.Equal(t, "Message 2", messages[1].Text)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetLastChatMessage(t provider.T) {
	t.WithNewStep("Test get last chat message", func(ctx provider.StepCtx) {
		now := time.Now()

		// Create test messages
		message1 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "First message",
			CreatedAt: now.Add(-5 * time.Minute),
			UpdatedAt: now.Add(-5 * time.Minute),
		}

		message2 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser2,
			Text:      "Last message",
			CreatedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}

		err := s.repository.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		err = s.repository.SaveMessage(context.Background(), message2)
		require.NoError(t, err)

		// Get last message
		lastMessage, err := s.repository.GetLastChatMessage(context.Background(), s.testChat)
		require.NoError(t, err)
		require.NotNil(t, lastMessage)
		require.Equal(t, "Last message", lastMessage.Text)
		require.Equal(t, s.testUser2, lastMessage.SenderID)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetLastChatMessage_NoMessages(t provider.T) {
	t.WithNewStep("Test get last chat message when no messages", func(ctx provider.StepCtx) {
		lastMessage, err := s.repository.GetLastChatMessage(context.Background(), s.testChat)
		require.NoError(t, err)
		require.Nil(t, lastMessage)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetMessageById(t provider.T) {
	t.WithNewStep("Test get message by ID", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		message := models.Message{
			ID:        messageID,
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Test message",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.repository.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Get message by ID
		retrievedMessage, err := s.repository.GetMessageById(context.Background(), messageID)
		require.NoError(t, err)
		require.Equal(t, messageID, retrievedMessage.ID)
		require.Equal(t, "Test message", retrievedMessage.Text)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetMessageById_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent message by ID", func(ctx provider.StepCtx) {
		nonExistentID := uuid.New()

		_, err := s.repository.GetMessageById(context.Background(), nonExistentID)
		require.Error(t, err)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestDeleteMessage(t provider.T) {
	t.WithNewStep("Test delete message", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		message := models.Message{
			ID:        messageID,
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Message to delete",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err := s.repository.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Verify message exists
		_, err = s.repository.GetMessageById(context.Background(), messageID)
		require.NoError(t, err)

		// Delete message
		err = s.repository.DeleteMessage(context.Background(), messageID)
		require.NoError(t, err)

		// Verify message was deleted
		_, err = s.repository.GetMessageById(context.Background(), messageID)
		require.Error(t, err)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetLastReadTs_NotSet(t provider.T) {
	t.WithNewStep("Test get last read timestamp when not set", func(ctx provider.StepCtx) {
		lastRead, err := s.repository.GetLastReadTs(context.Background(), s.testChat, s.testUser1)
		require.NoError(t, err)
		require.Nil(t, lastRead)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetNumUnreadMessages(t provider.T) {
	t.WithNewStep("Test get number of unread messages", func(ctx provider.StepCtx) {
		now := time.Now()

		// Create messages from other user
		message1 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser2, // Other user
			Text:      "Unread message 1",
			CreatedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now.Add(-3 * time.Minute),
		}

		message2 := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser2, // Other user
			Text:      "Unread message 2",
			CreatedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}

		err := s.repository.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		err = s.repository.SaveMessage(context.Background(), message2)
		require.NoError(t, err)

		// Set last read to before first message
		err = s.repository.UpdateLastReadTs(context.Background(), now.Add(-5*time.Minute), s.testChat, s.testUser1)
		require.NoError(t, err)

		// Get number of unread messages
		numUnread, err := s.repository.GetNumUnreadMessages(context.Background(), s.testChat, s.testUser1)
		require.NoError(t, err)
		require.Equal(t, 2, numUnread)
	})
}

func (s *PostgresMessageRepositoryTestSuite) TestGetNumUnreadMessages_NoUnread(t provider.T) {
	t.WithNewStep("Test get number of unread messages when all read", func(ctx provider.StepCtx) {
		now := time.Now()

		// Create message and mark as read
		message := models.Message{
			ID:        uuid.New(),
			ChatID:    s.testChat,
			SenderID:  s.testUser2,
			Text:      "Read message",
			CreatedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now.Add(-2 * time.Minute),
		}

		err := s.repository.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Set last read to after message
		err = s.repository.UpdateLastReadTs(context.Background(), now, s.testChat, s.testUser1)
		require.NoError(t, err)

		// Get number of unread messages
		numUnread, err := s.repository.GetNumUnreadMessages(context.Background(), s.testChat, s.testUser1)
		require.NoError(t, err)
		require.Equal(t, 0, numUnread)
	})
}

// Helper methods
func (s *PostgresMessageRepositoryTestSuite) insertTestUsers() error {
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
	return nil
}

func (s *PostgresMessageRepositoryTestSuite) createTestChat() error {
	// Create private chat
	_, err := s.db.Exec(`
		INSERT INTO chat (id, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, s.testChat, models.ChatTypePrivate, time.Now(), time.Now())
	if err != nil {
		return err
	}

	// Add users to chat
	_, err = s.db.Exec(`
		INSERT INTO chat_user (chat_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, s.testChat, s.testUser1, s.testUser2)
	return err
}

func (s *PostgresMessageRepositoryTestSuite) cleanupMessageData() error {
	queries := []string{
		`DELETE FROM message_file`,
		`DELETE FROM message`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresMessageRepositoryTestSuite) cleanupTestData() error {
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

func TestPostgresMessageRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresMessageRepositoryTestSuite))
}
