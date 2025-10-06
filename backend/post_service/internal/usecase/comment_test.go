//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"quickflow/post_service/utils/validation"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/post_service/internal/errors"
	"quickflow/post_service/internal/usecase"
	"quickflow/post_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

type CommentUseCaseSuite struct {
	suite.Suite
}

func TestCommentUseCase(t *testing.T) {
	suite.RunSuite(t, new(CommentUseCaseSuite))
}

func (s *CommentUseCaseSuite) TestCommentUseCase(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment UseCase")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование usecase комментариев с различными сценариями")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Моки
	commentRepo := mocks.NewMockCommentRepository(ctrl)
	fileService := mocks.NewMockFileService(ctrl)
	validator := mocks.NewMockPostValidator(ctrl)

	// Создание объекта usecase
	service := usecase.NewCommentUseCase(commentRepo, fileService, validator)

	t.Run("DeleteComment", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Delete Comment")
		t.Severity(allure.BLOCKER)
		t.Description("Тестирование успешного удаления комментария с файлами")

		commentId := uuid.New()
		userId := uuid.New()

		// Ожидаемый вызов репозитория
		commentRepo.EXPECT().GetCommentFiles(gomock.Any(), commentId).Return([]string{"file1", "file2"}, nil)
		commentRepo.EXPECT().DeleteComment(gomock.Any(), commentId).Return(nil)
		fileService.EXPECT().DeleteFile(gomock.Any(), "file1").Return(nil)
		fileService.EXPECT().DeleteFile(gomock.Any(), "file2").Return(nil)

		// Вызов функции
		err := service.DeleteComment(context.Background(), userId, commentId)

		// Проверки
		assert.NoError(t, err)
	})

	t.Run("FetchCommentsForPost_InvalidNumComments", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Fetch Comments")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование ошибки при некорректном количестве комментариев")

		// Ожидаемая ошибка для некорректного числа комментариев
		validator.EXPECT().ValidateFeedParams(gomock.Any(), gomock.Any()).Return(validation.ErrInvalidNumPosts)

		// Вызов функции
		_, err := service.FetchCommentsForPost(context.Background(), uuid.New(), 0, time.Now())

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidNumComments, err)
	})

	t.Run("LikeComment", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Like Comment")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование успешного лайка комментария")

		postId := uuid.New()
		userId := uuid.New()

		// Ожидаемый вызов репозитория
		commentRepo.EXPECT().CheckIfCommentLiked(gomock.Any(), postId, userId).Return(false, nil)
		commentRepo.EXPECT().LikeComment(gomock.Any(), postId, userId).Return(nil)

		// Вызов функции
		err := service.LikeComment(context.Background(), postId, userId)

		// Проверки
		assert.NoError(t, err)
	})

	t.Run("UnlikeComment", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Unlike Comment")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование успешного удаления лайка с комментария")

		postId := uuid.New()
		userId := uuid.New()

		// Ожидаемый вызов репозитория
		commentRepo.EXPECT().CheckIfCommentLiked(gomock.Any(), postId, userId).Return(true, nil)
		commentRepo.EXPECT().UnlikeComment(gomock.Any(), postId, userId).Return(nil)

		// Вызов функции
		err := service.UnlikeComment(context.Background(), postId, userId)

		// Проверки
		assert.NoError(t, err)
	})

	t.Run("GetComment", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Get Comment")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование успешного получения комментария по ID")

		commentId := uuid.New()

		// Ожидаемый вызов репозитория
		comment := models.Comment{
			Id:        commentId,
			UserId:    uuid.New(),
			PostId:    uuid.New(),
			Text:      "This is a comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LikeCount: 0,
			IsLiked:   false,
		}
		commentRepo.EXPECT().GetComment(gomock.Any(), commentId).Return(comment, nil)

		// Вызов функции
		result, err := service.GetComment(context.Background(), commentId, uuid.New())

		// Проверки
		assert.NoError(t, err)
		assert.Equal(t, comment, *result)
	})

	t.Run("GetLastPostComment", func(t provider.T) {
		t.Parallel()
		t.Epic("Unit")
		t.Feature("Get Last Post Comment")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование успешного получения последнего комментария поста")

		postId := uuid.New()

		// Ожидаемый вывоз репозитория
		comment := models.Comment{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			PostId:    postId,
			Text:      "This is the last comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LikeCount: 10,
			IsLiked:   true,
		}
		commentRepo.EXPECT().GetLastPostComment(gomock.Any(), postId).Return(&comment, nil)

		// Вызов функции
		result, err := service.GetLastPostComment(context.Background(), postId)

		// Проверки
		assert.NoError(t, err)
		assert.Equal(t, comment, *result)
	})
}
