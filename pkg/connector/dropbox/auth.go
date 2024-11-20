package dropbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// returns access token and expiry time
//
//	curl https://api.dropbox.com/oauth2/token \
//	    -d grant_type=refresh_token \
//	    -d refresh_token=<refresh_token> \
//	    -d client_id=<app_key> \
//	    -d client_secret=<app_secret>
func (c *Client) RequestAccessToken(ctx context.Context, appKey, appSecret, refreshToken string) (string, *time.Time, error) {
	// get an access token using the refresh token
	grantType := "refresh_token"

	form := url.Values{}
	form.Set("client_id", appKey)
	form.Set("client_secret", appSecret)
	form.Set("refresh_token", refreshToken)
	form.Set("grant_type", grantType)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var target struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
	)
	defer res.Body.Close()
	if err != nil {
		logBody(ctx, res.Body)
		return "", nil, err
	}

	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return "", nil, fmt.Errorf("error getting access token: %s", res.Status)
	}

	expiresIn := time.Now().Add(time.Duration(target.ExpiresIn) * time.Second)
	return target.AccessToken, &expiresIn, nil

}
