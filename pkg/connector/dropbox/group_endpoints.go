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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/list"), reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupsPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/list/continue"), reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupsPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/members/list"), reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/members/list/continue"), reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListGroupMembersPayload
	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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

func (c *Client) RemoveUserFromGroup(ctx context.Context, groupId string, teamMemberID string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := RemoveUserFromGroupBody{
		Group: GroupIdTag{
			GroupID: groupId,
			Tag:     "group_id",
		},
		Users: []TeamMemberIdTag{
			{
				Tag:          "team_member_id",
				TeamMemberID: teamMemberID,
			},
		},
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/members/remove"), buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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

func (c *Client) AddUserToGroup(ctx context.Context, groupId, teamMemberID, accessType string) (*v2.RateLimitDescription, error) {
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
				AccessLevel: Tag{Tag: accessType},
				User: TeamMemberIdTag{
					Tag:          "team_member_id",
					TeamMemberID: teamMemberID,
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/2/team/groups/members/add"), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
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
