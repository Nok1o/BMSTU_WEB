//go:build unit
// +build unit

package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"

	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/messenger_service/internal/usecase"
	"quickflow/messenger_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

type FileBuilder struct {
	file *models.File
}

func NewFileBuilder() *FileBuilder {
	return &FileBuilder{
		file: &models.File{
			URL:         "sticker_url",
			Name:        "sticker",
			DisplayType: models.DisplayTypeSticker,
		},
	}
}

func (b *FileBuilder) WithURL(url string) *FileBuilder {
	b.file.URL = url
	return b
}

func (b *FileBuilder) WithName(name string) *FileBuilder {
	b.file.Name = name
	return b
}

func (b *FileBuilder) Build() *models.File {
	return b.file
}

type StickerPackBuilder struct {
	pack *models.StickerPack
}

func NewStickerPackBuilder() *StickerPackBuilder {
	return &StickerPackBuilder{
		pack: &models.StickerPack{
			Id:       uuid.New(),
			Name:     "Funny Stickers",
			Stickers: []*models.File{NewFileBuilder().Build()},
		},
	}
}

func (b *StickerPackBuilder) WithID(id uuid.UUID) *StickerPackBuilder {
	b.pack.Id = id
	return b
}

func (b *StickerPackBuilder) WithName(name string) *StickerPackBuilder {
	b.pack.Name = name
	return b
}

func (b *StickerPackBuilder) WithStickers(stickers []*models.File) *StickerPackBuilder {
	b.pack.Stickers = stickers
	return b
}

func (b *StickerPackBuilder) Build() *models.StickerPack {
	return b.pack
}

type StickerServiceSuite struct {
	suite.Suite
}

func TestStickerService(t *testing.T) {
	suite.RunSuite(t, new(StickerServiceSuite))
}

func (s *StickerServiceSuite) TestStickerService(t provider.T) {
	t.Epic("Unit")
	t.Feature("Sticker Service")
	t.Severity(allure.CRITICAL)
	t.Description("Тестирование сервиса стикеров с различными сценариями")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	// Моки
	stickerRepo := mocks.NewMockStickerRepository(ctrl)
	fileRepo := mocks.NewMockFileService(ctrl)
	validator := mocks.NewMockStickerPackValidator(ctrl)
	service := usecase.NewStickerService(stickerRepo, fileRepo, validator)

	t.Run("AddStickerPack", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Add Sticker Pack")
		t.Severity(allure.BLOCKER)
		t.Description("Тестирование добавления стикерпаков с различными сценариями")

		pack := NewStickerPackBuilder().Build()

		tests := []struct {
			name      string
			setupMock func()
			wantError bool
		}{
			{
				name: "Success",
				setupMock: func() {
					validator.EXPECT().ValidateStickerPack(pack).Return(nil)
					stickerRepo.EXPECT().AddStickerPack(ctx, *pack).Return(nil)
				},
				wantError: false,
			},
			{
				name: "ValidationError",
				setupMock: func() {
					validator.EXPECT().ValidateStickerPack(pack).Return(errors.New("validation error"))
				},
				wantError: true,
			},
			{
				name: "RepoError",
				setupMock: func() {
					validator.EXPECT().ValidateStickerPack(pack).Return(nil)
					stickerRepo.EXPECT().AddStickerPack(ctx, *pack).Return(errors.New("repo error"))
				},
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				if tt.setupMock != nil {
					tt.setupMock()
				}
				result, err := service.AddStickerPack(ctx, pack)
				if tt.wantError {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, pack, result)
				}
			})
		}
	})

	t.Run("GetStickerPack", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Get Sticker Pack")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование получения стикерпака по ID с различными сценариями")

		packID := uuid.New()
		pack := *NewStickerPackBuilder().WithID(packID).Build()

		tests := []struct {
			name      string
			mockError error
			wantError bool
		}{
			{"Success", nil, false},
			{"NotFound", messenger_errors.ErrNotFound, true},
			{"RepoError", errors.New("repo error"), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				stickerRepo.EXPECT().GetStickerPack(ctx, packID).Return(pack, tt.mockError)
				result, err := service.GetStickerPack(ctx, packID)
				if tt.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, pack, result)
				}
			})
		}
	})

	t.Run("DeleteStickerPack", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Delete Sticker Pack")
		t.Severity(allure.CRITICAL)
		t.Description("Тестирование удаления стикерпаков с различными сценариями")

		userID := uuid.New()
		packID := uuid.New()

		tests := []struct {
			name       string
			belongs    bool
			belongsErr error
			deleteErr  error
			wantError  bool
		}{
			{"Success", true, nil, nil, false},
			{"NotOwner", false, nil, nil, true},
			{"BelongsRepoError", false, errors.New("repo error"), nil, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				stickerRepo.EXPECT().BelongsTo(ctx, userID, packID).Return(tt.belongs, tt.belongsErr)
				if tt.belongs {
					stickerRepo.EXPECT().DeleteStickerPack(ctx, userID, packID).Return(tt.deleteErr).AnyTimes()
				}

				err := service.DeleteStickerPack(ctx, userID, packID)
				if tt.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("GetStickerPackByName", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Get Sticker Pack By Name")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование получения стикерпака по имени с различными сценариями")

		name := "Funny Stickers"
		pack := *NewStickerPackBuilder().WithName(name).Build()

		tests := []struct {
			name      string
			mockError error
			wantError bool
		}{
			{"Success", nil, false},
			{"NotFound", messenger_errors.ErrNotFound, true},
			{"RepoError", errors.New("repo error"), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				stickerRepo.EXPECT().GetStickerPackByName(ctx, name).Return(pack, tt.mockError)
				result, err := service.GetStickerPackByName(ctx, name)
				if tt.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, pack, result)
				}
			})
		}
	})

	t.Run("BelongsTo", func(t provider.T) {

		t.Epic("Unit")
		t.Feature("Check Sticker Pack Ownership")
		t.Severity(allure.NORMAL)
		t.Description("Тестирование проверки принадлежности стикерпака пользователю")

		userID := uuid.New()
		packID := uuid.New()

		tests := []struct {
			name      string
			belongs   bool
			mockError error
			wantError bool
		}{
			{"Success", true, nil, false},
			{"NotFound", false, messenger_errors.ErrNotFound, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t provider.T) {

				stickerRepo.EXPECT().BelongsTo(ctx, userID, packID).Return(tt.belongs, tt.mockError)
				b, err := service.BelongsTo(ctx, userID, packID)
				if tt.wantError {
					assert.Error(t, err)
					assert.False(t, b)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.belongs, b)
				}
			})
		}
	})
}
