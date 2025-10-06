package client

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/url"
	"quickflow/cmd/tech_ui/models"
	"strconv"
)

func (c *APIClient) SearchUsers(query string, count int) ([]models.PublicUserInfoOut, error) {
	path := fmt.Sprintf("/api/v2/users?to_search=%s&count=%d", url.QueryEscape(query), count)

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var users []models.PublicUserInfoOut
	if err := c.parseResponse(resp, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *APIClient) GetProfile(username string) (*models.ProfileForm, error) {
	path := fmt.Sprintf("/api/v2/profiles/%s", url.QueryEscape(username))

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var profile models.ProfileForm
	if err := c.parseResponse(resp, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (c *APIClient) UpdateProfile(firstname, lastname, bio string, birthDate string, sex int) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("firstname", firstname)
	writer.WriteField("lastname", lastname)
	writer.WriteField("birth_date", birthDate)
	writer.WriteField("sex", strconv.Itoa(sex))
	writer.WriteField("bio", bio)

	// Закрываем writer чтобы добавить boundary
	err := writer.Close()
	if err != nil {
		return err
	}

	resp, err := c.doRequest("PATCH", "/api/v2/profiles/me", &body, map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}
