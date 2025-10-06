package inMemory

import (
	"context"

	"quickflow/file_service/internal/errors"
	"quickflow/shared/models"
)

type FileUrlsRepository struct {
	fileUrls map[string]string
}

func NewFileUrlsRepository() *FileUrlsRepository {
	return &FileUrlsRepository{
		fileUrls: make(map[string]string),
	}
}

func (f *FileUrlsRepository) AddFileRecord(_ context.Context, file *models.File) error {
	f.fileUrls[file.Name] = file.URL
	return nil
}

func (f *FileUrlsRepository) AddFilesRecords(_ context.Context, files []*models.File) error {
	for _, file := range files {
		f.fileUrls[file.Name] = file.URL
	}
	return nil
}

func (f *FileUrlsRepository) GetFileURL(_ context.Context, filename string) (string, error) {
	url, exists := f.fileUrls[filename]
	if !exists {
		return "", errors.FileNotFound
	}
	return url, nil
}

func (f *FileUrlsRepository) Clear() {
	f.fileUrls = make(map[string]string)
}
