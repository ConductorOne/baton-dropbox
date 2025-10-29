package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

type addRoleToUserBody struct {
	NewRoles   []string `json:"new_roles"`
	TeamMember EmailTag `json:"user"`
}

func (c *Client) AddRoleToUser(ctx context.Context, roleId, email string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := addRoleToUserBody{
		NewRoles:   []string{roleId},
		TeamMember: EmailTag{Tag: "email", Email: email},
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
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&ratelimitData),
	)

	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()

	// Check if the error indicates the user already has the role - treat as idempotent success
	if res != nil && res.StatusCode == http.StatusConflict {
		bodyBytes, readErr := io.ReadAll(res.Body)
		if readErr == nil {
			var errorResp struct {
				ErrorSummary string `json:"error_summary"`
				Error        struct {
					Tag string `json:".tag"`
				} `json:"error"`
			}
			if jsonErr := json.Unmarshal(bodyBytes, &errorResp); jsonErr == nil {
				// Treat certain errors as idempotent (user already has the role)
				if errorResp.Error.Tag == "duplicate_user" || errorResp.Error.Tag == "user_already_has_role" {
					return &ratelimitData, nil
				}
			}
			// Restore the body for logBody
			res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	if err != nil {
		logBody(ctx, res)
		return &ratelimitData, err
	}

	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return &ratelimitData, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return &ratelimitData, nil
}

// endpoint only allows removing all roles, not specific roles
// also removing them all leaves the user with the member role by default
// https://www.dropbox.com/developers/documentation/http/teams#team-members-set_admin_permissions
func (c *Client) ClearRoles(ctx context.Context, email string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := addRoleToUserBody{
		NewRoles:   []string{},
		TeamMember: EmailTag{Tag: "email", Email: email},
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
