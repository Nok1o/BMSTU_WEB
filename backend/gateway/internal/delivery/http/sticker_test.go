//go:build unit
// +build unit

package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/mock"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/shared/models"
)

type MockStickerUseCase struct {
	mock.Mock
}

func (m *MockStickerUseCase) AddStickerPack(ctx context.Context, stickerPack *models.StickerPack) (*models.StickerPack, error) {
	args := m.Called(ctx, stickerPack)
	return args.Get(0).(*models.StickerPack), args.Error(1)
}

func (m *MockStickerUseCase) GetStickerPack(ctx context.Context, packId uuid.UUID) (*models.StickerPack, error) {
	args := m.Called(ctx, packId)
	return args.Get(0).(*models.StickerPack), args.Error(1)
}

func (m *MockStickerUseCase) GetStickerPackByName(ctx context.Context, packName string) (*models.StickerPack, error) {
	args := m.Called(ctx, packName)
	return args.Get(0).(*models.StickerPack), args.Error(1)
}

func (m *MockStickerUseCase) GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]*models.StickerPack, error) {
	args := m.Called(ctx, userId, count, offset)
	return args.Get(0).([]*models.StickerPack), args.Error(1)
}

func (m *MockStickerUseCase) DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error {
	args := m.Called(ctx, userId, packId)
	return args.Error(0)
}

func TestStickerHandler_AddStickerPack(t *testing.T) {
	usecase := new(MockStickerUseCase)
	handler := NewStickerHandler(usecase, bluemonday.NewPolicy())

	userID := uuid.New()
	packID := uuid.New()

	usecase.On("AddStickerPack", mock.Anything, mock.AnythingOfType("*models.StickerPack")).
		Return(&models.StickerPack{Id: packID, CreatorId: userID}, nil)

	form := forms.StickerPackForm{
		Name:     "test",
		Stickers: []string{"url1", "url2"},
	}
	body, _ := easyjson.Marshal(form)

	req := httptest.NewRequest("POST", "/api/sticker_packs", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user", models.User{Id: userID}))

	rec := httptest.NewRecorder()
	handler.AddStickerPack(rec, req)
}

func TestStickerHandler_GetStickerPack(t *testing.T) {
	usecase := new(MockStickerUseCase)
	handler := NewStickerHandler(usecase, bluemonday.NewPolicy())

	packID := uuid.New()

	usecase.On("GetStickerPack", mock.Anything, packID).
		Return(&models.StickerPack{Id: packID}, nil)

	req := httptest.NewRequest("GET", "/api/sticker_packs/"+packID.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"pack_id": packID.String()})

	rec := httptest.NewRecorder()
	handler.GetStickerPack(rec, req)
}

func TestStickerHandler_GetStickerPackByName(t *testing.T) {
	usecase := new(MockStickerUseCase)
	handler := NewStickerHandler(usecase, bluemonday.NewPolicy())

	packName := "test-pack"

	usecase.On("GetStickerPackByName", mock.Anything, packName).
		Return(&models.StickerPack{Name: packName}, nil)

	req := httptest.NewRequest("GET", "/api/sticker_packs/"+packName, nil)
	req = mux.SetURLVars(req, map[string]string{"pack_name": packName})

	rec := httptest.NewRecorder()
	handler.GetStickerPackByName(rec, req)
}

func TestStickerHandler_GetStickerPacks(t *testing.T) {
	usecase := new(MockStickerUseCase)
	handler := NewStickerHandler(usecase, bluemonday.NewPolicy())

	userID := uuid.New()

	usecase.On("GetStickerPacks", mock.Anything, userID, 10, 0).
		Return([]*models.StickerPack{{Id: uuid.New()}}, nil)

	req := httptest.NewRequest("GET", "/api/sticker_packs?count=10&offset=0", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", models.User{Id: userID}))

	rec := httptest.NewRecorder()
	handler.GetStickerPacks(rec, req)
}

func TestStickerHandler_DeleteStickerPack(t *testing.T) {
	usecase := new(MockStickerUseCase)
	handler := NewStickerHandler(usecase, bluemonday.NewPolicy())

	userID := uuid.New()
	packID := uuid.New()

	usecase.On("DeleteStickerPack", mock.Anything, userID, packID).
		Return(nil)

	req := httptest.NewRequest("DELETE", "/api/sticker_packs/"+packID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", models.User{Id: userID}))
	req = mux.SetURLVars(req, map[string]string{"pack_id": packID.String()})

	rec := httptest.NewRecorder()
	handler.DeleteStickerPack(rec, req)
}

func TestStickerHandler_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockStickerUseCase)
		request func() *http.Request
		handler func(http.ResponseWriter, *http.Request)
	}{
		{
			name:  "AddStickerPack invalid body",
			setup: func(m *MockStickerUseCase) {},
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/api/sticker_packs", bytes.NewReader([]byte("invalid")))
			},
			handler: func(usecase *MockStickerUseCase) func(http.ResponseWriter, *http.Request) {
				h := NewStickerHandler(usecase, bluemonday.NewPolicy())
				return h.AddStickerPack
			}(new(MockStickerUseCase)),
		},
		{
			name:  "GetStickerPack invalid UUID",
			setup: func(m *MockStickerUseCase) {},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/sticker_packs/invalid", nil)
				return mux.SetURLVars(req, map[string]string{"pack_id": "invalid"})
			},
			handler: func(usecase *MockStickerUseCase) func(http.ResponseWriter, *http.Request) {
				h := NewStickerHandler(usecase, bluemonday.NewPolicy())
				return h.GetStickerPack
			}(new(MockStickerUseCase)),
		},
		{
			name:  "GetStickerPackByName empty name",
			setup: func(m *MockStickerUseCase) {},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/sticker_packs/", nil)
				return mux.SetURLVars(req, map[string]string{"pack_name": ""})
			},
			handler: func(usecase *MockStickerUseCase) func(http.ResponseWriter, *http.Request) {
				h := NewStickerHandler(usecase, bluemonday.NewPolicy())
				return h.GetStickerPackByName
			}(new(MockStickerUseCase)),
		},
		{
			name:  "GetStickerPacks invalid count",
			setup: func(m *MockStickerUseCase) {},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/api/sticker_packs?count=invalid", nil)
			},
			handler: func(usecase *MockStickerUseCase) func(http.ResponseWriter, *http.Request) {
				h := NewStickerHandler(usecase, bluemonday.NewPolicy())
				return h.GetStickerPacks
			}(new(MockStickerUseCase)),
		},
		{
			name:  "DeleteStickerPack invalid UUID",
			setup: func(m *MockStickerUseCase) {},
			request: func() *http.Request {
				req := httptest.NewRequest("DELETE", "/api/sticker_packs/invalid", nil)
				return mux.SetURLVars(req, map[string]string{"pack_id": "invalid"})
			},
			handler: func(usecase *MockStickerUseCase) func(http.ResponseWriter, *http.Request) {
				h := NewStickerHandler(usecase, bluemonday.NewPolicy())
				return h.DeleteStickerPack
			}(new(MockStickerUseCase)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := new(MockStickerUseCase)
			tt.setup(usecase)

			req := tt.request()
			req = req.WithContext(context.WithValue(req.Context(), "user", models.User{Id: uuid.New()}))

			rec := httptest.NewRecorder()
			tt.handler(rec, req)
		})
	}
}
