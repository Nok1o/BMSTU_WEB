//go:build unit
// +build unit

package community_service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/community_service"
)

func TestMapContactInfoToDTO(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.ContactInfo
		want    *pb.ContactInfo
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "complete contact info",
			input: &models.ContactInfo{
				Email: "test@example.com",
				Phone: "+123456789",
				City:  "New York",
			},
			want: &pb.ContactInfo{
				Email:       "test@example.com",
				PhoneNumber: "+123456789",
				City:        "New York",
			},
			wantNil: false,
		},
		{
			name: "partial contact info",
			input: &models.ContactInfo{
				Email: "test@example.com",
				City:  "New York",
			},
			want: &pb.ContactInfo{
				Email:       "test@example.com",
				PhoneNumber: "",
				City:        "New York",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapContactInfoToDTO(tt.input)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMapContactInfoDTOToModel(t *testing.T) {
	tests := []struct {
		name    string
		input   *pb.ContactInfo
		want    *models.ContactInfo
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "complete contact info",
			input: &pb.ContactInfo{
				Email:       "test@example.com",
				PhoneNumber: "+123456789",
				City:        "New York",
			},
			want: &models.ContactInfo{
				Email: "test@example.com",
				Phone: "+123456789",
				City:  "New York",
			},
			wantNil: false,
		},
		{
			name: "partial contact info",
			input: &pb.ContactInfo{
				Email: "test@example.com",
				City:  "New York",
			},
			want: &models.ContactInfo{
				Email: "test@example.com",
				Phone: "",
				City:  "New York",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapContactInfoDTOToModel(tt.input)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMapProtoCommunityToModel(t *testing.T) {
	validID := uuid.New().String()
	validOwnerID := uuid.New().String()
	now := timestamppb.Now()
	invalidID := "invalid-uuid"

	tests := []struct {
		name    string
		input   *pb.Community
		want    *models.Community
		wantErr bool
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "valid community",
			input: &pb.Community{
				Id:          validID,
				OwnerId:     validOwnerID,
				Name:        "Test Community",
				Description: "Test Description",
				CreatedAt:   now,
				AvatarUrl:   "http://example.com/avatar.jpg",
				CoverUrl:    "http://example.com/cover.jpg",
				Nickname:    "test-comm",
				ContactInfo: &pb.ContactInfo{
					Email:       "test@example.com",
					PhoneNumber: "+123456789",
					City:        "New York",
				},
			},
			want: &models.Community{
				ID:        uuid.MustParse(validID),
				OwnerID:   uuid.MustParse(validOwnerID),
				NickName:  "test-comm",
				CreatedAt: now.AsTime(),
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "Test Description",
					AvatarUrl:   "http://example.com/avatar.jpg",
					CoverUrl:    "http://example.com/cover.jpg",
				},
				ContactInfo: &models.ContactInfo{
					Email: "test@example.com",
					Phone: "+123456789",
					City:  "New York",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid community id",
			input: &pb.Community{
				Id:      invalidID,
				OwnerId: validOwnerID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid owner id",
			input: &pb.Community{
				Id:      validID,
				OwnerId: invalidID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "minimal community",
			input: &pb.Community{
				Id:      validID,
				OwnerId: validOwnerID,
				Name:    "Minimal Community",
			},
			want: &models.Community{
				ID:       uuid.MustParse(validID),
				OwnerID:  uuid.MustParse(validOwnerID),
				NickName: "",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Minimal Community",
					Description: "",
					AvatarUrl:   "",
					CoverUrl:    "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapProtoCommunityToModel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.OwnerID, got.OwnerID)
			assert.Equal(t, tt.want.NickName, got.NickName)
			assert.Equal(t, tt.want.BasicInfo.Name, got.BasicInfo.Name)
			assert.Equal(t, tt.want.BasicInfo.Description, got.BasicInfo.Description)

			if tt.input.ContactInfo != nil {
				assert.Equal(t, tt.want.ContactInfo.Email, got.ContactInfo.Email)
				assert.Equal(t, tt.want.ContactInfo.Phone, got.ContactInfo.Phone)
				assert.Equal(t, tt.want.ContactInfo.City, got.ContactInfo.City)
			}
		})
	}
}

func TestMapModelCommunityToProto(t *testing.T) {
	validID := uuid.New()
	validOwnerID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		input   *models.Community
		want    *pb.Community
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "complete community",
			input: &models.Community{
				ID:        validID,
				OwnerID:   validOwnerID,
				NickName:  "test-comm",
				CreatedAt: now,
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "Test Description",
					AvatarUrl:   "http://example.com/avatar.jpg",
					CoverUrl:    "http://example.com/cover.jpg",
				},
				ContactInfo: &models.ContactInfo{
					Email: "test@example.com",
					Phone: "+123456789",
					City:  "New York",
				},
			},
			want: &pb.Community{
				Id:          validID.String(),
				OwnerId:     validOwnerID.String(),
				Name:        "Test Community",
				Description: "Test Description",
				CreatedAt:   timestamppb.New(now),
				AvatarUrl:   "http://example.com/avatar.jpg",
				CoverUrl:    "http://example.com/cover.jpg",
				Nickname:    "test-comm",
				ContactInfo: &pb.ContactInfo{
					Email:       "test@example.com",
					PhoneNumber: "+123456789",
					City:        "New York",
				},
			},
			wantNil: false,
		},
		{
			name: "minimal community",
			input: &models.Community{
				ID:      validID,
				OwnerID: validOwnerID,
				BasicInfo: &models.BasicCommunityInfo{
					Name: "Minimal Community",
				},
			},
			want: &pb.Community{
				Id:      validID.String(),
				OwnerId: validOwnerID.String(),
				Name:    "Minimal Community",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapModelCommunityToProto(tt.input)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.OwnerId, got.OwnerId)
			assert.Equal(t, tt.want.Name, got.Name)

			if tt.input.BasicInfo.Description != "" {
				assert.Equal(t, tt.want.Description, got.Description)
			}
			if tt.input.ContactInfo != nil {
				assert.Equal(t, tt.want.ContactInfo.Email, got.ContactInfo.Email)
				assert.Equal(t, tt.want.ContactInfo.PhoneNumber, got.ContactInfo.PhoneNumber)
				assert.Equal(t, tt.want.ContactInfo.City, got.ContactInfo.City)
			}
		})
	}
}

func TestMapModelMemberToProto(t *testing.T) {
	userID := uuid.New()
	communityID := uuid.New()
	joinedAt := time.Now()

	tests := []struct {
		name    string
		input   *models.CommunityMember
		want    *pb.CommunityMember
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "complete member",
			input: &models.CommunityMember{
				UserID:      userID,
				CommunityID: communityID,
				Role:        models.CommunityRoleAdmin,
				JoinedAt:    joinedAt,
			},
			want: &pb.CommunityMember{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
				Role:        pb.CommunityRole_COMMUNITY_ROLE_ADMIN,
				JoinedAt:    timestamppb.New(joinedAt),
			},
			wantNil: false,
		},
		{
			name: "member with default role",
			input: &models.CommunityMember{
				UserID:      userID,
				CommunityID: communityID,
				JoinedAt:    joinedAt,
			},
			want: &pb.CommunityMember{
				UserId:      userID.String(),
				CommunityId: communityID.String(),
				Role:        pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
				JoinedAt:    timestamppb.New(joinedAt),
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapModelMemberToProto(tt.input)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.want.UserId, got.UserId)
			assert.Equal(t, tt.want.CommunityId, got.CommunityId)
			assert.Equal(t, tt.want.Role, got.Role)
			assert.Equal(t, tt.want.JoinedAt.AsTime(), got.JoinedAt.AsTime())
		})
	}
}

func TestMapProtoMemberToModel(t *testing.T) {
	validUserID := uuid.New().String()
	validCommunityID := uuid.New().String()
	invalidID := "invalid-uuid"
	joinedAt := timestamppb.Now()

	tests := []struct {
		name    string
		input   *pb.CommunityMember
		want    *models.CommunityMember
		wantErr bool
		wantNil bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantNil: true,
		},
		{
			name: "valid member",
			input: &pb.CommunityMember{
				UserId:      validUserID,
				CommunityId: validCommunityID,
				Role:        pb.CommunityRole_COMMUNITY_ROLE_ADMIN,
				JoinedAt:    joinedAt,
			},
			want: &models.CommunityMember{
				UserID:      uuid.MustParse(validUserID),
				CommunityID: uuid.MustParse(validCommunityID),
				Role:        models.CommunityRoleAdmin,
				JoinedAt:    joinedAt.AsTime(),
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			input: &pb.CommunityMember{
				UserId:      invalidID,
				CommunityId: validCommunityID,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid community id",
			input: &pb.CommunityMember{
				UserId:      validUserID,
				CommunityId: invalidID,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapProtoMemberToModel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.want.UserID, got.UserID)
			assert.Equal(t, tt.want.CommunityID, got.CommunityID)
			assert.Equal(t, tt.want.Role, got.Role)
			assert.Equal(t, tt.want.JoinedAt, got.JoinedAt)
		})
	}
}

func TestConvertRoleToProto(t *testing.T) {
	tests := []struct {
		name  string
		input models.CommunityRole
		want  pb.CommunityRole
	}{
		{
			name:  "member",
			input: models.CommunityRoleMember,
			want:  pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
		},
		{
			name:  "admin",
			input: models.CommunityRoleAdmin,
			want:  pb.CommunityRole_COMMUNITY_ROLE_ADMIN,
		},
		{
			name:  "owner",
			input: models.CommunityRoleOwner,
			want:  pb.CommunityRole_COMMUNITY_ROLE_OWNER,
		},
		{
			name:  "unknown",
			input: models.CommunityRole(rune(999)),
			want:  pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertRoleToProto(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertRoleFromProto(t *testing.T) {
	tests := []struct {
		name  string
		input pb.CommunityRole
		want  models.CommunityRole
	}{
		{
			name:  "member",
			input: pb.CommunityRole_COMMUNITY_ROLE_MEMBER,
			want:  models.CommunityRoleMember,
		},
		{
			name:  "admin",
			input: pb.CommunityRole_COMMUNITY_ROLE_ADMIN,
			want:  models.CommunityRoleAdmin,
		},
		{
			name:  "owner",
			input: pb.CommunityRole_COMMUNITY_ROLE_OWNER,
			want:  models.CommunityRoleOwner,
		},
		{
			name:  "unknown",
			input: pb.CommunityRole(999),
			want:  models.CommunityRoleMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertRoleFromProto(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
