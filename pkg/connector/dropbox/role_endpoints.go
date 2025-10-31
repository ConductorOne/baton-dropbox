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

func (c *Client) AddRoleToUser(ctx context.Context, roleId string, teamMemberID string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := addRoleToUserBody{
		NewRoles:   []string{roleId},
		TeamMember: TeamMemberIdTag{Tag: "team_member_id", TeamMemberID: teamMemberID},
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SetRoleURL, buffer)
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

// endpoint only allows removing all roles, not specific roles
// also removing them all leaves the user with the member role by default
// https://www.dropbox.com/developers/documentation/http/teams#team-members-set_admin_permissions
func (c *Client) ClearRoles(ctx context.Context, teamMemberID string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := addRoleToUserBody{
		NewRoles:   []string{},
		TeamMember: TeamMemberIdTag{Tag: "team_member_id", TeamMemberID: teamMemberID},
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SetRoleURL, buffer)
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
