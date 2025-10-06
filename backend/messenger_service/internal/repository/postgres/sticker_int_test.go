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

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"quickflow/shared/models"
)

type PostgresStickerRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.StickerRepository
	testUser1  uuid.UUID
	testUser2  uuid.UUID
}

func (s *PostgresStickerRepositoryTestSuite) BeforeAll(t provider.T) {
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

		// Insert test users and profiles
		err = s.insertTestUsers()
		if err != nil {
			log.Fatalf("Failed to insert test users: %v", err)
		}

		s.repository = postgres.NewPostgresStickerRepository(s.db)
	})
}

func (s *PostgresStickerRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *PostgresStickerRepositoryTestSuite) BeforeEach(t provider.T) {
	t.Epic("Integration")
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupStickerData()
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestAddStickerPack(t provider.T) {
	t.WithNewStep("Test add sticker pack", func(ctx provider.StepCtx) {
		packID := uuid.New()
		stickerPack := models.StickerPack{
			Id:        packID,
			Name:      "Test Pack",
			CreatorId: s.testUser1,
			Stickers: []*models.File{
				{
					URL:         "https://example.com/sticker1.png",
					DisplayType: "image",
					Name:        "sticker1.png",
				},
				{
					URL:         "https://example.com/sticker2.png",
					DisplayType: "image",
					Name:        "sticker2.png",
				},
			},
		}

		err := s.repository.AddStickerPack(context.Background(), stickerPack)
		require.NoError(t, err)

		// Verify sticker pack was created
		retrievedPack, err := s.repository.GetStickerPack(context.Background(), packID)
		require.NoError(t, err)
		require.Equal(t, packID, retrievedPack.Id)
		require.Equal(t, "Test Pack", retrievedPack.Name)
		require.Equal(t, s.testUser1, retrievedPack.CreatorId)
		require.Len(t, retrievedPack.Stickers, 2)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPack(t provider.T) {
	t.WithNewStep("Test get sticker pack", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
			"https://example.com/sticker2.png",
		})
		require.NoError(t, err)

		stickerPack, err := s.repository.GetStickerPack(context.Background(), packID)
		require.NoError(t, err)
		require.Equal(t, packID, stickerPack.Id)
		require.Equal(t, "Test Pack", stickerPack.Name)
		require.Equal(t, s.testUser1, stickerPack.CreatorId)
		require.Len(t, stickerPack.Stickers, 2)
		require.Equal(t, "https://example.com/sticker1.png", stickerPack.Stickers[0].URL)
		require.Equal(t, "https://example.com/sticker2.png", stickerPack.Stickers[1].URL)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPack_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent sticker pack", func(ctx provider.StepCtx) {
		nonExistentID := uuid.New()

		_, err := s.repository.GetStickerPack(context.Background(), nonExistentID)
		require.Error(t, err)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPackByName(t provider.T) {
	t.WithNewStep("Test get sticker pack by name", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		stickerPack, err := s.repository.GetStickerPackByName(context.Background(), "Test Pack")
		require.NoError(t, err)
		require.Equal(t, packID, stickerPack.Id)
		require.Equal(t, "Test Pack", stickerPack.Name)
		require.Equal(t, s.testUser1, stickerPack.CreatorId)
		require.Len(t, stickerPack.Stickers, 1)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPackByName_NotExists(t provider.T) {
	t.WithNewStep("Test get non-existent sticker pack by name", func(ctx provider.StepCtx) {
		_, err := s.repository.GetStickerPackByName(context.Background(), "NonExistentPack")
		require.Error(t, err)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPacks(t provider.T) {
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

		stickerPacks, err := s.repository.GetStickerPacks(context.Background(), s.testUser1, 10, 0)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 2)

		// Verify packs are returned in correct order (newest first)
		require.Equal(t, "Pack 2", stickerPacks[0].Name)
		require.Equal(t, "Pack 1", stickerPacks[1].Name)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPacks_Empty(t provider.T) {
	t.WithNewStep("Test get sticker packs when empty", func(ctx provider.StepCtx) {
		stickerPacks, err := s.repository.GetStickerPacks(context.Background(), s.testUser1, 10, 0)
		require.NoError(t, err)
		require.Empty(t, stickerPacks)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestGetStickerPacks_Pagination(t provider.T) {
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
		stickerPacks, err := s.repository.GetStickerPacks(context.Background(), s.testUser1, 2, 0)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 2)
		require.Equal(t, "Pack 3", stickerPacks[0].Name) // Newest first
		require.Equal(t, "Pack 2", stickerPacks[1].Name)

		// Get second page (1 item)
		stickerPacks, err = s.repository.GetStickerPacks(context.Background(), s.testUser1, 2, 2)
		require.NoError(t, err)
		require.Len(t, stickerPacks, 1)
		require.Equal(t, "Pack 1", stickerPacks[0].Name)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestDeleteStickerPack(t provider.T) {
	t.WithNewStep("Test delete sticker pack", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Verify pack exists
		_, err = s.repository.GetStickerPack(context.Background(), packID)
		require.NoError(t, err)

		// Delete pack
		err = s.repository.DeleteStickerPack(context.Background(), s.testUser1, packID)
		require.NoError(t, err)

		// Verify pack was deleted
		_, err = s.repository.GetStickerPack(context.Background(), packID)
		require.Error(t, err)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestDeleteStickerPack_NotOwner(t provider.T) {
	t.WithNewStep("Test delete sticker pack when not owner", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack owned by user1
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Try to delete with user2 (not owner)
		err = s.repository.DeleteStickerPack(context.Background(), s.testUser2, packID)
		require.Error(t, err)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestBelongsTo(t provider.T) {
	t.WithNewStep("Test check if sticker pack belongs to user", func(ctx provider.StepCtx) {
		packID := uuid.New()

		// Create test sticker pack owned by user1
		err := s.createTestStickerPack(packID, "Test Pack", s.testUser1, []string{
			"https://example.com/sticker1.png",
		})
		require.NoError(t, err)

		// Check if pack belongs to user1
		belongs, err := s.repository.BelongsTo(context.Background(), s.testUser1, packID)
		require.NoError(t, err)
		require.True(t, belongs)

		// Check if pack belongs to user2
		belongs, err = s.repository.BelongsTo(context.Background(), s.testUser2, packID)
		require.NoError(t, err)
		require.False(t, belongs)
	})
}

func (s *PostgresStickerRepositoryTestSuite) TestBelongsTo_NotExists(t provider.T) {
	t.WithNewStep("Test check if non-existent sticker pack belongs to user", func(ctx provider.StepCtx) {
		nonExistentID := uuid.New()

		_, err := s.repository.BelongsTo(context.Background(), s.testUser1, nonExistentID)
		require.Error(t, err)
	})
}

// Helper methods
func (s *PostgresStickerRepositoryTestSuite) insertTestUsers() error {
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

func (s *PostgresStickerRepositoryTestSuite) createTestStickerPack(packID uuid.UUID, name string, creatorID uuid.UUID, stickerURLs []string) error {
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

func (s *PostgresStickerRepositoryTestSuite) cleanupStickerData() error {
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

func (s *PostgresStickerRepositoryTestSuite) cleanupTestData() error {
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

func TestPostgresStickerRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresStickerRepositoryTestSuite))
}
