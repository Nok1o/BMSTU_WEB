//go:build unit
// +build unit

package messenger_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/messenger_service"
	"quickflow/shared/proto/messenger_service/mocks"
)

func TestAddStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockStickerServiceClient(ctrl)
	client := &StickerServiceClient{client: mockClient}

	ctx := context.Background()
	now := time.Now()
	testPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := MapStickerPackToProto(testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		wantResp    *models.StickerPack
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					AddStickerPack(ctx, &pb.AddStickerPackRequest{
						StickerPack: testProtoPack,
					}).
					Return(&pb.AddStickerPackResponse{
						StickerPack: testProtoPack,
					}, nil)
			},
			wantResp: testPack,
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().
					AddStickerPack(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			wantErr: true,
		},
		{
			name: "Mapping Error",
			mockSetup: func() {
				mockClient.EXPECT().
					AddStickerPack(ctx, gomock.Any()).
					Return(&pb.AddStickerPackResponse{
						StickerPack: &pb.StickerPack{Id: "invalid-id"},
					}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := client.AddStickerPack(ctx, testPack)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.Name, resp.Name)
				assert.Equal(t, tt.wantResp.CreatorId, resp.CreatorId)
				assert.Equal(t, tt.wantResp.Id, resp.Id)
			}
		})
	}
}

func TestGetStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockStickerServiceClient(ctrl)
	client := &StickerServiceClient{client: mockClient}

	ctx := context.Background()
	now := time.Now()
	testPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := MapStickerPackToProto(testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		packId      uuid.UUID
		wantResp    *models.StickerPack
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPack(ctx, &pb.GetStickerPackRequest{
						Id: testPack.Id.String(),
					}).
					Return(&pb.GetStickerPackResponse{
						StickerPack: testProtoPack,
					}, nil)
			},
			packId:   testPack.Id,
			wantResp: testPack,
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPack(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			packId:  testPack.Id,
			wantErr: true,
		},
		{
			name: "Mapping Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPack(ctx, gomock.Any()).
					Return(&pb.GetStickerPackResponse{
						StickerPack: &pb.StickerPack{Id: "invalid-id"},
					}, nil)
			},
			packId:  testPack.Id,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := client.GetStickerPack(ctx, tt.packId)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.Name, resp.Name)
				assert.Equal(t, tt.wantResp.CreatorId, resp.CreatorId)
				assert.Equal(t, tt.wantResp.Id, resp.Id)
			}
		})
	}
}

func TestGetStickerPacks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockStickerServiceClient(ctrl)
	client := &StickerServiceClient{client: mockClient}

	ctx := context.Background()
	now := time.Now()
	testUserId := uuid.New()
	testPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: testUserId,
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := MapStickerPackToProto(testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		userId      uuid.UUID
		count       int
		offset      int
		wantResp    []*models.StickerPack
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPacks(ctx, &pb.GetStickerPacksRequest{
						UserId: testUserId.String(),
						Count:  10,
						Offset: 0,
					}).
					Return(&pb.GetStickerPacksResponse{
						StickerPacks: []*pb.StickerPack{testProtoPack},
					}, nil)
			},
			userId:   testUserId,
			count:    10,
			offset:   0,
			wantResp: []*models.StickerPack{testPack},
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPacks(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			userId:  testUserId,
			count:   10,
			offset:  0,
			wantErr: true,
		},
		{
			name: "Mapping Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPacks(ctx, gomock.Any()).
					Return(&pb.GetStickerPacksResponse{
						StickerPacks: []*pb.StickerPack{{Id: "invalid-id"}},
					}, nil)
			},
			userId:  testUserId,
			count:   10,
			offset:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := client.GetStickerPacks(ctx, tt.userId, tt.count, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
				for i := range tt.wantResp {
					assert.Equal(t, tt.wantResp[i].Name, resp[i].Name)
					assert.Equal(t, tt.wantResp[i].CreatorId, resp[i].CreatorId)
					assert.Equal(t, tt.wantResp[i].Id, resp[i].Id)
				}
			}
		})
	}
}

func TestDeleteStickerPack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockStickerServiceClient(ctrl)
	client := &StickerServiceClient{client: mockClient}

	ctx := context.Background()
	testUserId := uuid.New()
	testPackId := uuid.New()

	tests := []struct {
		name        string
		mockSetup   func()
		userId      uuid.UUID
		packId      uuid.UUID
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					DeleteStickerPack(ctx, &pb.DeleteStickerPackRequest{
						UserId: testUserId.String(),
						PackId: testPackId.String(),
					}).
					Return(&pb.DeleteStickerPackResponse{
						Success: true,
					}, nil)
			},
			userId: testUserId,
			packId: testPackId,
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().
					DeleteStickerPack(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			userId:  testUserId,
			packId:  testPackId,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := client.DeleteStickerPack(ctx, tt.userId, tt.packId)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStickerPackByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockStickerServiceClient(ctrl)
	client := &StickerServiceClient{client: mockClient}

	ctx := context.Background()
	now := time.Now()
	testPack := &models.StickerPack{
		Id:        uuid.New(),
		Name:      "test-pack",
		CreatorId: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	testProtoPack := MapStickerPackToProto(testPack)

	tests := []struct {
		name        string
		mockSetup   func()
		packName    string
		wantResp    *models.StickerPack
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPackByName(ctx, &pb.GetStickerPackByNameRequest{
						Name: testPack.Name,
					}).
					Return(&pb.GetStickerPackByNameResponse{
						StickerPack: testProtoPack,
					}, nil)
			},
			packName: testPack.Name,
			wantResp: testPack,
		},
		{
			name: "GRPC Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPackByName(ctx, gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			packName: testPack.Name,
			wantErr:  true,
		},
		{
			name: "Mapping Error",
			mockSetup: func() {
				mockClient.EXPECT().
					GetStickerPackByName(ctx, gomock.Any()).
					Return(&pb.GetStickerPackByNameResponse{
						StickerPack: &pb.StickerPack{Id: "invalid-id"},
					}, nil)
			},
			packName: testPack.Name,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			resp, err := client.GetStickerPackByName(ctx, tt.packName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)

				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.Name, resp.Name)
				assert.Equal(t, tt.wantResp.CreatorId, resp.CreatorId)
				assert.Equal(t, tt.wantResp.Id, resp.Id)
			}
		})
	}
}
