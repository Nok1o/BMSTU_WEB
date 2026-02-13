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
	"testing"

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

type ChatServiceTestSuite struct {
	suite.Suite
	db             *sql.DB
	chatRepo       *postgres.ChatRepository
	messageRepo    *postgres.MessageRepository
	chatService    usecase.ChatService
	fileService    usecase.FileService
	profileService *userclient.ProfileClient
	validator      *validation.ChatValidator
	testUser1      uuid.UUID
	testUser2      uuid.UUID
	testUser3      uuid.UUID
	testChats      map[uuid.UUID]*testChatData // Track test chats for cleanup
	grpcConns      []*grpc.ClientConn
}

type testChatData struct {
	id           uuid.UUID
	chatType     models.ChatType
	name         string
	participants []uuid.UUID
	messages     []uuid.UUID
}

func (s *ChatServiceTestSuite) BeforeAll(t provider.T) {
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
			interceptors.RequestIDClientInterceptor(),
		)
		if err != nil {
			log.Fatalf("failed to connect to file service: %v", err)
		}
		s.grpcConns = append(s.grpcConns, grpcConnFileService)

		grpcConnUserService, err := service_discovery.NewGRPCClient(
			addr.DefaultUserServiceName,
			interceptors.RequestIDClientInterceptor(),
		)
		if err != nil {
			log.Fatalf("failed to connect to user service: %v", err)
		}
		s.grpcConns = append(s.grpcConns, grpcConnUserService)

		// Initialize services
		s.validator = validation.NewChatValidator()
		s.fileService = file_service.NewFileClient(grpcConnFileService)
		s.profileService = userclient.NewProfileClient(grpcConnUserService)

		// Initialize repositories
		s.chatRepo = postgres.NewPostgresChatRepository(s.db)
		s.messageRepo = postgres.NewPostgresMessageRepository(s.db)

		// Generate test users
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()
		s.testUser3 = uuid.New()

		// Initialize test chats map
		s.testChats = make(map[uuid.UUID]*testChatData)

		// Insert test users and profiles
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		// Initialize chat service
		s.chatService = *usecase.NewChatUseCase(
			s.chatRepo,
			s.fileService,
			s.profileService,
			s.messageRepo,
			s.validator,
		)
	})
}

func (s *ChatServiceTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		// Close gRPC connections
		for _, conn := range s.grpcConns {
			if conn != nil {
				conn.Close()
			}
		}

		// Cleanup database
		if s.db != nil {
			s.cleanupAllTestData()
			s.db.Close()
		}
	})
}

func (s *ChatServiceTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupTestChats()
	})
	t.Epic("Integration")
}

func (s *ChatServiceTestSuite) TestCreateChat_Private(t provider.T) {
	t.WithNewStep("Test create private chat", func(ctx provider.StepCtx) {
		chatInfo := models.ChatCreationInfo{
			Type: models.ChatTypePrivate,
		}

		chat, err := s.chatService.CreateChat(context.Background(), chatInfo)
		require.NoError(t, err)
		require.Equal(t, models.ChatTypePrivate, chat.Type)
		require.NotEqual(t, uuid.Nil, chat.ID)
		require.Empty(t, chat.Name)
		require.Empty(t, chat.AvatarURL)

		// Track the created chat for cleanup
		s.testChats[chat.ID] = &testChatData{
			id:       chat.ID,
			chatType: chat.Type,
		}
	})
}

func (s *ChatServiceTestSuite) TestGetChat(t provider.T) {
	t.WithNewStep("Test get chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)

		chat, err := s.chatService.GetChat(context.Background(), chatID)
		require.NoError(t, err)
		require.Equal(t, chatID, chat.ID)
		require.Equal(t, "Test Chat", chat.Name)
		require.Equal(t, models.ChatTypeGroup, chat.Type)
	})
}

func (s *ChatServiceTestSuite) TestDeleteChat(t provider.T) {
	t.WithNewStep("Test delete chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Chat to delete")
		require.NoError(t, err)

		// Verify chat exists
		exists, err := s.chatRepo.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.True(t, exists)

		err = s.chatService.DeleteChat(context.Background(), chatID)
		require.NoError(t, err)

		// Verify chat was deleted
		exists, err = s.chatRepo.Exists(context.Background(), chatID)
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func (s *ChatServiceTestSuite) TestJoinAndLeaveChat(t provider.T) {
	t.WithNewStep("Test join and leave chat", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)

		// Join chat
		err = s.chatService.JoinChat(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)

		// Verify user is participant
		isParticipant, err := s.chatRepo.IsParticipant(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)
		require.True(t, isParticipant)

		// Leave chat
		err = s.chatService.LeaveChat(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)

		// Verify user is no longer participant
		isParticipant, err = s.chatRepo.IsParticipant(context.Background(), chatID, s.testUser1)
		require.NoError(t, err)
		require.False(t, isParticipant)
	})
}

func (s *ChatServiceTestSuite) TestGetChatParticipants(t provider.T) {
	t.WithNewStep("Test get chat participants", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)

		// Add participants
		err = s.addUserToChat(chatID, s.testUser1)
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser2)
		require.NoError(t, err)

		participants, err := s.chatService.GetChatParticipants(context.Background(), chatID)
		require.NoError(t, err)
		require.Len(t, participants, 2)
		require.Contains(t, participants, s.testUser1)
		require.Contains(t, participants, s.testUser2)
	})
}

func (s *ChatServiceTestSuite) TestGetPrivateChat(t provider.T) {
	t.WithNewStep("Test get private chat", func(ctx provider.StepCtx) {
		// Create private chat between users
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypePrivate, "uauauau")
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser1)
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser2)
		require.NoError(t, err)

		chat, err := s.chatService.GetPrivateChat(context.Background(), s.testUser1, s.testUser2)
		require.NoError(t, err)
		require.Equal(t, chatID, chat.ID)
		require.Equal(t, models.ChatTypePrivate, chat.Type)
	})
}

func (s *ChatServiceTestSuite) TestGetNumUnreadChats(t provider.T) {
	t.WithNewStep("Test get number of unread chats", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)

		// Add user to chat
		err = s.addUserToChat(chatID, s.testUser1)
		require.NoError(t, err)

		// Create message from other user
		messageID := uuid.New()
		err = s.createTestMessage(messageID, chatID, s.testUser2, "Unread message")
		require.NoError(t, err)

		// Update chat timestamp to make it unread
		_, err = s.db.Exec(`UPDATE chat SET updated_at = NOW() WHERE id = $1`, chatID)
		require.NoError(t, err)

		numUnread, err := s.chatService.GetNumUnreadChats(context.Background(), s.testUser1)
		require.NoError(t, err)
		require.True(t, numUnread >= 1)
	})
}

func (s *ChatServiceTestSuite) TestJoinChat_AlreadyParticipant(t provider.T) {
	t.WithNewStep("Test join chat when already participant", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)
		err = s.addUserToChat(chatID, s.testUser1)
		require.NoError(t, err)

		err = s.chatService.JoinChat(context.Background(), chatID, s.testUser1)
		require.Error(t, err)
		require.Equal(t, messenger_errors.ErrAlreadyInChat, err)
	})
}

func (s *ChatServiceTestSuite) TestLeaveChat_NotParticipant(t provider.T) {
	t.WithNewStep("Test leave chat when not participant", func(ctx provider.StepCtx) {
		chatID := uuid.New()
		err := s.createTestChat(chatID, models.ChatTypeGroup, "Test Chat")
		require.NoError(t, err)

		err = s.chatService.LeaveChat(context.Background(), chatID, s.testUser1)
		require.Error(t, err)
		require.Equal(t, messenger_errors.ErrNotFound, err)
	})
}

// Helper methods for test data management
func (s *ChatServiceTestSuite) insertTestUsers() error {
	users := []struct {
		id        string
		username  string
		firstname string
		lastname  string
	}{
		{s.testUser1.String(), "testuser1", "Test", "User1"},
		{s.testUser2.String(), "testuser2", "Test", "User2"},
		{s.testUser3.String(), "testuser3", "Test", "User3"},
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
		`, user.id, user.firstname, user.lastname)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ChatServiceTestSuite) createTestChat(chatID uuid.UUID, chatType models.ChatType, name string) error {
	_, err := s.db.Exec(`
		INSERT INTO chat (id, type, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, chatID, chatType, name)
	if err != nil {
		return err
	}

	// Track the created chat
	s.testChats[chatID] = &testChatData{
		id:       chatID,
		chatType: chatType,
		name:     name,
	}
	return nil
}

func (s *ChatServiceTestSuite) addUserToChat(chatID, userID uuid.UUID) error {
	_, err := s.db.Exec(`
		INSERT INTO chat_user (chat_id, user_id)
		VALUES ($1, $2)
	`, chatID, userID)
	if err != nil {
		return err
	}

	// Track participant
	if chatData, exists := s.testChats[chatID]; exists {
		chatData.participants = append(chatData.participants, userID)
	}
	return nil
}

func (s *ChatServiceTestSuite) createTestMessage(messageID, chatID, senderID uuid.UUID, text string) error {
	_, err := s.db.Exec(`
		INSERT INTO message (id, chat_id, sender_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, messageID, chatID, senderID, text)
	if err != nil {
		return err
	}

	// Track message
	if chatData, exists := s.testChats[chatID]; exists {
		chatData.messages = append(chatData.messages, messageID)
	}
	return nil
}

func (s *ChatServiceTestSuite) cleanupTestChats() {
	for chatID := range s.testChats {
		s.cleanupChat(chatID)
	}
	s.testChats = make(map[uuid.UUID]*testChatData)
}

func (s *ChatServiceTestSuite) cleanupChat(chatID uuid.UUID) {
	// Cleanup messages first (due to foreign key constraints)
	_, _ = s.db.Exec(`DELETE FROM message_file WHERE message_id IN (SELECT id FROM message WHERE chat_id = $1)`, chatID)
	_, _ = s.db.Exec(`DELETE FROM message WHERE chat_id = $1`, chatID)

	// Cleanup participants
	_, _ = s.db.Exec(`DELETE FROM chat_user WHERE chat_id = $1`, chatID)

	// Cleanup chat
	_, _ = s.db.Exec(`DELETE FROM chat WHERE id = $1`, chatID)

	delete(s.testChats, chatID)
}

func (s *ChatServiceTestSuite) cleanupAllTestData() {
	// Cleanup all test data in correct order to respect foreign keys
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
			log.Printf("Error cleaning up test data: %v", err)
		}
	}

	s.testChats = make(map[uuid.UUID]*testChatData)
}

func TestChatService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(ChatServiceTestSuite))
}
