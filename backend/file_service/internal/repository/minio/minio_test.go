//go:build unit
// +build unit

package minio_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	minioConfig "quickflow/file_service/config/minio"
	"quickflow/file_service/internal/repository"
	minio2 "quickflow/file_service/internal/repository/minio"
	"quickflow/shared/models"
)

type fakeMinioClient struct {
	putObjectFn    func(ctx context.Context, bucket, object string, data io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	statObjectFn   func(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	removeObjectFn func(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error
}

func (f *fakeMinioClient) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if f.putObjectFn != nil {
		return f.putObjectFn(ctx, bucketName, objectName, reader, objectSize, opts)
	}
	return minio.UploadInfo{}, nil
}

func (f *fakeMinioClient) StatObject(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	if f.statObjectFn != nil {
		return f.statObjectFn(ctx, bucket, object, opts)
	}
	return minio.ObjectInfo{}, nil
}

func (f *fakeMinioClient) RemoveObject(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error {
	if f.removeObjectFn != nil {
		return f.removeObjectFn(ctx, bucket, object, opts)
	}
	return nil
}

func TestMinioRepository_UploadFile(t *testing.T) {
	runner.Run(t, "UploadFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadFile")
		t.Severity(allure.CRITICAL)
		t.Description("Test uploading files to Minio with success and failure cases")

		tests := []struct {
			name       string
			clientMock *fakeMinioClient
			file       *models.File
			wantErr    bool
		}{
			{
				name: "Success Upload",
				clientMock: &fakeMinioClient{
					putObjectFn: func(ctx context.Context, bucket, object string, data io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
						return minio.UploadInfo{}, nil
					},
				},
				file: repository.NewFileBuilder().
					WithName("file1").
					WithExt(".png").
					WithSize(100).
					WithDisplayType(models.DisplayTypeMedia).
					Build(),
				wantErr: false,
			},
			{
				name: "Upload Error",
				clientMock: &fakeMinioClient{
					putObjectFn: func(ctx context.Context, bucket, object string, data io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
						return minio.UploadInfo{}, errors.New("put error")
					},
				},
				file: repository.NewFileBuilder().
					WithName("file2").
					WithExt(".jpg").
					WithSize(200).
					WithDisplayType(models.DisplayTypeMedia).
					Build(),
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				cfg := &minioConfig.MinioConfig{
					PostsBucketName:     "posts",
					MinioPublicEndpoint: "http://localhost",
				}
				repo, err := minio2.NewMinioRepository(cfg, tt.clientMock)
				assert.NoError(t, err)

				url, err := repo.UploadFile(context.Background(), tt.file)

				if tt.wantErr {
					stepCtx.WithNewStep("Verify error", func(stepCtx provider.StepCtx) {
						assert.Error(t, err)
						assert.Empty(t, url)
					})
				} else {
					stepCtx.WithNewStep("Verify success", func(stepCtx provider.StepCtx) {
						assert.NoError(t, err)
						assert.Contains(t, url, "http://localhost/posts/")
					})
				}
			})
		}
	})
}

func TestMinioRepository_GetFileURL(t *testing.T) {
	runner.Run(t, "GetFileURL Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetFileURL")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving file URLs from Minio with success and failure cases")

		tests := []struct {
			name       string
			clientMock *fakeMinioClient
			wantErr    bool
		}{
			{
				name: "File Exists",
				clientMock: &fakeMinioClient{
					statObjectFn: func(_ context.Context, _, _ string, _ minio.StatObjectOptions) (minio.ObjectInfo, error) {
						return minio.ObjectInfo{}, nil
					},
				},
				wantErr: false,
			},
			{
				name: "File Not Found",
				clientMock: &fakeMinioClient{
					statObjectFn: func(_ context.Context, _, _ string, _ minio.StatObjectOptions) (minio.ObjectInfo, error) {
						return minio.ObjectInfo{}, errors.New("not found")
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				cfg := &minioConfig.MinioConfig{
					PostsBucketName:     "posts",
					MinioPublicEndpoint: "http://localhost",
				}
				repo, err := minio2.NewMinioRepository(cfg, tt.clientMock)
				assert.NoError(t, err)

				url, err := repo.GetFileURL(context.Background(), "file.png")

				if tt.wantErr {
					stepCtx.WithNewStep("Verify error", func(stepCtx provider.StepCtx) {
						assert.Error(t, err)
						assert.Empty(t, url)
					})
				} else {
					stepCtx.WithNewStep("Verify success", func(stepCtx provider.StepCtx) {
						assert.NoError(t, err)
						assert.Contains(t, url, "http://localhost/posts/")
					})
				}
			})
		}
	})
}

func TestMinioRepository_DeleteFile(t *testing.T) {
	runner.Run(t, "DeleteFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("DeleteFile")
		t.Severity(allure.CRITICAL)
		t.Description("Test deleting files from Minio with success and failure cases")

		tests := []struct {
			name       string
			clientMock *fakeMinioClient
			wantErr    bool
		}{
			{
				name: "Delete Success",
				clientMock: &fakeMinioClient{
					removeObjectFn: func(_ context.Context, _, _ string, _ minio.RemoveObjectOptions) error { return nil },
				},
				wantErr: false,
			},
			{
				name: "Delete Error",
				clientMock: &fakeMinioClient{
					removeObjectFn: func(_ context.Context, _, _ string, _ minio.RemoveObjectOptions) error {
						return errors.New("delete failed")
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				cfg := &minioConfig.MinioConfig{
					PostsBucketName:     "posts",
					MinioPublicEndpoint: "http://localhost",
				}
				repo, err := minio2.NewMinioRepository(cfg, tt.clientMock)
				assert.NoError(t, err)

				err = repo.DeleteFile(context.Background(), "file.png")

				if tt.wantErr {
					stepCtx.WithNewStep("Verify error", func(stepCtx provider.StepCtx) {
						assert.Error(t, err)
					})
				} else {
					stepCtx.WithNewStep("Verify success", func(stepCtx provider.StepCtx) {
						assert.NoError(t, err)
					})
				}
			})
		}
	})
}
