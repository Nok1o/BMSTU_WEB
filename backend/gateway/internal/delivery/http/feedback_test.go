//go:build unit
// +build unit

package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/require"

	"quickflow/gateway/internal/delivery/http/mocks"
	"quickflow/shared/models"
)

func TestFeedbackHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		setupRequest   func() *http.Request
		mockSetup      func(*mocks.MockFeedbackUseCase, *mocks.MockProfileUseCase)
		expectedStatus int
	}{
		// SaveFeedback tests
		{
			name:   "SaveFeedback success",
			method: http.MethodPost,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				body := bytes.NewBufferString(`{"type":"general","text":"test","rating":5}`)
				req := httptest.NewRequest(http.MethodPost, "/feedback", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(fu *mocks.MockFeedbackUseCase, pu *mocks.MockProfileUseCase) {
				fu.EXPECT().SaveFeedback(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "SaveFeedback no user in context",
			method: http.MethodPost,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				body := bytes.NewBufferString(`{"type":"general","text":"test","rating":5}`)
				req := httptest.NewRequest(http.MethodPost, "/feedback", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			mockSetup:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "SaveFeedback invalid json",
			method: http.MethodPost,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				body := bytes.NewBufferString(`invalid json`)
				req := httptest.NewRequest(http.MethodPost, "/feedback", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "SaveFeedback invalid type",
			method: http.MethodPost,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				body := bytes.NewBufferString(`{"type":"invalid","text":"test","rating":5}`)
				req := httptest.NewRequest(http.MethodPost, "/feedback", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "SaveFeedback usecase error",
			method: http.MethodPost,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				body := bytes.NewBufferString(`{"type":"general","text":"test","rating":5}`)
				req := httptest.NewRequest(http.MethodPost, "/feedback", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), "user", models.User{Id: uuid.New()})
				return req.WithContext(ctx)
			},
			mockSetup: func(fu *mocks.MockFeedbackUseCase, pu *mocks.MockProfileUseCase) {
				fu.EXPECT().SaveFeedback(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},

		// GetAllFeedbackType tests
		{
			name:   "GetAllFeedbackType success",
			method: http.MethodGet,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/feedback?feedback_count=10&type=general", nil)
				return req
			},
			mockSetup: func(fu *mocks.MockFeedbackUseCase, pu *mocks.MockProfileUseCase) {
				fu.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any(), gomock.Any(), 10).
					Return([]models.Feedback{
						{RespondentId: uuid.New()},
					}, nil)
				pu.EXPECT().GetPublicUserInfo(gomock.Any(), gomock.Any()).Return(models.PublicUserInfo{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "GetAllFeedbackType missing count param",
			method: http.MethodGet,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/feedback?type=general", nil)
				return req
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GetAllFeedbackType invalid count param",
			method: http.MethodGet,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/feedback?feedback_count=invalid&type=general", nil)
				return req
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GetAllFeedbackType usecase error",
			method: http.MethodGet,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/feedback?feedback_count=10&type=general", nil)
				return req
			},
			mockSetup: func(fu *mocks.MockFeedbackUseCase, pu *mocks.MockProfileUseCase) {
				fu.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any(), gomock.Any(), 10).
					Return(nil, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "GetAllFeedbackType profile info error",
			method: http.MethodGet,
			path:   "/feedback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/feedback?feedback_count=10&type=general", nil)
				return req
			},
			mockSetup: func(fu *mocks.MockFeedbackUseCase, pu *mocks.MockProfileUseCase) {
				fu.EXPECT().GetAllFeedbackType(gomock.Any(), gomock.Any(), gomock.Any(), 10).
					Return([]models.Feedback{
						{RespondentId: uuid.New()},
					}, nil)
				pu.EXPECT().GetPublicUserInfo(gomock.Any(), gomock.Any()).Return(models.PublicUserInfo{}, errors.New("profile error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFeedbackUC := mocks.NewMockFeedbackUseCase(ctrl)
			mockProfileUC := mocks.NewMockProfileUseCase(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockFeedbackUC, mockProfileUC)
			}

			handler := NewFeedbackHandler(
				mockFeedbackUC,
				mockProfileUC,
				bluemonday.UGCPolicy(),
			)

			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			switch tt.path {
			case "/feedback":
				if tt.method == http.MethodPost {
					handler.SaveFeedback(rr, req)
				} else {
					handler.GetAllFeedbackType(rr, req)
				}
			}

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
