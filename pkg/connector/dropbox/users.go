package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type ListUserArgs struct {
	Limit          int  `json:"limit"`
	IncludeRemoved bool `json:"include_removed"`
}

func DefaultListUserBody() ListUserArgs {
	return ListUserArgs{
		Limit:          100,
		IncludeRemoved: false,
	}

}

type ListUserOption func(*ListUserArgs)

func WithLimit(limit int) ListUserOption {
	return func(args *ListUserArgs) {
		args.Limit = limit
	}
}

func WithIncludeRemoved() ListUserOption {
	return func(args *ListUserArgs) {
		args.IncludeRemoved = true
	}
}

func (c *Client) ListUsers(ctx context.Context, opts ...ListUserOption) (*ListUsersPayload, error) {
	// l := ctxzap.Extract(ctx)

	body := DefaultListUserBody()
	for _, opt := range opts {
		opt(&body)
	}

	reader := new(bytes.Buffer)
	err := json.NewEncoder(reader).Encode(body)
	req, err := http.NewRequest("POST", ListUsersURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return nil, err
	}
	var listUsersPayload ListUsersPayload
	err = json.NewDecoder(res.Body).Decode(&listUsersPayload)
	if err != nil {
		return nil, err
	}

	return &listUsersPayload, nil
}
