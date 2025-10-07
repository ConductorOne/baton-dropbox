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

const listLimitDefault = 100

type ListUserBody struct {
	Limit          int  `json:"limit"`
	IncludeRemoved bool `json:"include_removed"`
}

func DefaultListUserBody() ListUserBody {
	return ListUserBody{
		Limit:          listLimitDefault,
		IncludeRemoved: false,
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
	res, err := c.Do(req,
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
	res, err := c.Do(req,
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

// AddMember creates a new user in Dropbox Team.
// https://www.dropbox.com/developers/documentation/http/teams#team-members-add
func (c *Client) AddMember(ctx context.Context, email string) (*AddMemberResponse, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := AddMemberRequest{
		NewMembers: []NewMemberInfo{
			{
				MemberEmail: email,
			},
		},
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AddMemberURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target AddMemberResponse
	var rateLimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, &rateLimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &rateLimitData, fmt.Errorf("error adding member: %s", res.Status)
	}

	return &target, &rateLimitData, nil
}

// RemoveMember removes a user from Dropbox Team using email.
// https://www.dropbox.com/developers/documentation/http/teams#team-members-remove
func (c *Client) RemoveMember(ctx context.Context, email string) (*RemoveMemberResponse, *v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	body := RemoveMemberRequest{
		User: EmailTag{
			Tag:   "email",
			Email: email,
		},
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RemoveMemberURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var target RemoveMemberResponse
	var rateLimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return nil, &rateLimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return nil, &rateLimitData, fmt.Errorf("error removing member: %s", res.Status)
	}

	return &target, &rateLimitData, nil
}

// SuspendMember suspends a user from Dropbox Team using email.
// https://www.dropbox.com/developers/documentation/http/teams#team-members-suspend
func (c *Client) SuspendMember(ctx context.Context, email string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := RemoveMemberRequest{
		User: EmailTag{
			Tag:   "email",
			Email: email,
		},
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SuspendMemberURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var rateLimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return &rateLimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return &rateLimitData, fmt.Errorf("error suspending member: %s", res.Status)
	}

	return &rateLimitData, nil
}

// UnsuspendMember unsuspends (reactivates) a user from Dropbox Team using email.
// https://www.dropbox.com/developers/documentation/http/teams#team-members-unsuspend
func (c *Client) UnsuspendMember(ctx context.Context, email string) (*v2.RateLimitDescription, error) {
	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	body := RemoveMemberRequest{
		User: EmailTag{
			Tag:   "email",
			Email: email,
		},
	}

	reader := new(bytes.Buffer)
	err = json.NewEncoder(reader).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, UnsuspendMemberURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	var rateLimitData v2.RateLimitDescription
	res, err := c.Do(req,
		uhttp.WithRatelimitData(&rateLimitData),
	)
	if err != nil {
		logBody(ctx, res)
		return &rateLimitData, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res)
		return &rateLimitData, fmt.Errorf("error unsuspending member: %s", res.Status)
	}

	return &rateLimitData, nil
}
