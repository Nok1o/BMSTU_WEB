//go:build integration
// +build integration

package usecase

import (
	"context"
	"database/sql"
	"quickflow/config/test"

	"fmt"
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

	"quickflow/feedback_service/internal/repository/postgres"
	"quickflow/shared/models"
)

type FeedbackObjectMother struct {
	userId uuid.UUID
}

func NewFeedbackObjectMother(userId uuid.UUID) *FeedbackObjectMother {
	return &FeedbackObjectMother{userId: userId}
}

// ValidFeedback создает валидный feedback
func (om *FeedbackObjectMother) ValidFeedback() *models.Feedback {
	return &models.Feedback{
		Id:           uuid.New(),
		Rating:       4,
		RespondentId: om.userId,
		Text:         "This is a valid feedback text",
		Type:         models.FeedbackGeneral,
		CreatedAt:    time.Now(),
	}
}

// FeedbackWithHighRating создает feedback с высоким рейтингом
func (om *FeedbackObjectMother) FeedbackWithHighRating() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Rating = 5
	fb.Text = "Excellent service!"
	fb.Type = models.FeedbackGeneral
	return fb
}

// FeedbackWithLowRating создает feedback с низким рейтингом
func (om *FeedbackObjectMother) FeedbackWithLowRating() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Rating = 3
	fb.Text = "Needs improvement"
	fb.Type = models.FeedbackGeneral
	return fb
}

// FeedbackWithoutText создает feedback без текста
func (om *FeedbackObjectMother) FeedbackWithoutText() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Text = ""
	return fb
}

// FeedbackWithLongText создает feedback с длинным текстом
func (om *FeedbackObjectMother) FeedbackWithLongText() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Text = `This is a very long feedback text that exceeds the normal length. 
	It contains multiple sentences and should be properly handled by the system.
	The validation should allow this kind of`
	return fb
}

// FeedbackWithInvalidRating создает feedback с невалидным рейтингом
func (om *FeedbackObjectMother) FeedbackWithInvalidRating() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Rating = 8 // Невалидный рейтинг
	return fb
}

// FeedbackWithNegativeRating создает feedback с отрицательным рейтингом
func (om *FeedbackObjectMother) FeedbackWithNegativeRating() *models.Feedback {
	fb := om.ValidFeedback()
	fb.Rating = -2 // Отрицательный рейтинг
	return fb
}

// FeedbackWithDifferentType создает feedback с указанным типом
func (om *FeedbackObjectMother) FeedbackWithDifferentType(feedbackType models.FeedbackType) *models.Feedback {
	fb := om.ValidFeedback()
	fb.Type = feedbackType
	return fb
}

// FeedbackWithPastDate создает feedback с прошедшей датой
func (om *FeedbackObjectMother) FeedbackWithPastDate() *models.Feedback {
	fb := om.ValidFeedback()
	fb.CreatedAt = time.Now().Add(-24 * time.Hour)
	return fb
}

// FeedbackWithFutureDate создает feedback с будущей датой
func (om *FeedbackObjectMother) FeedbackWithFutureDate() *models.Feedback {
	fb := om.ValidFeedback()
	fb.CreatedAt = time.Now().Add(24 * time.Hour)
	return fb
}

type FeedbackUseCaseIntegrationSuite struct {
	suite.Suite
	db           *sql.DB
	repository   *postgres.FeedbackRepository
	useCase      *FeedbackUseCase
	objectMother *FeedbackObjectMother
	testData     TestData
	cleanupIDs   []uuid.UUID
}

type TestData struct {
	userID        uuid.UUID
	feedbackIDs   []uuid.UUID
	testFeedbacks []*models.Feedback
}

func TestFeedbackUseCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.RunSuite(t, new(FeedbackUseCaseIntegrationSuite))
}

func (s *FeedbackUseCaseIntegrationSuite) BeforeAll(t provider.T) {
	t.Feature("Feedback UseCase Integration")
	t.Severity(allure.CRITICAL)
	t.Description("Подготовка тестовой среды для интеграционных тестов UseCase")

	connString := getEnv.GetEnv(test.TestDbConnStringEnvVar, test.DefaultDatabaseTestUrl)
	require.NotEmpty(t, connString, "Connection string must not be empty")

	var err error
	s.db, err = sql.Open("pgx", connString)
	require.NoError(t, err, "Failed to connect to test database")

	err = s.db.Ping()
	require.NoError(t, err, "Failed to ping database")

	// Инициализация репозитория
	s.repository = postgres.NewFeedbackRepository(s.db)

	// Инициализация use case
	s.useCase = NewFeedBackUseCase(s.repository)
}

func (s *FeedbackUseCaseIntegrationSuite) AfterAll(t provider.T) {
	t.Description("Очистка тестовой среды")
	s.cleanupTestData(t)
	if s.db != nil {
		s.db.Close()
	}
}

func (s *FeedbackUseCaseIntegrationSuite) BeforeEach(t provider.T) {
	s.cleanupTestData(t)
	s.testData = s.setupTestData(t)
	t.Epic("Integration")
}

func (s *FeedbackUseCaseIntegrationSuite) setupTestData(t provider.T) TestData {
	// Создание тестового пользователя
	userID := uuid.New()
	_, err := s.db.Exec(
		`INSERT INTO "user" (id, username, psw_hash, salt) VALUES ($1, 'testuser', 'hash', 'salt')`,
		userID,
	)
	require.NoError(t, err, "Failed to create test user")
	s.cleanupIDs = append(s.cleanupIDs, userID)

	s.objectMother = NewFeedbackObjectMother(userID)

	// Создание тестовых feedback записей разных типов
	feedbackIDs := make([]uuid.UUID, 0)
	testFeedbacks := make([]*models.Feedback, 0)
	now := time.Now()

	// Создаем feedback записи для тестирования
	feedbacksToCreate := []struct {
		feedback *models.Feedback
		delay    time.Duration
	}{
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackGeneral), -2 * time.Hour},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackGeneral), -90 * time.Minute},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackGeneral), -1 * time.Hour},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackAuth), -30 * time.Minute},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackAuth), -15 * time.Minute},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackGeneral), -5 * time.Minute},
		{s.objectMother.FeedbackWithDifferentType(models.FeedbackGeneral), -2 * time.Minute},
	}

	for _, fb := range feedbacksToCreate {
		fb.feedback.RespondentId = userID
		fb.feedback.CreatedAt = now.Add(fb.delay)
		fb.feedback.Id = uuid.New()

		_, err := s.db.Exec(`
			INSERT INTO feedback (id, rating, respondent_id, text, type, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)
		`, fb.feedback.Id, fb.feedback.Rating, fb.feedback.RespondentId,
			fb.feedback.Text, string(fb.feedback.Type), fb.feedback.CreatedAt)
		require.NoError(t, err, "Failed to create test feedback")

		feedbackIDs = append(feedbackIDs, fb.feedback.Id)
		testFeedbacks = append(testFeedbacks, fb.feedback)
		s.cleanupIDs = append(s.cleanupIDs, fb.feedback.Id)
	}

	return TestData{
		userID:        userID,
		feedbackIDs:   feedbackIDs,
		testFeedbacks: testFeedbacks,
	}
}

func (s *FeedbackUseCaseIntegrationSuite) cleanupTestData(t provider.T) {
	// Удаление feedback записей
	_, err := s.db.Exec(`
		DELETE FROM feedback
	`)
	if err != nil {
		t.Error("Failed to clean up feedback:", err)
	}

	// Удаление пользователя
	_, err = s.db.Exec(`
		DELETE FROM "user"
	`)
	if err != nil {
		t.Error("Failed to clean up user:", err)
	}
}

func (s *FeedbackUseCaseIntegrationSuite) TestSaveFeedback(t provider.T) {
	t.Tags("create", "feedback", "integration")
	t.Description("Тестирование сохранения feedback через use case")

	testCases := []struct {
		name          string
		feedback      *models.Feedback
		expectError   bool
		expectedError error
	}{
		{
			name:        "Успешное сохранение валидного feedback",
			feedback:    s.objectMother.ValidFeedback(),
			expectError: false,
		},
		{
			name:        "Успешное сохранение feedback с высоким рейтингом",
			feedback:    s.objectMother.FeedbackWithHighRating(),
			expectError: false,
		},
		{
			name:        "Успешное сохранение feedback с низким рейтингом",
			feedback:    s.objectMother.FeedbackWithLowRating(),
			expectError: false,
		},
		{
			name:        "Успешное сохранение feedback без текста",
			feedback:    s.objectMother.FeedbackWithoutText(),
			expectError: false,
		},
		{
			name:        "Успешное сохранение feedback с длинным текстом",
			feedback:    s.objectMother.FeedbackWithLongText(),
			expectError: false,
		},
		{
			name:          "Ошибка при невалидном рейтинге",
			feedback:      s.objectMother.FeedbackWithInvalidRating(),
			expectError:   true,
			expectedError: fmt.Errorf("invalid"),
		},
		{
			name:          "Ошибка при отрицательном рейтинге",
			feedback:      s.objectMother.FeedbackWithNegativeRating(),
			expectError:   true,
			expectedError: fmt.Errorf("invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			err := s.useCase.SaveFeedback(ctx, tc.feedback)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Contains(t, err.Error(), tc.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)

				// Verify that feedback was actually saved
				feedbacks, err := s.useCase.GetAllFeedbackType(
					ctx,
					tc.feedback.Type,
					time.Now().Add(time.Minute),
					10,
				)
				assert.NoError(t, err)

				// Ищем наш feedback по тексту или рейтингу
				found := false
				for _, fb := range feedbacks {
					if (fb.Text == tc.feedback.Text && fb.Text != "") ||
						(fb.Rating == tc.feedback.Rating && tc.feedback.Text == "") {
						found = true
						assert.Equal(t, tc.feedback.Type, fb.Type)
						assert.Equal(t, tc.feedback.Rating, fb.Rating)
						break
					}
				}
				assert.True(t, found, "Feedback should be found in database")
			}
		})
	}
}

func (s *FeedbackUseCaseIntegrationSuite) TestGetAllFeedbackType(t provider.T) {
	t.Tags("read", "feedback", "integration")
	t.Description("Тестирование получения feedback по типу через use case")

	testCases := []struct {
		name          string
		feedbackType  models.FeedbackType
		timestamp     time.Time
		count         int
		expectedCount int
		expectError   bool
		checkOrder    bool
	}{
		{
			name:          "Успешное получение feature requests",
			feedbackType:  models.FeedbackAuth,
			timestamp:     time.Now(),
			count:         10,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "Успешное получение general feedback",
			feedbackType:  models.FeedbackGeneral,
			timestamp:     time.Now(),
			count:         10,
			expectedCount: 5,
			expectError:   false,
		},
		{
			name:          "Ограничение количества результатов",
			feedbackType:  models.FeedbackGeneral,
			timestamp:     time.Now(),
			count:         2,
			expectedCount: 2,
			expectError:   false,
			checkOrder:    true,
		},
		{
			name:          "Получение только старых записей",
			feedbackType:  models.FeedbackGeneral,
			timestamp:     time.Now().Add(-45 * time.Minute),
			count:         10,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "Count = 0 возвращает nil",
			feedbackType:  models.FeedbackGeneral,
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
			feedbacks, err := s.useCase.GetAllFeedbackType(ctx, tc.feedbackType, tc.timestamp, tc.count)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
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

					// Проверяем порядок сортировки
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

func (s *FeedbackUseCaseIntegrationSuite) TestGetAllFeedbackType_EdgeCases(t provider.T) {
	t.Tags("read", "feedback", "edge-cases", "integration")
	t.Description("Тестирование пограничных случаев при получении feedback")

	testCases := []struct {
		name          string
		feedbackType  models.FeedbackType
		timestamp     time.Time
		count         int
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Пустой результат для несуществующего типа",
			feedbackType:  models.FeedbackType("nonexistent_type"),
			timestamp:     time.Now(),
			count:         10,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Пустой результат для будущего timestamp",
			feedbackType:  models.FeedbackGeneral,
			timestamp:     time.Now().AddDate(-1, 0, 0), // Год назад
			count:         10,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "Отрицательный count",
			feedbackType:  models.FeedbackGeneral,
			timestamp:     time.Now(),
			count:         -1,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			feedbacks, err := s.useCase.GetAllFeedbackType(ctx, tc.feedbackType, tc.timestamp, tc.count)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tc.count <= 0 {
					assert.Nil(t, feedbacks)
				} else {
					assert.Len(t, feedbacks, tc.expectedCount)
				}
			}
		})
	}
}

func (s *FeedbackUseCaseIntegrationSuite) TestValidationIntegration(t provider.T) {
	t.Tags("validation", "integration")
	t.Description("Тестирование интеграции валидации в use case")

	testCases := []struct {
		name          string
		feedback      *models.Feedback
		expectError   bool
		errorContains string
	}{
		{
			name:        "Валидный feedback проходит валидацию",
			feedback:    s.objectMother.ValidFeedback(),
			expectError: false,
		},
		{
			name:          "Рейтинг > 5 не проходит валидацию",
			feedback:      s.objectMother.FeedbackWithInvalidRating(),
			expectError:   true,
			errorContains: "rating",
		},
		{
			name:          "Рейтинг < 0 не проходит валидацию",
			feedback:      s.objectMother.FeedbackWithNegativeRating(),
			expectError:   true,
			errorContains: "rating",
		},
		{
			name: "Отсутствует respondent Id",
			feedback: func() *models.Feedback {
				fb := s.objectMother.ValidFeedback()
				fb.RespondentId = uuid.Nil
				return fb
			}(),
			expectError:   true,
			errorContains: "respondent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Epic("Integration")
			ctx := context.Background()

			// Act
			err := s.useCase.SaveFeedback(ctx, tc.feedback)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *FeedbackUseCaseIntegrationSuite) TestErrorPropagation(t provider.T) {
	t.Tags("error-handling", "integration")
	t.Description("Тестирование пробрасывания ошибок из репозитория")

	// Создаем use case с невалидным репозиторием для тестирования ошибок
	invalidRepo := &InvalidFeedbackRepository{}
	useCaseWithInvalidRepo := NewFeedBackUseCase(invalidRepo)

	t.Run("Ошибка из репозитория пробрасывается наверх", func(t provider.T) {
		ctx := context.Background()
		feedback := s.objectMother.ValidFeedback()

		// Act
		err := useCaseWithInvalidRepo.SaveFeedback(ctx, feedback)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
	})

	t.Run("Ошибка при получении feedback пробрасывается наверх", func(t provider.T) {
		ctx := context.Background()

		// Act
		_, err := useCaseWithInvalidRepo.GetAllFeedbackType(ctx, models.FeedbackGeneral, time.Now(), 10)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get all feedback type")
	})
}

// InvalidFeedbackRepository - мок репозитория для тестирования ошибок
type InvalidFeedbackRepository struct{}

func (r *InvalidFeedbackRepository) SaveFeedback(ctx context.Context, feedback *models.Feedback) error {
	return fmt.Errorf("repository error: cannot save feedback")
}

func (r *InvalidFeedbackRepository) GetAllFeedbackType(ctx context.Context, feedbackType models.FeedbackType, ts time.Time, count int) ([]models.Feedback, error) {
	return nil, fmt.Errorf("repository error: cannot get feedback")
}

func (s *FeedbackUseCaseIntegrationSuite) TestConcurrentOperations(t provider.T) {
	t.Tags("concurrency", "integration")
	t.Description("Тестирование конкурентных операций с feedback")

	ctx := context.Background()
	concurrentCount := 3
	errorsChan := make(chan error, concurrentCount)
	resultsChan := make(chan []models.Feedback, concurrentCount)

	// Конкурентное сохранение
	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			feedback := s.objectMother.ValidFeedback()
			feedback.Text = fmt.Sprintf("Concurrent feedback %d", index)
			err := s.useCase.SaveFeedback(ctx, feedback)
			errorsChan <- err
		}(i)
	}

	// Конкурентное чтение
	for i := 0; i < concurrentCount; i++ {
		go func() {
			feedbacks, err := s.useCase.GetAllFeedbackType(ctx, models.FeedbackGeneral, time.Now(), 10)
			if err != nil {
				resultsChan <- nil
			} else {
				resultsChan <- feedbacks
			}
		}()
	}

	// Проверяем результаты сохранения
	for i := 0; i < concurrentCount; i++ {
		err := <-errorsChan
		assert.NoError(t, err)
	}

	// Проверяем результаты чтения
	for i := 0; i < concurrentCount; i++ {
		feedbacks := <-resultsChan
		assert.NotNil(t, feedbacks)
	}
}
