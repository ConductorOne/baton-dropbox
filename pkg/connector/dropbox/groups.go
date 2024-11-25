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

const defaultLimit = 100

type ListGroupsBody struct {
	Limit int `json:"limit"`
}

func DefaultListGroupsBody() ListGroupsBody {
	return ListGroupsBody{
		Limit: defaultLimit,
	}
}

// docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-list
func (c *Client) ListGroups(ctx context.Context, limit int) (*ListGroupsPayload, error) {
	body := DefaultListGroupsBody()
	if limit != 0 {
		body.Limit = limit
	}

	reader := new(bytes.Buffer)
	err := json.NewEncoder(reader).Encode(body)
	req, err := http.NewRequest("POST", ListGroupsURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupsPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return nil, err
	}

	return &target, nil
}

type ListGroupMembersBody struct {
	Group GroupMembersBody `json:"group"`
	Limit int              `json:"limit"`
}

type GroupMembersBody struct {
	GroupID string `json:"group_id"`
	Tag     string `json:".tag"`
}

func DefaultGroupMembersBody() ListGroupMembersBody {
	return ListGroupMembersBody{
		Group: GroupMembersBody{
			Tag: "group_id",
		},
		Limit: defaultLimit,
	}
}

func (c *Client) ListGroupMembers(ctx context.Context, groupId string, limit int) (*ListGroupMembersPayload, *v2.RateLimitDescription, error) {
	body := DefaultGroupMembersBody()
	if groupId == "" {
		return nil, nil, fmt.Errorf("groupId is required")
	}
	body.Group.GroupID = groupId

	if limit != 0 {
		body.Limit = limit
	}

	reader := new(bytes.Buffer)
	err := json.NewEncoder(reader).Encode(body)
	req, err := http.NewRequest("POST", ListGroupMembersURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}

func (c *Client) ListGroupMembersContinue(ctx context.Context, cursor string) (*ListGroupMembersPayload, *v2.RateLimitDescription, error) {
	body := struct {
		Cursor string `json:"cursor"`
	}{Cursor: cursor}

	reader := new(bytes.Buffer)
	err := json.NewEncoder(reader).Encode(body)
	req, err := http.NewRequest("POST", ListGroupMembersURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		return nil, &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return nil, &ratelimitData, err
	}

	return &target, &ratelimitData, nil
}
