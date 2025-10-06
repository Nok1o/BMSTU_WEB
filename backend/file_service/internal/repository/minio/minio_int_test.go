//go:build integration
// +build integration

package minio_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"

	minioconfig "quickflow/file_service/config/minio"
	minio2 "quickflow/file_service/internal/repository/minio"
	"quickflow/shared/models"
)

type MinioRepositoryTestSuite struct {
	suite.Suite
	client        *minio.Client
	fileStorage   *minio2.MinioRepository
	minioCfg      *minioconfig.MinioConfig
	testBucket    string
	testFileNames []string
}

func (s *MinioRepositoryTestSuite) BeforeAll(t provider.T) {
	t.WithNewStep("Setup MinIO client and repository", func(ctx provider.StepCtx) {
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

		var err error
		s.client, err = minio.New(s.minioCfg.MinioInternalEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s.minioCfg.MinioRootUser, s.minioCfg.MinioRootPassword, ""),
			Secure: s.minioCfg.MinioUseSSL,
		})
		if err != nil {
			t.Fatalf("could not create minio client: %v", err)
		}

		// Create test buckets
		buckets := []string{s.minioCfg.PostsBucketName, s.minioCfg.StickerBuckerName}
		for _, bucket := range buckets {
			exists, err := s.client.BucketExists(context.Background(), bucket)
			if err != nil {
				t.Fatalf("could not check if bucket %s exists: %v", bucket, err)
			}
			if !exists {
				err = s.client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
				if err != nil {
					t.Fatalf("could not create bucket %s: %v", bucket, err)
				}
			}
		}

		s.fileStorage, err = minio2.NewMinioRepository(s.minioCfg, s.client)
		if err != nil {
			t.Fatalf("failed to create minio repository: %v", err)
		}

		s.testBucket = s.minioCfg.PostsBucketName
		s.testFileNames = []string{}
	})
}

func (s *MinioRepositoryTestSuite) BeforeEach(t provider.T) {
	t.Epic("Integration")
}

func (s *MinioRepositoryTestSuite) AfterAll(t provider.T) {
	t.WithNewStep("Cleanup test data", func(ctx provider.StepCtx) {
		// Remove all test files
		for _, fileName := range s.testFileNames {
			err := s.client.RemoveObject(context.Background(), s.testBucket, fileName, minio.RemoveObjectOptions{})
			if err != nil {
				t.Logf("failed to remove test file %s: %v", fileName, err)
			}
		}
	})
}

func (s *MinioRepositoryTestSuite) TestUploadFile_Success(t provider.T) {
	testCases := []struct {
		name        string
		file        *models.File
		bucketType  string
		description string
	}{
		{
			name: "Upload regular file to posts bucket",
			file: &models.File{
				Name:        "test.txt",
				Ext:         ".txt",
				MimeType:    "text/plain",
				Size:        int64(len("test content")),
				Reader:      strings.NewReader("test content"),
				DisplayType: models.DisplayTypeFile,
			},
			bucketType:  "posts",
			description: "Should successfully upload text file to posts bucket",
		},
		{
			name: "Upload sticker file to stickers bucket",
			file: &models.File{
				Name:        "sticker.png",
				Ext:         ".png",
				MimeType:    "image/png",
				Size:        int64(len("sticker content")),
				Reader:      strings.NewReader("sticker content"),
				DisplayType: models.DisplayTypeSticker,
			},
			bucketType:  "stickers",
			description: "Should successfully upload sticker to stickers bucket",
		},
	}

	for _, tc := range testCases {
		t.WithNewStep(tc.name, func(ctx provider.StepCtx) {
			ctx.WithNewAttachment("test_case", allure.Text, []byte(fmt.Sprintf("%+v", tc)))

			// Execute
			url, err := s.fileStorage.UploadFile(context.Background(), tc.file)

			// Verify
			ctx.Require().NoError(err)
			ctx.Require().NotEmpty(url)
			ctx.Require().True(strings.Contains(url, tc.bucketType))

			// Verify file exists in MinIO
			fileName := strings.TrimSuffix(strings.Split(url, "/")[4], tc.file.Ext)
			_, err = s.client.StatObject(context.Background(), getBucketName(tc.bucketType, s.minioCfg), fileName+tc.file.Ext, minio.StatObjectOptions{})
			ctx.Require().NoError(err)

			s.testFileNames = append(s.testFileNames, fileName+tc.file.Ext)
		})
	}
}

func (s *MinioRepositoryTestSuite) TestUploadManyFiles_Success(t provider.T) {
	t.WithNewStep("Test uploading multiple files concurrently", func(ctx provider.StepCtx) {
		files := []*models.File{
			{
				Name:        "file1.txt",
				Ext:         ".txt",
				MimeType:    "text/plain",
				Size:        int64(len("content1")),
				Reader:      strings.NewReader("content1"),
				DisplayType: models.DisplayTypeFile,
			},
			{
				Name:        "file2.txt",
				Ext:         ".txt",
				MimeType:    "text/plain",
				Size:        int64(len("content2")),
				Reader:      strings.NewReader("content2"),
				DisplayType: models.DisplayTypeFile,
			},
			{
				Name:        "sticker1.png",
				Ext:         ".png",
				MimeType:    "image/png",
				Size:        int64(len("sticker1")),
				Reader:      strings.NewReader("sticker1"),
				DisplayType: models.DisplayTypeSticker,
			},
		}

		urls, err := s.fileStorage.UploadManyFiles(context.Background(), files)

		ctx.Require().NoError(err)
		ctx.Require().Len(urls, 3)

		// Verify all files were uploaded
		for i, url := range urls {
			ctx.Require().NotEmpty(url)

			fileName := strings.TrimSuffix(strings.Split(url, "/")[4], files[i].Ext)
			_, statErr := s.client.StatObject(context.Background(), getBucketNameFromURL(url, s.minioCfg), fileName+files[i].Ext, minio.StatObjectOptions{})
			ctx.Require().NoError(statErr)

			s.testFileNames = append(s.testFileNames, fileName+files[i].Ext)
		}
	})
}

func (s *MinioRepositoryTestSuite) TestGetFileURL_Success(t provider.T) {
	t.WithNewStep("Test getting file URL for existing file", func(ctx provider.StepCtx) {
		// Setup: Upload a test file first
		testContent := "test content for URL"
		testFile := &models.File{
			Name:        "url_test.txt",
			Ext:         ".txt",
			MimeType:    "text/plain",
			Size:        int64(len(testContent)),
			Reader:      strings.NewReader(testContent),
			DisplayType: models.DisplayTypeFile,
		}

		uploadURL, err := s.fileStorage.UploadFile(context.Background(), testFile)
		ctx.Require().NoError(err)

		fileName := strings.Split(uploadURL, "/")[4]
		s.testFileNames = append(s.testFileNames, fileName)

		// Execute
		retrievedURL, err := s.fileStorage.GetFileURL(context.Background(), fileName)

		// Verify
		ctx.Require().NoError(err)
		ctx.Require().Equal(uploadURL, retrievedURL)
	})
}

func (s *MinioRepositoryTestSuite) TestGetFileURL_NotFound(t provider.T) {
	t.WithNewStep("Test getting file URL for non-existent file", func(ctx provider.StepCtx) {
		nonExistentFile := "non-existent-file.txt"

		url, err := s.fileStorage.GetFileURL(context.Background(), nonExistentFile)

		ctx.Require().Error(err)
		ctx.Require().Empty(url)
	})
}

func (s *MinioRepositoryTestSuite) TestDeleteFile_Success(t provider.T) {
	t.WithNewStep("Test deleting existing file", func(ctx provider.StepCtx) {
		// Setup: Upload a test file first
		testContent := "test content for deletion"
		testFile := &models.File{
			Name:        "delete_test.txt",
			Ext:         ".txt",
			MimeType:    "text/plain",
			Size:        int64(len(testContent)),
			Reader:      strings.NewReader(testContent),
			DisplayType: models.DisplayTypeFile,
		}

		uploadURL, err := s.fileStorage.UploadFile(context.Background(), testFile)
		ctx.Require().NoError(err)

		fileName := strings.Split(uploadURL, "/")[4]

		// Verify file exists before deletion
		_, err = s.client.StatObject(context.Background(), s.testBucket, fileName, minio.StatObjectOptions{})
		ctx.Require().NoError(err)

		// Execute
		err = s.fileStorage.DeleteFile(context.Background(), fileName)

		// Verify
		ctx.Require().NoError(err)

		// Verify file no longer exists
		_, err = s.client.StatObject(context.Background(), s.testBucket, fileName, minio.StatObjectOptions{})
		ctx.Require().Error(err)
	})
}

func (s *MinioRepositoryTestSuite) TestDeleteFile_NonExistent(t provider.T) {
	t.WithNewStep("Test deleting non-existent file", func(ctx provider.StepCtx) {
		nonExistentFile := "non-existent-delete.txt"

		err := s.fileStorage.DeleteFile(context.Background(), nonExistentFile)

		// Note: Minio's RemoveObject doesn't return error for non-existent files
		ctx.Require().NoError(err)
	})
}

func (s *MinioRepositoryTestSuite) TestUploadFile_ContextCancellation(t provider.T) {
	t.WithNewStep("Test upload with cancelled context", func(ctx provider.StepCtx) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		testFile := &models.File{
			Name:        "cancelled_test.txt",
			Ext:         ".txt",
			MimeType:    "text/plain",
			Size:        int64(len("test content")),
			Reader:      strings.NewReader("test content"),
			DisplayType: models.DisplayTypeFile,
		}

		url, err := s.fileStorage.UploadFile(cancelledCtx, testFile)

		ctx.Require().Error(err)
		ctx.Require().Empty(url)
	})
}

// Helper functions
func getBucketName(bucketType string, cfg *minioconfig.MinioConfig) string {
	if bucketType == "stickers" {
		return cfg.StickerBuckerName
	}
	return cfg.PostsBucketName
}

func getBucketNameFromURL(url string, cfg *minioconfig.MinioConfig) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		bucketName := parts[3]
		if bucketName == cfg.StickerBuckerName {
			return cfg.StickerBuckerName
		}
	}
	return cfg.PostsBucketName
}

func TestMinioRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(MinioRepositoryTestSuite))
}
