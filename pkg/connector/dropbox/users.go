package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const defaultListUserLimit = 100

type ListUserBody struct {
	Limit          int  `json:"limit"`
	IncludeRemoved bool `json:"include_removed"`
}

func DefaultListUserBody() ListUserBody {
	return ListUserBody{
		Limit:          defaultListUserLimit,
		IncludeRemoved: false,
	}
}

// TODO: https://www.dropbox.com/developers/documentation/http/teams#team-members-list-continue
func (c *Client) ListUsers(ctx context.Context, limit int, includeRemoved bool) (*ListUsersPayload, error) {
	body := DefaultListUserBody()
	if limit != 0 {
		body.Limit = limit
	}

	if includeRemoved {
		body.IncludeRemoved = true
	}

	reader := new(bytes.Buffer)
	err := json.NewEncoder(reader).Encode(body)
	req, err := http.NewRequest("POST", ListUsersURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	var target ListUsersPayload
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
