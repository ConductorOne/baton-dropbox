package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cfg "github.com/conductorone/baton-dropbox/pkg/config"
	"github.com/conductorone/baton-dropbox/pkg/connector"
	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfigurationV2(
		ctx,
		"baton-dropbox",
		getConnector,
		cfg.ConfigurationSchema,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, dropboxCfg *cfg.Dropbox, runtimeOpts cli.RunTimeOpts) (types.ConnectorServer, error) {
	if err := field.Validate(cfg.ConfigurationSchema, dropboxCfg); err != nil {
		return nil, err
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

	opts := connector.WithRefreshToken(
		ctx,
		dropboxCfg.AppKey,
		dropboxCfg.AppSecret,
		dropboxCfg.RefreshToken,
	)

	if dropboxCfg.RefreshToken == "" {
		opts = connector.WithTokenSource(
			ctx,
			dropboxCfg.AppKey,
			runtimeOpts.TokenSource,
		)
	}
	cb, err := connector.New(ctx, opts)

	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
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
