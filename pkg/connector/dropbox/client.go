package dropbox

import (
	"context"
	"fmt"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Client struct {
	wrapper     *uhttp.BaseHttpClient
	config      Config
	TokenSource oauth2.TokenSource
}

type Config struct {
	AppKey       string
	AppSecret    string
	RefreshToken string
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	httpClient, err := uhttp.NewClient(
		ctx,
		uhttp.WithLogger(
			true,
			ctxzap.Extract(ctx),
		),
	)
	if err != nil {
		return nil, err
	}

	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	client := &Client{
		wrapper: wrapper,
		config:  config,
	}
	return client, nil
}

// doRequest executes an HTTP request and decodes the response into the provided result.
// It handles authentication, headers, rate limiting, and error handling consistently.
func (c *Client) doRequest(
	ctx context.Context,
	endpointURL string,
	method string,
	result any,
	body any,
) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	token, err := c.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var reqOptions []uhttp.RequestOption
	if body != nil {
		reqOptions = append(reqOptions, uhttp.WithJSONBody(body))
	}

	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Build header options - only set Content-Type when there's a body
	headerOpts := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithBearerToken(token.AccessToken),
	}
	if body != nil {
		headerOpts = append(headerOpts, uhttp.WithContentTypeJSONHeader())
	}
	reqOptions = append(reqOptions, headerOpts...)

	request, err := c.wrapper.NewRequest(ctx, method, parsedURL, reqOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var doOptions []uhttp.DoOption
	var ratelimitData v2.RateLimitDescription
	var errResp DropboxError

	if result != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(result))
	}
	doOptions = append(doOptions,
		uhttp.WithRatelimitData(&ratelimitData),
		uhttp.WithErrorResponse(&errResp),
	)

	response, err := c.wrapper.Do(request, doOptions...)
	if err != nil {
		l.Debug("request failed",
			zap.String("url", endpointURL),
			zap.Error(err),
			zap.String("dropbox_error", errResp.Message()),
		)
		return nil, err
	}
	defer response.Body.Close()

	annos := annotations.Annotations{}
	annos.WithRateLimiting(&ratelimitData)

	return annos, nil
}
