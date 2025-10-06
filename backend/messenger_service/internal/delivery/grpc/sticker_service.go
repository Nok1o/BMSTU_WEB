package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	dto "quickflow/shared/client/messenger_service"
	"quickflow/shared/logger"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

type StickerServiceUseCase interface {
	AddStickerPack(ctx context.Context, stickerPack *models.StickerPack) (*models.StickerPack, error)
	GetStickerPack(ctx context.Context, packId uuid.UUID) (models.StickerPack, error)
	GetStickerPackByName(ctx context.Context, packName string) (models.StickerPack, error)
	GetStickerPacks(ctx context.Context, userId uuid.UUID, count, offset int) ([]models.StickerPack, error)
	DeleteStickerPack(ctx context.Context, userId, packId uuid.UUID) error
}

type StickerServiceServer struct {
	pb.UnimplementedStickerServiceServer
	stickerUseCase StickerServiceUseCase
}

func NewStickerServiceServer(stickerUseCase StickerServiceUseCase) *StickerServiceServer {
	return &StickerServiceServer{stickerUseCase: stickerUseCase}
}

func (s *StickerServiceServer) AddStickerPack(ctx context.Context, req *pb.AddStickerPackRequest) (*pb.AddStickerPackResponse, error) {
	logger.Info(ctx, "Received AddStickerPack request")

	stickerPack, err := dto.MapProtoToStickerPack(req.StickerPack)
	if err != nil {
		logger.Error(ctx, "Failed to map Proto to StickerPack: %v", err)
		return nil, fmt.Errorf("failed to map Proto to StickerPack: %v", err)
	}

	createdStickerPack, err := s.stickerUseCase.AddStickerPack(ctx, stickerPack)
	if err != nil {
		logger.Error(ctx, "Failed to add sticker pack: %v", err)
		return nil, fmt.Errorf("failed to add sticker pack: %w", err)
	}

	return &pb.AddStickerPackResponse{
		StickerPack: dto.MapStickerPackToProto(createdStickerPack),
	}, nil
}

func (s *StickerServiceServer) GetStickerPack(ctx context.Context, req *pb.GetStickerPackRequest) (*pb.GetStickerPackResponse, error) {
	logger.Info(ctx, "Received GetStickerPack request")

	packId, err := uuid.Parse(req.Id)
	if err != nil {
		logger.Error(ctx, "Invalid sticker pack ID: %v", err)
		return nil, fmt.Errorf("invalid sticker pack ID: %w", err)
	}

	stickerPack, err := s.stickerUseCase.GetStickerPack(ctx, packId)
	if err != nil {
		logger.Error(ctx, "Failed to get sticker pack: %v", err)
		return nil, fmt.Errorf("failed to get sticker pack: %w", err)
	}

	return &pb.GetStickerPackResponse{
		StickerPack: dto.MapStickerPackToProto(&stickerPack),
	}, nil
}

func (s *StickerServiceServer) GetStickerPacks(ctx context.Context, req *pb.GetStickerPacksRequest) (*pb.GetStickerPacksResponse, error) {
	logger.Info(ctx, "Received GetStickerPacks request")

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: %v", err)
		return nil, fmt.Errorf("invalid UserId: %w", err)
	}

	stickerPacks, err := s.stickerUseCase.GetStickerPacks(ctx, userId, int(req.Count), int(req.Offset))
	if err != nil {
		logger.Error(ctx, "Failed to get sticker packs: %v", err)
		return nil, fmt.Errorf("failed to get sticker packs: %w", err)
	}

	var protoStickerPacks []*pb.StickerPack
	for _, pack := range stickerPacks {
		protoStickerPacks = append(protoStickerPacks, dto.MapStickerPackToProto(&pack))
	}

	return &pb.GetStickerPacksResponse{
		StickerPacks: protoStickerPacks,
	}, nil
}

func (s *StickerServiceServer) DeleteStickerPack(ctx context.Context, req *pb.DeleteStickerPackRequest) (*pb.DeleteStickerPackResponse, error) {
	logger.Info(ctx, "Received DeleteStickerPack request")

	packId, err := uuid.Parse(req.PackId)
	if err != nil {
		logger.Error(ctx, "Invalid sticker pack ID: %v", err)
		return nil, fmt.Errorf("invalid sticker pack ID: %w", err)
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.Error(ctx, "Invalid UserId: %v", err)
		return nil, fmt.Errorf("invalid UserId: %w", err)
	}

	err = s.stickerUseCase.DeleteStickerPack(ctx, userId, packId)
	if err != nil {
		logger.Error(ctx, "Failed to delete sticker pack: %v", err)
		return nil, fmt.Errorf("failed to delete sticker pack: %w", err)
	}

	return &pb.DeleteStickerPackResponse{
		Success: true,
	}, nil
}

func (s *StickerServiceServer) GetStickerPackByName(ctx context.Context, req *pb.GetStickerPackByNameRequest) (*pb.GetStickerPackByNameResponse, error) {
	logger.Info(ctx, "Received GetStickerPackByName request")

	if len(req.Name) == 0 {
		logger.Error(ctx, "Name is required")
		return nil, fmt.Errorf("name is required")
	}

	stickerPack, err := s.stickerUseCase.GetStickerPackByName(ctx, req.Name)
	if err != nil {
		logger.Error(ctx, "Failed to get sticker pack by name: %v", err)
		return nil, fmt.Errorf("failed to get sticker pack by name: %w", err)
	}

	return &pb.GetStickerPackByNameResponse{
		StickerPack: dto.MapStickerPackToProto(&stickerPack),
	}, nil
}
