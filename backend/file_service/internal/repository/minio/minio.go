package minio

import (
    "context"
    "fmt"
    `io`

    "github.com/google/uuid"
    "github.com/minio/minio-go/v7"
    "golang.org/x/sync/errgroup"

    minioconfig "quickflow/file_service/config/minio"
    `quickflow/file_service/internal/errors`
    threadsafeslice "quickflow/pkg/thread-safe-slice"
    "quickflow/shared/logger"
    "quickflow/shared/models"
)

type S3Repository interface {
    PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
    StatObject(ctx context.Context, bucket, object string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
    RemoveObject(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error
}

type MinioRepository struct {
    client                S3Repository
    PostsBucketName       string
    AttachmentsBucketName string
    ProfileBucketName     string
    StickerBuckerName     string
    PublicUrlRoot         string
}

func NewMinioRepository(cfg *minioconfig.MinioConfig, client S3Repository) (*MinioRepository, error) {
    return &MinioRepository{
        client:                client,
        PostsBucketName:       cfg.PostsBucketName,
        AttachmentsBucketName: cfg.AttachmentsBucketName,
        ProfileBucketName:     cfg.ProfileBucketName,
        StickerBuckerName:     cfg.StickerBuckerName,
        PublicUrlRoot:         fmt.Sprintf("%s://%s", cfg.Scheme, cfg.MinioPublicEndpoint),
    }, nil
}

// UploadFile uploads file to MinIO and returns a public URL.
func (m *MinioRepository) UploadFile(ctx context.Context, file *models.File) (string, error) {
    var err error
    uuID := uuid.New()
    fileName := uuID.String() + file.Ext

    var bucketName string
    if file.DisplayType == models.DisplayTypeSticker {
        bucketName = m.StickerBuckerName
    } else {
        bucketName = m.PostsBucketName
    }
    _, err = m.client.PutObject(ctx, bucketName, fileName, file.Reader, file.Size, minio.PutObjectOptions{
        ContentType: file.MimeType,
    })
    if err != nil {
        logger.Error(ctx, "could not upload file %v: %v", file.Name, err)
        return "", fmt.Errorf("could not upload file: %v", err)
    }

    publicURL := fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, bucketName, fileName)
    logger.Info(ctx, "File successfully loaded: %v, url: %v", file.Name, publicURL)
    return publicURL, nil
}

func (m *MinioRepository) UploadManyFiles(ctx context.Context, files []*models.File) ([]string, error) {
    urls := threadsafeslice.NewThreadSafeSliceN[string](len(files))

    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    wg, ctx := errgroup.WithContext(ctx)

    for i, file := range files {
        i := i
        file := file // https://golang.org/doc/faq#closures_and_goroutines
        uuID := uuid.New()
        fileName := uuID.String() + file.Ext

        wg.Go(func() error {
            var err error
            var bucketName string
            if file.DisplayType == models.DisplayTypeSticker {
                bucketName = m.StickerBuckerName
            } else {
                bucketName = m.PostsBucketName
            }
            _, err = m.client.PutObject(ctx, bucketName, fileName, file.Reader, file.Size, minio.PutObjectOptions{
                ContentType: file.MimeType,
            })
            if err != nil {
                logger.Error(ctx, "could not upload file %v: %v", file.Name, err)
                return fmt.Errorf("could not upload file: %v", err)
            }

            publicURL := fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, bucketName, fileName)
            err = urls.SetByIdx(i, publicURL)
            if err != nil {
                return fmt.Errorf("could not upload file: %v, err: %v", file.Name, err)
            }
            return nil
        })
    }

    if err := wg.Wait(); err != nil {
        return nil, err
    }
    return urls.GetSliceCopy(), nil
}

// GetFileURL returns a public URL for the file.
func (m *MinioRepository) GetFileURL(_ context.Context, fileName string) (string, error) {
    // Check if file exists
    _, err := m.client.StatObject(context.Background(), m.PostsBucketName, fileName, minio.StatObjectOptions{})
    if err != nil {
        return "", errors.FileNotFound
    }
    return fmt.Sprintf("%s/%s/%s", m.PublicUrlRoot, m.PostsBucketName, fileName), nil
}

// DeleteFile deletes a file from MinIO.
func (m *MinioRepository) DeleteFile(ctx context.Context, fileName string) error {
    err := m.client.RemoveObject(ctx, m.PostsBucketName, fileName, minio.RemoveObjectOptions{})
    if err != nil {
        return fmt.Errorf("could not delete file: %v", err)
    }
    return nil
}
