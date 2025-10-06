//go:build unit
// +build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"quickflow/gateway/internal/delivery/http/forms"
	"quickflow/gateway/internal/delivery/http/mocks"
	mocks2 "quickflow/gateway/internal/delivery/ws/mocks"
	errors2 "quickflow/gateway/internal/errors"
	"quickflow/shared/models"
)

type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) GetMessagesForChat(ctx context.Context, chatID uuid.UUID, count int, ts time.Time, userID uuid.UUID) ([]*models.Message, error) {
	args := m.Called(ctx, chatID, count, ts, userID)
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageService) GetLastReadTs(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) (time.Time, error) {
	args := m.Called(ctx, chatID, userID)
	return args.Get(0).(time.Time), args.Error(1)
}

type MockProfileUseCase struct {
	mock.Mock
}

func (m *MockProfileUseCase) GetPublicUsersInfo(ctx context.Context, userIDs []uuid.UUID) ([]models.PublicUserInfo, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).([]models.PublicUserInfo), args.Error(1)
}

func TestMessageHandler_GetMessagesForChat(t *testing.T) {
	// Создаем тестовые данные
	chatID := uuid.New()
	userID := uuid.New()

	testUser := models.User{
		Id:       userID,
		Username: "testuser",
	}

	// Создаем моки
	tests := []struct {
		name              string
		chatID            string
		queryParams       url.Values
		setupMocks        func(*MockMessageService, *MockProfileUseCase)
		expectedStatus    int
		expectedErrorCode string
	}{
		{
			name:              "missing chat_id",
			chatID:            "",
			setupMocks:        func(ms *MockMessageService, pu *MockProfileUseCase) {},
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: errors2.BadRequestErrorCode,
		},
		{
			name:              "invalid chat_id format",
			chatID:            "invalid-uuid",
			setupMocks:        func(ms *MockMessageService, pu *MockProfileUseCase) {},
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: errors2.BadRequestErrorCode,
		},
		{
			name:              "missing messages_count parameter",
			chatID:            chatID.String(),
			queryParams:       url.Values{},
			setupMocks:        func(ms *MockMessageService, pu *MockProfileUseCase) {},
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: errors2.BadRequestErrorCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем моки
			mockMessageService := mocks2.NewMockMessageService(gomock.NewController(t))
			mockProfileUseCase := mocks.NewMockProfileUseCase(gomock.NewController(t))

			// Создаем хендлер с моками
			handler := &MessageHandler{
				messageUseCase: mockMessageService,
				profileUseCase: mockProfileUseCase,
			}

			// Создаем запрос
			req, err := http.NewRequest("GET", "/api/chats/"+tt.chatID+"/messages?"+tt.queryParams.Encode(), nil)
			assert.NoError(t, err)

			// Устанавливаем контекст с пользователем
			ctx := req.Context()
			ctx = context.WithValue(ctx, "user", testUser)
			req = req.WithContext(ctx)

			// Устанавливаем параметры маршрута
			vars := map[string]string{"chat_id": tt.chatID}
			req = mux.SetURLVars(req, vars)

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем хендлер
			handler.GetMessagesForChat(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedErrorCode != "" {
				// Проверяем тело ответа с ошибкой
				var errorForm forms.ErrorForm
				err = easyjson.Unmarshal(rr.Body.Bytes(), &errorForm)
				assert.NoError(t, err)
			} else {
				// Проверяем успешный ответ
				var messagesOut forms.MessagesOut
				err = easyjson.Unmarshal(rr.Body.Bytes(), &messagesOut)
				assert.NoError(t, err)
				assert.NotEmpty(t, messagesOut.Messages)
			}
		})
	}
}
