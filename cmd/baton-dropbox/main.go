package main

import (
	"context"
	"fmt"
	"log"

	cfg "github.com/conductorone/baton-dropbox/pkg/config"
	"github.com/conductorone/baton-dropbox/pkg/connector"
	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(ctx, "baton-dropbox", version, cfg.ConfigurationSchema, connector.NewLambdaConnector, connectorrunner.WithSessionStoreEnabled())
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
