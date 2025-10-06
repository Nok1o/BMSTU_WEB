//go:build unit
// +build unit

package interceptor

import (
	"context"
	"errors"
	user_errors "quickflow/user_service/internal/errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		inputError     error
		expectedCode   codes.Code
		expectedReason string
		wantErr        bool
	}{
		{
			name:           "no error",
			inputError:     nil,
			expectedCode:   codes.OK,
			expectedReason: "",
			wantErr:        false,
		},
		{
			name:           "not found error",
			inputError:     user_errors.ErrNotFound,
			expectedCode:   codes.NotFound,
			expectedReason: "NOT_FOUND",
			wantErr:        true,
		},
		{
			name:           "profile not found",
			inputError:     user_errors.ErrProfileNotFound,
			expectedCode:   codes.NotFound,
			expectedReason: "NOT_FOUND",
			wantErr:        true,
		},
		{
			name:           "user not found",
			inputError:     user_errors.ErrUserNotFound,
			expectedCode:   codes.NotFound,
			expectedReason: "NOT_FOUND",
			wantErr:        true,
		},
		{
			name:           "already exists",
			inputError:     user_errors.ErrAlreadyExists,
			expectedCode:   codes.AlreadyExists,
			expectedReason: "ALREADY_EXISTS",
			wantErr:        true,
		},
		{
			name:           "username taken",
			inputError:     user_errors.ErrUsernameTaken,
			expectedCode:   codes.AlreadyExists,
			expectedReason: "ALREADY_EXISTS",
			wantErr:        true,
		},
		{
			name:           "invalid user id",
			inputError:     user_errors.ErrInvalidUserId,
			expectedCode:   codes.InvalidArgument,
			expectedReason: "INVALID_ARGUMENT",
			wantErr:        true,
		},
		{
			name:           "invalid profile info",
			inputError:     user_errors.ErrInvalidProfileInfo,
			expectedCode:   codes.InvalidArgument,
			expectedReason: "INVALID_ARGUMENT",
			wantErr:        true,
		},
		{
			name:           "user validation error",
			inputError:     user_errors.ErrUserValidation,
			expectedCode:   codes.InvalidArgument,
			expectedReason: "INVALID_ARGUMENT",
			wantErr:        true,
		},
		{
			name:           "profile validation error",
			inputError:     user_errors.ErrProfileValidation,
			expectedCode:   codes.InvalidArgument,
			expectedReason: "INVALID_ARGUMENT",
			wantErr:        true,
		},
		{
			name:           "unknown error",
			inputError:     errors.New("some unknown error"),
			expectedCode:   codes.Internal,
			expectedReason: "INTERNAL",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock handler that returns our test error
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, tt.inputError
			}

			// Call the interceptor
			_, err := ErrorInterceptor(
				context.Background(),
				nil,
				&grpc.UnaryServerInfo{},
				handler,
			)

			// Check error conditions
			if (err != nil) != tt.wantErr {
				t.Errorf("ErrorInterceptor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				// Check gRPC status code
				st, ok := status.FromError(err)
				if !ok {
					t.Fatal("expected gRPC status error")
				}

				if st.Code() != tt.expectedCode {
					t.Errorf("expected code %v, got %v", tt.expectedCode, st.Code())
				}

				// Check error details
				for _, detail := range st.Details() {
					if errInfo, ok := detail.(interface{ GetReason() string }); ok {
						if errInfo.GetReason() != tt.expectedReason {
							t.Errorf("expected reason %q, got %q", tt.expectedReason, errInfo.GetReason())
						}
						return
					}
				}

				if tt.expectedReason != "" {
					t.Error("expected error details, but none found")
				}
			}
		})
	}
}

func TestWithErrorInfo(t *testing.T) {
	tests := []struct {
		name         string
		code         codes.Code
		reason       string
		message      string
		wantContains string
	}{
		{
			name:         "with details",
			code:         codes.NotFound,
			reason:       "TEST_REASON",
			message:      "test message",
			wantContains: "test message",
		},
		{
			name:         "fallback when details fail",
			code:         codes.InvalidArgument,
			reason:       "", // empty reason might cause failure in WithDetails
			message:      "fallback message",
			wantContains: "fallback message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := withErrorInfo(tt.code, tt.reason, tt.message)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Verify it's a gRPC status error
			st, ok := status.FromError(err)
			if !ok {
				t.Error("expected gRPC status error")
			}

			if st.Code() != tt.code {
				t.Errorf("expected code %v, got %v", tt.code, st.Code())
			}
		})
	}
}
