//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ozontech/allure-go/pkg/framework/asserts_wrapper/require"
	"quickflow/config/test"
	getEnv "quickflow/utils/get-env"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	"quickflow/file_service/internal/repository/postgres"
	"quickflow/shared/models"
)

type PostgresFileRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repository *postgres.PostgresFileRepository
}

func (s *PostgresFileRepositoryTestSuite) BeforeAll(t provider.T) {
	t.WithNewStep("Setup PostgreSQL database connection", func(ctx provider.StepCtx) {
		connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
		require.NotEmpty(t, connString, "Connection string must not be empty")

		var err error
		s.db, err = sql.Open("pgx", connString)
		require.NoError(t, err, "Failed to connect to test database")

		ctx.WithNewAttachment("connection_string", allure.Text, []byte(connString))

		// Verify connection
		err = s.db.Ping()
		ctx.Require().NoError(err, "Failed to ping database")

		// Create test table
		err = s.createTestTable()
		ctx.Require().NoError(err, "Failed to create test table")

		s.repository = postgres.NewPostgresFileRepository(s.db)
	})
}

func (s *PostgresFileRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup database resources", func(ctx provider.StepCtx) {
		if s.db != nil {
			err := s.cleanupTestData()
			if err != nil {
				t.Logf("Failed to cleanup test data: %v", err)
			}
			s.db.Close()
		}
	})
}

func (s *PostgresFileRepositoryTestSuite) BeforeEach(t provider.T) {
	t.Epic("Integration")
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		err := s.cleanupTestData()
		if err != nil {
			t.Logf("Failed to cleanup before test: %v", err)
		}
	})
}

func (s *PostgresFileRepositoryTestSuite) TestAddFileRecord_Success(t provider.T) {
	testCases := []struct {
		name        string
		file        *models.File
		description string
	}{
		{
			name: "Add valid file record",
			file: &models.File{
				Name: "test.txt",
				URL:  "http://example.com/files/test.txt",
			},
			description: "Should successfully add file record to database",
		},
		{
			name: "Add file with long URL",
			file: &models.File{
				Name: "image.png",
				URL:  "http://example.com/files/very/long/path/to/image.png?query=param&another=param",
			},
			description: "Should successfully add file record with long URL",
		},
		{
			name: "Add file with special characters in name",
			file: &models.File{
				Name: "file-with-special-chars_123@test.txt",
				URL:  "http://example.com/files/special.txt",
			},
			description: "Should successfully add file record with special characters in filename",
		},
	}

	for _, tc := range testCases {
		t.WithNewStep(tc.name, func(ctx provider.StepCtx) {
			ctx.WithNewAttachment("test_case", allure.Text, []byte(fmt.Sprintf("%+v", tc)))

			// Execute
			err := s.repository.AddFileRecord(context.Background(), tc.file)

			// Verify
			ctx.Require().NoError(err)

			// Verify record was inserted
			var count int
			err = s.db.QueryRowContext(context.Background(),
				"SELECT COUNT(*) FROM files WHERE file_url = $1 AND filename = $2",
				tc.file.URL, tc.file.Name).Scan(&count)
			ctx.Require().NoError(err)
			ctx.Require().Equal(1, count, "Expected exactly one record to be inserted")
		})
	}
}

func (s *PostgresFileRepositoryTestSuite) TestAddFileRecord_NilFile(t provider.T) {
	t.WithNewStep("Test adding nil file record", func(ctx provider.StepCtx) {
		// Execute
		err := s.repository.AddFileRecord(context.Background(), nil)

		// Verify
		ctx.Require().Error(err)
		ctx.Require().Contains(err.Error(), "file cannot be nil")
	})
}

func (s *PostgresFileRepositoryTestSuite) TestAddFilesRecords_Success(t provider.T) {
	t.WithNewStep("Test adding multiple file records in transaction", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name: "file1.txt",
				URL:  "http://example.com/files/1.txt",
			},
			{
				Name: "file2.txt",
				URL:  "http://example.com/files/2.txt",
			},
			{
				Name: "file3.png",
				URL:  "http://example.com/files/3.png",
			},
		}

		// Execute
		err := s.repository.AddFilesRecords(context.Background(), files)

		// Verify
		ctx.Require().NoError(err)

		// Verify all records were inserted
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(len(files), count, "Expected all files to be inserted")

		// Verify each individual record
		for _, file := range files {
			var exists bool
			err = s.db.QueryRowContext(context.Background(),
				"SELECT EXISTS(SELECT 1 FROM files WHERE file_url = $1 AND filename = $2)",
				file.URL, file.Name).Scan(&exists)
			ctx.Require().NoError(err)
			ctx.Require().True(exists, "File %s should exist in database", file.Name)
		}
	})
}

func (s *PostgresFileRepositoryTestSuite) TestAddFilesRecords_EmptySlice(t provider.T) {
	t.WithNewStep("Test adding empty files slice", func(ctx provider.StepCtx) {
		// Execute
		err := s.repository.AddFilesRecords(context.Background(), []*models.File{})

		// Verify
		ctx.Require().NoError(err)

		// Verify no records were inserted
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(0, count, "Expected no files to be inserted")
	})
}

func (s *PostgresFileRepositoryTestSuite) TestAddFilesRecords_NilFileInSlice(t provider.T) {
	t.WithNewStep("Test adding slice with nil file", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name: "valid.txt",
				URL:  "http://example.com/valid.txt",
			},
			nil, // This should cause error
			{
				Name: "another.txt",
				URL:  "http://example.com/another.txt",
			},
		}

		// Execute
		err := s.repository.AddFilesRecords(context.Background(), files)

		// Verify
		ctx.Require().Error(err)
		ctx.Require().Contains(err.Error(), "file cannot be nil")

		// Verify transaction was rolled back (no records inserted)
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(0, count, "Expected transaction rollback - no files should be inserted")
	})
}

func (s *PostgresFileRepositoryTestSuite) TestAddFilesRecords_ContextCancellation(t provider.T) {
	t.WithNewStep("Test adding files with cancelled context", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name: "cancelled.txt",
				URL:  "http://example.com/cancelled.txt",
			},
		}

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		// Execute
		err := s.repository.AddFilesRecords(cancelledCtx, files)

		// Verify
		ctx.Require().Error(err)
		ctx.Require().Contains(err.Error(), "context")
	})
}

// Helper methods
func (s *PostgresFileRepositoryTestSuite) createTestTable() error {
	// Drop existing table if exists

	return nil
}

func (s *PostgresFileRepositoryTestSuite) cleanupTestData() error {
	_, err := s.db.Exec("DELETE FROM files")
	return err
}

func TestPostgresFileRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(PostgresFileRepositoryTestSuite))
}
