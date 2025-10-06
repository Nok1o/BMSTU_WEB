package messenger_service

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/client/file_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

// MapStickerPackToProto преобразует StickerPack из модели в Proto сообщение
func MapStickerPackToProto(stickerPack *models.StickerPack) *pb.StickerPack {
	if stickerPack == nil {
		return nil
	}
	return &pb.StickerPack{
		Id:        stickerPack.Id.String(),
		Name:      stickerPack.Name,
		CreatorId: stickerPack.CreatorId.String(),
		CreatedAt: timestamppb.New(stickerPack.CreatedAt),
		UpdatedAt: timestamppb.New(stickerPack.UpdatedAt),
		Stickers:  file_service.ModelFilesToProto(stickerPack.Stickers),
	}
}

// MapStickerPacksToProto преобразует стикерпак из списка моделей в список Proto сообщений
func MapStickerPacksToProto(stickerPacks []*models.StickerPack) []*pb.StickerPack {
	if stickerPacks == nil {
		return nil
	}
	res := make([]*pb.StickerPack, len(stickerPacks))
	for i, stickerPack := range stickerPacks {
		res[i] = MapStickerPackToProto(stickerPack)
	}
	return res
}

// MapProtoToStickerPack преобразует StickerPack из Proto сообщения в модель
func MapProtoToStickerPack(stickerPack *pb.StickerPack) (*models.StickerPack, error) {
	if stickerPack == nil {
		return nil, nil
	}
	id, err := uuid.Parse(stickerPack.Id)
	if err != nil {
		return nil, err
	}

	creatorId, err := uuid.Parse(stickerPack.CreatorId)
	if err != nil {
		return nil, err
	}

	stickers := file_service.ProtoFilesToModels(stickerPack.Stickers)

	return &models.StickerPack{
		Id:        id,
		Name:      stickerPack.Name,
		CreatorId: creatorId,
		Stickers:  stickers,
		CreatedAt: stickerPack.CreatedAt.AsTime(),
		UpdatedAt: stickerPack.UpdatedAt.AsTime(),
	}, nil
}
