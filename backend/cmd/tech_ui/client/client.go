package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"quickflow/cmd/tech_ui/models"
	"time"
)

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	SessionID  string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) SetSession(sessionID string) {
	c.SessionID = sessionID
}

func (c *APIClient) doRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set session cookie if available
	if c.SessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: c.SessionID,
		})
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.HTTPClient.Do(req)
}

func (c *APIClient) doJSONRequest(method, path string, data interface{}) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return c.doRequest(method, path, body, headers)
}

func (c *APIClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errorResp models.ErrorForm
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return &errorResp
	}

	if result != nil {
		err = json.Unmarshal(body, result)

		PyaloadWrapper := struct {
			Payload json.RawMessage `json:"payload"`
		}{}
		err2 := json.Unmarshal(body, &PyaloadWrapper)
		if err2 == nil && PyaloadWrapper.Payload != nil {
			return json.Unmarshal(PyaloadWrapper.Payload, result)
		}
		return err
	}

	return nil
}
