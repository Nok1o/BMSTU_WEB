//go:build unit
// +build unit

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

type UserRepositorySuite struct {
	suite.Suite
}

func TestUserRepository(t *testing.T) {
	suite.RunSuite(t, new(UserRepositorySuite))
}

func (s *UserRepositorySuite) TestSaveUser(t provider.T) {
	t.Epic("Unit")
	t.Feature("User Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование сохранения пользователя")

	tests := []struct {
		name    string
		user    models.User
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successful save user",
			user: models.User{
				Id:       uuid.New(),
				Username: "johndoe",
				Password: "password123",
				Salt:     "salt123",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into "user"`).
					WithArgs(
						sqlmock.AnyArg(), "johndoe", "password123", "salt123",
					).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Failed to save user",
			user: models.User{
				Id:       uuid.New(),
				Username: "johndoe",
				Password: "password123",
				Salt:     "salt123",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into "user"`).
					WithArgs(
						sqlmock.AnyArg(), "johndoe", "password123", "salt123",
					).WillReturnError(fmt.Errorf("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			userRepo := &PostgresUserRepository{connPool: mockDB}
			tt.mock(mock)

			_, err = userRepo.SaveUser(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (s *UserRepositorySuite) TestGetUser(t provider.T) {
	t.Epic("Unit")
	t.Feature("User Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование получения пользователя по логину и паролю")

	hash := sha256.Sum256([]byte("hashed_password" + "salt123"))
	uuid_new := uuid.New()
	tests := []struct {
		name      string
		loginData models.LoginData
		mock      func(mock sqlmock.Sqlmock)
		want      models.User
		wantErr   bool
	}{
		{
			name: "Successfully get user",
			loginData: models.LoginData{
				Username: "johndoe",
				Password: "hashed_password",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, username, psw_hash, salt`).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "psw_hash", "salt"}).
						AddRow(uuid_new, "johndoe", hex.EncodeToString(hash[:]), "salt123"))

			},
			want: models.User{
				Id:       uuid_new,
				Username: "johndoe",
				Password: hex.EncodeToString(hash[:]),
				Salt:     "salt123",
			},
			wantErr: false,
		},
		{
			name: "Failed to get user (incorrect password)",
			loginData: models.LoginData{
				Username: "johndoe",
				Password: "wrongpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, username, psw_hash, salt`).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "psw_hash", "salt"}).
						AddRow(uuid.New(), "johndoe", "hashed_password", "salt123"))
			},
			want:    models.User{},
			wantErr: true,
		},
		{
			name: "Failed to get user (user not found)",
			loginData: models.LoginData{
				Username: "nonexistentuser",
				Password: "password123",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, username, psw_hash, salt`).
					WithArgs("nonexistentuser").
					WillReturnError(fmt.Errorf("user not found"))
			},
			want:    models.User{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			userRepo := &PostgresUserRepository{connPool: mockDB}
			tt.mock(mock)

			got, err := userRepo.GetUser(context.Background(), tt.loginData)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (s *UserRepositorySuite) TestGetUserByUId(t provider.T) {
	t.Epic("Unit")
	t.Feature("User Repository")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование получения пользователя по UUID")

	uuid_ := uuid.New()
	tests := []struct {
		name    string
		userId  uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    models.User
		wantErr bool
	}{
		{
			name:   "Successfully get user by UID",
			userId: uuid_, // Используем сгенерированный UUID
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, username, psw_hash, salt`).
					WithArgs(uuid_). // UUID для запроса
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "psw_hash", "salt"}).
						AddRow(uuid_, "johndoe", "hashed_password", "salt123")) // Данные в мок-результате
			},
			want: models.User{
				Id:       uuid_,
				Username: "johndoe",
				Password: "hashed_password",
				Salt:     "salt123",
			},
			wantErr: false,
		},
		{
			name:   "Failed to get user by UID",
			userId: uuid_,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, username, psw_hash, salt`).
					WithArgs(uuid_).
					WillReturnError(fmt.Errorf("user not found"))
			},
			want:    models.User{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			userRepo := &PostgresUserRepository{connPool: mockDB}
			tt.mock(mock)

			got, err := userRepo.GetUserByUId(context.Background(), tt.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUId() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (s *UserRepositorySuite) TestSearchSimilar(t provider.T) {
	t.Epic("Unit")
	t.Feature("User Repository")
	t.Severity(allure.NORMAL)
	t.Description("Тестирование поиска похожих пользователей")

	uuid_ := uuid.New()
	tests := []struct {
		name       string
		toSearch   string
		postsCount uint
		mock       func(mock sqlmock.Sqlmock)
		want       []models.PublicUserInfo
		wantErr    bool
	}{
		{
			name:       "Successfully search similar users",
			toSearch:   "john",
			postsCount: 5,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, firstname, lastname, profile_avatar`).
					WithArgs("john", uint(5)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "firstname", "lastname", "profile_avatar"}).
						AddRow(uuid_, "johndoe", "John", "Doe", "http://avatar.url"))
			},
			want: []models.PublicUserInfo{
				{
					Id:        uuid_,
					Username:  "johndoe",
					Firstname: "John",
					Lastname:  "Doe",
					AvatarURL: "http://avatar.url",
				},
			},
			wantErr: false,
		},
		{
			name:       "Failed to search similar users",
			toSearch:   "john",
			postsCount: 5,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, firstname, lastname, profile_avatar`).
					WithArgs("john", uint(5)).
					WillReturnError(fmt.Errorf("query failed"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t provider.T) {
			t.Epic("Unit")
			t.Parallel()

			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			userRepo := &PostgresUserRepository{connPool: mockDB}
			tt.mock(mock)

			got, err := userRepo.SearchSimilar(context.Background(), tt.toSearch, tt.postsCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchSimilar() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
