package usecase

import (
	"context"

	"github.com/google/uuid"

	messenger_errors "quickflow/messenger_service/internal/errors"
	"quickflow/shared/models"
)

type StickerRepository interface {
	AddStickerPack(ctx context.Context, stickerPack models.StickerPack) error
	GetStickerPack(ctx context.Context, packId uuid.UUID) (models.StickerPack, error)
	GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]models.StickerPack, error)
	DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error
	GetStickerPackByName(ctx context.Context, name string) (models.StickerPack, error)
	BelongsTo(ctx context.Context, userId uuid.UUID, packId uuid.UUID) (bool, error)
}

type StickerPackValidator interface {
	ValidateStickerPack(stickerPack *models.StickerPack) error
}

type StickerService struct {
	fileRepo    FileService
	stickerRepo StickerRepository
	validator   StickerPackValidator
}

func NewStickerService(stickerRepo StickerRepository, fileRepo FileService, validator StickerPackValidator) *StickerService {
	return &StickerService{
		fileRepo:    fileRepo,
		stickerRepo: stickerRepo,
		validator:   validator,
	}
}

func (s *StickerService) AddStickerPack(ctx context.Context, stickerPack *models.StickerPack) (*models.StickerPack, error) {
	// Validate sticker pack
	if err := s.validator.ValidateStickerPack(stickerPack); err != nil {
		return nil, err
	}

	if stickerPack.Id == uuid.Nil {
		stickerPack.Id = uuid.New()
	}

	err := s.stickerRepo.AddStickerPack(ctx, *stickerPack)
	if err != nil {
		return nil, err
	}

	return stickerPack, nil
}

func (s *StickerService) GetStickerPack(ctx context.Context, packId uuid.UUID) (models.StickerPack, error) {
	stickerPack, err := s.stickerRepo.GetStickerPack(ctx, packId)
	if err != nil {
		return models.StickerPack{}, err
	}
	return stickerPack, nil
}

func (s *StickerService) GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]models.StickerPack, error) {
	return s.stickerRepo.GetStickerPacks(ctx, userId, count, offset)
}

func (s *StickerService) DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error {
	if belongs, err := s.stickerRepo.BelongsTo(ctx, userId, packId); err != nil {
		return err
	} else if !belongs {
		return messenger_errors.ErrNotOwnerOfStickerPack
	}

	return s.stickerRepo.DeleteStickerPack(ctx, userId, packId)
}

func (s *StickerService) GetStickerPackByName(ctx context.Context, name string) (models.StickerPack, error) {
	stickerPack, err := s.stickerRepo.GetStickerPackByName(ctx, name)
	if err != nil {
		return models.StickerPack{}, err
	}
	return stickerPack, nil
}

func (s *StickerService) BelongsTo(ctx context.Context, userId uuid.UUID, packId uuid.UUID) (bool, error) {
	belongs, err := s.stickerRepo.BelongsTo(ctx, userId, packId)
	if err != nil {
		return false, err
	}
	return belongs, nil
}
