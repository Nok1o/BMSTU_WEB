package inMemory

import (
	"context"

	"quickflow/file_service/internal/errors"
	"quickflow/shared/models"
)

type InMemoryFileStorage struct {
	files map[string]*models.File
}

func NewInMemoryFileStorage() *InMemoryFileStorage {
	return &InMemoryFileStorage{
		files: make(map[string]*models.File),
	}
}

func (fs *InMemoryFileStorage) UploadFile(_ context.Context, file *models.File) (string, error) {
	fs.files[file.Name] = file
	return "in-memory://" + file.Name, nil
}

func (fs *InMemoryFileStorage) UploadManyFiles(_ context.Context, files []*models.File) ([]string, error) {
	for _, file := range files {
		fs.files[file.Name] = file
	}
	urls := make([]string, len(files))
	for i, file := range files {
		urls[i] = "in-memory://" + file.Name
	}
	return urls, nil
}

func (fs *InMemoryFileStorage) GetFileURL(_ context.Context, filename string) (string, error) {
	if _, exists := fs.files[filename]; !exists {
		return "", errors.FileNotFound
	}

	return "in-memory://" + filename, nil
}

func (fs *InMemoryFileStorage) DeleteFile(_ context.Context, filename string) error {
	if _, exists := fs.files[filename]; !exists {
		return errors.FileNotFound
	}
	delete(fs.files, filename)
	return nil
}

func (fs *InMemoryFileStorage) Clear() {
	fs.files = make(map[string]*models.File)
}
