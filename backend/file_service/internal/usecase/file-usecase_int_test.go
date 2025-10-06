//go:build integration
// +build integration

package usecase_test

import (
	"context"
	"database/sql"
	minio2 "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/require"
	"quickflow/config/test"
	"quickflow/file_service/config/validation"
	validation2 "quickflow/file_service/utils/validation"
	getEnv "quickflow/utils/get-env"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	minioconfig "quickflow/file_service/config/minio"
	qf_errors "quickflow/file_service/internal/errors"
	"quickflow/file_service/internal/repository/minio"
	"quickflow/file_service/internal/repository/postgres"
	"quickflow/file_service/internal/usecase"
	"quickflow/shared/models"
)

type FileUseCaseIntegrationTestSuite struct {
	suite.Suite
	db            *sql.DB
	useCase       *usecase.FileUseCase
	minioClient   *minio2.Client
	minioCfg      *minioconfig.MinioConfig
	validationCfg *validation.ValidationConfig
	fileRepo      *postgres.PostgresFileRepository
}

func (s *FileUseCaseIntegrationTestSuite) BeforeAll(t provider.T) {
	t.WithNewStep("Setup integration test environment", func(ctx provider.StepCtx) {
		// Setup PostgreSQL
		connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
		require.NotEmpty(t, connString, "Connection string must not be empty")

		ctx.WithNewAttachment("connection_string", allure.Text, []byte(connString))

		var err error
		s.db, err = sql.Open("pgx", connString)
		ctx.Require().NoError(err, "Failed to connect to test database")

		err = s.db.Ping()
		ctx.Require().NoError(err, "Failed to ping database")

		// Create files table
		err = s.createFilesTable()
		ctx.Require().NoError(err, "Failed to create files table")

		// Setup MinIO config
		s.minioCfg = &minioconfig.MinioConfig{
			MinioInternalEndpoint: "minio:9000",
			MinioRootUser:         "admin",
			MinioRootPassword:     "adminpassword",
			MinioUseSSL:           false,
			PostsBucketName:       "posts",
			StickerBuckerName:     "stickers",
			MinioPublicEndpoint:   "localhost:9000",
			Scheme:                "http",
		}

		// Setup validation config
		s.validationCfg = &validation.ValidationConfig{
			MaxFileCount:      10,
			MaxPictureSize:    10 * 1024 * 1024,  // 10MB
			MaxVideoSize:      100 * 1024 * 1024, // 100MB
			MaxAudioSize:      50 * 1024 * 1024,  // 50MB
			MaxFileSize:       20 * 1024 * 1024,  // 20MB
			AllowedVideoExt:   []string{".mp4", ".avi", ".mov"},
			AllowedPictureExt: []string{".jpg", ".jpeg", ".png", ".gif"},
			AllowedFileExt:    []string{".txt", ".pdf", ".doc", ".docx"},
			AllowedAudioExt:   []string{".mp3", ".wav", ".ogg"},
		}

		s.minioClient, err = minio2.New(s.minioCfg.MinioInternalEndpoint, &minio2.Options{
			Creds:  credentials.NewStaticV4(s.minioCfg.MinioRootUser, s.minioCfg.MinioRootPassword, ""),
			Secure: s.minioCfg.MinioUseSSL,
		})
		if err != nil {
			t.Fatalf("could not create minio client: %v", err)
		}

		// Create test buckets
		buckets := []string{s.minioCfg.PostsBucketName, s.minioCfg.StickerBuckerName}
		for _, bucket := range buckets {
			exists, err := s.minioClient.BucketExists(context.Background(), bucket)
			if err != nil {
				t.Fatalf("could not check if bucket %s exists: %v", bucket, err)
			}
			if !exists {
				err = s.minioClient.MakeBucket(context.Background(), bucket, minio2.MakeBucketOptions{})
				if err != nil {
					t.Fatalf("could not create bucket %s: %v", bucket, err)
				}
			}
		}

		// Initialize repositories
		s.fileRepo = postgres.NewPostgresFileRepository(s.db)
		fileStorage, err := minio.NewMinioRepository(s.minioCfg, s.minioClient)
		ctx.Require().NoError(err, "Failed to create MinIO repository")

		fileValidator := validation2.NewFileValidator(s.validationCfg)

		// Initialize use case
		s.useCase = usecase.NewFileUseCase(fileStorage, s.fileRepo, fileValidator)
	})
}

func (s *FileUseCaseIntegrationTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test environment", func(ctx provider.StepCtx) {
		if s.db != nil {
			s.cleanupTestData()
			s.db.Close()
		}
	})
}

func (s *FileUseCaseIntegrationTestSuite) BeforeEach(t provider.T) {
	t.WithNewStep("Cleanup before each test", func(ctx provider.StepCtx) {
		s.cleanupTestData()
	})
	t.Epic("Integration")
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadFile_Success(t provider.T) {
	t.WithNewStep("Test successful file upload with database record", func(ctx provider.StepCtx) {
		testFile := &models.File{
			Name:     "test.txt",
			Ext:      ".txt",
			Size:     int64(len("test content")),
			Reader:   strings.NewReader("test content"),
			MimeType: "text/plain",
		}

		url, err := s.useCase.UploadFile(context.Background(), testFile)

		ctx.Require().NoError(err)
		ctx.Require().NotEmpty(url)
		ctx.Require().True(strings.Contains(url, s.minioCfg.PostsBucketName))

		// Verify file exists in MinIO
		fileName := extractFileNameFromURL(url)
		_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
		ctx.Require().NoError(err)

		// Verify record exists in database
		var count int
		err = s.db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM files WHERE file_url = $1", url).Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(1, count)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadFile_NilFile(t provider.T) {
	t.WithNewStep("Test upload nil file", func(ctx provider.StepCtx) {
		url, err := s.useCase.UploadFile(context.Background(), nil)

		ctx.Require().Error(err)
		ctx.Require().Empty(url)
		ctx.Require().Equal(qf_errors.ErrFileIsNil, err)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadFile_ValidationError(t provider.T) {
	t.WithNewStep("Test upload file with validation error", func(ctx provider.StepCtx) {
		// File with invalid extension
		testFile := &models.File{
			Name:     "test.exe",
			Ext:      ".exe",
			Size:     int64(len("test content")),
			Reader:   strings.NewReader("test content"),
			MimeType: "application/octet-stream",
		}

		url, err := s.useCase.UploadFile(context.Background(), testFile)

		ctx.Require().Error(err)
		ctx.Require().Empty(url)
		ctx.Require().Contains(err.Error(), "validation.ValidateFile")
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadManyMedia_Success(t provider.T) {
	t.WithNewStep("Test successful multiple media upload with database records", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name:        "image1.jpg",
				Ext:         ".jpg",
				Size:        int64(len("image content 1")),
				Reader:      strings.NewReader("image content 1"),
				MimeType:    "image/jpeg",
				DisplayType: models.DisplayTypeMedia,
			},
			{
				Name:        "image2.png",
				Ext:         ".png",
				Size:        int64(len("image content 2")),
				Reader:      strings.NewReader("image content 2"),
				MimeType:    "image/png",
				DisplayType: models.DisplayTypeMedia,
			},
		}

		urls, err := s.useCase.UploadManyMedia(context.Background(), files)

		ctx.Require().NoError(err)
		ctx.Require().Len(urls, 2)

		// Verify files exist in MinIO
		for _, url := range urls {
			fileName := extractFileNameFromURL(url)
			_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
			ctx.Require().NoError(err)
		}

		// Verify records exist in database
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(2, count)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadManyFiles_Success(t provider.T) {
	t.WithNewStep("Test successful multiple files upload without database records", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name:     "doc1.txt",
				Ext:      ".txt",
				Size:     int64(len("document content 1")),
				Reader:   strings.NewReader("document content 1"),
				MimeType: "text/plain",
			},
			{
				Name:     "doc2.pdf",
				Ext:      ".pdf",
				Size:     int64(len("document content 2")),
				Reader:   strings.NewReader("document content 2"),
				MimeType: "application/pdf",
			},
		}

		urls, err := s.useCase.UploadManyFiles(context.Background(), files)

		ctx.Require().NoError(err)
		ctx.Require().Len(urls, 2)

		// Verify files exist in MinIO
		for _, url := range urls {
			fileName := extractFileNameFromURL(url)
			_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
			ctx.Require().NoError(err)
		}

		// Verify no records in database (UploadManyFiles doesn't save to DB)
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(0, count)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestUploadManyAudios_Success(t provider.T) {
	t.WithNewStep("Test successful multiple audio files upload without database records", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name:        "audio1.mp3",
				Ext:         ".mp3",
				Size:        int64(len("audio content 1")),
				Reader:      strings.NewReader("audio content 1"),
				MimeType:    "audio/mpeg",
				DisplayType: models.DisplayTypeAudio,
			},
			{
				Name:        "audio2.wav",
				Ext:         ".wav",
				Size:        int64(len("audio content 2")),
				Reader:      strings.NewReader("audio content 2"),
				MimeType:    "audio/wav",
				DisplayType: models.DisplayTypeAudio,
			},
		}

		urls, err := s.useCase.UploadManyAudios(context.Background(), files)

		ctx.Require().NoError(err)
		ctx.Require().Len(urls, 2)

		// Verify files exist in MinIO
		for _, url := range urls {
			fileName := extractFileNameFromURL(url)
			_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
			ctx.Require().NoError(err)
		}

		// Verify no records in database
		var count int
		err = s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM files").Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(0, count)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestGetFileURL_Success(t provider.T) {
	t.WithNewStep("Test get file URL for existing file", func(ctx provider.StepCtx) {
		// First upload a file
		testFile := &models.File{
			Name:     "test-get.txt",
			Ext:      ".txt",
			Size:     int64(len("test content")),
			Reader:   strings.NewReader("test content"),
			MimeType: "text/plain",
		}

		uploadURL, err := s.useCase.UploadFile(context.Background(), testFile)
		ctx.Require().NoError(err)

		fileName := extractFileNameFromURL(uploadURL)

		// Then try to get its URL
		retrievedURL, err := s.useCase.GetFileURL(context.Background(), fileName)

		ctx.Require().NoError(err)
		ctx.Require().Equal(uploadURL, retrievedURL)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestGetFileURL_NotFound(t provider.T) {
	t.WithNewStep("Test get file URL for non-existent file", func(ctx provider.StepCtx) {
		nonExistentFile := "non-existent-file.txt"

		url, err := s.useCase.GetFileURL(context.Background(), nonExistentFile)

		ctx.Require().Error(err)
		ctx.Require().Empty(url)
		ctx.Require().Contains(err.Error(), "f.fileStorage.GetFileURL")
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestDeleteFile_Success(t provider.T) {
	t.WithNewStep("Test successful file deletion", func(ctx provider.StepCtx) {
		// First upload a file
		testFile := &models.File{
			Name:     "test-delete.txt",
			Ext:      ".txt",
			Size:     int64(len("test content")),
			Reader:   strings.NewReader("test content"),
			MimeType: "text/plain",
		}

		uploadURL, err := s.useCase.UploadFile(context.Background(), testFile)
		ctx.Require().NoError(err)

		fileName := extractFileNameFromURL(uploadURL)

		// Verify file exists before deletion
		_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
		ctx.Require().NoError(err)

		// Delete the file
		err = s.useCase.DeleteFile(context.Background(), fileName)

		ctx.Require().NoError(err)

		// Verify file no longer exists in MinIO
		_, err = s.minioClient.StatObject(context.Background(), s.minioCfg.PostsBucketName, fileName, minio2.StatObjectOptions{})
		ctx.Require().Error(err)

		// Verify record still exists in database (delete doesn't remove DB record)
		var count int
		err = s.db.QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM files WHERE file_url = $1", uploadURL).Scan(&count)
		ctx.Require().NoError(err)
		ctx.Require().Equal(1, count)
	})
}

func (s *FileUseCaseIntegrationTestSuite) TestDeleteFile_NonExistent(t provider.T) {
	t.WithNewStep("Test delete non-existent file", func(ctx provider.StepCtx) {
		nonExistentFile := "non-existent-delete.txt"

		err := s.useCase.DeleteFile(context.Background(), nonExistentFile)

		// MinIO doesn't return error for deleting non-existent files
		ctx.Require().NoError(err)
	})
}

// Helper methods
func (s *FileUseCaseIntegrationTestSuite) createFilesTable() error {
	return nil
}

func (s *FileUseCaseIntegrationTestSuite) dropFilesTable() error {
	return nil
}

func (s *FileUseCaseIntegrationTestSuite) cleanupTestData() error {
	// Clean database
	_, err := s.db.Exec("DELETE FROM files")
	if err != nil {
		return err
	}

	// Clean MinIO objects from test bucket
	objectsCh := make(chan minio2.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range s.minioClient.ListObjects(context.Background(), s.minioCfg.PostsBucketName, minio2.ListObjectsOptions{}) {
			if object.Err == nil {
				objectsCh <- object
			}
		}
	}()

	for object := range objectsCh {
		err := s.minioClient.RemoveObject(context.Background(), s.minioCfg.PostsBucketName, object.Key, minio2.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func TestFileUseCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(FileUseCaseIntegrationTestSuite))
}
