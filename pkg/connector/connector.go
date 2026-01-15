package connector

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	cfg "github.com/conductorone/baton-dropbox/pkg/config"
	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var _ connectorbuilder.GlobalActionProvider = (*Connector)(nil)

type Connector struct {
	client *dropbox.Client
}

// Option is a function that configures a Connector.
type Option func(*Connector) error

// WithRefreshToken configures the connector to use refresh token authentication.
func WithRefreshToken(ctx context.Context, appKey, appSecret, refreshToken string) Option {
	return func(c *Connector) error {
		if refreshToken == "" {
			return fmt.Errorf("refresh token is required, get it by running the connector with the --configure flag")
		}

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

func NewLambdaConnector(ctx context.Context, dropboxCfg *cfg.Dropbox, cliOpts *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	if err := field.Validate(cfg.ConfigurationSchema, dropboxCfg); err != nil {
		return nil, nil, err
	}

	// In production, `dropboxCfg.Configure` is always false.
	if dropboxCfg.Configure {
		if err := configure(ctx, dropboxCfg); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	l := ctxzap.Extract(ctx)

	var opts Option
	if dropboxCfg.RefreshToken == "" {
		opts = WithTokenSource(
			ctx,
			dropboxCfg.AppKey,
			cliOpts.TokenSource,
		)
	} else {
		opts = WithRefreshToken(
			ctx,
			dropboxCfg.AppKey,
			dropboxCfg.AppSecret,
			dropboxCfg.RefreshToken,
		)
	}

	cb, err := New(ctx, opts)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, nil, err
	}
	return cb, nil, nil
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

func configure(ctx context.Context, dropboxCfg *cfg.Dropbox) error {
	appKey := dropboxCfg.AppKey
	if dropboxCfg.AppKey == "" {
		return fmt.Errorf("app key is required")
	}

	appSecret := dropboxCfg.AppSecret
	if dropboxCfg.AppSecret == "" {
		return fmt.Errorf("app secret is required")
	}

	client, err := dropbox.NewClient(ctx, dropbox.Config{
		AppKey:    appKey,
		AppSecret: appSecret,
	})
	if err != nil {
		return err
	}
	code, err := client.Authorize(ctx, appKey, appSecret)
	if err != nil {
		return err
	}

	_, _, refreshToken, err := client.RequestAccessToken(ctx, code)
	if err != nil {
		return err
	}
	log.Printf("\nrefresh token: %s\n", refreshToken)
	return nil
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
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
		DisplayName: "Dropbox Business Connector",
		Description: "The Dropbox Business connector syncs users, groups, and roles with account provisioning and deprovisioning support.",
		AccountCreationSchema: &v2.ConnectorAccountCreationSchema{
			FieldMap: map[string]*v2.ConnectorAccountCreationSchema_Field{
				"email": {
					DisplayName: "Email",
					Required:    true,
					Description: "Email address for the user account. Dropbox will send an invitation to this address.",
					Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
						StringField: &v2.ConnectorAccountCreationSchema_StringField{},
					},
					Placeholder: "john@doe.com",
					Order:       0,
				},
			},
		},
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}
