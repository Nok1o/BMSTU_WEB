//go:build unit
// +build unit

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/post_service/internal/repository/postgres"
	"quickflow/shared/models"
)

type CommentRepositorySuite struct {
	suite.Suite
}

func TestCommentRepository(t *testing.T) {
	suite.RunSuite(t, new(CommentRepositorySuite))
}

func (s *CommentRepositorySuite) TestAddComment_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование успешного добавления комментария")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	comment := models.Comment{
		Id:        uuid.New(),
		PostId:    uuid.New(),
		UserId:    uuid.New(),
		Text:      "This is a comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LikeCount: 0,
	}

	// Моки
	mock.ExpectExec(`INSERT INTO comment \(id, post_id, user_id, created_at, text, like_count\)`).
		WithArgs(comment.Id, comment.PostId, comment.UserId, comment.CreatedAt, comment.Text, comment.LikeCount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызов метода
	_ = repo.AddComment(context.Background(), comment)
}

func (s *CommentRepositorySuite) TestGetCommentFiles_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование успешного получения файлов комментария")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	commentId := uuid.New()

	// Моки
	rows := sqlmock.NewRows([]string{"file_url", "file_type", "filename"}).
		AddRow("http://example.com/file1.jpg", "image", "file1.jpg").
		AddRow("http://example.com/file2.jpg", "image", "file2.jpg")

	mock.ExpectQuery(`select cf.file_url, cf.file_type, f.filename from comment_file cf`).
		WithArgs(commentId).
		WillReturnRows(rows)

	// Вызов метода
	files, err := repo.GetCommentFiles(context.Background(), commentId)

	// Проверка
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "http://example.com/file1.jpg", files[0])
	assert.Equal(t, "http://example.com/file2.jpg", files[1])
}

func (s *CommentRepositorySuite) TestDeleteComment_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование успешного удаления комментария")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	commentId := uuid.New()

	// Моки
	mock.ExpectExec(`delete from comment cascade where id = \$1`).
		WithArgs(commentId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызов метода
	err = repo.DeleteComment(context.Background(), commentId)

	// Проверка
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func (s *CommentRepositorySuite) TestGetCommentsForPost_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование успешного получения комментариев для поста")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	postId := uuid.New()
	numComments := 5
	timestamp := time.Now()

	// Моки для получения комментариев
	rows := sqlmock.NewRows([]string{"id", "post_id", "user_id", "text", "created_at", "updated_at", "like_count"}).
		AddRow(uuid.New(), postId, uuid.New(), "Comment text", time.Now(), time.Now(), 10)

	mock.ExpectQuery(`select id, post_id, user_id, text, created_at, updated_at, like_count`).
		WithArgs(postId, timestamp, numComments).
		WillReturnRows(rows)

	// Вызов метода
	_, _ = repo.GetCommentsForPost(context.Background(), postId, numComments, timestamp)
}

func (s *CommentRepositorySuite) TestGetComment_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование успешного получения комментария по ID")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	commentId := uuid.New()

	// Моки для получения комментария
	row := sqlmock.NewRows([]string{"id", "post_id", "user_id", "text", "created_at", "updated_at", "like_count"}).
		AddRow(commentId, uuid.New(), uuid.New(), "Comment text", time.Now(), time.Now(), 10)

	mock.ExpectQuery(`select id, post_id, user_id, text, created_at, updated_at, like_count`).
		WithArgs(commentId).
		WillReturnRows(row)

	// Вызов метода
	_, _ = repo.GetComment(context.Background(), commentId)
}

func (s *CommentRepositorySuite) TestLikeComment_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование успешного лайка комментария")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	commentId := uuid.New()
	userId := uuid.New()

	// Моки для проверки лайка
	mock.ExpectQuery(`select 1 from like_comment where comment_id = \$1 and user_id = \$2`).
		WithArgs(commentId, userId).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Моки для добавления лайка
	mock.ExpectExec(`insert into like_comment \(user_id, comment_id\)`).
		WithArgs(userId, commentId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызов метода
	err = repo.LikeComment(context.Background(), commentId, userId)

	// Проверка
	assert.NoError(t, err)
}

func (s *CommentRepositorySuite) TestUnlikeComment_Success(t provider.T) {
	t.Epic("Unit")
	t.Feature("Comment Repository")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование успешного удаления лайка с комментария")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Инициализация мока
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	// Создание репозитория
	repo := postgres.NewPostgresCommentRepository(db)

	// Тестовые данные
	commentId := uuid.New()
	userId := uuid.New()

	// Моки для проверки лайка
	mock.ExpectQuery(`select 1 from like_comment where comment_id = \$1 and user_id = \$2`).
		WithArgs(commentId, userId).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	// Моки для удаления лайка
	mock.ExpectExec(`delete from like_comment where comment_id = \$1 and user_id = \$2`).
		WithArgs(commentId, userId).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызов метода
	err = repo.UnlikeComment(context.Background(), commentId, userId)

	// Проверка
	assert.NoError(t, err)
}
