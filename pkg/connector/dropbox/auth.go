package dropbox

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"golang.org/x/term"
)

// returns access token and expiry time
//
//	curl https://api.dropbox.com/oauth2/token \
//	    -d grant_type=refresh_token \
//	    -d refresh_token=<refresh_token> \
//	    -d client_id=<app_key> \
//	    -d client_secret=<app_secret>
func (c *Client) RequestAccessTokenUsingRefreshToken(ctx context.Context, refreshToken, appKey, appSecret string) (string, *time.Time, error) {
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

func (c *Client) Authorize(ctx context.Context, appKey, appSecret string) (string, error) {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	if !isTTY {
		return "", fmt.Errorf("dropbox-connector: non-interactive mode not supported. Pass a refresh token as an argument ")
	}

	url := fmt.Sprintf(AuthURL + "?client_id=" + appKey + "&token_access_type=offline&response_type=code")
	fmt.Printf("\nOpen this link in your browser: " + url)
	fmt.Printf("\nPaste the code: ")

	var code string

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		code = scanner.Text()
		code = strings.TrimSpace(code) // Remove any leading or trailing whitespace
		fmt.Printf("You entered: %s\n", code)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading from stdin: %w", err)
	}

	return code, nil
}

func (c *Client) RequestAccessToken(ctx context.Context, appKey, appSecret, code string) (string, *time.Time, string, error) {
	grantType := "authorization_code"

	form := url.Values{}
	form.Set("grant_type", grantType)
	form.Set("client_id", appKey)
	form.Set("client_secret", appSecret)
	form.Set("code", code)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var target struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	res, err := c.Do(req,
		uhttp.WithJSONResponse(&target),
	)
	defer res.Body.Close()
	if err != nil {
		logBody(ctx, res.Body)
		return "", nil, "", err
	}

	if res.StatusCode != http.StatusOK {
		logBody(ctx, res.Body)
		return "", nil, "", fmt.Errorf("error getting access token: %s", res.Status)
	}

	logBody(ctx, res.Body)
	// __AUTO_GENERATED_PRINT_VAR_START__
	fmt.Println(fmt.Sprintf("RequestAccessToken target: %+v", target)) // __AUTO_GENERATED_PRINT_VAR_END__

	accessTokenexpiresIn := time.Now().Add(time.Duration(target.ExpiresIn) * time.Second)
	return target.AccessToken, &accessTokenexpiresIn, target.RefreshToken, nil
}