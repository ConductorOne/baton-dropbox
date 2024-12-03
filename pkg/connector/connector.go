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

//NOTE: validate doesn't get called when running grant and revoke, which is why they are not working

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, appKey, appSecret, refreshToken string) (*Connector, error) {
	client, err := dropbox.NewClient(ctx, dropbox.Config{
		AppKey:       appKey,
		AppSecret:    appSecret,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating dropbox client: %w", err)
	}

	accessToken, _, err := client.RequestAccessTokenUsingRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("dropbox-connector: error getting access token using refresh token: %w", err)
	}
	client.AccessToken = accessToken
	return &Connector{
		client: client,
	}, nil
}
