package client

import (
	"fmt"
	"quickflow/cmd/tech_ui/models"
)

func (c *APIClient) GetFriends(userID, requestType string, count, offset int) ([]models.FriendsInfoOut, error) {
	path := fmt.Sprintf("/api/v2/users/%s/friends?count=%d&offset=%d&request_type=%s",
		userID, count, offset, requestType)

	resp, err := c.doRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var friends = struct {
		Frnds []models.FriendsInfoOut `json:"friends"`
	}{}
	if err := c.parseResponse(resp, &friends); err != nil {
		return nil, err
	}

	return friends.Frnds, nil
}

func (c *APIClient) SendFriendRequest(receiverID string) (string, error) {
	form := models.FriendRequest{
		ReceiverID: receiverID,
	}

	resp, err := c.doJSONRequest("POST", "/api/v2/friend_requests", form)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := c.parseResponse(resp, &result); err != nil {
		return "", err
	}

	return result["request_id"], nil
}

func (c *APIClient) RespondToFriendRequest(requestID, status string) error {
	form := models.FriendRequestStatus{
		Status: status,
	}

	resp, err := c.doJSONRequest("PUT", "/api/v2/friend_requests/"+requestID, form)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("response failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) DeleteFriend(friendID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/friends/"+friendID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode > 204 {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) DeleteFriendRequest(requestID string) error {
	resp, err := c.doRequest("DELETE", "/api/v2/friend_requests/"+requestID, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}
