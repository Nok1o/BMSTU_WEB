package validation

import (
    "errors"

    "quickflow/shared/models"
)

var (
    ErrEmptyStickerPackName     = errors.New("sticker pack name cannot be empty")
    ErrStickerPackNameTooLong   = errors.New("sticker pack name cannot be longer than 50 characters")
    ErrEmptyStickerPackStickers = errors.New("sticker pack must contain at least one sticker")
    ErrInvalidStickerType       = errors.New("sticker must be of type 'sticker'")
)

type StickerValidator struct{}

func NewStickerValidator() *StickerValidator {
    return &StickerValidator{}
}

func (s *StickerValidator) ValidateStickerPack(pack *models.StickerPack) error {
    if pack == nil {
        return errors.New("sticker pack cannot be nil")
    }

    if len(pack.Name) == 0 {
        return ErrEmptyStickerPackName
    }
    if len(pack.Name) > 50 {
        return ErrStickerPackNameTooLong
    }
    if len(pack.Stickers) == 0 {
        return ErrEmptyStickerPackStickers
    }
    for _, sticker := range pack.Stickers {
        if sticker.DisplayType != models.DisplayTypeSticker {
            return ErrInvalidStickerType
        }
    }
    return nil
}
