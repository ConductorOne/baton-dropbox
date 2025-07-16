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
		if res != nil && res.Body != nil {
			logBody(ctx, res.Body)
		}
		return nil, nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
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
		if res != nil && res.Body != nil {
			logBody(ctx, res.Body)
		}
		return nil, nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return nil, nil, err
	}

	return &target, &rateLimitData, nil
}
