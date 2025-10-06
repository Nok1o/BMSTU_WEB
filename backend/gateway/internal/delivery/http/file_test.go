//go:build unit
// +build unit

package http

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestFileHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		setupRequest   func() *http.Request
		mockSetup      func(*mocks.MockFileService)
		expectedStatus int
	}{
		{
			name:   "AddFiles success",
			method: http.MethodPost,
			path:   "/files",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add test files
				part, _ := writer.CreateFormFile("media", "test1.jpg")
				part.Write([]byte("test"))
				part, _ = writer.CreateFormFile("audio", "test2.mp3")
				part.Write([]byte("test"))
				part, _ = writer.CreateFormFile("stickers", "test3.png")
				part.Write([]byte("test"))
				part, _ = writer.CreateFormFile("files", "test4.txt")
				part.Write([]byte("test"))

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/files", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), "user", models.User{Username: "testuser"})
				return req.WithContext(ctx)
			},
			mockSetup: func(fs *mocks.MockFileService) {
				fs.EXPECT().UploadManyFiles(gomock.Any(), gomock.Any()).Return([]string{"url1"}, nil).Times(4)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "AddFiles no user in context",
			method: http.MethodPost,
			path:   "/files",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/files", nil)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "AddFiles too many files",
			method: http.MethodPost,
			path:   "/files",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add more than 10 files for one type
				for i := 0; i < 11; i++ {
					part, _ := writer.CreateFormFile("media", "test.jpg")
					part.Write([]byte("test"))
				}

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/files", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), "user", models.User{Username: "testuser"})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "AddFiles upload error",
			method: http.MethodPost,
			path:   "/files",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				part, _ := writer.CreateFormFile("media", "test.jpg")
				part.Write([]byte("test"))

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/files", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), "user", models.User{Username: "testuser"})
				return req.WithContext(ctx)
			},
			mockSetup: func(fs *mocks.MockFileService) {
				fs.EXPECT().UploadManyFiles(gomock.Any(), gomock.Any()).Return(nil, errors.New("upload error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "AddFiles invalid form",
			method: http.MethodPost,
			path:   "/files",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/files", nil)
				ctx := context.WithValue(req.Context(), "user", models.User{Username: "testuser"})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFileService := mocks.NewMockFileService(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockFileService)
			}

			handler := NewFileHandler(mockFileService, bluemonday.UGCPolicy())

			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			handler.AddFiles(rr, req)
		})
	}
}
