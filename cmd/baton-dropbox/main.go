package main

import (
	"context"

	cfg "github.com/conductorone/baton-dropbox/pkg/config"
	"github.com/conductorone/baton-dropbox/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(ctx, "baton-dropbox", version, cfg.ConfigurationSchema, connector.NewLambdaConnector, connectorrunner.WithSessionStoreEnabled())
}
