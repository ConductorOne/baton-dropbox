package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

type addRoleToUserBody struct {
	NewRoles   []string `json:"new_roles"`
	TeamMember EmailTag `json:"user"`
}

func (c *Client) AddRoleToUser(ctx context.Context, roleId, email string) (*v2.RateLimitDescription, error) {
	body := addRoleToUserBody{
		NewRoles:   []string{roleId},
		TeamMember: EmailTag{Tag: "email", Email: email},
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SetRoleURL, buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		return &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return &ratelimitData, err
	}

	return &ratelimitData, nil
}

// endpoint only allows removing all roles, not specific roles
// also removing them all leaves the user with the member role by default
// https://www.dropbox.com/developers/documentation/http/teams#team-members-set_admin_permissions
func (c *Client) ClearRoles(ctx context.Context, email string) (*v2.RateLimitDescription, error) {
	body := addRoleToUserBody{
		NewRoles:   []string{},
		TeamMember: EmailTag{Tag: "email", Email: email},
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SetRoleURL, buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var ratelimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&ratelimitData),
	)

	if err != nil {
		logBody(ctx, res.Body)
		return &ratelimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return &ratelimitData, err
	}

	return &ratelimitData, nil
}
