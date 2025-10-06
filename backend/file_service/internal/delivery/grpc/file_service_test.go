//go:build unit
// +build unit

package grpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/file_service/internal/delivery/grpc/mocks"
	dto "quickflow/shared/client/file_service"
	shared_models "quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
	mocks2 "quickflow/shared/proto/file_service/mocks"
)

func TestUploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockFileUseCase(ctrl)
	server := NewFileServiceServer(mockUC)

	tests := []struct {
		name        string
		setupMock   func()
		streamSetup func(fileServer *mocks2.MockFileService_UploadFileServer)
		expectedURL string
		expectedErr error
	}{
		{
			name: "successful upload",
			setupMock: func() {
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f *shared_models.File) (string, error) {
						assert.Equal(t, "test.txt", f.Name)
						assert.Equal(t, "text/plain", f.MimeType)
						return "http://storage.example.com/test.txt", nil
					})
			},
			streamSetup: func(stream *mocks2.MockFileService_UploadFileServer) {
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Info{
							Info: &pb.File{
								FileName: "test.txt",
								FileType: "text/plain",
							},
						},
					}, nil)
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Chunk{
							Chunk: []byte("test content"),
						},
					}, nil)
				stream.EXPECT().Recv().Return(nil, io.EOF)
				stream.EXPECT().SendAndClose(gomock.Any()).
					Do(func(resp *pb.UploadFileResponse) error {
						assert.Equal(t, "http://storage.example.com/test.txt", resp.FileUrl)
						return nil
					})
			},
			expectedURL: "http://storage.example.com/test.txt",
		},
		{
			name: "invalid chunk before info",
			streamSetup: func(stream *mocks2.MockFileService_UploadFileServer) {
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Chunk{
							Chunk: []byte("test"),
						},
					}, nil)
			},
			expectedErr: status.Error(codes.InvalidArgument, "FileInfo must be sent before chunks"),
		},
		{
			name: "upload usecase failure",
			setupMock: func() {
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("", errors.New("storage error"))
			},
			streamSetup: func(stream *mocks2.MockFileService_UploadFileServer) {
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Info{
							Info: &pb.File{
								FileName: "test.txt",
								FileType: "text/plain",
							},
						},
					}, nil)
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Chunk{
							Chunk: []byte("test"),
						},
					}, nil)
				stream.EXPECT().Recv().Return(nil, io.EOF)
			},
			expectedErr: status.Error(codes.Internal, "Upload usecase failed: storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			stream := mocks2.NewMockFileService_UploadFileServer(ctrl)
			stream.EXPECT().Context().Return(context.Background()).AnyTimes()
			if tt.streamSetup != nil {
				tt.streamSetup(stream)
			}

			err := server.UploadFile(stream)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadManyFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockFileUseCase(ctrl)
	server := NewFileServiceServer(mockUC)

	tests := []struct {
		name        string
		setupMock   func()
		streamSetup func(stream *mocks2.MockFileService_UploadManyFilesServer)
		expected    []string
		expectedErr error
	}{
		{
			name: "successful multiple upload",
			setupMock: func() {
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("http://storage.example.com/file1", nil).Times(1)
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("http://storage.example.com/file2", nil).Times(1)
			},
			streamSetup: func(stream *mocks2.MockFileService_UploadManyFilesServer) {
				// First file
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Info{
							Info: &pb.File{
								FileName: "file1.txt",
								FileType: "text/plain",
							},
						},
					}, nil)
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Chunk{
							Chunk: []byte("content1"),
						},
					}, nil)

				// Second file
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Info{
							Info: &pb.File{
								FileName: "file2.txt",
								FileType: "text/plain",
							},
						},
					}, nil)
				stream.EXPECT().Recv().
					Return(&pb.UploadFileRequest{
						Data: &pb.UploadFileRequest_Chunk{
							Chunk: []byte("content2"),
						},
					}, nil)

				stream.EXPECT().Recv().Return(nil, io.EOF)
				stream.EXPECT().Send(gomock.Any()).Return(nil).Times(2)
			},
			expected: []string{
				"http://storage.example.com/file1",
				"http://storage.example.com/file2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			stream := mocks2.NewMockFileService_UploadManyFilesServer(ctrl)
			stream.EXPECT().Context().Return(context.Background()).AnyTimes()
			if tt.streamSetup != nil {
				tt.streamSetup(stream)
			}

			err := server.UploadManyFiles(stream)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockFileUseCase(ctrl)
	server := NewFileServiceServer(mockUC)

	tests := []struct {
		name        string
		setupMock   func()
		req         *pb.DeleteFileRequest
		expectedRes *pb.DeleteFileResponse
		expectedErr error
	}{
		{
			name: "successful delete",
			setupMock: func() {
				mockUC.EXPECT().
					DeleteFile(gomock.Any(), "http://storage.example.com/file.txt").
					Return(nil)
			},
			req: &pb.DeleteFileRequest{
				FileUrl: "http://storage.example.com/file.txt",
			},
			expectedRes: &pb.DeleteFileResponse{Success: true},
		},
		{
			name: "delete failed",
			setupMock: func() {
				mockUC.EXPECT().
					DeleteFile(gomock.Any(), "http://storage.example.com/file.txt").
					Return(errors.New("file not found"))
			},
			req: &pb.DeleteFileRequest{
				FileUrl: "http://storage.example.com/file.txt",
			},
			expectedRes: &pb.DeleteFileResponse{Success: false},
			expectedErr: status.Error(codes.Internal, "Failed to delete file: file not found"),
		},
		{
			name: "empty file URL",
			req: &pb.DeleteFileRequest{
				FileUrl: "",
			},
			expectedRes: nil,
			expectedErr: status.Error(codes.InvalidArgument, "file url is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			res, err := server.DeleteFile(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestProtoFileToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.File
		expected *shared_models.File
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full file info",
			input: &pb.File{
				FileName:    "test.txt",
				FileSize:    123,
				FileType:    "text/plain",
				AccessMode:  pb.AccessMode_ACCESS_PRIVATE,
				Url:         "http://example.com/file",
				File:        []byte("content"),
				DisplayType: "document",
			},
			expected: &shared_models.File{
				Name:        "test.txt",
				Size:        123,
				MimeType:    "text/plain",
				AccessMode:  shared_models.AccessPrivate,
				URL:         "http://example.com/file",
				Reader:      bytes.NewReader([]byte("content")),
				DisplayType: shared_models.DisplayType("document"),
				Ext:         ".txt",
			},
		},
		{
			name: "only URL",
			input: &pb.File{
				Url:         "http://example.com/file",
				DisplayType: "image",
			},
			expected: &shared_models.File{
				URL:         "http://example.com/file",
				DisplayType: shared_models.DisplayType("image"),
				Reader:      bytes.NewReader(nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dto.ProtoFileToModel(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Size, result.Size)
			assert.Equal(t, tt.expected.MimeType, result.MimeType)
			assert.Equal(t, tt.expected.AccessMode, result.AccessMode)
			assert.Equal(t, tt.expected.URL, result.URL)
			assert.Equal(t, tt.expected.DisplayType, result.DisplayType)
			assert.Equal(t, tt.expected.Ext, result.Ext)

			if tt.input != nil && tt.input.File != nil {
				content, _ := io.ReadAll(result.Reader)
				assert.Equal(t, tt.input.File, content)
			}
		})
	}
}

// Helpers for testing file operations
var osCreateTemp = os.CreateTemp

func TestFinalizeUploadedFile(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mockUC *mocks.MockFileUseCase)
		prepareFile func() (*os.File, string)
		info        *pb.File
		expectedURL string
		expectedErr error
	}{
		{
			name: "upload failed",
			setupMock: func(mockUC *mocks.MockFileUseCase) {
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("", errors.New("storage error"))
			},
			prepareFile: func() (*os.File, string) {
				tempFile, err := os.CreateTemp("", "test-*")
				require.NoError(t, err)
				_, err = tempFile.Write([]byte("test content"))
				require.NoError(t, err)
				tempFile.Close()
				return tempFile, tempFile.Name()
			},
			info: &pb.File{
				FileName: "test.txt",
				FileType: "text/plain",
			},
			expectedErr: errors.New("Upload usecase failed: storage error"),
		},
		{
			name: "successful finalize",
			setupMock: func(mockUC *mocks.MockFileUseCase) {
				mockUC.EXPECT().UploadFile(gomock.Any(), gomock.Any()).
					Return("http://storage.example.com/file", nil)
			},
			prepareFile: func() (*os.File, string) {
				tempFile, err := os.CreateTemp("", "test-*")
				require.NoError(t, err)
				_, err = tempFile.Write([]byte("test content"))
				require.NoError(t, err)
				tempFile.Close()
				return tempFile, tempFile.Name()
			},
			info: &pb.File{
				FileName: "test.txt",
				FileType: "text/plain",
			},
			expectedURL: "http://storage.example.com/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockFileUseCase(ctrl)
			server := NewFileServiceServer(mockUC)
			if tt.setupMock != nil {
				tt.setupMock(mockUC)
			}

			_, tempFileName := tt.prepareFile()
			defer os.Remove(tempFileName)

			// Reopen temp file for reading
			f, err := os.Open(tempFileName)
			require.NoError(t, err)
			defer f.Close()

			url, err := server.finalizeUploadedFile(context.Background(), tt.info, f)

			if tt.expectedErr != nil {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
				assert.Equal(t, []byte("test content"), tt.info.File)
			}
		})
	}
}
