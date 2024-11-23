package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *dropbox.Client
	config config
}

type config struct {
	appKey       string
	appSecret    string
	refreshToken string
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(c.client),
		newRoleBuilder(c.client),
		newGroupBuilder(c.client),
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
	accessToken, _, err := c.client.RequestAccessTokenUsingRefreshToken(ctx, c.config.refreshToken, c.config.appKey, c.config.appSecret)
	if err != nil {
		return nil, fmt.Errorf("dropbox-connector: error getting access token using refresh token: %w", err)
	}
	c.client.AccessToken = accessToken
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, appKey, appSecret, refreshToken string) (*Connector, error) {

	client, err := dropbox.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating dropbox client: %w", err)
	}

	return &Connector{
		client: client,
		config: config{
			appKey:       appKey,
			appSecret:    appSecret,
			refreshToken: refreshToken,
		}}, nil
}
