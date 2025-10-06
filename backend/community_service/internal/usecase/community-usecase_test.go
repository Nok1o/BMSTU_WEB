//go:build unit
// +build unit

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	community_errors "quickflow/community_service/internal/errors"
	"quickflow/community_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestCommunityUseCase_CreateCommunity(t *testing.T) {
	runner.Run(t, "CreateCommunity", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()

		tests := []struct {
			name      string
			setup     func(repo *mocks.MockCommunityRepository, fs *mocks.MockFileService, v *mocks.MockCommunityValidator)
			community models.Community
			wantErr   bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
					r.EXPECT().GetCommunityByName(ctx, "Test").Return(models.Community{}, community_errors.ErrNotFound)
					r.EXPECT().CreateCommunity(ctx, gomock.Any()).Return(nil)
				},
				community: models.Community{
					BasicInfo: &models.BasicCommunityInfo{Name: "Test"},
					OwnerID:   uuid.New(),
				},
				wantErr: false,
			},
			{
				name: "fails validation",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(errors.New("bad"))
				},
				community: models.Community{BasicInfo: &models.BasicCommunityInfo{Name: "X"}},
				wantErr:   true,
			},
			{
				name: "already exists",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
					r.EXPECT().GetCommunityByName(ctx, "Dup").Return(models.Community{ID: uuid.New()}, nil)
				},
				community: models.Community{BasicInfo: &models.BasicCommunityInfo{Name: "Dup"}},
				wantErr:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Epic("Unit")
				t.Parallel()
				t.Feature("CreateCommunity")
				t.Story(tt.name)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				tt.setup(repo, fs, v)

				uc := NewCommunityUseCase(repo, fs, v)
				res, err := uc.CreateCommunity(ctx, tt.community)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, res)
					t.WithParameters(allure.NewParameter("expectedError", "true"))
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, res)
					t.WithParameters(allure.NewParameter("result", res.ID.String()))
				}
			})
		}
	})
}

func TestCommunityUseCase_GetCommunityById(t *testing.T) {
	runner.Run(t, "GetCommunityById", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		id := uuid.New()

		tests := []struct {
			name    string
			id      uuid.UUID
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				id:   id,
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityById(ctx, id).Return(models.Community{ID: id}, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty id",
				id:      uuid.Nil,
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				id:   id,
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityById(ctx, id).Return(models.Community{}, errors.New("db"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("GetCommunityById")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("communityID", tt.id.String()),
					allure.NewParameter("expectError", tt.wantErr),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				tt.setup(repo)

				uc := NewCommunityUseCase(repo, fs, v)
				got, err := uc.GetCommunityById(ctx, tt.id)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Equal(t, models.Community{}, got)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.id, got.ID)
				}
			})
		}
	})
}

func TestCommunityUseCase_GetCommunityByName(t *testing.T) {
	runner.Run(t, "GetCommunityByName", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		name := "test-community"

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityByName(ctx, name).Return(models.Community{BasicInfo: &models.BasicCommunityInfo{Name: name}}, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty name",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityByName(ctx, name).Return(models.Community{}, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("GetCommunityByName")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("communityName", name),
					allure.NewParameter("expectError", tt.wantErr),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty name" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var got models.Community
				var err error

				if tt.name == "empty name" {
					got, err = uc.GetCommunityByName(ctx, "")
				} else {
					got, err = uc.GetCommunityByName(ctx, name)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, name, got.BasicInfo.Name)
				}
			})
		}
	})
}

func TestCommunityUseCase_GetCommunityMembers(t *testing.T) {
	runner.Run(t, "GetCommunityMembers", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		communityID := uuid.New()
		numMembers := 10
		ts := time.Now()

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					members := []models.CommunityMember{
						{UserID: uuid.New(), CommunityID: communityID},
						{UserID: uuid.New(), CommunityID: communityID},
					}
					r.EXPECT().GetCommunityMembers(ctx, communityID, numMembers, ts).Return(members, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty community id",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityMembers(ctx, communityID, numMembers, ts).Return(nil, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("GetCommunityMembers")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("communityID", communityID.String()),
					allure.NewParameter("limit", numMembers),
					allure.NewParameter("timestamp", ts.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty community id" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var got []models.CommunityMember
				var err error

				if tt.name == "empty community id" {
					got, err = uc.GetCommunityMembers(ctx, uuid.Nil, numMembers, ts)
				} else {
					got, err = uc.GetCommunityMembers(ctx, communityID, numMembers, ts)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, got, 2)
				}
			})
		}
	})
}

func TestCommunityUseCase_IsCommunityMember(t *testing.T) {
	runner.Run(t, "IsCommunityMember", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		communityID := uuid.New()
		role := models.CommunityRoleAdmin

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success - is member",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().IsCommunityMember(ctx, userID, communityID).Return(true, &role, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty ids",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().IsCommunityMember(ctx, userID, communityID).Return(false, nil, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("IsCommunityMember")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("communityID", communityID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty ids" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var isMember bool
				var gotRole *models.CommunityRole
				var err error

				if tt.name == "empty ids" {
					isMember, gotRole, err = uc.IsCommunityMember(ctx, uuid.Nil, uuid.Nil)
				} else {
					isMember, gotRole, err = uc.IsCommunityMember(ctx, userID, communityID)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else if tt.name == "success - is member" {
					assert.NoError(t, err)
					assert.True(t, isMember)
					assert.Equal(t, role, *gotRole)
				}
			})
		}
	})
}

func TestCommunityUseCase_DeleteCommunity(t *testing.T) {
	runner.Run(t, "DeleteCommunity", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		communityID := uuid.New()
		community := models.Community{
			ID: communityID,
			BasicInfo: &models.BasicCommunityInfo{
				AvatarUrl: "avatar.jpg",
				CoverUrl:  "cover.jpg",
			},
		}

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository, f *mocks.MockFileService)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService) {
					r.EXPECT().GetCommunityById(ctx, communityID).Return(community, nil)
					r.EXPECT().DeleteCommunity(ctx, communityID).Return(nil)
					f.EXPECT().DeleteFile(ctx, "avatar.jpg").Return(nil)
					f.EXPECT().DeleteFile(ctx, "cover.jpg").Return(nil)
				},
				wantErr: false,
			},
			{
				name:    "empty community id",
				setup:   func(r *mocks.MockCommunityRepository, f *mocks.MockFileService) {},
				wantErr: true,
			},
			{
				name: "repo get error",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService) {
					r.EXPECT().GetCommunityById(ctx, communityID).Return(models.Community{}, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("DeleteCommunity")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("communityID", communityID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty community id" {
					tt.setup(repo, fs)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var err error

				if tt.name == "empty community id" {
					err = uc.DeleteCommunity(ctx, uuid.Nil)
				} else {
					err = uc.DeleteCommunity(ctx, communityID)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCommunityUseCase_UpdateCommunity(t *testing.T) {
	runner.Run(t, "UpdateCommunity", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		communityID := uuid.New()
		community := models.Community{
			ID: communityID,
			BasicInfo: &models.BasicCommunityInfo{
				Name: "updated-name",
			},
		}
		role := models.CommunityRoleAdmin

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
					r.EXPECT().GetCommunityByName(ctx, "updated-name").Return(models.Community{}, community_errors.ErrNotFound)
					r.EXPECT().IsCommunityMember(ctx, userID, communityID).Return(true, &role, nil)
					r.EXPECT().GetCommunityById(ctx, communityID).Return(models.Community{ID: communityID}, nil)
					r.EXPECT().UpdateCommunityTextInfo(ctx, gomock.Any()).Return(nil)
					r.EXPECT().GetCommunityById(ctx, communityID).Return(community, nil)
				},
				wantErr: false,
			},
			{
				name: "validation error",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(errors.New("validation error"))
				},
				wantErr: true,
			},
			{
				name: "not a member",
				setup: func(r *mocks.MockCommunityRepository, f *mocks.MockFileService, v *mocks.MockCommunityValidator) {
					v.EXPECT().ValidateCommunity(gomock.Any()).Return(nil)
					r.EXPECT().GetCommunityByName(ctx, "updated-name").Return(models.Community{}, community_errors.ErrNotFound)
					r.EXPECT().IsCommunityMember(ctx, userID, communityID).Return(false, nil, nil)
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("UpdateCommunity")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("communityID", communityID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				tt.setup(repo, fs, v)

				uc := NewCommunityUseCase(repo, fs, v)
				_, err := uc.UpdateCommunity(ctx, community, userID)

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCommunityUseCase_JoinCommunity(t *testing.T) {
	runner.Run(t, "JoinCommunity", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		member := models.CommunityMember{
			UserID:      uuid.New(),
			CommunityID: uuid.New(),
		}
		community := models.Community{
			ID:      member.CommunityID,
			OwnerID: uuid.New(), // different from member.UserID
		}

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityById(ctx, member.CommunityID).Return(community, nil)
					r.EXPECT().JoinCommunity(ctx, gomock.Any()).Return(nil)
				},
				wantErr: false,
			},
			{
				name:    "empty ids",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo get error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetCommunityById(ctx, member.CommunityID).Return(models.Community{}, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("JoinCommunity")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", member.UserID.String()),
					allure.NewParameter("communityID", member.CommunityID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty ids" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var err error

				if tt.name == "empty ids" {
					err = uc.JoinCommunity(ctx, models.CommunityMember{})
				} else {
					err = uc.JoinCommunity(ctx, member)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCommunityUseCase_LeaveCommunity(t *testing.T) {
	runner.Run(t, "LeaveCommunity", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		communityID := uuid.New()

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().LeaveCommunity(ctx, userID, communityID).Return(nil)
				},
				wantErr: false,
			},
			{
				name:    "empty ids",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().LeaveCommunity(ctx, userID, communityID).Return(errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("LeaveCommunity")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("communityID", communityID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty ids" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var err error

				if tt.name == "empty ids" {
					err = uc.LeaveCommunity(ctx, uuid.Nil, uuid.Nil)
				} else {
					err = uc.LeaveCommunity(ctx, userID, communityID)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCommunityUseCase_GetUserCommunities(t *testing.T) {
	runner.Run(t, "GetUserCommunities", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		count := 5
		ts := time.Now()

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					communities := []models.Community{
						{ID: uuid.New(), BasicInfo: &models.BasicCommunityInfo{Name: "Community 1"}},
						{ID: uuid.New(), BasicInfo: &models.BasicCommunityInfo{Name: "Community 2"}},
					}
					r.EXPECT().GetUserCommunities(ctx, userID, count, ts).Return(communities, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty user id",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetUserCommunities(ctx, userID, count, ts).Return(nil, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("GetUserCommunities")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("count", count),
					allure.NewParameter("timestamp", ts.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty user id" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var got []models.Community
				var err error

				if tt.name == "empty user id" {
					got, err = uc.GetUserCommunities(ctx, uuid.Nil, count, ts)
				} else {
					got, err = uc.GetUserCommunities(ctx, userID, count, ts)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, got, 2)
				}
			})
		}
	})
}

func TestCommunityUseCase_SearchSimilarCommunities(t *testing.T) {
	runner.Run(t, "SearchSimilarCommunities", func(t provider.T) {
		ctx := context.Background()
		t.Epic("Unit")
		name := "test"
		count := 5

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					communities := []models.Community{
						{ID: uuid.New(), BasicInfo: &models.BasicCommunityInfo{Name: "test-community"}},
					}
					r.EXPECT().SearchSimilarCommunities(ctx, name, count).Return(communities, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty name",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name:    "invalid count",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().SearchSimilarCommunities(ctx, name, count).Return(nil, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("SearchSimilarCommunities")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("name", name),
					allure.NewParameter("count", count),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty name" && tt.name != "invalid count" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var got []models.Community
				var err error

				switch tt.name {
				case "empty name":
					got, err = uc.SearchSimilarCommunities(ctx, "", count)
				case "invalid count":
					got, err = uc.SearchSimilarCommunities(ctx, name, 0)
				default:
					got, err = uc.SearchSimilarCommunities(ctx, name, count)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, got, 1)
				}
			})
		}
	})
}

func TestCommunityUseCase_ChangeUserRole(t *testing.T) {
	runner.Run(t, "ChangeUserRole", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		communityID := uuid.New()
		requesterID := uuid.New()
		role := models.CommunityRoleAdmin
		requesterRole := models.CommunityRoleAdmin

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().IsCommunityMember(ctx, requesterID, communityID).Return(true, &requesterRole, nil)
					r.EXPECT().ChangeUserRole(ctx, userID, communityID, role).Return(nil)
				},
				wantErr: false,
			},
			{
				name:    "empty ids",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "requester not admin",
				setup: func(r *mocks.MockCommunityRepository) {
					memberRole := models.CommunityRoleMember
					r.EXPECT().IsCommunityMember(ctx, requesterID, communityID).Return(true, &memberRole, nil)
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("ChangeUserRole")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("communityID", communityID.String()),
					allure.NewParameter("role", string(role)),
					allure.NewParameter("requesterID", requesterID.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty ids" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var err error

				if tt.name == "empty ids" {
					err = uc.ChangeUserRole(ctx, uuid.Nil, uuid.Nil, role, uuid.Nil)
				} else {
					err = uc.ChangeUserRole(ctx, userID, communityID, role, requesterID)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestCommunityUseCase_GetControlledCommunities(t *testing.T) {
	runner.Run(t, "GetControlledCommunities", func(t provider.T) {
		t.Epic("Unit")
		ctx := context.Background()
		userID := uuid.New()
		count := 5
		ts := time.Now()

		tests := []struct {
			name    string
			setup   func(r *mocks.MockCommunityRepository)
			wantErr bool
		}{
			{
				name: "success",
				setup: func(r *mocks.MockCommunityRepository) {
					communities := []models.Community{
						{ID: uuid.New(), BasicInfo: &models.BasicCommunityInfo{Name: "Controlled Community 1"}},
						{ID: uuid.New(), BasicInfo: &models.BasicCommunityInfo{Name: "Controlled Community 2"}},
					}
					r.EXPECT().GetControlledCommunities(ctx, userID, count, ts).Return(communities, nil)
				},
				wantErr: false,
			},
			{
				name:    "empty user id",
				setup:   func(r *mocks.MockCommunityRepository) {},
				wantErr: true,
			},
			{
				name: "repo error",
				setup: func(r *mocks.MockCommunityRepository) {
					r.EXPECT().GetControlledCommunities(ctx, userID, count, ts).Return(nil, errors.New("db error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {
				t.Parallel()
				t.Epic("Unit")
				t.Feature("GetControlledCommunities")
				t.Story(tt.name)
				t.WithParameters(
					allure.NewParameter("userID", userID.String()),
					allure.NewParameter("count", count),
					allure.NewParameter("timestamp", ts.String()),
				)

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				repo := mocks.NewMockCommunityRepository(ctrl)
				fs := mocks.NewMockFileService(ctrl)
				v := mocks.NewMockCommunityValidator(ctrl)

				if tt.name != "empty user id" {
					tt.setup(repo)
				}

				uc := NewCommunityUseCase(repo, fs, v)
				var got []models.Community
				var err error

				if tt.name == "empty user id" {
					got, err = uc.GetControlledCommunities(ctx, uuid.Nil, count, ts)
				} else {
					got, err = uc.GetControlledCommunities(ctx, userID, count, ts)
				}

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, got, 2)
				}
			})
		}
	})
}
