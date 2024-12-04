package dropbox

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type Client struct {
	uhttp.BaseHttpClient
	Config
	AccessToken string
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
		BaseHttpClient: *wrapper,
		Config:         config,
	}
	return client, nil
}
