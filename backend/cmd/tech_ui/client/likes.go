package client

import (
	"fmt"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) LikePost(postID string) (*models.LikeForm, error) {
	_, err := c.doRequest("POST", "/api/v2/posts/"+postID+"/likes", nil, nil)
	if err != nil {
		return nil, err
	}

	//var like models.LikeForm
	//if err := c.parseResponse(resp, &like); err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (c *APIClient) LikeComment(postID, commentID string) (*models.LikeForm, error) {
	_, err := c.doRequest("POST", "/api/v2/posts/"+postID+"/comments/"+commentID+"/likes", nil, nil)
	if err != nil {
		return nil, err
	}

	//var like models.LikeForm
	//if err := c.parseResponse(resp, &like); err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (c *APIClient) UnlikePost(postID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/posts/"+postID+"/likes/me", nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 205 {
		return fmt.Errorf("unlike failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) UnlikeComment(postID, commentID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/posts/"+postID+"/comments/"+commentID+"/likes/me", nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("unlike failed with status: %d", resp.StatusCode)
	}

	return nil
}
