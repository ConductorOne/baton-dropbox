package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const groupDefaultLimit = 100

type ListGroupsBody struct {
	Limit int `json:"limit"`
}

func DefaultListGroupsBody() ListGroupsBody {
	return ListGroupsBody{
		Limit: groupDefaultLimit,
	}
}

// docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-list
func (c *Client) ListGroups(ctx context.Context, limit int) (*ListGroupsPayload, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := DefaultListGroupsBody()
	if limit != 0 {
		body.Limit = limit
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListGroupsURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupsPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}

func (c *Client) ListGroupsContinue(ctx context.Context, cursor string) (*ListGroupsPayload, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := struct {
		Cursor string `json:"cursor"`
	}{Cursor: cursor}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListGroupsContinueURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupsPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}

type ListGroupMembersBody struct {
	Group GroupIdTag `json:"group"`
	Limit int        `json:"limit"`
}

type GroupIdTag struct {
	GroupID string `json:"group_id"`
	Tag     string `json:".tag"`
}

func DefaultGroupMembersBody() ListGroupMembersBody {
	return ListGroupMembersBody{
		Group: GroupIdTag{
			Tag: "group_id",
		},
		Limit: groupDefaultLimit,
	}
}

func (c *Client) ListGroupMembers(ctx context.Context, groupId string, limit int) (*ListGroupMembersPayload, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := DefaultGroupMembersBody()
	if groupId == "" {
		return nil, nil, fmt.Errorf("groupId is required")
	}
	body.Group.GroupID = groupId

	if limit != 0 {
		body.Limit = limit
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListGroupMembersURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}

func (c *Client) ListGroupMembersContinue(ctx context.Context, cursor string) (*ListGroupMembersPayload, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := struct {
		Cursor string `json:"cursor"`
	}{Cursor: cursor}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListGroupMembersContinueURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}

type RemoveUserFromGroupBody struct {
	Group         GroupIdTag `json:"group"`
	Users         []EmailTag `json:"users"`
	ReturnMembers bool       `json:"return_members"`
}

type EmailTag struct {
	Tag   string `json:".tag"`
	Email string `json:"email"`
}

func (c *Client) RemoveUserFromGroup(ctx context.Context, groupId, email string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := RemoveUserFromGroupBody{
		Group: GroupIdTag{
			GroupID: groupId,
			Tag:     "group_id",
		},
		Users: []EmailTag{
			{
				Tag:   "email",
				Email: email,
			},
		},
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RemoveUserFromGroupURL, buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		logBody(ctx, res)
		return &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return &ratelimitData, err
	}

	return &ratelimitData, nil
}

func (c *Client) GetTeamMemberID(ctx context.Context, groupId, userId string) (string, error) {
	var payload *ListGroupMembersPayload
	var err error
	var limit int = 100

	payload, _, err = c.ListGroupMembers(ctx, groupId, limit)
	if err != nil {
		return "", fmt.Errorf("baton-dropbox: failed to list group members: %s", err.Error())
	}

	for _, member := range payload.Members {
		if member.Profile.AccountID == userId {
			return member.Profile.TeamMemberID, nil
		}
	}

	for payload.HasMore {
		payload, _, err = c.ListGroupMembersContinue(ctx, payload.Cursor)
		if err != nil {
			return "", fmt.Errorf("baton-dropbox: failed to list group members: %s", err.Error())
		}

		for _, member := range payload.Members {
			if member.Profile.AccountID == userId {
				return member.Profile.TeamMemberID, nil
			}
		}
	}

	return "", fmt.Errorf("baton-dropbox: user not found in group")
}

type AddUserToGroupBody struct {
	Group         GroupIdTag          `json:"group"`
	Members       []AddToGroupMembers `json:"members"`
	ReturnMembers bool                `json:"return_members"`
}

type AddToGroupMembers struct {
	AccessLevel string   `json:"access_type"`
	User        EmailTag `json:"user"`
}

func (c *Client) AddUserToGroup(ctx context.Context, groupId, email, accessType string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := AddUserToGroupBody{
		Group: GroupIdTag{
			Tag:     "group_id",
			GroupID: groupId,
		},
		Members: []AddToGroupMembers{
			{
				AccessLevel: accessType,
				User: EmailTag{
					Tag:   "email",
					Email: email,
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AddUserToGroupURL, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return &ratelimitData, err
	}

	return &ratelimitData, nil
}
