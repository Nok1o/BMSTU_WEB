//go:build unit
// +build unit

package inMemory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"quickflow/file_service/internal/errors"
	"quickflow/file_service/internal/repository/inMemory"
	"quickflow/shared/models"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestInMemoryFileStorage_UploadFileAndGetFileURL(t *testing.T) {
	runner.Run(t, "UploadFile and GetFileURL Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("InMemoryFileStorage")

		tests := []struct {
			name    string
			file    *models.File
			query   string
			wantURL string
			wantErr error
		}{
			{
				name:    "success - upload and get file",
				file:    &models.File{Name: "test.txt"},
				query:   "test.txt",
				wantURL: "in-memory://test.txt",
				wantErr: nil,
			},
			{
				name:    "error - file not found",
				file:    &models.File{Name: "test.txt"},
				query:   "not_exist.txt",
				wantURL: "",
				wantErr: errors.FileNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание репозитория", func(s provider.StepCtx) {
					storage := inMemory.NewInMemoryFileStorage()

					s.WithNewStep("Загрузка файла", func(s provider.StepCtx) {
						_, _ = storage.UploadFile(context.Background(), tt.file)
					})

					t.WithNewStep("Получение URL файла", func(s provider.StepCtx) {
						gotURL, err := storage.GetFileURL(context.Background(), tt.query)

						if tt.wantErr != nil {
							assert.ErrorIs(t, err, tt.wantErr)
							assert.Equal(t, tt.wantURL, gotURL)
						} else {
							assert.NoError(t, err)
							assert.Equal(t, tt.wantURL, gotURL)
						}
					})
				})
			})
		}
	})
}

func TestInMemoryFileStorage_UploadManyFiles(t *testing.T) {
	runner.Run(t, "UploadManyFiles Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("InMemoryFileStorage")

		tests := []struct {
			name      string
			files     []*models.File
			wantURLs  []string
			queryName string
			wantErr   error
		}{
			{
				name: "success - multiple files",
				files: []*models.File{
					{Name: "a.txt"},
					{Name: "b.txt"},
				},
				wantURLs:  []string{"in-memory://a.txt", "in-memory://b.txt"},
				queryName: "a.txt",
				wantErr:   nil,
			},
			{
				name:      "empty list - no files uploaded",
				files:     []*models.File{},
				wantURLs:  []string{},
				queryName: "any.txt",
				wantErr:   errors.FileNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание репозитория", func(s provider.StepCtx) {
					storage := inMemory.NewInMemoryFileStorage()

					s.WithNewStep("Загрузка файлов", func(s provider.StepCtx) {
						gotURLs, err := storage.UploadManyFiles(context.Background(), tt.files)
						assert.NoError(t, err)
						assert.Equal(t, tt.wantURLs, gotURLs)
					})

					t.WithNewStep("Проверка URL файла", func(s provider.StepCtx) {
						gotURL, err := storage.GetFileURL(context.Background(), tt.queryName)
						if tt.wantErr != nil {
							assert.ErrorIs(t, err, tt.wantErr)
							assert.Equal(t, "", gotURL)
						} else {
							assert.NoError(t, err)
							assert.Contains(t, tt.wantURLs, gotURL)
						}
					})
				})
			})
		}
	})
}

func TestInMemoryFileStorage_DeleteFile(t *testing.T) {
	runner.Run(t, "DeleteFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("InMemoryFileStorage")

		tests := []struct {
			name       string
			prepare    func(storage *inMemory.InMemoryFileStorage)
			filename   string
			wantErr    error
			shouldKeep bool
		}{
			{
				name: "success - delete existing file",
				prepare: func(storage *inMemory.InMemoryFileStorage) {
					_, _ = storage.UploadFile(context.Background(), &models.File{Name: "to_delete.txt"})
				},
				filename:   "to_delete.txt",
				wantErr:    nil,
				shouldKeep: false,
			},
			{
				name:       "error - delete non-existing file",
				prepare:    func(_ *inMemory.InMemoryFileStorage) {},
				filename:   "not_exist.txt",
				wantErr:    errors.FileNotFound,
				shouldKeep: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Создание репозитория", func(s provider.StepCtx) {
					storage := inMemory.NewInMemoryFileStorage()
					tt.prepare(storage)

					s.WithNewStep("Удаление файла", func(s provider.StepCtx) {
						err := storage.DeleteFile(context.Background(), tt.filename)
						if tt.wantErr != nil {
							assert.ErrorIs(t, err, tt.wantErr)
						} else {
							assert.NoError(t, err)

							_, err := storage.GetFileURL(context.Background(), tt.filename)
							assert.ErrorIs(t, err, errors.FileNotFound)
						}
					})
				})
			})
		}
	})
}
