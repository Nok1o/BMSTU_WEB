//go:build unit
// +build unit

package usecase

import (
	"context"
	"testing"

	cfg "quickflow/file_service/config/validation"
	"quickflow/file_service/internal/repository/inMemory"
	"quickflow/file_service/utils/validation"
	"quickflow/shared/models"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
)

func createDefaultValidationConfig() *cfg.ValidationConfig {
	return &cfg.ValidationConfig{
		MaxFileCount:   2,
		MaxFileSize:    10 * 1024 * 1024, // 10 MB
		MaxPictureSize: 5 * 1024 * 1024,  // 5 MB
		MaxVideoSize:   50 * 1024 * 1024, // 50 MB
		MaxAudioSize:   15 * 1024 * 1024, // 15 MB

		AllowedFileExt:    []string{".pdf", ".docx", ".txt"},
		AllowedPictureExt: []string{".jpg", ".jpeg", ".png", ".gif"},
		AllowedVideoExt:   []string{".mp4", ".avi", ".mov"},
		AllowedAudioExt:   []string{".mp3", ".wav", ".aac"},
	}
}

type FileUseCaseSuite struct{}

func (s *FileUseCaseSuite) TestUploadFile(t *testing.T) {
	runner.Run(t, "UploadFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadFile")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		tests := []struct {
			name            string
			fileName        string
			fileSize        int64
			fileExt         string
			fileDisplayType models.DisplayType
			wantError       bool
		}{
			{"Valid file upload", "document.pdf", 2 * 1024 * 1024, ".pdf", models.DisplayTypeFile, false},
			{"File too large", "large_video.mp4", 100 * 1024 * 1024, ".mp4", models.DisplayTypeMedia, true},
			{"Unsupported file type", "script.exe", 1 * 1024 * 1024, ".exe", models.DisplayTypeFile, true},
			{"Empty file name", "", 1 * 1024 * 1024, ".txt", models.DisplayTypeFile, true},
			{"Zero file size", "empty.txt", 0, ".txt", models.DisplayTypeFile, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Upload file and verify repository", func(s provider.StepCtx) {
					file := &models.File{
						Name:        tt.fileName,
						Size:        tt.fileSize,
						Ext:         tt.fileExt,
						DisplayType: tt.fileDisplayType,
					}

					_, err := fileUseCase.UploadFile(nil, file)
					if (err != nil) != tt.wantError {
						s.Errorf("UploadFile() error = %v, wantError %v", err, tt.wantError)
					}

					if err == nil {
						gotURL, repoErr := inMemoryFileRepository.GetFileURL(context.Background(), tt.fileName)
						assert.NoError(s, repoErr)
						assert.NotEmpty(s, gotURL)
					}
				})
			})
		}
	})
}

func (s *FileUseCaseSuite) TestUploadManyMedia(t *testing.T) {
	runner.Run(t, "UploadManyMedia Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyMedia")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		validFiles := []*models.File{
			{Name: "img1.jpg", Size: 1000, Ext: ".jpg", DisplayType: models.DisplayTypeMedia},
			{Name: "img2.png", Size: 2000, Ext: ".png", DisplayType: models.DisplayTypeMedia},
		}

		invalidFiles := []*models.File{
			{Name: "bad.exe", Size: 1000, Ext: ".exe", DisplayType: models.DisplayTypeMedia},
		}

		tests := []struct {
			name      string
			files     []*models.File
			wantError bool
		}{
			{"Valid media upload", validFiles, false},
			{"Invalid media file", invalidFiles, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Upload media files", func(s provider.StepCtx) {
					_, err := fileUseCase.UploadManyMedia(context.Background(), tt.files)
					if (err != nil) != tt.wantError {
						s.Errorf("UploadManyMedia() error = %v, wantError %v", err, tt.wantError)
					}
				})
			})
		}
	})
}

func (s *FileUseCaseSuite) TestUploadManyFiles(t *testing.T) {
	runner.Run(t, "UploadManyFiles Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyFiles")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		validFiles := []*models.File{
			{Name: "doc1.txt", Size: 500, Ext: ".txt", DisplayType: models.DisplayTypeFile},
		}

		tooBigFile := []*models.File{
			{Name: "big.pdf", Size: 20 * 1024 * 1024, Ext: ".pdf", DisplayType: models.DisplayTypeFile},
		}

		tests := []struct {
			name      string
			files     []*models.File
			wantError bool
		}{
			{"Valid file batch", validFiles, false},
			{"File too large", tooBigFile, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Upload files batch", func(s provider.StepCtx) {
					_, err := fileUseCase.UploadManyFiles(context.Background(), tt.files)
					if (err != nil) != tt.wantError {
						s.Errorf("UploadManyFiles() error = %v, wantError %v", err, tt.wantError)
					}
				})
			})
		}
	})
}

func (s *FileUseCaseSuite) TestUploadManyAudios(t *testing.T) {
	runner.Run(t, "UploadManyAudios Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyAudios")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		validFiles := []*models.File{
			{Name: "track1.mp3", Size: 1000, Ext: ".mp3", DisplayType: models.DisplayTypeAudio},
		}

		invalidFiles := []*models.File{
			{Name: "wrong.wavv", Size: 1000, Ext: ".wavv", DisplayType: models.DisplayTypeMedia},
		}

		tests := []struct {
			name      string
			files     []*models.File
			wantError bool
		}{
			{"Valid audio upload", validFiles, false},
			{"Invalid audio ext", invalidFiles, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Upload audio files", func(s provider.StepCtx) {
					_, err := fileUseCase.UploadManyAudios(context.Background(), tt.files)
					if (err != nil) != tt.wantError {
						s.Errorf("UploadManyAudios() error = %v, wantError %v", err, tt.wantError)
					}
				})
			})
		}
	})
}

func (s *FileUseCaseSuite) TestGetFileURL(t *testing.T) {
	runner.Run(t, "GetFileURL Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetFileURL")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		file := &models.File{Name: "doc.pdf", Size: 1024, Ext: ".pdf"}
		inMemoryFileStorage.UploadFile(context.Background(), file)

		tests := []struct {
			name      string
			filename  string
			wantError bool
		}{
			{"Valid file URL", "doc.pdf", false},
			{"Invalid name", "notfound.pdf", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Get file URL", func(s provider.StepCtx) {
					_, err := fileUseCase.GetFileURL(context.Background(), tt.filename)
					if (err != nil) != tt.wantError {
						s.Errorf("GetFileURL() error = %v, wantError %v", err, tt.wantError)
					}
				})
			})
		}
	})
}

func (s *FileUseCaseSuite) TestDeleteFile(t *testing.T) {
	runner.Run(t, "DeleteFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("DeleteFile")

		inMemoryFileStorage := inMemory.NewInMemoryFileStorage()
		inMemoryFileRepository := inMemory.NewFileUrlsRepository()
		fileValidator := validation.NewFileValidator(createDefaultValidationConfig())
		fileUseCase := NewFileUseCase(inMemoryFileStorage, inMemoryFileRepository, fileValidator)

		t.Cleanup(func() {
			inMemoryFileStorage.Clear()
			inMemoryFileRepository.Clear()
		})

		file := &models.File{Name: "doc.pdf", Size: 1024, Ext: ".pdf"}
		inMemoryFileStorage.UploadFile(context.Background(), file)

		tests := []struct {
			name      string
			filename  string
			wantError bool
		}{
			{"Delete existing file", "doc.pdf", false},
			{"Delete non-existing file", "ghost.pdf", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Delete file", func(s provider.StepCtx) {
					err := fileUseCase.DeleteFile(context.Background(), tt.filename)
					if (err != nil) != tt.wantError {
						s.Errorf("DeleteFile() error = %v, wantError %v", err, tt.wantError)
					}
				})
			})
		}
	})
}

func TestFileUseCaseSuite(t *testing.T) {
	suite := FileUseCaseSuite{}

	t.Run("UploadFile", suite.TestUploadFile)
	t.Run("UploadManyMedia", suite.TestUploadManyMedia)
	t.Run("UploadManyFiles", suite.TestUploadManyFiles)
	t.Run("UploadManyAudios", suite.TestUploadManyAudios)
	t.Run("GetFileURL", suite.TestGetFileURL)
	t.Run("DeleteFile", suite.TestDeleteFile)
}
