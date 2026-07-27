//go:build integration
// +build integration

package usecase_test

import (
	"context"
	"database/sql"
	"log"
	addr "quickflow/config/micro-addr"
	"quickflow/config/test"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/messenger_service/internal/usecase"
	"quickflow/messenger_service/utils/validation"
	"quickflow/shared/client/file_service"
	userclient "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	getEnv "quickflow/utils/get-env"
	service_discovery "quickflow/utils/service-discovery"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"google.golang.org/grpc"
	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/shared/models"
)

type MessageServiceTestSuite struct {
	suite.Suite
	db             *sql.DB
	messageRepo    *postgres.MessageRepository
	chatRepo       *postgres.ChatRepository
	messageService usecase.MessageService
	fileService    usecase.FileService
	profileService *userclient.ProfileClient
	validator      *validation.MessageValidator
	testUser1      uuid.UUID
	testUser2      uuid.UUID
	testChat       uuid.UUID
	grpcConns      []*grpc.ClientConn
}

func (s *MessageServiceTestSuite) BeforeAll(t provider.T) {
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

		// Setup gRPC connections
		grpcConnFileService, err := service_discovery.NewGRPCClient(
			addr.DefaultFileServiceName,
			service_discovery.ModeFailover,
			interceptors.RequestIDClientInterceptor(),
		)
		if err != nil {
			log.Fatalf("failed to connect to file service: %v", err)
		}
		s.grpcConns = append(s.grpcConns, grpcConnFileService)

		grpcConnUserService, err := service_discovery.NewGRPCClient(
			addr.DefaultUserServiceName,
			service_discovery.ModeFailover,
			interceptors.RequestIDClientInterceptor(),
		)
		if err != nil {
			log.Fatalf("failed to connect to user service: %v", err)
		}
		s.grpcConns = append(s.grpcConns, grpcConnUserService)

		// Initialize services
		s.validator = validation.NewMessageValidator()
		s.fileService = file_service.NewFileClient(grpcConnFileService)
		s.profileService = userclient.NewProfileClient(grpcConnUserService)

		// Initialize repositories
		s.messageRepo = postgres.NewPostgresMessageRepository(s.db)
		s.chatRepo = postgres.NewPostgresChatRepository(s.db)

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

		// Initialize message service
		s.messageService = *usecase.NewMessageService(
			s.messageRepo,
			s.fileService,
			s.chatRepo,
			s.validator,
		)
	})
}

func (s *MessageServiceTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		// Close gRPC connections
		for _, conn := range s.grpcConns {
			if conn != nil {
				conn.Close()
			}
		}

		// Cleanup database
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *MessageServiceTestSuite) BeforeEach(t provider.T) {
	t.Epic("Integration")
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupMessageData()
	})
}

func (s *MessageServiceTestSuite) TestSaveMessage(t provider.T) {
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

		savedMessage, err := s.messageService.SaveMessage(context.Background(), message)
		require.NoError(t, err)
		require.Equal(t, messageID, savedMessage.ID)
		require.Equal(t, "Hello, world!", savedMessage.Text)
		require.Equal(t, s.testUser1, savedMessage.SenderID)
	})
}

func (s *MessageServiceTestSuite) TestSaveMessage_WithAttachments(t provider.T) {
	t.WithNewStep("Test save message with attachments", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		message := models.Message{
			ID:        messageID,
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "Message with files",
			CreatedAt: now,
			UpdatedAt: now,
			Attachments: []*models.File{
				{
					URL:         "https://example.com/file1.jpg",
					DisplayType: "image",
					Name:        "file1.jpg",
				},
			},
		}

		savedMessage, err := s.messageService.SaveMessage(context.Background(), message)
		require.NoError(t, err)
		require.Len(t, savedMessage.Attachments, 1)
		require.Equal(t, "https://example.com/file1.jpg", savedMessage.Attachments[0].URL)
	})
}

func (s *MessageServiceTestSuite) TestSaveMessage_CreatePrivateChat(t provider.T) {
	t.WithNewStep("Test save message creates private chat", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		// Message without chat ID but with receiver - should create private chat
		message := models.Message{
			ID:         messageID,
			SenderID:   s.testUser1,
			ReceiverID: s.testUser2,
			Text:       "Private message",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		savedMessage, err := s.messageService.SaveMessage(context.Background(), message)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, savedMessage.ChatID)
		require.Equal(t, "Private message", savedMessage.Text)

		// Verify chat was created and both users are participants
		isParticipant1, err := s.chatRepo.IsParticipant(context.Background(), savedMessage.ChatID, s.testUser1)
		require.NoError(t, err)
		require.True(t, isParticipant1)

		isParticipant2, err := s.chatRepo.IsParticipant(context.Background(), savedMessage.ChatID, s.testUser2)
		require.NoError(t, err)
		require.True(t, isParticipant2)
	})
}

func (s *MessageServiceTestSuite) TestGetMessagesForChatOlder(t provider.T) {
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

		_, err := s.messageService.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		_, err = s.messageService.SaveMessage(context.Background(), message2)
		require.NoError(t, err)

		// Get messages older than 2 minutes
		messages, err := s.messageService.GetMessagesForChatOlder(
			context.Background(),
			s.testChat,
			s.testUser1,
			10,
			now.Add(-2*time.Minute),
		)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		// Verify messages are in correct order (oldest first)
		require.Equal(t, "Message 1", messages[0].Text)
		require.Equal(t, "Message 2", messages[1].Text)
	})
}

func (s *MessageServiceTestSuite) TestGetMessagesForChatOlder_NotParticipant(t provider.T) {
	t.WithNewStep("Test get messages when not participant", func(ctx provider.StepCtx) {
		nonParticipant := uuid.New()

		_, err := s.messageService.GetMessagesForChatOlder(
			context.Background(),
			s.testChat,
			nonParticipant,
			10,
			time.Now(),
		)
		require.Error(t, err)
		require.Equal(t, messenger_errors.ErrNotParticipant, err)
	})
}

func (s *MessageServiceTestSuite) TestGetLastChatMessage(t provider.T) {
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

		_, err := s.messageService.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		_, err = s.messageService.SaveMessage(context.Background(), message2)
		require.NoError(t, err)

		// Get last message
		lastMessage, err := s.messageService.GetLastChatMessage(context.Background(), s.testChat)
		require.NoError(t, err)
		require.NotNil(t, lastMessage)
		require.Equal(t, "Last message", lastMessage.Text)
	})
}

func (s *MessageServiceTestSuite) TestGetMessageById(t provider.T) {
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

		_, err := s.messageService.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Get message by ID
		retrievedMessage, err := s.messageService.GetMessageById(context.Background(), messageID)
		require.NoError(t, err)
		require.Equal(t, messageID, retrievedMessage.ID)
		require.Equal(t, "Test message", retrievedMessage.Text)
	})
}

func (s *MessageServiceTestSuite) TestDeleteMessage(t provider.T) {
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

		_, err := s.messageService.SaveMessage(context.Background(), message)
		require.NoError(t, err)

		// Verify message exists
		_, err = s.messageService.GetMessageById(context.Background(), messageID)
		require.NoError(t, err)

		// Delete message
		err = s.messageService.DeleteMessage(context.Background(), messageID)
		require.NoError(t, err)

		// Verify message was deleted
		_, err = s.messageService.GetMessageById(context.Background(), messageID)
		require.Error(t, err)
	})
}

func (s *MessageServiceTestSuite) TestGetNumUnreadMessages(t provider.T) {
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

		_, err := s.messageService.SaveMessage(context.Background(), message1)
		require.NoError(t, err)
		_, err = s.messageService.SaveMessage(context.Background(), message2)
		require.NoError(t, err)

		// Set last read to before first message
		err = s.messageService.UpdateLastReadTs(context.Background(), now.Add(-5*time.Minute), s.testChat, s.testUser1)
		require.NoError(t, err)

		// Get number of unread messages
		numUnread, err := s.messageService.GetNumUnreadMessages(context.Background(), s.testChat, s.testUser1)
		require.NoError(t, err)
		require.Equal(t, 2, numUnread)
	})
}

func (s *MessageServiceTestSuite) TestSaveMessage_ValidationError(t provider.T) {
	t.WithNewStep("Test save message with validation error", func(ctx provider.StepCtx) {
		messageID := uuid.New()
		now := time.Now()

		// Empty message without text or attachments
		message := models.Message{
			ID:        messageID,
			ChatID:    s.testChat,
			SenderID:  s.testUser1,
			Text:      "",
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err := s.messageService.SaveMessage(context.Background(), message)
		require.Error(t, err)
	})
}

func (s *MessageServiceTestSuite) insertTestUsers() error {
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

func (s *MessageServiceTestSuite) createTestChat() error {
	// Create private chat
	_, err := s.db.Exec(`
		INSERT INTO chat (id, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, s.testChat, models.ChatTypeGroup, time.Now(), time.Now())
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

func (s *MessageServiceTestSuite) cleanupMessageData() error {
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

func (s *MessageServiceTestSuite) cleanupTestData() error {
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

func TestMessageServiceInt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(MessageServiceTestSuite))
}
