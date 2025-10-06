package client

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) UploadFiles(mediaPaths, audioPaths, filePaths []string) (*models.FileUploadResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add media files
	for _, path := range mediaPaths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("media", filepath.Base(path))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	// Add audio files
	for _, path := range audioPaths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("audio", filepath.Base(path))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	// Add other files
	for _, path := range filePaths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("files", filepath.Base(path))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}

	writer.Close()

	resp, err := c.doRequest("POST", "/api/v2/files", &body, map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})
	if err != nil {
		return nil, err
	}

	var uploadResp models.FileUploadResponse
	if err := c.parseResponse(resp, &uploadResp); err != nil {
		return nil, err
	}

	return &uploadResp, nil
}
