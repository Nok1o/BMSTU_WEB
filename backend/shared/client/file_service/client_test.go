//go:build unit
// +build unit

package file_service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
	"quickflow/shared/proto/file_service/mocks"
)

func TestFileClient_UploadFile(t *testing.T) {
	tests := []struct {
		name        string
		file        *models.File
		setupMock   func(stream *mocks.MockFileService_UploadFileClient)
		expectedURL string
		expectedErr string
	}{
		{
			name: "successful upload",
			file: &models.File{
				Name:        "test.txt",
				MimeType:    "text/plain",
				Size:        123,
				AccessMode:  models.AccessPrivate,
				DisplayType: models.DisplayType("document"),
				Reader:      strings.NewReader("test content"),
			},
			setupMock: func(stream *mocks.MockFileService_UploadFileClient) {
				stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(req *pb.UploadFileRequest) error {
					assert.NotNil(t, req.GetInfo())
					return nil
				}).Times(1)
				stream.EXPECT().Send(gomock.Any()).DoAndReturn(func(req *pb.UploadFileRequest) error {
					assert.NotNil(t, req.GetChunk())
					return nil
				}).AnyTimes()
				stream.EXPECT().CloseAndRecv().Return(&pb.UploadFileResponse{
					FileUrl: "http://storage.example.com/test.txt",
				}, nil)
			},
			expectedURL: "http://storage.example.com/test.txt",
		},
		{
			name: "failed to send metadata",
			file: &models.File{
				Name:     "test.txt",
				MimeType: "text/plain",
				Reader:   strings.NewReader("test"),
			},
			setupMock: func(stream *mocks.MockFileService_UploadFileClient) {
				stream.EXPECT().Send(gomock.Any()).Return(errors.New("connection error"))
			},
			expectedErr: "send metadata: connection error",
		},
		{
			name: "failed to send chunk",
			file: &models.File{
				Name:     "test.txt",
				MimeType: "text/plain",
				Reader:   strings.NewReader("test"),
			},
			setupMock: func(stream *mocks.MockFileService_UploadFileClient) {
				stream.EXPECT().Send(gomock.Any()).Return(nil).Times(1)
				stream.EXPECT().Send(gomock.Any()).Return(errors.New("chunk error"))
				stream.EXPECT().CloseAndRecv().Return(nil, errors.New("stream closed")).AnyTimes()
			},
			expectedErr: "send chunk: chunk error",
		},
		{
			name: "failed to close stream",
			file: &models.File{
				Name:     "test.txt",
				MimeType: "text/plain",
				Reader:   strings.NewReader("test"),
			},
			setupMock: func(stream *mocks.MockFileService_UploadFileClient) {
				stream.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
				stream.EXPECT().CloseAndRecv().Return(nil, errors.New("server error"))
			},
			expectedErr: "receive response: server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			stream := mocks.NewMockFileService_UploadFileClient(ctrl)
			stream.EXPECT().Context().Return(context.Background()).AnyTimes()
			tt.setupMock(stream)

			mockClient := mocks.NewMockFileServiceClient(ctrl)
			mockClient.EXPECT().UploadFile(gomock.Any()).Return(stream, nil)

			client := &FileClient{client: mockClient}
			url, err := client.UploadFile(context.Background(), tt.file)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestFileClient_UploadManyFiles(t *testing.T) {
	tests := []struct {
		name         string
		files        []*models.File
		setupMock    func(stream *mocks.MockFileService_UploadManyFilesClient)
		expectedURLs []string
		expectedErr  string
	}{
		{
			name: "successful multiple upload",
			files: []*models.File{
				{
					Name:     "file1.txt",
					MimeType: "text/plain",
					Reader:   strings.NewReader("content1"),
				},
				{
					Name:     "file2.txt",
					MimeType: "text/plain",
					Reader:   strings.NewReader("content2"),
				},
			},
			setupMock: func(stream *mocks.MockFileService_UploadManyFilesClient) {
				// Expect 2 info sends and multiple chunks
				stream.EXPECT().Send(gomock.Any()).Return(nil).Times(4) // 2 info + 2 chunks
				stream.EXPECT().CloseSend().Return(nil)
				stream.EXPECT().Recv().Return(&pb.UploadFileResponse{
					FileUrl: "http://storage.example.com/file1",
				}, nil)
				stream.EXPECT().Recv().Return(&pb.UploadFileResponse{
					FileUrl: "http://storage.example.com/file2",
				}, nil)
				stream.EXPECT().Recv().Return(nil, io.EOF)
			},
			expectedURLs: []string{
				"http://storage.example.com/file1",
				"http://storage.example.com/file2",
			},
		},
		{
			name: "failed to send file info",
			files: []*models.File{
				{
					Name:     "file1.txt",
					MimeType: "text/plain",
					Reader:   strings.NewReader("content1"),
				},
			},
			setupMock: func(stream *mocks.MockFileService_UploadManyFilesClient) {
				stream.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))
				stream.EXPECT().Recv().Return(nil, io.EOF).AnyTimes()
			},
			expectedErr: "send info: send error",
		},
		{
			name: "failed to receive response",
			files: []*models.File{
				{
					Name:     "file1.txt",
					MimeType: "text/plain",
					Reader:   strings.NewReader("content1"),
				},
			},
			setupMock: func(stream *mocks.MockFileService_UploadManyFilesClient) {
				stream.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
				stream.EXPECT().CloseSend().Return(nil)
				stream.EXPECT().Recv().Return(nil, errors.New("receive error"))
			},
			expectedErr: "receive response: receive error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			stream := mocks.NewMockFileService_UploadManyFilesClient(ctrl)
			stream.EXPECT().Context().Return(context.Background()).AnyTimes()
			tt.setupMock(stream)

			mockClient := mocks.NewMockFileServiceClient(ctrl)
			mockClient.EXPECT().UploadManyFiles(gomock.Any()).Return(stream, nil)

			client := &FileClient{client: mockClient}
			urls, err := client.UploadManyFiles(context.Background(), tt.files)

			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURLs, urls)
			}
		})
	}
}

func TestFileClient_DeleteFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		setupMock   func(client *mocks.MockFileServiceClient)
		expectedErr string
	}{
		{
			name:     "successful delete",
			filename: "http://storage.example.com/file.txt",
			setupMock: func(client *mocks.MockFileServiceClient) {
				client.EXPECT().DeleteFile(gomock.Any(), &pb.DeleteFileRequest{
					FileUrl: "http://storage.example.com/file.txt",
				}).Return(&pb.DeleteFileResponse{Success: true}, nil)
			},
		},
		{
			name:     "empty filename",
			filename: "",
			setupMock: func(client *mocks.MockFileServiceClient) {
				// No expectations as it should fail before calling the client
			},
			expectedErr: "fileClient.DeleteFile: file URL cannot be empty",
		},
		{
			name:     "server error",
			filename: "http://storage.example.com/file.txt",
			setupMock: func(client *mocks.MockFileServiceClient) {
				client.EXPECT().DeleteFile(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("server error"))
			},
			expectedErr: "fileClient.DeleteFile: server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockFileServiceClient(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			client := &FileClient{client: mockClient}
			err := client.DeleteFile(context.Background(), tt.filename)

			if tt.expectedErr != "" {
				assert.Error(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
