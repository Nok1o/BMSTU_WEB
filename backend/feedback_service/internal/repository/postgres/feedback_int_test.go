//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"quickflow/config/test"
	getEnv "quickflow/utils/get-env"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickflow/shared/models"
)

// Генерация случайных данных
func generateRandomText(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), rand.Intn(10000))
}

func generateRandomUsername() string {
	return fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(10000))
}

type FeedbackBuilder struct {
	feedback models.Feedback
}

func NewFeedbackBuilder() *FeedbackBuilder {
	return &FeedbackBuilder{
		feedback: models.Feedback{
			Id:           uuid.New(),
			Rating:       rand.Intn(11), // Случайный рейтинг от 0 до 10
			RespondentId: uuid.New(),
			Text:         generateRandomText("feedback"),
			Type:         models.FeedbackGeneral,
			CreatedAt:    time.Now(),
		},
	}
}

func (b *FeedbackBuilder) WithID(id uuid.UUID) *FeedbackBuilder {
	b.feedback.Id = id
	return b
}

func (b *FeedbackBuilder) WithRating(rating int) *FeedbackBuilder {
	b.feedback.Rating = rating
	return b
}

func (b *FeedbackBuilder) WithRespondentID(respondentID uuid.UUID) *FeedbackBuilder {
	b.feedback.RespondentId = respondentID
	return b
}

func (b *FeedbackBuilder) WithText(text string) *FeedbackBuilder {
	b.feedback.Text = text
	return b
}

func (b *FeedbackBuilder) WithType(feedbackType models.FeedbackType) *FeedbackBuilder {
	b.feedback.Type = feedbackType
	return b
}

func (b *FeedbackBuilder) WithCreatedAt(createdAt time.Time) *FeedbackBuilder {
	b.feedback.CreatedAt = createdAt
	return b
}

func (b *FeedbackBuilder) Build() models.Feedback {
	return b.feedback
}

type FeedbackRepositoryIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	repository  *FeedbackRepository
	testData    TestData
	cleanupData CleanupData
}

type TestData struct {
	userID      uuid.UUID
	feedbackIDs []uuid.UUID
	sessionID   string
}

type CleanupData struct {
	userIDs     []uuid.UUID
	feedbackIDs []uuid.UUID
}

func TestFeedbackRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())
	suite.RunSuite(t, new(FeedbackRepositoryIntegrationSuite))
}

func (s *FeedbackRepositoryIntegrationSuite) BeforeAll(t provider.T) {
	t.Feature("Feedback Repository Integration")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов репозитория feedback")

	// Инициализация БД
	connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
	require.NotEmpty(t, connString, "Connection string must not be empty")

	var err error
	s.db, err = sql.Open("pgx", connString)
	require.NoError(t, err, "Failed to connect to test database")

	err = s.db.Ping()
	require.NoError(t, err, "Failed to ping database")

	// Инициализация репозитория
	s.repository = NewFeedbackRepository(s.db)
}

func (s *FeedbackRepositoryIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	if s.db != nil {
		s.db.Close()
	}
}

func (s *FeedbackRepositoryIntegrationSuite) BeforeEach(t provider.T) {
	// Очистка данных предыдущего теста
	s.cleanupTestData(t)
	// Создание новых случайных тестовых данных
	s.testData = s.setupTestData(t)
	s.cleanupData = CleanupData{}
	t.Epic("Integration")
}

func (s *FeedbackRepositoryIntegrationSuite) AfterEach(t provider.T) {
	// Очистка только данных, созданных в текущем тесте
	s.cleanupTestData(t)
}

func (s *FeedbackRepositoryIntegrationSuite) setupTestData(t provider.T) TestData {
	// Генерация уникального идентификатора сессии
	sessionID := fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

	// Создание тестового пользователя со случайным именем
	userID := uuid.New()
	username := generateRandomUsername()

	_, err := s.db.Exec(
		`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, $2, 'hash', 'salt')`,
		userID, username,
	)
	require.NoError(t, err, "Failed to create test user")
	s.cleanupData.userIDs = append(s.cleanupData.userIDs, userID)

	// Создание тестовых feedback записей со случайными данными
	feedbackIDs := make([]uuid.UUID, 0)
	now := time.Now()

	// Создаем feedback записи разных типов и с разным временем
	feedbacks := []struct {
		rating       int
		respondentID uuid.UUID
		text         string
		feedbackType models.FeedbackType
		createdAt    time.Time
	}{
		// Bug reports (самые старые)
		{
			rand.Intn(11), userID,
			generateRandomText("bug-report"),
			models.FeedbackMessenger,
			now.Add(-2 * time.Hour),
		},
		{
			rand.Intn(11), userID,
			generateRandomText("bug-report"),
			models.FeedbackMessenger,
			now.Add(-90 * time.Minute),
		},
		{
			rand.Intn(11), userID,
			generateRandomText("bug-report"),
			models.FeedbackMessenger,
			now.Add(-1 * time.Hour),
		},

		// Feature requests
		{
			rand.Intn(11), userID,
			generateRandomText("feature-request"),
			models.FeedbackAuth,
			now.Add(-30 * time.Minute),
		},
		{
			rand.Intn(11), userID,
			generateRandomText("feature-request"),
			models.FeedbackAuth,
			now.Add(-15 * time.Minute),
		},

		// General feedback (самые новые)
		{
			rand.Intn(11), userID,
			generateRandomText("general-feedback"),
			models.FeedbackGeneral,
			now.Add(-5 * time.Minute),
		},
		{
			rand.Intn(11), userID,
			generateRandomText("general-feedback"),
			models.FeedbackGeneral,
			now.Add(-2 * time.Minute),
		},
	}

	for _, fb := range feedbacks {
		feedbackID := uuid.New()
		_, err := s.db.Exec(`
			INSERT INTO feedback (id, rating, respondent_id, text, type, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)
		`, feedbackID, fb.rating, fb.respondentID, fb.text, string(fb.feedbackType), fb.createdAt)
		require.NoError(t, err, "Failed to create test feedback")

		feedbackIDs = append(feedbackIDs, feedbackID)
		s.cleanupData.feedbackIDs = append(s.cleanupData.feedbackIDs, feedbackID)
	}

	return TestData{
		userID:      userID,
		feedbackIDs: feedbackIDs,
		sessionID:   sessionID,
	}
}

func (s *FeedbackRepositoryIntegrationSuite) cleanupTestData(t provider.T) {
	if len(s.cleanupData.feedbackIDs) == 0 && len(s.cleanupData.userIDs) == 0 {
		return
	}

	// Удаление feedback записей (только созданных в этом тесте)
	if len(s.cleanupData.feedbackIDs) > 0 {
		_, err := s.db.Exec(`
			DELETE FROM feedback WHERE id = ANY($1)
		`, s.cleanupData.feedbackIDs)
		if err != nil {
			t.Logf("Failed to clean up feedback: %v", err)
		}
	}

	// Удаление пользователей (только созданных в этом тесте)
	if len(s.cleanupData.userIDs) > 0 {
		_, err := s.db.Exec(`
			DELETE FROM "user" WHERE id = ANY($1)
		`, s.cleanupData.userIDs)
		if err != nil {
			t.Logf("Failed to clean up user: %v", err)
		}
	}

	// Очищаем списки ID после удаления
	s.cleanupData = CleanupData{}
}

func (s *FeedbackRepositoryIntegrationSuite) TestSaveFeedback(t provider.T) {
	t.Tags("create", "feedback", "integration")
	t.Description("Тестирование сохранения feedback")

	testCases := []struct {
		name          string
		feedback      models.Feedback
		expectError   bool
		expectedError error
	}{
		{
			name: "Успешное сохранение feedback с рейтингом",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(9).
				WithText("Great app!").
				WithType(models.FeedbackGeneral).
				Build(),
			expectError: false,
		},
		{
			name: "Успешное сохранение feedback без текста",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(10).
				WithText(""). // Пустой текст
				WithType(models.FeedbackMessenger).
				Build(),
			expectError: false,
		},
		{
			name: "Успешное сохранение feedback с максимальным рейтингом",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(10). // Максимальный рейтинг
				WithText("Perfect!").
				WithType(models.FeedbackAuth).
				Build(),
			expectError: false,
		},
		{
			name: "Успешное сохранение feedback с минимальным рейтингом",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(0). // Минимальный рейтинг
				WithText("Needs improvement").
				WithType(models.FeedbackGeneral).
				Build(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			err := s.repository.SaveFeedback(ctx, &tc.feedback)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				// Verify that feedback was actually saved
				// Проверяем, что feedback действительно сохранен
				feedbacks, err := s.repository.GetAllFeedbackType(
					ctx,
					tc.feedback.Type,
					time.Now().Add(time.Minute), // Будущее время чтобы получить все записи
					10,
				)
				assert.NoError(t, err)

				// Ищем наш feedback по тексту
				found := false
				for _, fb := range feedbacks {
					if fb.Text == tc.feedback.Text && fb.Rating == tc.feedback.Rating {
						found = true
						assert.Equal(t, tc.feedback.Type, fb.Type)
						assert.Equal(t, tc.feedback.RespondentId, fb.RespondentId)
						break
					}
				}
				assert.True(t, found, "Feedback should be found in database")
			}
		})
	}
}

func (s *FeedbackRepositoryIntegrationSuite) TestGetAllFeedbackType(t provider.T) {
	t.Tags("read", "feedback", "integration")
	t.Description("Тестирование получения feedback по типу")

	testCases := []struct {
		name             string
		feedbackType     models.FeedbackType
		timestamp        time.Time
		count            int
		expectedCount    int
		expectError      bool
		expectedError    error
		checkOrder       bool
		checkOldestFirst bool
	}{
		{
			name:          "Ограничение количества результатов",
			feedbackType:  models.FeedbackMessenger,
			timestamp:     time.Now(),
			count:         2,
			expectedCount: 2, // Ограничиваем 2 результатами
			expectError:   false,
			checkOrder:    true,
		},
		{
			name:          "Count = 0 возвращает nil",
			feedbackType:  models.FeedbackMessenger,
			timestamp:     time.Now(),
			count:         0,
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			feedbacks, err := s.repository.GetAllFeedbackType(ctx, tc.feedbackType, tc.timestamp, tc.count)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)

				if tc.count == 0 {
					assert.Nil(t, feedbacks)
				} else {
					assert.NotNil(t, feedbacks)
					assert.Len(t, feedbacks, tc.expectedCount)

					// Проверяем, что все feedback имеют правильный тип
					for _, fb := range feedbacks {
						assert.Equal(t, tc.feedbackType, fb.Type)
						assert.True(t, fb.CreatedAt.Before(tc.timestamp) || fb.CreatedAt.Equal(tc.timestamp))
					}

					// Проверяем порядок сортировки (должны быть отсортированы от новых к старым)
					if tc.checkOrder && len(feedbacks) > 1 {
						for i := 0; i < len(feedbacks)-1; i++ {
							assert.True(t,
								feedbacks[i].CreatedAt.After(feedbacks[i+1].CreatedAt) ||
									feedbacks[i].CreatedAt.Equal(feedbacks[i+1].CreatedAt),
								"Feedbacks should be ordered from newest to oldest")
						}
					}
				}
			}
		})
	}
}

func (s *FeedbackRepositoryIntegrationSuite) TestSaveFeedback_EdgeCases(t provider.T) {
	t.Tags("create", "feedback", "edge-cases", "integration")
	t.Description("Тестирование пограничных случаев при сохранении feedback")

	testCases := []struct {
		name          string
		feedback      models.Feedback
		expectError   bool
		expectedError error
	}{
		{
			name: "Ошибка при рейтинге больше 10",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(11). // Невалидный рейтинг
				WithText("Invalid rating").
				WithType(models.FeedbackGeneral).
				Build(),
			expectError: true,
		},
		{
			name: "Ошибка при рейтинге меньше 0",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(-1). // Невалидный рейтинг
				WithText("Invalid rating").
				WithType(models.FeedbackGeneral).
				Build(),
			expectError: true,
		},
		{
			name: "Успешное сохранение с очень длинным текстом",
			feedback: NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating(8).
				WithText("This is a very long feedback text that should be stored properly in the database without any issues. " +
					"PostgreSQL can handle large text values without problems.").
				WithType(models.FeedbackGeneral).
				Build(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			err := s.repository.SaveFeedback(ctx, &tc.feedback)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *FeedbackRepositoryIntegrationSuite) TestGetAllFeedbackType_NotFound(t provider.T) {
	t.Tags("read", "feedback", "not-found", "integration")
	t.Description("Тестирование случаев когда feedback не найден")

	testCases := []struct {
		name          string
		feedbackType  models.FeedbackType
		timestamp     time.Time
		count         int
		expectError   bool
		expectedError error
	}{
		{
			name:         "Пустой результат для несуществующего типа",
			feedbackType: models.FeedbackType("invalid_type"),
			timestamp:    time.Now(),
			count:        10,
			expectError:  false,
		},
		{
			name:         "Пустой результат для очень старого timestamp",
			feedbackType: models.FeedbackMessenger,
			timestamp:    time.Now().AddDate(-1, 0, 0), // Год назад
			count:        10,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			feedbacks, err := s.repository.GetAllFeedbackType(ctx, tc.feedbackType, tc.timestamp, tc.count)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError))
				}
			} else {
				assert.NoError(t, err)
				assert.Empty(t, feedbacks)
			}
		})
	}
}

func (s *FeedbackRepositoryIntegrationSuite) TestConcurrentAccess(t provider.T) {
	t.Tags("concurrency", "feedback", "integration")
	t.Description("Тестирование конкурентного доступа к репозиторию")

	ctx := context.Background()
	feedbackType := models.FeedbackGeneral

	// Сначала получаем текущее количество записей
	initialFeedbacks, err := s.repository.GetAllFeedbackType(ctx, models.FeedbackType(feedbackType), time.Now(), 100)
	require.NoError(t, err)
	initialCount := len(initialFeedbacks)

	// Создаем несколько feedback параллельно
	concurrentCount := 5
	errorsChan := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			feedback := NewFeedbackBuilder().
				WithRespondentID(s.testData.userID).
				WithRating((7 + index) % 11).
				WithText(fmt.Sprintf("Concurrent feedback %d", index)).
				WithType(models.FeedbackType(feedbackType)).
				Build()

			err := s.repository.SaveFeedback(ctx, &feedback)
			errorsChan <- err
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < concurrentCount; i++ {
		err := <-errorsChan
		assert.NoError(t, err)
	}

	// Проверяем, что все записи сохранены
	finalFeedbacks, err := s.repository.GetAllFeedbackType(ctx, models.FeedbackType(feedbackType), time.Now(), 100)
	assert.NoError(t, err)
	assert.Len(t, finalFeedbacks, initialCount+concurrentCount)
}
