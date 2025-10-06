//go:build unit
// +build unit

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	postgres "quickflow/friends_service/internal/repository"
	"quickflow/shared/models"
)

func TestPostgresFriendsRepository_GetFriendsPublicInfo(t *testing.T) {
	runner.Run(t, "Get Friends Public Info Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("GetFriendsPublicInfo")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving friends public info for various scenarios")

		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := postgres.NewPostgresFriendsRepository(db)

		userID := "user-123"
		limit, offset := 10, 0
		reqType := "all"

		tests := []struct {
			name    string
			setup   func()
			wantErr bool
		}{
			{
				name: "success",
				setup: func() {
					rows := sqlmock.NewRows([]string{"id", "username", "firstname", "lastname", "profile_avatar", "university"}).
						AddRow(uuid.New(), "user1", "John", "Doe", "avatar1.jpg", "UnivA")
					mock.ExpectQuery("with related_users").
						WithArgs(userID, models.RelationFriend, models.RelationFriend, limit, offset, true, false).
						WillReturnRows(rows)

					mock.ExpectQuery("select count").
						WithArgs(userID, models.RelationFriend, models.RelationFriend, true, false).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				},
				wantErr: false,
			},
			{
				name: "query error",
				setup: func() {
					mock.ExpectQuery("with related_users").
						WithArgs(userID, models.RelationFriend, models.RelationFriend, limit, offset, true, false).
						WillReturnError(errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				tt.setup()
				_, _, err := repo.GetFriendsPublicInfo(context.Background(), userID, limit, offset, reqType)
				if tt.wantErr {
					assert.Error(stepCtx, err)
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.NoError(stepCtx, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestPostgresFriendsRepository_SendFriendRequest(t *testing.T) {
	runner.Run(t, "Send Friend Request Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("SendFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test sending friend request with success and error cases")

		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := postgres.NewPostgresFriendsRepository(db)

		senderID, receiverID := "user-123", "user-456"

		tests := []struct {
			name    string
			setup   func()
			wantErr bool
		}{
			{
				name: "success",
				setup: func() {
					mock.ExpectExec("insert into friendship").
						WithArgs(senderID, receiverID, models.RelationFollowing).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "db error",
				setup: func() {
					mock.ExpectExec("insert into friendship").
						WithArgs(senderID, receiverID, models.RelationFollowing).
						WillReturnError(errors.New("insert failed"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				tt.setup()
				err := repo.SendFriendRequest(context.Background(), senderID, receiverID)
				if tt.wantErr {
					assert.Error(stepCtx, err)
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.NoError(stepCtx, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestPostgresFriendsRepository_IsExistsFriendRequest(t *testing.T) {
	runner.Run(t, "Check Friend Request Exists Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("IsExistsFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test checking existence of a friend request")

		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := postgres.NewPostgresFriendsRepository(db)

		senderID, receiverID := "user-123", "user-456"

		tests := []struct {
			name     string
			setup    func()
			wantErr  bool
			expected bool
		}{
			{
				name: "exists",
				setup: func() {
					mock.ExpectQuery("select status from friendship").
						WithArgs(senderID, receiverID).
						WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RelationFollowing))
				},
				wantErr:  false,
				expected: true,
			},
			{
				name: "no rows",
				setup: func() {
					mock.ExpectQuery("select status from friendship").
						WithArgs(senderID, receiverID).
						WillReturnError(sql.ErrNoRows)
				},
				wantErr:  false,
				expected: false,
			},
			{
				name: "db error",
				setup: func() {
					mock.ExpectQuery("select status from friendship").
						WithArgs(senderID, receiverID).
						WillReturnError(errors.New("db fail"))
				},
				wantErr:  true,
				expected: false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				tt.setup()
				exists, err := repo.IsExistsFriendRequest(context.Background(), senderID, receiverID)
				if tt.wantErr {
					assert.Error(stepCtx, err)
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.Equal(stepCtx, tt.expected, exists)
				assert.NoError(stepCtx, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestPostgresFriendsRepository_AcceptFriendRequest(t *testing.T) {
	runner.Run(t, "Accept Friend Request Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("AcceptFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test accepting friend requests with success and failure scenarios")

		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := postgres.NewPostgresFriendsRepository(db)

		senderID, receiverID := "user-123", "user-456"

		tests := []struct {
			name    string
			setup   func()
			wantErr bool
		}{
			{
				name: "success",
				setup: func() {
					mock.ExpectExec("update friendship").
						WithArgs(senderID, receiverID, models.RelationFriend).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "no rows affected",
				setup: func() {
					mock.ExpectExec("update friendship").
						WithArgs(senderID, receiverID, models.RelationFriend).
						WillReturnResult(sqlmock.NewResult(1, 0))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				tt.setup()
				err := repo.AcceptFriendRequest(context.Background(), senderID, receiverID)
				if tt.wantErr {
					assert.Error(stepCtx, err)
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.NoError(stepCtx, mock.ExpectationsWereMet())
			})
		}
	})
}

func TestPostgresFriendsRepository_Unfollow(t *testing.T) {
	runner.Run(t, "Unfollow Tests", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Unfollow")
		t.Severity(allure.CRITICAL)
		t.Description("Test unfollowing a friend with success and failure scenarios")

		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo := postgres.NewPostgresFriendsRepository(db)

		userID, friendID := "user-123", "user-456"

		tests := []struct {
			name    string
			setup   func()
			wantErr bool
		}{
			{
				name: "success",
				setup: func() {
					mock.ExpectExec("delete from friendship").
						WithArgs(userID, friendID, models.RelationFollowedBy, models.RelationFollowing).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "no rows affected",
				setup: func() {
					mock.ExpectExec("delete from friendship").
						WithArgs(userID, friendID, models.RelationFollowedBy, models.RelationFollowing).
						WillReturnResult(sqlmock.NewResult(1, 0))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				tt.setup()
				err := repo.Unfollow(context.Background(), userID, friendID)
				if tt.wantErr {
					assert.Error(stepCtx, err)
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.NoError(stepCtx, mock.ExpectationsWereMet())
			})
		}
	})
}
