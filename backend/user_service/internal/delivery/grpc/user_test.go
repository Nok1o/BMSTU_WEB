//go:build unit
// +build unit

package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/user_service"
)

type mockUserUseCase struct {
	mock.Mock
}

func (m *mockUserUseCase) CreateUser(ctx context.Context, user models.User, profile models.Profile) (uuid.UUID, models.Session, error) {
	args := m.Called(ctx, user, profile)
	return args.Get(0).(uuid.UUID), args.Get(1).(models.Session), args.Error(2)
}

func (m *mockUserUseCase) AuthUser(ctx context.Context, authData models.LoginData) (models.Session, error) {
	args := m.Called(ctx, authData)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *mockUserUseCase) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUserUseCase) LookupUserSession(ctx context.Context, session models.Session) (models.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUserUseCase) DeleteUserSession(ctx context.Context, session string) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockUserUseCase) GetUserById(ctx context.Context, userId uuid.UUID) (models.User, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *mockUserUseCase) SearchSimilarUser(ctx context.Context, toSearch string, usersCount uint) ([]models.PublicUserInfo, error) {
	args := m.Called(ctx, toSearch, usersCount)
	return args.Get(0).([]models.PublicUserInfo), args.Error(1)
}

func TestUserServiceServer_SignUp(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	expiry := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		req         *pb.SignUpRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.SignUpResponse
		expectedErr error
	}{
		{
			name: "successful signup",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Id:       userID.String(),
					Username: "testuser",
					Password: "password123",
					Salt:     "salt",
				},
				Profile: &pb.Profile{
					Id:       userID.String(),
					Username: "testuser",
					BasicInfo: &pb.BasicInfo{
						Firstname: "testuser",
						Lastname:  "testuser",
						BirthDate: timestamppb.New(expiry),
					},
				},
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).
					Return(userID, models.Session{
						SessionId:  sessionID,
						ExpireDate: expiry,
					}, nil)
			},
			expected: &pb.SignUpResponse{
				Session: &pb.Session{
					Id:     sessionID.String(),
					Expiry: timestamppb.New(expiry),
				},
			},
		},
		{
			name: "invalid user data",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Username: "", // invalid empty username
				},
			},
			mockSetup:   func(m *mockUserUseCase) {},
			expectedErr: status.Error(codes.InvalidArgument, "invalid user data"),
		},
		{
			name: "use case error",
			req: &pb.SignUpRequest{
				User: &pb.User{
					Username: "testuser",
					Password: "password123",
					Salt:     "salt",
				},
				Profile: &pb.Profile{
					BasicInfo: &pb.BasicInfo{
						Firstname: "testuser",
						Lastname:  "testuser",
						BirthDate: timestamppb.New(expiry),
					},
				},
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).
					Return(uuid.Nil, models.Session{}, errors.New("user already exists")).Maybe()
			},
			expectedErr: errors.New("user already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.SignUp(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_SignIn(t *testing.T) {
	sessionID := uuid.New()
	expiry := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		req         *pb.SignInRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.SignInResponse
		expectedErr error
	}{
		{
			name: "successful signin",
			req: &pb.SignInRequest{
				SignIn: &pb.SignIn{
					Username: "testuser",
					Password: "password123",
				},
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("AuthUser", mock.Anything, models.LoginData{
					Username: "testuser",
					Password: "password123",
				}).Return(models.Session{
					SessionId:  sessionID,
					ExpireDate: expiry,
				}, nil)
			},
			expected: &pb.SignInResponse{
				Session: &pb.Session{
					Id:     sessionID.String(),
					Expiry: timestamppb.New(expiry),
				},
			},
		},
		{
			name: "invalid credentials",
			req: &pb.SignInRequest{
				SignIn: &pb.SignIn{
					Username: "testuser",
					Password: "wrongpassword",
				},
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("AuthUser", mock.Anything, mock.Anything).
					Return(models.Session{}, errors.New("invalid credentials"))
			},
			expectedErr: errors.New("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.SignIn(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_SignOut(t *testing.T) {
	sessionID := uuid.New().String()

	tests := []struct {
		name        string
		req         *pb.SignOutRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.SignOutResponse
		expectedErr error
	}{
		{
			name: "successful signout",
			req: &pb.SignOutRequest{
				SessionId: sessionID,
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("DeleteUserSession", mock.Anything, sessionID).
					Return(nil)
			},
			expected: &pb.SignOutResponse{Success: true},
		},
		{
			name: "session not found",
			req: &pb.SignOutRequest{
				SessionId: sessionID,
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("DeleteUserSession", mock.Anything, sessionID).
					Return(errors.New("session not found"))
			},
			expected:    &pb.SignOutResponse{Success: false},
			expectedErr: errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.SignOut(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_GetUserByUsername(t *testing.T) {
	userID := uuid.New()
	lastSeen := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetUserByUsernameRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.GetUserByUsernameResponse
		expectedErr error
	}{
		{
			name: "successful get by username",
			req: &pb.GetUserByUsernameRequest{
				Username: "testuser",
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("GetUserByUsername", mock.Anything, "testuser").
					Return(models.User{
						Id:       userID,
						Username: "testuser",
						LastSeen: lastSeen,
					}, nil)
			},
			expected: &pb.GetUserByUsernameResponse{
				User: &pb.User{
					Id:       userID.String(),
					Username: "testuser",
					LastSeen: timestamppb.New(lastSeen),
				},
			},
		},
		{
			name: "user not found",
			req: &pb.GetUserByUsernameRequest{
				Username: "nonexistent",
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("GetUserByUsername", mock.Anything, "nonexistent").
					Return(models.User{}, errors.New("user not found"))
			},
			expectedErr: errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.GetUserByUsername(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_GetUserById(t *testing.T) {
	userID := uuid.New()
	lastSeen := time.Now()

	tests := []struct {
		name        string
		req         *pb.GetUserByIdRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.GetUserByIdResponse
		expectedErr error
	}{
		{
			name: "successful get by id",
			req: &pb.GetUserByIdRequest{
				Id: userID.String(),
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("GetUserById", mock.Anything, userID).
					Return(models.User{
						Id:       userID,
						Username: "testuser",
						LastSeen: lastSeen,
					}, nil)
			},
			expected: &pb.GetUserByIdResponse{
				User: &pb.User{
					Id:       userID.String(),
					Username: "testuser",
					LastSeen: timestamppb.New(lastSeen),
				},
			},
		},
		{
			name: "invalid user id",
			req: &pb.GetUserByIdRequest{
				Id: "invalid-uuid",
			},
			mockSetup:   func(m *mockUserUseCase) {},
			expectedErr: status.Error(codes.InvalidArgument, "invalid user id"),
		},
		{
			name: "user not found",
			req: &pb.GetUserByIdRequest{
				Id: userID.String(),
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("GetUserById", mock.Anything, userID).
					Return(models.User{}, errors.New("user not found"))
			},
			expectedErr: errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.GetUserById(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_LookupUserSession(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name        string
		req         *pb.LookupUserSessionRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.LookupUserSessionResponse
		expectedErr error
	}{
		{
			name: "successful session lookup",
			req: &pb.LookupUserSessionRequest{
				SessionId: sessionID.String(),
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("LookupUserSession", mock.Anything, models.Session{SessionId: sessionID}).
					Return(models.User{
						Id:       userID,
						Username: "testuser",
					}, nil)
			},
			expected: &pb.LookupUserSessionResponse{
				UserId:   userID.String(),
				Username: "testuser",
			},
		},
		{
			name: "invalid session id",
			req: &pb.LookupUserSessionRequest{
				SessionId: "invalid-uuid",
			},
			mockSetup:   func(m *mockUserUseCase) {},
			expectedErr: status.Error(codes.InvalidArgument, "invalid session id"),
		},
		{
			name: "session not found",
			req: &pb.LookupUserSessionRequest{
				SessionId: sessionID.String(),
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("LookupUserSession", mock.Anything, models.Session{SessionId: sessionID}).
					Return(models.User{}, errors.New("session not found"))
			},
			expectedErr: errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.LookupUserSession(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUserServiceServer_SearchSimilarUser(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()

	tests := []struct {
		name        string
		req         *pb.SearchSimilarUserRequest
		mockSetup   func(*mockUserUseCase)
		expected    *pb.SearchSimilarUserResponse
		expectedErr error
	}{
		{
			name: "successful search",
			req: &pb.SearchSimilarUserRequest{
				ToSearch: "test",
				NumUsers: 5,
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("SearchSimilarUser", mock.Anything, "test", uint(5)).
					Return([]models.PublicUserInfo{
						{Id: userID1, Username: "testuser1"},
						{Id: userID2, Username: "testuser2"},
					}, nil)
			},
			expected: &pb.SearchSimilarUserResponse{
				UsersInfo: []*pb.PublicUserInfo{
					{Id: userID1.String(), Username: "testuser1"},
					{Id: userID2.String(), Username: "testuser2"},
				},
			},
		},
		{
			name: "empty result",
			req: &pb.SearchSimilarUserRequest{
				ToSearch: "nonexistent",
				NumUsers: 5,
			},
			mockSetup: func(m *mockUserUseCase) {
				m.On("SearchSimilarUser", mock.Anything, "nonexistent", uint(5)).
					Return([]models.PublicUserInfo{}, nil)
			},
			expected: &pb.SearchSimilarUserResponse{
				UsersInfo: []*pb.PublicUserInfo{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mockUserUseCase)
			tt.mockSetup(mockUC)
			server := NewUserServiceServer(mockUC)

			resp, err := server.SearchSimilarUser(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				for i := range tt.expected.UsersInfo {
					assert.Equal(t, tt.expected.UsersInfo[i].Id, resp.UsersInfo[i].Id)
					assert.Equal(t, tt.expected.UsersInfo[i].Username, resp.UsersInfo[i].Username)
				}
			}
			mockUC.AssertExpectations(t)
		})
	}
}
