package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"golang.org/x/oauth2"
)

type Connector struct {
	client *dropbox.Client
}

// Option is a function that configures a Connector.
type Option func(*Connector) error

// WithRefreshToken configures the connector to use refresh token authentication.
func WithRefreshToken(ctx context.Context, appKey, appSecret, refreshToken string) Option {
	return func(c *Connector) error {
		client, err := dropbox.NewClient(ctx, dropbox.Config{
			AppKey:       appKey,
			AppSecret:    appSecret,
			RefreshToken: refreshToken,
		})
		if err != nil {
			return fmt.Errorf("error creating dropbox client: %w", err)
		}

		accessToken, _, err := client.RequestAccessTokenUsingRefreshToken(ctx)
		if err != nil {
			return fmt.Errorf("dropbox-connector: error getting access token using refresh token: %w", err)
		}
		client.TokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
		c.client = client
		return nil
	}
}

// WithTokenSource configures the connector to use a pre-configured token source.
func WithTokenSource(ctx context.Context, appKey string, tokenSource oauth2.TokenSource) Option {
	return func(c *Connector) error {
		client, err := dropbox.NewClient(ctx, dropbox.Config{
			AppKey: appKey,
		})
		if err != nil {
			return fmt.Errorf("error creating dropbox client: %w", err)
		}
		client.TokenSource = tokenSource
		c.client = client
		return nil
	}
}

// New returns a new instance of the connector.
func New(ctx context.Context, opts ...Option) (*Connector, error) {
	c := &Connector{}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if c.client == nil {
		return nil, fmt.Errorf("no client configuration provided")
	}

	return c, nil
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(c.client),
		newRoleBuilder(c.client),
		newGroupBuilder(c.client),
		// newFolderBuilder(c.client), // WIP
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (c *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (c *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "My Baton Connector",
		Description: "The template implementation of a baton connector",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}
