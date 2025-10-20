package client

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/url"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) SearchCommunities(query string, count int) ([]models.CommunityForm, error) {
	path := fmt.Sprintf("/api/v2/communities?to_search=%s&count=%d", url.QueryEscape(query), count)

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var communities []models.CommunityForm
	if err := c.parseResponse(resp, &communities); err != nil {
		return nil, err
	}

	return communities, nil
}

func (c *APIClient) GetUserCommunities(username string, count int, ts, role string) ([]models.CommunityForm, error) {
	path := fmt.Sprintf("/api/v2/users/%s/communities?count=%d", url.QueryEscape(username), count)
	if ts != "" {
		path += "&ts=" + ts
	}
	if role != "" {
		path += "&role=" + role
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var communities []models.CommunityForm
	if err := c.parseResponse(resp, &communities); err != nil {
		return nil, err
	}

	return communities, nil
}

func (c *APIClient) GetCommunityByID(communityID string) (*models.CommunityForm, error) {
	resp, err := c.doRequest("GET", "/api/v2/communities/"+communityID, nil, nil)
	if err != nil {
		return nil, err
	}

	var community models.CommunityForm
	if err := c.parseResponse(resp, &community); err != nil {
		return nil, err
	}

	return &community, nil
}

func (c *APIClient) GetCommunityByName(name string) (*models.CommunityForm, error) {
	resp, err := c.doRequest("GET", "/api/v2/communities/"+name, nil, nil)
	if err != nil {
		return nil, err
	}

	var community models.CommunityForm
	if err := c.parseResponse(resp, &community); err != nil {
		return nil, err
	}

	return &community, nil
}

func (c *APIClient) CreateCommunity(nickname, name, description string) (*models.CommunityForm, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("nickname", nickname)
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("description", description)

	// ВАЖНО: закрыть writer, чтобы записать закрывающее boundary!
	if err := writer.Close(); err != nil {
		return nil, err
	}

	resp, err := c.doRequest("POST", "/api/v2/communities", &body, map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})
	if err != nil {
		return nil, err
	}

	var community models.CommunityForm
	if err := c.parseResponse(resp, &community); err != nil {
		return nil, err
	}

	return &community, nil
}

func (c *APIClient) DeleteCommunity(communityID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/communities/"+communityID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) JoinCommunity(communityID string) error {
	resp, err := c.doRequest("POST", "/api/v2/communities/"+communityID+"/members", nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode > 204 {
		return fmt.Errorf("join failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) LeaveCommunity(communityID, userID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/communities/"+communityID+"/members/"+userID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("leave failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) GetCommunityMembers(communityID string, count int, ts string) ([]models.CommunityMemberOut, error) {
	path := fmt.Sprintf("/api/v2/communities/%s/members?count=%d", communityID, count)
	if ts != "" {
		path += "&ts=" + ts
	}

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var members []models.CommunityMemberOut
	if err := c.parseResponse(resp, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (c *APIClient) GetCommunityPosts(name string, count int, ts string) ([]models.PostOut, error) {
	path := fmt.Sprintf("/api/v2/communities/%s/posts?count=%d", url.QueryEscape(name), count)
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

func (c *APIClient) CreateCommunityPost(communityName string, form models.CommentForm) (*models.PostOut, error) {
	resp, err := c.doJSONRequest("POST", "/api/v2/communities/"+communityName+"/posts", form)
	if err != nil {
		return nil, err
	}

	var post models.PostOut
	if err := c.parseResponse(resp, &post); err != nil {
		return nil, err
	}

	return &post, nil
}

// ChangeCommunityMemberRole меняет роль участника в сообществе
func (c *APIClient) ChangeCommunityMemberRole(communityID, userID, role string) error {
	reqBody := struct {
		Role string `json:"role"`
	}{Role: role}

	resp, err := c.doJSONRequest(
		"PATCH",
		fmt.Sprintf("/api/v2/communities/%s/members/%s", communityID, userID),
		reqBody,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("change role failed with status %d", resp.StatusCode)
	}
	return nil
}
