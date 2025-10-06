package client

import (
	"fmt"
	"net/url"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) GetUserPosts(username string, count int, ts string) ([]models.PostOut, error) {
	path := fmt.Sprintf("/api/v2/users/%s/posts?count=%d", url.QueryEscape(username), count)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var posts []models.PostOut
	if err := c.parseResponse(resp, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (c *APIClient) GetFeed(count int, ts, feedType string) ([]models.PostOut, error) {
	path := fmt.Sprintf("/api/v2/posts?count=%d&type=%s", count, feedType)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var posts []models.PostOut
	if err := c.parseResponse(resp, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (c *APIClient) GetPost(postID string) (*models.PostOut, error) {
	resp, err := c.doRequest("GET", "/api/v2/posts/"+postID, nil, nil)
	if err != nil {
		return nil, err
	}

	var post models.PostOut
	if err := c.parseResponse(resp, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

func (c *APIClient) CreatePost(form models.CommentForm) (*models.PostOut, error) {
	resp, err := c.doJSONRequest("POST", "/api/v2/posts", form)
	if err != nil {
		return nil, err
	}

	var post models.PostOut
	if err := c.parseResponse(resp, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

func (c *APIClient) UpdatePost(postID string, form models.UpdatePostForm) error {
	resp, err := c.doJSONRequest("PATCH", "/api/v2/posts/"+postID, form)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) DeletePost(postID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/posts/"+postID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}
