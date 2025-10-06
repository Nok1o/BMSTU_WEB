//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/messenger_service/internal/delivery/grpc/mocks"
	dto "quickflow/shared/client/messenger_service"
	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
)

func TestAddStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockStickerServiceUseCase(ctrl)
	server := NewStickerServiceServer(mockUseCase)

	ctx := context.Background()
	now := time.Now()
	testPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := dto.MapStickerPackToProto(testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		req         *pb.AddStickerPackRequest
		wantResp    *pb.AddStickerPackResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					AddStickerPack(ctx, gomock.Any()).
					Return(testPack, nil)
			},
			req: &pb.AddStickerPackRequest{
				StickerPack: testProtoPack,
			},
			wantResp: &pb.AddStickerPackResponse{
				StickerPack: testProtoPack,
			},
		},
		{
			name: "Mapping Error",
			mockSetup: func() {
				// No mock setup needed as error occurs before usecase call
			},
			req: &pb.AddStickerPackRequest{
				StickerPack: &pb.StickerPack{Id: "invalid-id"},
			},
			wantErr: true,
		},
		{
			name: "UseCase Error",
			mockSetup: func() {
				mockUseCase.EXPECT().
					AddStickerPack(ctx, gomock.Any()).
					Return(nil, errors.New("usecase error"))
			},
			req: &pb.AddStickerPackRequest{
				StickerPack: testProtoPack,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := server.AddStickerPack(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

func TestGetStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockStickerServiceUseCase(ctrl)
	server := NewStickerServiceServer(mockUseCase)

	ctx := context.Background()
	now := time.Now()
	testPack := models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := dto.MapStickerPackToProto(&testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		req         *pb.GetStickerPackRequest
		wantResp    *pb.GetStickerPackResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPack(ctx, testPack.Id).
					Return(testPack, nil)
			},
			req: &pb.GetStickerPackRequest{
				Id: testPack.Id.String(),
			},
			wantResp: &pb.GetStickerPackResponse{
				StickerPack: testProtoPack,
			},
		},
		{
			name: "Invalid Pack ID",
			req: &pb.GetStickerPackRequest{
				Id: "invalid-id",
			},
			wantErr: true,
		},
		{
			name: "UseCase Error",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPack(ctx, testPack.Id).
					Return(models.StickerPack{}, errors.New("usecase error"))
			},
			req: &pb.GetStickerPackRequest{
				Id: testPack.Id.String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := server.GetStickerPack(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

func TestGetStickerPacks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockStickerServiceUseCase(ctrl)
	server := NewStickerServiceServer(mockUseCase)

	ctx := context.Background()
	now := time.Now()
	testUserId := uuid.New()
	testPack := models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: testUserId,
		CreatedAt: now,
		UpdatedAt: now,
	}
	testPacks := []models.StickerPack{testPack}
	testProtoPacks := []*pb.StickerPack{dto.MapStickerPackToProto(&testPack)}

	tests := []struct {
		name        string
		mockSetup   func()
		req         *pb.GetStickerPacksRequest
		wantResp    *pb.GetStickerPacksResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPacks(ctx, testUserId, 10, 0).
					Return(testPacks, nil)
			},
			req: &pb.GetStickerPacksRequest{
				UserId: testUserId.String(),
				Count:  10,
				Offset: 0,
			},
			wantResp: &pb.GetStickerPacksResponse{
				StickerPacks: testProtoPacks,
			},
		},
		{
			name: "Invalid User ID",
			req: &pb.GetStickerPacksRequest{
				UserId: "invalid-id",
				Count:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "UseCase Error",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPacks(ctx, testUserId, 10, 0).
					Return(nil, errors.New("usecase error"))
			},
			req: &pb.GetStickerPacksRequest{
				UserId: testUserId.String(),
				Count:  10,
				Offset: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := server.GetStickerPacks(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

func TestDeleteStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockStickerServiceUseCase(ctrl)
	server := NewStickerServiceServer(mockUseCase)

	ctx := context.Background()
	testUserId := uuid.New()
	testPackId := uuid.New()

	tests := []struct {
		name        string
		mockSetup   func()
		req         *pb.DeleteStickerPackRequest
		wantResp    *pb.DeleteStickerPackResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					DeleteStickerPack(ctx, testUserId, testPackId).
					Return(nil)
			},
			req: &pb.DeleteStickerPackRequest{
				UserId: testUserId.String(),
				PackId: testPackId.String(),
			},
			wantResp: &pb.DeleteStickerPackResponse{
				Success: true,
			},
		},
		{
			name: "Invalid User ID",
			req: &pb.DeleteStickerPackRequest{
				UserId: "invalid-id",
				PackId: testPackId.String(),
			},
			wantErr: true,
		},
		{
			name: "Invalid Pack ID",
			req: &pb.DeleteStickerPackRequest{
				UserId: testUserId.String(),
				PackId: "invalid-id",
			},
			wantErr: true,
		},
		{
			name: "UseCase Error",
			mockSetup: func() {
				mockUseCase.EXPECT().
					DeleteStickerPack(ctx, testUserId, testPackId).
					Return(errors.New("usecase error"))
			},
			req: &pb.DeleteStickerPackRequest{
				UserId: testUserId.String(),
				PackId: testPackId.String(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := server.DeleteStickerPack(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

func TestGetStickerPackByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockStickerServiceUseCase(ctrl)
	server := NewStickerServiceServer(mockUseCase)

	ctx := context.Background()
	now := time.Now()
	testPack := models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := dto.MapStickerPackToProto(&testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		req         *pb.GetStickerPackByNameRequest
		wantResp    *pb.GetStickerPackByNameResponse
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPackByName(ctx, testPack.Name).
					Return(testPack, nil)
			},
			req: &pb.GetStickerPackByNameRequest{
				Name: testPack.Name,
			},
			wantResp: &pb.GetStickerPackByNameResponse{
				StickerPack: testProtoPack,
			},
		},
		{
			name: "Empty Name",
			req: &pb.GetStickerPackByNameRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "UseCase Error",
			mockSetup: func() {
				mockUseCase.EXPECT().
					GetStickerPackByName(ctx, testPack.Name).
					Return(models.StickerPack{}, errors.New("usecase error"))
			},
			req: &pb.GetStickerPackByNameRequest{
				Name: testPack.Name,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := server.GetStickerPackByName(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}
