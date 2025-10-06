package client

import (
	"fmt"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) GetPostComments(postID string, count int, ts string) ([]models.CommentOut, error) {
	path := fmt.Sprintf("/api/v2/posts/%s/comments?count=%d", postID, count)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var comments []models.CommentOut
	if err := c.parseResponse(resp, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

func (c *APIClient) CreateComment(postID string, form models.CommentForm) (*models.CommentOut, error) {
	resp, err := c.doJSONRequest("POST", "/api/v2/posts/"+postID+"/comments", form)
	if err != nil {
		return nil, err
	}

	var comment models.CommentOut
	if err := c.parseResponse(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (c *APIClient) UpdateComment(postID, commentID string, form models.CommentForm) (*models.CommentOut, error) {
	resp, err := c.doJSONRequest("PATCH", "/api/v2/posts/"+postID+"/comments/"+commentID, form)
	if err != nil {
		return nil, err
	}

	var comment models.CommentOut
	if err := c.parseResponse(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (c *APIClient) DeleteComment(postID, commentID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/posts/"+postID+"/comments/"+commentID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 205 {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}
