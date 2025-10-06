package client

import (
	"fmt"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) Login(username, password string) (*models.Session, error) {
	authForm := models.AuthForm{
		Username: username,
		Password: password,
	}

	resp, err := c.doJSONRequest("POST", "/api/v2/sessions", authForm)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 201 {
		// Extract session from cookies
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session" {
				c.SessionID = cookie.Value
			}
		}
	}

	var session models.Session
	if err := c.parseResponse(resp, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *APIClient) Logout() error {
	resp, err := c.doRequest("DELETE", "/api/v2/sessions/me", nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
	}

	c.SessionID = ""
	return nil
}

func (c *APIClient) SignUp(form models.SignUpForm) (*models.SignupResponse, error) {
	resp, err := c.doJSONRequest("POST", "/api/v2/users", form)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 201 {
		// Extract session from cookies
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session" {
				c.SessionID = cookie.Value
			}
		}
	}

	var signupResp models.SignupResponse
	if err := c.parseResponse(resp, &signupResp); err != nil {
		return nil, err
	}

	return &signupResp, nil
}
