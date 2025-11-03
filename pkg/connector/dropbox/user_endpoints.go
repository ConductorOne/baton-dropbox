package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const listLimitDefault = 100

func DefaultListUserBody() ListUserBody {
	return ListUserBody{
		Limit:          listLimitDefault,
		IncludeRemoved: false,
	}
}

// userActionRequest represents a request body for user-related actions (suspend/unsuspend).
type userActionRequest struct {
	User TeamMemberIdTag `json:"user"`
}

// newUserActionRequest creates a user action request with the given team member ID.
func newUserActionRequest(teamMemberID string) userActionRequest {
	return userActionRequest{
		User: TeamMemberIdTag{
			Tag:          "team_member_id",
			TeamMemberID: teamMemberID,
		},
	}
}

func (c *Client) ListUsers(ctx context.Context, limit int) (*ListUsersPayload, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := DefaultListUserBody()
	if limit != 0 {
		body.Limit = limit
	}
	body.IncludeRemoved = true

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListUsersURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListUsersPayload
	var rateLimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, nil, err
	}

	return &target, &rateLimitData, nil
}

func (c *Client) ListUsersContinue(ctx context.Context, cursor string) (*ListUsersPayload, *v2.RateLimitDescription, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListUsersContinueURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target ListUsersPayload
	var rateLimitData v2.RateLimitDescription
	res, err := c.wrapper.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, nil, err
	}

	return &target, &rateLimitData, nil
}

// AddMember provisions a new team member.
// Based on API: POST /2/team/members/add_v2.
func (c *Client) AddMember(ctx context.Context, email string) (*AddMemberResponse, *v2.RateLimitDescription, error) {
	member := NewMemberInfo{MemberEmail: email}
	requestBody := AddMemberRequest{NewMembers: []NewMemberInfo{member}}

	result := &AddMemberResponse{}
	annos, err := c.doRequest(ctx, AddMemberURL, http.MethodPost, result, requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add member: %w", err)
	}

	return result, getRateLimitFromAnnos(annos), nil
}

// RemoveMember deprovisions a team member using their team_member_id.
// Based on API: POST /2/team/members/remove.
func (c *Client) RemoveMember(ctx context.Context, teamMemberID string) (*RemoveMemberResponse, *v2.RateLimitDescription, error) {
	requestBody := RemoveMemberRequest{
		User: TeamMemberIdTag{
			Tag:          "team_member_id",
			TeamMemberID: teamMemberID,
		},
	}

	result := &RemoveMemberResponse{}
	annos, err := c.doRequest(ctx, RemoveMemberURL, http.MethodPost, result, requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to remove member: %w", err)
	}

	if result.Tag == "" {
		return nil, getRateLimitFromAnnos(annos), fmt.Errorf("received empty response from Dropbox API")
	}

	return result, getRateLimitFromAnnos(annos), nil
}

// SuspendMember suspends a team member's access using their team_member_id.
// Based on API: POST /2/team/members/suspend.
func (c *Client) SuspendMember(ctx context.Context, teamMemberID string) (*v2.RateLimitDescription, error) {
	annos, err := c.doRequest(ctx, SuspendMemberURL, http.MethodPost, nil, newUserActionRequest(teamMemberID))
	if err != nil {
		return nil, fmt.Errorf("failed to suspend member: %w", err)
	}

	return getRateLimitFromAnnos(annos), nil
}

// UnsuspendMember reactivates a suspended team member using their team_member_id.
// Based on API: POST /2/team/members/unsuspend.
func (c *Client) UnsuspendMember(ctx context.Context, teamMemberID string) (*v2.RateLimitDescription, error) {
	annos, err := c.doRequest(ctx, UnsuspendMemberURL, http.MethodPost, nil, newUserActionRequest(teamMemberID))
	if err != nil {
		return nil, fmt.Errorf("failed to unsuspend member: %w", err)
	}

	return getRateLimitFromAnnos(annos), nil
}

// getRateLimitFromAnnos extracts rate limit data from annotations.
func getRateLimitFromAnnos(annos annotations.Annotations) *v2.RateLimitDescription {
	if annos == nil {
		return nil
	}

	var rateLimit v2.RateLimitDescription
	if _, err := annos.Pick(&rateLimit); err == nil {
		return &rateLimit
	}

	return nil
}
