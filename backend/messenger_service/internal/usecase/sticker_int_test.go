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
	"quickflow/shared/interceptors"
	getEnv "quickflow/utils/get-env"
	service_discovery "quickflow/utils/service-discovery"
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

type StickerServiceTestSuite struct {
	suite.Suite
	db             *sql.DB
	stickerRepo    *postgres.StickerRepository
	stickerService usecase.StickerService
	fileService    usecase.FileService
	validator      usecase.StickerPackValidator
	testUser1      uuid.UUID
	testUser2      uuid.UUID
	grpcConns      []*grpc.ClientConn
}

func (s *StickerServiceTestSuite) BeforeAll(t provider.T) {
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

		// Setup gRPC connection for file service
		grpcConnFileService, err := service_discovery.NewGRPCClient(
			addr.DefaultFileServiceName,
			service_discovery.ModeFailover,
			interceptors.RequestIDClientInterceptor(),
		)
		if err != nil {
			log.Fatalf("failed to connect to file service: %v", err)
		}
		s.grpcConns = append(s.grpcConns, grpcConnFileService)

		// Initialize services
		s.validator = validation.NewStickerValidator()
		s.fileService = file_service.NewFileClient(grpcConnFileService)

		// Initialize repository
		s.stickerRepo = postgres.NewPostgresStickerRepository(s.db)

		// Generate test users
		s.testUser1 = uuid.New()
		s.testUser2 = uuid.New()

		// Insert test users and profiles
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		// Initialize sticker service
		s.stickerService = *usecase.NewStickerService(
			s.stickerRepo,
			s.fileService,
			s.validator,
		)
	})
}

func (s *StickerServiceTestSuite) AfterAll(t provider.T) {
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

func (s *StickerServiceTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupStickerData()
	})
	t.Epic("Integration")
}

func (s *StickerServiceTestSuite) TestAddStickerPack(t provider.T) {
	t.WithNewStep("Test add sticker pack", func(ctx provider.StepCtx) {
		stickerPack := &models.StickerPack{
			Id:        uuid.New(),
			Name:      "Test Pack",
			CreatorId: s.testUser1,
			Stickers: []*models.File{
				{
					URL:         "https://example.com/sticker1.png",
					DisplayType: models.DisplayTypeSticker,
					Name:        "sticker1.png",
				},
				{
					URL:         "https://example.com/sticker2.png",
					DisplayType: models.DisplayTypeSticker,
					Name:        "sticker2.png",
				},
			},
		}

		result, err := s.stickerService.AddStickerPack(context.Background(), stickerPack)
		require.NoError(t, err)
		require.Equal(t, stickerPack.Id, result.Id)
		require.Equal(t, "Test Pack", result.Name)
		require.Equal(t, s.testUser1, result.CreatorId)
		require.Len(t, result.Stickers, 2)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPack(t provider.T) {
	t.WithNewStep("Test get sticker pack", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
			"https://example.com/sticker2.png",
		})
		require.NoError(t, err)

		stickerPack, err := s.stickerService.GetStickerPack(context.Background(), packID)
		require.NoError(t, err)
		require.Equal(t, packID, stickerPack.Id)
		require.Equal(t, "Test Pack", stickerPack.Name)
		require.Equal(t, s.testUser1, stickerPack.CreatorId)
		require.Len(t, stickerPack.Stickers, 2)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPack_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent sticker pack", func(ctx provider.StepCtx) {
		nonExistentID := uuid.New()

		_, err := s.stickerService.GetStickerPack(context.Background(), nonExistentID)
		require.Error(t, err)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPacks(t provider.T) {
	t.WithNewStep("Test get sticker packs", func(ctx provider.StepCtx) {
		packID1 := uuid.New()
		packID2 := uuid.New()

		// Create test sticker packs
		err := s.createTestStickerPack(packID1, "Pack 1", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		err = s.createTestStickerPack(packID2, "Pack 2", s.testUser1, []string{
			"https://example.com/sticker2.png",
		})
		require.NoError(t, err)

		stickerPacks, err := s.stickerService.GetStickerPacks(context.Background(), s.testUser1, 10, 0)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 2)

		// Verify packs are returned in correct order (newest first)
		require.Equal(t, "Pack 2", stickerPacks[0].Name)
		require.Equal(t, "Pack 1", stickerPacks[1].Name)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPacks_Empty(t provider.T) {
	t.WithNewStep("Test get sticker packs when empty", func(ctx provider.StepCtx) {
		stickerPacks, err := s.stickerService.GetStickerPacks(context.Background(), s.testUser1, 10, 0)
		require.NoError(t, err)
		require.Empty(t, stickerPacks)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPacks_Pagination(t provider.T) {
	t.WithNewStep("Test get sticker packs with pagination", func(ctx provider.StepCtx) {
		packID1 := uuid.New()
		packID2 := uuid.New()
		packID3 := uuid.New()

		// Create test sticker packs
		err := s.createTestStickerPack(packID1, "Pack 1", s.testUser1, []string{"sticker1.png"})
		require.NoError(t, err)
		err = s.createTestStickerPack(packID2, "Pack 2", s.testUser1, []string{"sticker2.png"})
		require.NoError(t, err)
		err = s.createTestStickerPack(packID3, "Pack 3", s.testUser1, []string{"sticker3.png"})
		require.NoError(t, err)

		// Get first page (2 items)
		stickerPacks, err := s.stickerService.GetStickerPacks(context.Background(), s.testUser1, 2, 0)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 2)
		require.Equal(t, "Pack 3", stickerPacks[0].Name) // Newest first
		require.Equal(t, "Pack 2", stickerPacks[1].Name)

		// Get second page (1 item)
		stickerPacks, err = s.stickerService.GetStickerPacks(context.Background(), s.testUser1, 2, 2)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 1)
		require.Equal(t, "Pack 1", stickerPacks[0].Name)
	})
}

func (s *StickerServiceTestSuite) TestDeleteStickerPack(t provider.T) {
	t.WithNewStep("Test delete sticker pack", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Verify pack exists
		_, err = s.stickerService.GetStickerPack(context.Background(), packID)
		require.NoError(t, err)

		// Delete pack
		err = s.stickerService.DeleteStickerPack(context.Background(), s.testUser1, packID)
		require.NoError(t, err)

		// Verify pack was deleted
		_, err = s.stickerService.GetStickerPack(context.Background(), packID)
		require.Error(t, err)
	})
}

func (s *StickerServiceTestSuite) TestDeleteStickerPack_NotOwner(t provider.T) {
	t.WithNewStep("Test delete sticker pack when not owner", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack owned by user1
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Try to delete with user2 (not owner)
		err = s.stickerService.DeleteStickerPack(context.Background(), s.testUser2, packID)
		require.Error(t, err)
		require.Equal(t, messenger_errors.ErrNotOwnerOfStickerPack, err)
	})
}

func (s *StickerServiceTestSuite) TestGetStickerPackByName(t provider.T) {
	t.WithNewStep("Test get sticker pack by name", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		stickerPack, err := s.stickerService.GetStickerPackByName(context.Background(), "Test Pack")
		require.NoError(t, err)
		require.Equal(t, packID, stickerPack.Id)
		require.Equal(t, "Test Pack", stickerPack.Name)
		require.Equal(t, s.testUser1, stickerPack.CreatorId)
		require.Len(t, stickerPack.Stickers, 1)
	})
}

func (s *StickerServiceTestSuite) TestBelongsTo(t provider.T) {
	t.WithNewStep("Test check if sticker pack belongs to user", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack owned by user1
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Check if pack belongs to user1
		belongs, err := s.stickerService.BelongsTo(context.Background(), s.testUser1, packID)
		require.NoError(t, err)
		require.True(t, belongs)

		// Check if pack belongs to user2
		belongs, err = s.stickerService.BelongsTo(context.Background(), s.testUser2, packID)
		require.NoError(t, err)
		require.False(t, belongs)
	})
}

func (s *StickerServiceTestSuite) TestAddStickerPack_ValidationError(t provider.T) {
	t.WithNewStep("Test add sticker pack with validation error", func(ctx provider.StepCtx) {
		// Empty sticker pack without name
		stickerPack := &models.StickerPack{
			Id:        uuid.New(),
			Name:      "", // Empty name should cause validation error
			CreatorId: s.testUser1,
			Stickers: []*models.File{
				{
					URL:         "https://example.com/sticker1.png",
					DisplayType: "image",
					Name:        "sticker1.png",
				},
			},
		}

		_, err := s.stickerService.AddStickerPack(context.Background(), stickerPack)
		require.Error(t, err)
	})
}

// Helper methods
func (s *StickerServiceTestSuite) insertTestUsers() error {
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

func (s *StickerServiceTestSuite) createTestStickerPack(packID uuid.UUID, name string, creatorID uuid.UUID, stickerURLs []string) error {
	// Create sticker pack
	_, err := s.db.Exec(`
		INSERT INTO sticker_pack (id, name, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, packID, name, creatorID)
	if err != nil {
		return err
	}

	// Add stickers
	for _, url := range stickerURLs {
		_, err = s.db.Exec(`
			INSERT INTO sticker (sticker_pack_id, sticker_url)
			VALUES ($1, $2)
		`, packID, url)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StickerServiceTestSuite) cleanupStickerData() error {
	queries := []string{
		`DELETE FROM sticker`,
		`DELETE FROM sticker_pack`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StickerServiceTestSuite) cleanupTestData() error {
	queries := []string{
		`DELETE FROM sticker`,
		`DELETE FROM sticker_pack`,
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

func TestStickerServiceInt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(StickerServiceTestSuite))
}
