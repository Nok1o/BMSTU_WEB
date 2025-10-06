//go:build unit
// +build unit

package userclient

import (
	proto "quickflow/shared/proto/file_service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	shared_models "quickflow/shared/models"
	db "quickflow/shared/proto/user_service"
)

func TestMapSchoolEducation(t *testing.T) {
	tests := []struct {
		name     string
		input    *shared_models.SchoolEducation
		expected *db.SchoolEducation
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.SchoolEducation{
				City:   "Moscow",
				School: "School 123",
			},
			expected: &db.SchoolEducation{
				City: "Moscow",
				Name: "School 123",
			},
		},
		{
			name: "empty fields",
			input: &shared_models.SchoolEducation{
				City:   "",
				School: "",
			},
			expected: &db.SchoolEducation{
				City: "",
				Name: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapSchoolEducationToDTO(tt.input)
			assert.Equal(t, tt.expected, resultDTO)

			// ToModel (only if not nil)
			if tt.input != nil {
				resultModel := MapSchoolEducationDTOToModel(resultDTO)
				assert.Equal(t, tt.input, resultModel)
			}
		})
	}
}

func TestMapUniversityEducation(t *testing.T) {
	tests := []struct {
		name     string
		input    *shared_models.UniversityEducation
		expected *db.UniversityEducation
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.UniversityEducation{
				City:           "St. Petersburg",
				University:     "SPbSU",
				Faculty:        "Mathematics",
				GraduationYear: 2020,
			},
			expected: &db.UniversityEducation{
				City:           "St. Petersburg",
				University:     "SPbSU",
				Faculty:        "Mathematics",
				GraduationYear: 2020,
			},
		},
		{
			name: "empty fields",
			input: &shared_models.UniversityEducation{
				City:           "",
				University:     "",
				Faculty:        "",
				GraduationYear: 0,
			},
			expected: &db.UniversityEducation{
				City:           "",
				University:     "",
				Faculty:        "",
				GraduationYear: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapUniversityEducationToDTO(tt.input)
			assert.Equal(t, tt.expected, resultDTO)

			// ToModel (only if not nil)
			if tt.input != nil {
				resultModel := MapUniversityEducationDTOToModel(resultDTO)
				assert.Equal(t, tt.input, resultModel)
			}
		})
	}
}

func TestMapContactInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    *shared_models.ContactInfo
		expected *db.ContactInfo
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.ContactInfo{
				Email: "test@example.com",
				Phone: "+1234567890",
				City:  "New York",
			},
			expected: &db.ContactInfo{
				Email:       "test@example.com",
				PhoneNumber: "+1234567890",
				City:        "New York",
			},
		},
		{
			name: "empty fields",
			input: &shared_models.ContactInfo{
				Email: "",
				Phone: "",
				City:  "",
			},
			expected: &db.ContactInfo{
				Email:       "",
				PhoneNumber: "",
				City:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapContactInfoToDTO(tt.input)
			assert.Equal(t, tt.expected, resultDTO)

			// ToModel (only if not nil)
			if tt.input != nil {
				resultModel := MapContactInfoDTOToModel(resultDTO)
				assert.Equal(t, tt.input, resultModel)
			}
		})
	}
}

func TestMapBasicInfo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		input    *shared_models.BasicInfo
		expected *db.BasicInfo
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.BasicInfo{
				Name:          "John",
				Surname:       "Doe",
				Sex:           shared_models.MALE,
				DateOfBirth:   now,
				Bio:           "Software developer",
				AvatarUrl:     "avatar.jpg",
				BackgroundUrl: "cover.jpg",
			},
			expected: &db.BasicInfo{
				Firstname: "John",
				Lastname:  "Doe",
				Sex:       int32(shared_models.MALE),
				BirthDate: timestamppb.New(now),
				Bio:       "Software developer",
				AvatarUrl: "avatar.jpg",
				CoverUrl:  "cover.jpg",
			},
		},
		{
			name: "empty fields",
			input: &shared_models.BasicInfo{
				Name:          "",
				Surname:       "",
				Sex:           0,
				DateOfBirth:   time.Time{},
				Bio:           "",
				AvatarUrl:     "",
				BackgroundUrl: "",
			},
			expected: &db.BasicInfo{
				Firstname: "",
				Lastname:  "",
				Sex:       0,
				BirthDate: timestamppb.New(time.Time{}),
				Bio:       "",
				AvatarUrl: "",
				CoverUrl:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapBasicInfoToDTO(tt.input)
			if tt.input == nil {
				assert.Nil(t, resultDTO)
			} else {
				assert.Equal(t, tt.expected.Firstname, resultDTO.Firstname)
				assert.Equal(t, tt.expected.Lastname, resultDTO.Lastname)
				assert.Equal(t, tt.expected.Sex, resultDTO.Sex)
				assert.Equal(t, tt.expected.Bio, resultDTO.Bio)
				assert.Equal(t, tt.expected.AvatarUrl, resultDTO.AvatarUrl)
				assert.Equal(t, tt.expected.CoverUrl, resultDTO.CoverUrl)
				assert.True(t, tt.expected.BirthDate.AsTime().Equal(resultDTO.BirthDate.AsTime()))
			}

			// ToModel (only if not nil)
			if tt.input != nil {
				resultModel := MapBasicInfoDTOToModel(resultDTO)
				assert.Equal(t, tt.input.Name, resultModel.Name)
				assert.Equal(t, tt.input.Surname, resultModel.Surname)
				assert.Equal(t, tt.input.Sex, resultModel.Sex)
				assert.Equal(t, tt.input.Bio, resultModel.Bio)
				assert.Equal(t, tt.input.AvatarUrl, resultModel.AvatarUrl)
				assert.Equal(t, tt.input.BackgroundUrl, resultModel.BackgroundUrl)
				assert.True(t, tt.input.DateOfBirth.Equal(resultModel.DateOfBirth))
			}
		})
	}
}

func TestMapProfile(t *testing.T) {
	now := time.Now()
	userId := uuid.New()

	tests := []struct {
		name     string
		input    *shared_models.Profile
		expected *db.Profile
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.Profile{
				UserId:   userId,
				Username: "johndoe",
				BasicInfo: &shared_models.BasicInfo{
					Name:          "John",
					Surname:       "Doe",
					Sex:           shared_models.MALE,
					DateOfBirth:   now,
					Bio:           "Bio",
					AvatarUrl:     "avatar.jpg",
					BackgroundUrl: "cover.jpg",
				},
				ContactInfo: &shared_models.ContactInfo{
					Email: "john@example.com",
					Phone: "+1234567890",
					City:  "New York",
				},
				SchoolEducation: &shared_models.SchoolEducation{
					City:   "NY",
					School: "School 1",
				},
				UniversityEducation: &shared_models.UniversityEducation{
					City:           "Boston",
					University:     "MIT",
					Faculty:        "CS",
					GraduationYear: 2020,
				},
				Avatar: &shared_models.File{
					URL: "path/to/avatar",
				},
				Background: &shared_models.File{
					URL: "path/to/background",
				},
				LastSeen: now,
			},
			expected: &db.Profile{
				Id:       userId.String(),
				Username: "johndoe",
				BasicInfo: &db.BasicInfo{
					Firstname: "John",
					Lastname:  "Doe",
					Sex:       int32(shared_models.MALE),
					BirthDate: timestamppb.New(now),
					Bio:       "Bio",
					AvatarUrl: "avatar.jpg",
					CoverUrl:  "cover.jpg",
				},
				ContactInfo: &db.ContactInfo{
					Email:       "john@example.com",
					PhoneNumber: "+1234567890",
					City:        "New York",
				},
				SchoolEducation: &db.SchoolEducation{
					City: "NY",
					Name: "School 1",
				},
				UniversityEducation: &db.UniversityEducation{
					City:           "Boston",
					University:     "MIT",
					Faculty:        "CS",
					GraduationYear: 2020,
				},
				Avatar: &proto.File{
					Url: "path/to/avatar",
				},
				Cover: &proto.File{
					Url: "path/to/background",
				},
				LastSeen: timestamppb.New(now),
			},
		},
		{
			name: "partial data",
			input: &shared_models.Profile{
				UserId:   userId,
				Username: "johndoe",
				BasicInfo: &shared_models.BasicInfo{
					Name:    "John",
					Surname: "Doe",
				},
				LastSeen: now,
			},
			expected: &db.Profile{
				Id:       userId.String(),
				Username: "johndoe",
				BasicInfo: &db.BasicInfo{
					Firstname: "John",
					Lastname:  "Doe",
				},
				LastSeen: timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapProfileToProfileDTO(tt.input)
			if tt.input == nil {
				assert.Nil(t, resultDTO)
				return
			}

			assert.Equal(t, tt.expected.Id, resultDTO.Id)
			assert.Equal(t, tt.expected.Username, resultDTO.Username)
			assert.Equal(t, tt.expected.BasicInfo.Firstname, resultDTO.BasicInfo.Firstname)
			assert.Equal(t, tt.expected.BasicInfo.Lastname, resultDTO.BasicInfo.Lastname)
			if tt.input.ContactInfo != nil {
				assert.Equal(t, tt.expected.ContactInfo.Email, resultDTO.ContactInfo.Email)
			}
			if tt.input.SchoolEducation != nil {
				assert.Equal(t, tt.expected.SchoolEducation.Name, resultDTO.SchoolEducation.Name)
			}
			if tt.input.UniversityEducation != nil {
				assert.Equal(t, tt.expected.UniversityEducation.University, resultDTO.UniversityEducation.University)
			}
			assert.True(t, tt.expected.LastSeen.AsTime().Equal(resultDTO.LastSeen.AsTime()))

			// ToModel
			resultModel, err := MapProfileDTOToProfile(resultDTO)
			assert.NoError(t, err)
			assert.Equal(t, tt.input.UserId, resultModel.UserId)
			assert.Equal(t, tt.input.Username, resultModel.Username)
			if tt.input.BasicInfo != nil {
				assert.Equal(t, tt.input.BasicInfo.Name, resultModel.BasicInfo.Name)
			}
			if tt.input.ContactInfo != nil {
				assert.Equal(t, tt.input.ContactInfo.Email, resultModel.ContactInfo.Email)
			}
			if tt.input.SchoolEducation != nil {
				assert.Equal(t, tt.input.SchoolEducation.School, resultModel.SchoolEducation.School)
			}
			if tt.input.UniversityEducation != nil {
				assert.Equal(t, tt.input.UniversityEducation.University, resultModel.UniversityEducation.University)
			}
			assert.True(t, tt.input.LastSeen.Equal(resultModel.LastSeen))
		})
	}
}

func TestMapPublicUserInfo(t *testing.T) {
	now := time.Now()
	userId := uuid.New()

	tests := []struct {
		name     string
		input    *shared_models.PublicUserInfo
		expected *db.PublicUserInfo
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full data",
			input: &shared_models.PublicUserInfo{
				Id:        userId,
				Username:  "johndoe",
				Firstname: "John",
				Lastname:  "Doe",
				AvatarURL: "avatar.jpg",
				LastSeen:  now,
			},
			expected: &db.PublicUserInfo{
				Id:        userId.String(),
				Username:  "johndoe",
				Firstname: "John",
				Lastname:  "Doe",
				AvatarUrl: "avatar.jpg",
				LastSeen:  timestamppb.New(now),
			},
		},
		{
			name: "empty fields",
			input: &shared_models.PublicUserInfo{
				Id:        userId,
				Username:  "",
				Firstname: "",
				Lastname:  "",
				AvatarURL: "",
				LastSeen:  time.Time{},
			},
			expected: &db.PublicUserInfo{
				Id:        userId.String(),
				Username:  "",
				Firstname: "",
				Lastname:  "",
				AvatarUrl: "",
				LastSeen:  timestamppb.New(time.Time{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ToDTO
			resultDTO := MapPublicUserInfoToDTO(tt.input)
			assert.Equal(t, tt.expected, resultDTO)

			// ToModel (only if not nil)
			if tt.input != nil {
				resultModel, err := MapPublicUserInfoDTOToModel(resultDTO)
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Id, resultModel.Id)
				assert.Equal(t, tt.input.Username, resultModel.Username)
				assert.Equal(t, tt.input.Firstname, resultModel.Firstname)
				assert.Equal(t, tt.input.Lastname, resultModel.Lastname)
				assert.Equal(t, tt.input.AvatarURL, resultModel.AvatarURL)
				assert.True(t, tt.input.LastSeen.Equal(resultModel.LastSeen))
			}
		})
	}
}

func TestMapProfileDTOToProfile_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       *db.Profile
		expectedErr string
	}{
		{
			name: "invalid uuid",
			input: &db.Profile{
				Id: "invalid-uuid",
			},
			expectedErr: "invalid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MapProfileDTOToProfile(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestMapPublicUserInfoDTOToModel_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       *db.PublicUserInfo
		expectedErr string
	}{
		{
			name: "invalid uuid",
			input: &db.PublicUserInfo{
				Id: "invalid-uuid",
			},
			expectedErr: "invalid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MapPublicUserInfoDTOToModel(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
