//go:build unit
// +build unit

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestCreateCommunity(t *testing.T) {
	runner.Run(t, "Create Community Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("Create Community")

		tests := []struct {
			name      string
			community models.Community
			mock      func(mock sqlmock.Sqlmock)
			wantErr   bool
		}{
			{
				name: "Successfully create community",
				community: models.Community{
					ID:        uuid.New(),
					OwnerID:   uuid.New(),
					NickName:  "TestCommunity",
					CreatedAt: time.Now(),
					BasicInfo: &models.BasicCommunityInfo{
						Name:        "Test Community",
						Description: "A description of the community",
						AvatarUrl:   "http://avatar.url",
						CoverUrl:    "http://cover.url",
					},
				},
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(`insert into community`).WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), "Test Community", "A description of the community",
						sqlmock.AnyArg(), "http://avatar.url", "http://cover.url", "TestCommunity").
						WillReturnResult(sqlmock.NewResult(1, 1))

					mock.ExpectExec(`insert into community_user`).WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), "owner", sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "Failed to create community",
				community: models.Community{
					ID:        uuid.New(),
					OwnerID:   uuid.New(),
					NickName:  "TestCommunity",
					CreatedAt: time.Now(),
					BasicInfo: &models.BasicCommunityInfo{
						Name:        "Test Community",
						Description: "A description of the community",
						AvatarUrl:   "http://avatar.url",
						CoverUrl:    "http://cover.url",
					},
				},
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(`insert into community`).WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), "Test Community", "A description of the community",
						sqlmock.AnyArg(), "http://avatar.url", "http://cover.url", "TestCommunity").
						WillReturnError(fmt.Errorf("insert failed"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Parallel()
				t.Severity(allure.CRITICAL)

				mockDB, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to open mock DB: %v", err)
				}
				defer mockDB.Close()

				repo := &SqlCommunityRepository{connPool: mockDB}
				tt.mock(mock)

				err = repo.CreateCommunity(context.Background(), tt.community)

				if tt.wantErr {
					t.WithNewStep("Verify error occurred", func(ctx provider.StepCtx) {
						assert.Error(t, err)
					})
				} else {
					t.WithNewStep("Verify success", func(ctx provider.StepCtx) {
						assert.NoError(t, err)
					})
				}

				t.WithNewStep("Verify mock expectations", func(ctx provider.StepCtx) {
					assert.NoError(t, mock.ExpectationsWereMet())
				})
			})
		}
	})
}

func TestGetCommunityById(t *testing.T) {
	runner.Run(t, "Get Community By ID Tests", func(t provider.T) {
		t.Feature("Get Community By ID")
		t.Epic("Unit")

		uuid_ := uuid.New()
		time_ := time.Now()

		tests := []struct {
			name    string
			id      uuid.UUID
			mock    func(mock sqlmock.Sqlmock)
			want    models.Community
			wantErr bool
		}{
			{
				name: "Successfully get community by id",
				id:   uuid_,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`select id, owner_id, name, description`).WithArgs(uuid_).
						WillReturnRows(sqlmock.NewRows([]string{
							"id", "owner_id", "name", "description", "created_at", "avatar_url", "cover_url", "contact_info", "nickname",
						}).
							AddRow(uuid_, uuid_, "Test Community", "Description of community", time_, "http://avatar.url", "http://cover.url", nil, "TestCommunity"))
				},
				want: models.Community{
					ID:        uuid_,
					OwnerID:   uuid_,
					NickName:  "TestCommunity",
					CreatedAt: time_,
					BasicInfo: &models.BasicCommunityInfo{
						Name:        "Test Community",
						Description: "Description of community",
						AvatarUrl:   "http://avatar.url",
						CoverUrl:    "http://cover.url",
					},
				},
				wantErr: false,
			},
			{
				name: "Failed to get community by id",
				id:   uuid_,
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`select id, owner_id, name, description`).WithArgs(uuid_).
						WillReturnError(fmt.Errorf("select failed"))
				},
				want:    models.Community{},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Parallel()
				t.Severity(allure.CRITICAL)

				mockDB, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to open mock DB: %v", err)
				}
				defer mockDB.Close()

				repo := &SqlCommunityRepository{connPool: mockDB}
				tt.mock(mock)

				var got models.Community
				got, err = repo.GetCommunityById(context.Background(), tt.id)

				if tt.wantErr {
					t.WithNewStep("Verify error occurred", func(ctx provider.StepCtx) {
						assert.Error(t, err)
					})
				} else {
					t.WithNewStep("Verify success and compare results", func(ctx provider.StepCtx) {
						assert.NoError(t, err)

						// Compare times using Equal to avoid failure due to small differences
						if !got.CreatedAt.Equal(tt.want.CreatedAt) {
							t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
						}

						// Compare all fields except time
						tt.want.CreatedAt = got.CreatedAt
						assert.Equal(t, tt.want, got)
					})
				}

				t.WithNewStep("Verify mock expectations", func(ctx provider.StepCtx) {
					assert.NoError(t, mock.ExpectationsWereMet())
				})
			})
		}
	})
}

func TestUpdateCommunityTextInfo(t *testing.T) {
	runner.Run(t, "Update Community Text Info Tests", func(t provider.T) {
		t.Feature("Update Community Text Info")
		t.Epic("Unit")

		tests := []struct {
			name      string
			community models.Community
			mock      func(mock sqlmock.Sqlmock)
			wantErr   bool
		}{
			{
				name: "Successfully update community",
				community: models.Community{
					ID:       uuid.New(),
					NickName: "UpdatedCommunity",
					BasicInfo: &models.BasicCommunityInfo{
						Name:        "Updated Name",
						Description: "Updated Description",
					},
				},
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(`update community set nickname`).WithArgs(
						"UpdatedCommunity", "Updated Name", "Updated Description", sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
				},
				wantErr: false,
			},
			{
				name: "Failed to update community",
				community: models.Community{
					ID:       uuid.New(),
					NickName: "UpdatedCommunity",
					BasicInfo: &models.BasicCommunityInfo{
						Name:        "Updated Name",
						Description: "Updated Description",
					},
				},
				mock: func(mock sqlmock.Sqlmock) {
					mock.ExpectExec(`update community set nickname`).WithArgs(
						"UpdatedCommunity", "Updated Name", "Updated Description", sqlmock.AnyArg()).
						WillReturnError(fmt.Errorf("update failed"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				t.Epic("Unit")
				t.Parallel()
				t.Severity(allure.NORMAL)
				t.Description("Test updating community text information")

				mockDB, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to open mock DB: %v", err)
				}
				defer mockDB.Close()

				repo := &SqlCommunityRepository{connPool: mockDB}
				tt.mock(mock)

				err = repo.UpdateCommunityTextInfo(context.Background(), tt.community)

				if tt.wantErr {
					t.WithNewStep("Verify error occurred", func(ctx provider.StepCtx) {
						assert.Error(t, err)
					})
				} else {
					t.WithNewStep("Verify success", func(ctx provider.StepCtx) {
						assert.NoError(t, err)
					})
				}

				t.WithNewStep("Verify mock expectations", func(ctx provider.StepCtx) {
					assert.NoError(t, mock.ExpectationsWereMet())
				})
			})
		}
	})
}
