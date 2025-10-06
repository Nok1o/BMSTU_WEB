package messenger_service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type StickerServiceClient struct {
	client pb.StickerServiceClient
}

func NewStickerServiceClient(conn *grpc.ClientConn) *StickerServiceClient {
	return &StickerServiceClient{
		client: pb.NewStickerServiceClient(conn),
	}
}

func (c *StickerServiceClient) AddStickerPack(ctx context.Context, stickerPack *models.StickerPack) (*models.StickerPack, error) {
	protoStickerPack := MapStickerPackToProto(stickerPack)

	resp, err := c.client.AddStickerPack(ctx, &pb.AddStickerPackRequest{
		StickerPack: protoStickerPack,
	})
	if err != nil {
		logger.Error(ctx, "Failed to add sticker pack: %v", err)
		return nil, err
	}

	createdStickerPack, err := MapProtoToStickerPack(resp.StickerPack)
	if err != nil {
		logger.Error(ctx, "Failed to convert Proto to StickerPack: %v", err)
		return nil, err
	}

	return createdStickerPack, nil
}

func (c *StickerServiceClient) GetStickerPack(ctx context.Context, packId uuid.UUID) (*models.StickerPack, error) {
	resp, err := c.client.GetStickerPack(ctx, &pb.GetStickerPackRequest{
		Id: packId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get sticker pack: %v", err)
		return nil, err
	}

	stickerPack, err := MapProtoToStickerPack(resp.StickerPack)
	if err != nil {
		logger.Error(ctx, "Failed to convert Proto to StickerPack: %v", err)
		return nil, err
	}

	return stickerPack, nil
}

func (c *StickerServiceClient) GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]*models.StickerPack, error) {
	resp, err := c.client.GetStickerPacks(ctx, &pb.GetStickerPacksRequest{
		UserId: userId.String(),
		Count:  int32(count),
		Offset: int32(offset),
	})
	if err != nil {
		logger.Error(ctx, "Failed to get sticker packs: %v", err)
		return nil, err
	}

	var stickerPacks []*models.StickerPack
	for _, pack := range resp.StickerPacks {
		stickerPack, err := MapProtoToStickerPack(pack)
		if err != nil {
			logger.Error(ctx, "Failed to convert Proto to StickerPack: %v", err)
			return nil, err
		}
		stickerPacks = append(stickerPacks, stickerPack)
	}

	return stickerPacks, nil
}

func (c *StickerServiceClient) DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error {
	_, err := c.client.DeleteStickerPack(ctx, &pb.DeleteStickerPackRequest{
		UserId: userId.String(),
		PackId: packId.String(),
	})
	if err != nil {
		logger.Error(ctx, "Failed to delete sticker pack: %v", err)
		return err
	}

	return nil
}

func (c *StickerServiceClient) GetStickerPackByName(ctx context.Context, packName string) (*models.StickerPack, error) {
	resp, err := c.client.GetStickerPackByName(ctx, &pb.GetStickerPackByNameRequest{
		Name: packName,
	})
	if err != nil {
		logger.Error(ctx, "Failed to get sticker pack by name: %v", err)
		return nil, err
	}

	stickerPack, err := MapProtoToStickerPack(resp.StickerPack)
	if err != nil {
		logger.Error(ctx, "Failed to convert Proto to StickerPack: %v", err)
		return nil, err
	}

	return stickerPack, nil
}
