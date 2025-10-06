//go:build unit
// +build unit

package file_service

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
	pb "quickflow/shared/proto/file_service"
)

func TestProtoFileToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.File
		expected *models.File
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
			expected: &models.File{
				Name:        "test.txt",
				Size:        123,
				MimeType:    "text/plain",
				AccessMode:  models.AccessPrivate,
				URL:         "http://example.com/file",
				Reader:      bytes.NewReader([]byte("content")),
				DisplayType: models.DisplayType("document"),
				Ext:         ".txt",
			},
		},
		{
			name: "only URL",
			input: &pb.File{
				Url:         "http://example.com/file",
				DisplayType: "image",
			},
			expected: &models.File{
				URL:         "http://example.com/file",
				DisplayType: models.DisplayType("image"),
				Reader:      bytes.NewReader(nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoFileToModel(tt.input)
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

func TestModelFileToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.File
		expected *pb.File
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full file info",
			input: &models.File{
				Name:        "test.txt",
				Size:        123,
				MimeType:    "text/plain",
				AccessMode:  models.AccessPrivate,
				URL:         "http://example.com/file",
				Reader:      strings.NewReader("content"),
				DisplayType: models.DisplayType("document"),
			},
			expected: &pb.File{
				FileName:    "test.txt",
				FileSize:    123,
				FileType:    "text/plain",
				AccessMode:  pb.AccessMode_ACCESS_PRIVATE,
				Url:         "http://example.com/file",
				File:        []byte("content"),
				DisplayType: "document",
			},
		},
		{
			name: "only URL",
			input: &models.File{
				URL:         "http://example.com/file",
				DisplayType: models.DisplayType("image"),
			},
			expected: &pb.File{
				Url:         "http://example.com/file",
				DisplayType: "image",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ModelFileToProto(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, tt.expected.FileName, result.FileName)
			assert.Equal(t, tt.expected.FileSize, result.FileSize)
			assert.Equal(t, tt.expected.FileType, result.FileType)
			assert.Equal(t, tt.expected.AccessMode, result.AccessMode)
			assert.Equal(t, tt.expected.Url, result.Url)
			assert.Equal(t, tt.expected.DisplayType, result.DisplayType)

			if tt.input != nil && tt.input.Reader != nil {
				assert.Equal(t, tt.expected.File, result.File)
			}
		})
	}
}

func TestProtoFilesToModels(t *testing.T) {
	tests := []struct {
		name     string
		input    []*pb.File
		expected []*models.File
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "multiple files",
			input: []*pb.File{
				{
					FileName: "file1.txt",
					FileType: "text/plain",
				},
				{
					FileName: "file2.txt",
					FileType: "text/plain",
				},
			},
			expected: []*models.File{
				{
					Name:     "file1.txt",
					MimeType: "text/plain",
					Reader:   bytes.NewReader(nil),
				},
				{
					Name:     "file2.txt",
					MimeType: "text/plain",
					Reader:   bytes.NewReader(nil),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoFilesToModels(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, tt.expected[i].Name, result[i].Name)
				assert.Equal(t, tt.expected[i].MimeType, result[i].MimeType)
			}
		})
	}
}

func TestModelFilesToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.File
		expected []*pb.File
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "multiple files",
			input: []*models.File{
				{
					Name: "file1.txt",
					URL:  "http://example.com/file",
				},
				{
					Name: "file2.txt",
					URL:  "http://example.com/file2",
				},
			},
			expected: []*pb.File{
				{
					FileName: "file1.txt",
					Url:      "http://example.com/file",
				},
				{
					FileName: "file2.txt",
					Url:      "http://example.com/file2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ModelFilesToProto(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.Equal(t, tt.expected[i].FileName, result[i].FileName)
				assert.Equal(t, tt.expected[i].FileType, result[i].FileType)
			}
		})
	}
}
