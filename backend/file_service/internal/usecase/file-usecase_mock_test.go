//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	file_errors "quickflow/file_service/internal/errors"
	"quickflow/file_service/internal/usecase"
	"quickflow/file_service/internal/usecase/mocks"
	"quickflow/shared/models"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
)

func TestFileUseCase_UploadFile(t *testing.T) {
	runner.Run(t, "UploadFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadFile")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileRepo := mocks.NewMockFileRepository(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, mockFileRepo, mockFileValidator)

		tests := []struct {
			name          string
			file          *models.File
			setupMocks    func(file *models.File)
			expectedURL   string
			expectedError error
		}{
			{
				name: "success",
				file: &models.File{Name: "test.txt", Size: 1024},
				setupMocks: func(file *models.File) {
					mockFileValidator.EXPECT().ValidateFile(file).Return(nil)
					mockFileStorage.EXPECT().UploadFile(gomock.Any(), file).Return("http://example.com/test.txt", nil)
					mockFileRepo.EXPECT().AddFileRecord(gomock.Any(), file).Return(nil)
				},
				expectedURL:   "http://example.com/test.txt",
				expectedError: nil,
			},
			{
				name:          "file is nil",
				file:          nil,
				setupMocks:    func(_ *models.File) {},
				expectedURL:   "",
				expectedError: file_errors.ErrFileIsNil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute UploadFile", func(s provider.StepCtx) {
					tt.setupMocks(tt.file)
					url, err := fileUseCase.UploadFile(context.Background(), tt.file)

					if tt.expectedError != nil {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err)
						assert.Equal(s, "", url)
					} else {
						assert.NoError(s, err)
						assert.Equal(s, tt.expectedURL, url)
					}
				})
			})
		}
	})
}

func TestFileUseCase_UploadManyMedia(t *testing.T) {
	runner.Run(t, "UploadManyMedia Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyMedia")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileRepo := mocks.NewMockFileRepository(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, mockFileRepo, mockFileValidator)

		tests := []struct {
			name          string
			files         []*models.File
			setupMocks    func(files []*models.File)
			expectedURLs  []string
			expectedError string
		}{
			{
				name:  "success",
				files: []*models.File{{Name: "img1.jpg"}, {Name: "img2.png"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(nil)
					mockFileStorage.EXPECT().UploadManyFiles(gomock.Any(), files).Return(
						[]string{"http://example.com/img1.jpg", "http://example.com/img2.png"}, nil)
					mockFileRepo.EXPECT().AddFilesRecords(gomock.Any(), files).Return(nil)
				},
				expectedURLs:  []string{"http://example.com/img1.jpg", "http://example.com/img2.png"},
				expectedError: "",
			},
			{
				name:  "validation error",
				files: []*models.File{{Name: "bad.exe"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(errors.New("invalid files"))
				},
				expectedURLs:  nil,
				expectedError: "validation.ValidateFiles: invalid files",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute UploadManyMedia", func(s provider.StepCtx) {
					tt.setupMocks(tt.files)
					urls, err := fileUseCase.UploadManyMedia(context.Background(), tt.files)

					if tt.expectedError != "" {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err.Error())
						assert.Nil(s, urls)
					} else {
						assert.NoError(s, err)
						assert.Equal(s, tt.expectedURLs, urls)
					}
				})
			})
		}
	})
}

func TestFileUseCase_UploadManyFiles(t *testing.T) {
	runner.Run(t, "UploadManyFiles Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyFiles")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileRepo := mocks.NewMockFileRepository(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, mockFileRepo, mockFileValidator)

		tests := []struct {
			name          string
			files         []*models.File
			setupMocks    func(files []*models.File)
			expectedURLs  []string
			expectedError string
		}{
			{
				name:  "success",
				files: []*models.File{{Name: "doc1.txt"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(nil)
					mockFileStorage.EXPECT().UploadManyFiles(gomock.Any(), files).Return([]string{"http://example.com/doc1.txt"}, nil)
				},
				expectedURLs:  []string{"http://example.com/doc1.txt"},
				expectedError: "",
			},
			{
				name:  "storage error",
				files: []*models.File{{Name: "doc1.txt"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(nil)
					mockFileStorage.EXPECT().UploadManyFiles(gomock.Any(), files).Return(nil, errors.New("storage error"))
				},
				expectedURLs:  nil,
				expectedError: "f.fileStorage.UploadManyFiles: storage error",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute UploadManyFiles", func(s provider.StepCtx) {
					tt.setupMocks(tt.files)
					urls, err := fileUseCase.UploadManyFiles(context.Background(), tt.files)

					if tt.expectedError != "" {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err.Error())
						assert.Nil(s, urls)
					} else {
						assert.NoError(s, err)
						assert.Equal(s, tt.expectedURLs, urls)
					}
				})
			})
		}
	})
}

func TestFileUseCase_UploadManyAudios(t *testing.T) {
	runner.Run(t, "UploadManyAudios Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("UploadManyAudios")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileRepo := mocks.NewMockFileRepository(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, mockFileRepo, mockFileValidator)

		tests := []struct {
			name          string
			files         []*models.File
			setupMocks    func(files []*models.File)
			expectedURLs  []string
			expectedError string
		}{
			{
				name:  "success",
				files: []*models.File{{Name: "song.mp3"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(nil)
					mockFileStorage.EXPECT().UploadManyFiles(gomock.Any(), files).Return([]string{"http://example.com/song.mp3"}, nil)
				},
				expectedURLs:  []string{"http://example.com/song.mp3"},
				expectedError: "",
			},
			{
				name:  "validation error",
				files: []*models.File{{Name: "wrong.wavv"}},
				setupMocks: func(files []*models.File) {
					mockFileValidator.EXPECT().ValidateFiles(files).Return(errors.New("invalid audio files"))
				},
				expectedURLs:  nil,
				expectedError: "validation.ValidateFiles: invalid audio files",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute UploadManyAudios", func(s provider.StepCtx) {
					tt.setupMocks(tt.files)
					urls, err := fileUseCase.UploadManyAudios(context.Background(), tt.files)

					if tt.expectedError != "" {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err.Error())
						assert.Nil(s, urls)
					} else {
						assert.NoError(s, err)
						assert.Equal(s, tt.expectedURLs, urls)
					}
				})
			})
		}
	})
}

func TestFileUseCase_GetFileURL(t *testing.T) {
	runner.Run(t, "GetFileURL Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetFileURL")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, nil, mockFileValidator)

		tests := []struct {
			name          string
			fileName      string
			setupMocks    func(fileName string)
			expectedURL   string
			expectedError string
		}{
			{
				name:     "success",
				fileName: "test.txt",
				setupMocks: func(fileName string) {
					mockFileValidator.EXPECT().ValidateFileName(fileName).Return(nil)
					mockFileStorage.EXPECT().GetFileURL(gomock.Any(), fileName).Return("http://example.com/test.txt", nil)
				},
				expectedURL:   "http://example.com/test.txt",
				expectedError: "",
			},
			{
				name:     "validation error",
				fileName: "test.txt",
				setupMocks: func(fileName string) {
					mockFileValidator.EXPECT().ValidateFileName(fileName).Return(errors.New("invalid filename"))
				},
				expectedURL:   "",
				expectedError: "validation.ValidateFileName: invalid filename",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute GetFileURL", func(s provider.StepCtx) {
					tt.setupMocks(tt.fileName)
					url, err := fileUseCase.GetFileURL(context.Background(), tt.fileName)

					if tt.expectedError != "" {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err.Error())
						assert.Equal(s, "", url)
					} else {
						assert.NoError(s, err)
						assert.Equal(s, tt.expectedURL, url)
					}
				})
			})
		}
	})
}

func TestFileUseCase_DeleteFile(t *testing.T) {
	runner.Run(t, "DeleteFile Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("DeleteFile")

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFileStorage := mocks.NewMockFileStorage(ctrl)
		mockFileValidator := mocks.NewMockFileValidator(ctrl)
		fileUseCase := usecase.NewFileUseCase(mockFileStorage, nil, mockFileValidator)

		tests := []struct {
			name          string
			fileName      string
			setupMocks    func(fileName string)
			expectedError string
		}{
			{
				name:     "success",
				fileName: "test.txt",
				setupMocks: func(fileName string) {
					mockFileValidator.EXPECT().ValidateFileName(fileName).Return(nil)
					mockFileStorage.EXPECT().DeleteFile(gomock.Any(), fileName).Return(nil)
				},
				expectedError: "",
			},
			{
				name:     "validation error",
				fileName: "test.txt",
				setupMocks: func(fileName string) {
					mockFileValidator.EXPECT().ValidateFileName(fileName).Return(errors.New("invalid filename"))
				},
				expectedError: "validation.ValidateFileName: invalid filename",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.WithNewStep("Execute DeleteFile", func(s provider.StepCtx) {
					tt.setupMocks(tt.fileName)
					err := fileUseCase.DeleteFile(context.Background(), tt.fileName)

					if tt.expectedError != "" {
						assert.Error(s, err)
						assert.Equal(s, tt.expectedError, err.Error())
					} else {
						assert.NoError(s, err)
					}
				})
			})
		}
	})
}
