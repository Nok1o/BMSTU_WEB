//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"

	"quickflow/friends_service/internal/usecase"
	"quickflow/friends_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestFriendsService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFriendsRepository(ctrl)
	service := usecase.NewFriendsService(mockRepo)
	ctx := context.Background()
	validUserID := uuid.New().String()
	validFriendID := uuid.New().String()

	// -------------------- GetFriendsInfo --------------------
	runner.Run(t, "GetFriendsInfo Tests", func(t provider.T) {
		t.Epic("Unit")
		t.Feature("GetFriendsInfo")
		t.Severity(allure.CRITICAL)
		t.Description("Test retrieving friends info with various scenarios")

		tests := []struct {
			name          string
			limit         string
			offset        string
			reqType       string
			mockSetup     func()
			expectedResp  []models.FriendInfo
			expectedCount int
			expectedErr   error
		}{
			{
				name:    "Success",
				limit:   "10",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					mockRepo.EXPECT().GetFriendsPublicInfo(ctx, validUserID, 10, 0, "all").
						Return([]models.FriendInfo{
							{Id: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Username: "user1"},
						}, 1, nil)
				},
				expectedResp: []models.FriendInfo{
					{Id: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Username: "user1"},
				},
				expectedCount: 1,
				expectedErr:   nil,
			},
			{
				name:    "Invalid limit",
				limit:   "invalid",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					// no repo call
				},
				expectedResp:  nil,
				expectedCount: 0,
				expectedErr:   fmt.Errorf("strconv.Atoi: parsing \"invalid\": invalid syntax"),
			},
			{
				name:    "Repository error",
				limit:   "10",
				offset:  "0",
				reqType: "all",
				mockSetup: func() {
					mockRepo.EXPECT().GetFriendsPublicInfo(ctx, validUserID, 10, 0, "all").
						Return(nil, 0, errors.New("db error"))
				},
				expectedResp:  []models.FriendInfo{},
				expectedCount: 0,
				expectedErr:   errors.New("db error"),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}

				resp, count, err := service.GetFriendsInfo(ctx, validUserID, tt.limit, tt.offset, tt.reqType)
				if tt.expectedErr != nil {
					assert.ErrorContains(stepCtx, err, tt.expectedErr.Error())
				} else {
					assert.NoError(stepCtx, err)
				}

				assert.Equal(stepCtx, tt.expectedResp, resp)
				assert.Equal(stepCtx, tt.expectedCount, count)
			})
		}
	})

	// -------------------- SendFriendRequest --------------------
	runner.Run(t, "SendFriendRequest Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("SendFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test sending friend requests with success and error cases")

		tests := []struct {
			name        string
			mockSetup   func()
			expectedErr error
		}{
			{
				name: "Success",
				mockSetup: func() {
					mockRepo.EXPECT().SendFriendRequest(ctx, validUserID, validFriendID).Return(nil)
				},
				expectedErr: nil,
			},
			{
				name: "Repository error",
				mockSetup: func() {
					mockRepo.EXPECT().SendFriendRequest(ctx, validUserID, validFriendID).Return(errors.New("db error"))
				},
				expectedErr: errors.New("db error"),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}
				err := service.SendFriendRequest(ctx, validUserID, validFriendID)
				if tt.expectedErr != nil {
					assert.ErrorContains(stepCtx, err, tt.expectedErr.Error())
				} else {
					assert.NoError(stepCtx, err)
				}
			})
		}
	})

	// -------------------- GetUserRelation --------------------
	runner.Run(t, "GetUserRelation Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("GetUserRelation")
		t.Severity(allure.CRITICAL)
		t.Description("Test getting relation between two users")

		user1 := uuid.New()
		user2 := uuid.New()

		tests := []struct {
			name        string
			user1       uuid.UUID
			user2       uuid.UUID
			mockSetup   func()
			expectedRel models.UserRelation
			expectedErr error
		}{
			{
				name:  "Success - Friends",
				user1: user1,
				user2: user2,
				mockSetup: func() {
					mockRepo.EXPECT().GetUserRelation(ctx, user1, user2).Return(models.RelationFriend, nil)
				},
				expectedRel: models.RelationFriend,
				expectedErr: nil,
			},
			{
				name:  "Same user",
				user1: user1,
				user2: user1,
				mockSetup: func() {
					// no repo call
				},
				expectedRel: models.RelationSelf,
				expectedErr: nil,
			},
			{
				name:  "Nil UUID",
				user1: uuid.Nil,
				user2: user2,
				mockSetup: func() {
					// no repo call
				},
				expectedRel: models.RelationStranger,
				expectedErr: fmt.Errorf("userID is empty"),
			},
			{
				name:  "Repository error",
				user1: user1,
				user2: user2,
				mockSetup: func() {
					mockRepo.EXPECT().GetUserRelation(ctx, user1, user2).Return(models.RelationStranger, errors.New("db error"))
				},
				expectedRel: models.RelationStranger,
				expectedErr: fmt.Errorf("f.friendsRepo.GetUserRelation"),
			},
		}

		for _, tt := range tests {
			tt := tt
			t.WithNewStep(tt.name, func(stepCtx provider.StepCtx) {
				if tt.mockSetup != nil {
					tt.mockSetup()
				}

				rel, err := service.GetUserRelation(ctx, tt.user1, tt.user2)
				if tt.expectedErr != nil {
					assert.ErrorContains(stepCtx, err, tt.expectedErr.Error())
				} else {
					assert.NoError(stepCtx, err)
				}
				assert.Equal(stepCtx, tt.expectedRel, rel)
			})
		}
	})

	// -------------------- AcceptFriendRequest --------------------
	runner.Run(t, "AcceptFriendRequest Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("AcceptFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test accepting friend requests")

		t.WithNewStep("Success", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().AcceptFriendRequest(ctx, validUserID, validFriendID).Return(nil)
			err := service.AcceptFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(stepCtx, err)
		})

		t.WithNewStep("Error", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().AcceptFriendRequest(ctx, validUserID, validFriendID).Return(errors.New("db error"))
			err := service.AcceptFriendRequest(ctx, validUserID, validFriendID)
			assert.Error(stepCtx, err)
		})
	})

	// -------------------- DeleteFriend --------------------
	runner.Run(t, "DeleteFriend Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("DeleteFriend")
		t.Severity(allure.CRITICAL)
		t.Description("Test deleting friends")

		t.WithNewStep("Success", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().DeleteFriend(ctx, validUserID, validFriendID).Return(nil)
			err := service.DeleteFriend(ctx, validUserID, validFriendID)
			assert.NoError(stepCtx, err)
		})

		t.WithNewStep("Error", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().DeleteFriend(ctx, validUserID, validFriendID).Return(errors.New("db error"))
			err := service.DeleteFriend(ctx, validUserID, validFriendID)
			assert.Error(stepCtx, err)
		})
	})

	// -------------------- Unfollow --------------------
	runner.Run(t, "Unfollow Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("Unfollow")
		t.Severity(allure.CRITICAL)
		t.Description("Test unfollowing friends")

		t.WithNewStep("Success", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().Unfollow(ctx, validUserID, validFriendID).Return(nil)
			err := service.Unfollow(ctx, validUserID, validFriendID)
			assert.NoError(stepCtx, err)
		})

		t.WithNewStep("Error", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().Unfollow(ctx, validUserID, validFriendID).Return(errors.New("db error"))
			err := service.Unfollow(ctx, validUserID, validFriendID)
			assert.Error(stepCtx, err)
		})
	})

	// -------------------- IsExistsFriendRequest --------------------
	runner.Run(t, "IsExistsFriendRequest Tests", func(t provider.T) {
		t.Epic("Friends Service")
		t.Feature("IsExistsFriendRequest")
		t.Severity(allure.CRITICAL)
		t.Description("Test checking if friend request exists")

		t.WithNewStep("Exists", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).Return(true, nil)
			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(stepCtx, err)
			assert.True(stepCtx, exists)
		})

		t.WithNewStep("Not exists", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).Return(false, nil)
			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.NoError(stepCtx, err)
			assert.False(stepCtx, exists)
		})

		t.WithNewStep("Error", func(stepCtx provider.StepCtx) {
			mockRepo.EXPECT().IsExistsFriendRequest(ctx, validUserID, validFriendID).Return(false, errors.New("db error"))
			exists, err := service.IsExistsFriendRequest(ctx, validUserID, validFriendID)
			assert.Error(stepCtx, err)
			assert.False(stepCtx, exists)
		})
	})
}
