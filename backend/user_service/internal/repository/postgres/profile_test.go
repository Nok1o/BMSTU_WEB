//go:build unit
// +build unit

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

type ProfileRepositorySuite struct {
	suite.Suite
}

func TestProfileRepository(t *testing.T) {
	suite.RunSuite(t, new(ProfileRepositorySuite))
}

func (s *ProfileRepositorySuite) TestSaveProfile(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Test saving a user profile")

	tests := []struct {
		name        string
		profile     models.Profile
		mockSetup   func(sqlmock.Sqlmock)
		expectedErr bool
	}{
		{
			name: "Successful profile save",
			profile: models.Profile{
				UserId: uuid.New(),
				BasicInfo: &models.BasicInfo{
					Name:        "Test",
					Surname:     "User",
					Sex:         models.MALE,
					DateOfBirth: time.Now(),
					Bio:         "Test bio",
					AvatarUrl:   "avatar.jpg",
				},
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("insert into profile").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: false,
		},
		{
			name: "Error saving profile",
			profile: models.Profile{
				UserId: uuid.New(),
				BasicInfo: &models.BasicInfo{
					Name:        "Test",
					Surname:     "User",
					Sex:         models.MALE,
					DateOfBirth: time.Now(),
					Bio:         "Test bio",
				},
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("insert into profile").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock)

			err = repo.SaveProfile(context.Background(), tt.profile)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to save profile")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestGetProfile(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Test retrieving a user profile")

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(sqlmock.Sqlmock, uuid.UUID)
		expectedErr bool
	}{
		{
			name:   "Successful profile retrieval",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "bio", "profile_avatar", "profile_background",
					"firstname", "lastname", "sex", "birth_date", "school_id", "contact_info_id", "last_seen"}).
					AddRow(userId, "bio", "avatar.jpg", nil, "Test", "User", 1, now, nil, nil, now)
				m.ExpectQuery("select id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date, school_id, contact_info_id, last_seen").
					WithArgs(userId).WillReturnRows(rows)
				m.ExpectQuery("select u.name, u.city, f.name, e.graduation_year").
					WithArgs(userId).WillReturnError(sql.ErrNoRows)
			},
			expectedErr: false,
		},
		{
			name:   "Profile not found",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				m.ExpectQuery("select id, bio, profile_avatar, profile_background, firstname, lastname, sex, birth_date, school_id, contact_info_id, last_seen").
					WithArgs(userId).WillReturnError(sql.ErrNoRows)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userId)

			result, err := repo.GetProfile(context.Background(), tt.userId)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to get profile")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userId, result.UserId)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestUpdateProfileTextInfo(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Test updating profile text information")

	tests := []struct {
		name        string
		profile     models.Profile
		mockSetup   func(sqlmock.Sqlmock)
		expectedErr bool
	}{
		{
			name: "Successful profile update",
			profile: models.Profile{
				UserId: uuid.New(),
				BasicInfo: &models.BasicInfo{
					Name:        "Test",
					Surname:     "User",
					Sex:         models.MALE,
					DateOfBirth: time.Now(),
					Bio:         "Test bio",
				},
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec("update profile").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			expectedErr: false,
		},
		{
			name: "Error updating profile",
			profile: models.Profile{
				UserId: uuid.New(),
				BasicInfo: &models.BasicInfo{
					Name:        "Test",
					Surname:     "User",
					Sex:         models.MALE,
					DateOfBirth: time.Now(),
					Bio:         "Test bio",
				},
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec("update profile").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("update error"))
				m.ExpectRollback()
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock)

			err = repo.UpdateProfileTextInfo(context.Background(), tt.profile)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to update profile")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestGetPublicUserInfo(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.NORMAL)
	t.Description("Test retrieving public user information")

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(sqlmock.Sqlmock, uuid.UUID)
		expectedErr bool
	}{
		{
			name:   "Successful public info retrieval",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "firstname", "lastname", "profile_avatar", "username", "last_seen"}).
					AddRow(userId, "Test", "User", "avatar.jpg", "testuser", now)
				m.ExpectQuery("select u.id, firstname, lastname, profile_avatar, username, last_seen").
					WithArgs(userId).WillReturnRows(rows)
			},
			expectedErr: false,
		},
		{
			name:   "Error retrieving public info",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				m.ExpectQuery("select u.id, firstname, lastname, profile_avatar, username, last_seen").
					WithArgs(userId).WillReturnError(errors.New("query error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userId)

			result, err := repo.GetPublicUserInfo(context.Background(), tt.userId)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to get public user info")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.userId, result.Id)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestGetPublicUsersInfo(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.NORMAL)
	t.Description("Test retrieving public info for multiple users")

	tests := []struct {
		name        string
		userIds     []uuid.UUID
		mockSetup   func(sqlmock.Sqlmock, []uuid.UUID)
		expectedErr bool
	}{
		{
			name:    "Empty user list",
			userIds: []uuid.UUID{},
			mockSetup: func(m sqlmock.Sqlmock, userIds []uuid.UUID) {
				// No setup needed
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userIds)

			result, err := repo.GetPublicUsersInfo(context.Background(), tt.userIds)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.userIds) == 0 {
					assert.Nil(t, result)
				} else {
					assert.Len(t, result, len(tt.userIds))
				}
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestUpdateProfileAvatar(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.NORMAL)
	t.Description("Test updating profile avatar")

	tests := []struct {
		name        string
		userId      uuid.UUID
		avatarUrl   string
		mockSetup   func(sqlmock.Sqlmock, uuid.UUID, string)
		expectedErr bool
	}{
		{
			name:      "Error updating avatar",
			userId:    uuid.New(),
			avatarUrl: "new_avatar.jpg",
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID, avatarUrl string) {
				m.ExpectExec("update profile set profile_avatar = \\$1 where id = \\$2").
					WithArgs(avatarUrl, userId).
					WillReturnError(errors.New("update error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userId, tt.avatarUrl)

			err = repo.UpdateProfileAvatar(context.Background(), tt.userId, tt.avatarUrl)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to update profile avatar")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestUpdateProfileCover(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.NORMAL)
	t.Description("Test updating profile cover image")

	tests := []struct {
		name        string
		userId      uuid.UUID
		coverUrl    string
		mockSetup   func(sqlmock.Sqlmock, uuid.UUID, string)
		expectedErr bool
	}{
		{
			name:     "Error updating cover",
			userId:   uuid.New(),
			coverUrl: "new_cover.jpg",
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID, coverUrl string) {
				m.ExpectExec("update profile set profile_background = \\$1 where id = \\$2").
					WithArgs(coverUrl, userId).
					WillReturnError(errors.New("update error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userId, tt.coverUrl)

			err = repo.UpdateProfileCover(context.Background(), tt.userId, tt.coverUrl)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unable to update profile background")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *ProfileRepositorySuite) TestUpdateLastSeen(t provider.T) {
	t.Epic("Unit")
	t.Feature("Profile Repository")
	t.Severity(allure.MINOR)
	t.Description("Test updating last seen time")

	tests := []struct {
		name        string
		userId      uuid.UUID
		mockSetup   func(sqlmock.Sqlmock, uuid.UUID)
		expectedErr bool
	}{
		{
			name:   "Successful last seen update",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				m.ExpectExec("update profile").
					WithArgs(userId, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: false,
		},
		{
			name:   "Error updating last seen",
			userId: uuid.New(),
			mockSetup: func(m sqlmock.Sqlmock, userId uuid.UUID) {
				m.ExpectExec("update profile").
					WithArgs(userId, sqlmock.AnyArg()).
					WillReturnError(errors.New("update error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewPostgresProfileRepository(db)
			tt.mockSetup(mock, tt.userId)

			err = repo.UpdateLastSeen(context.Background(), tt.userId)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "u.connPool.Exec")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
