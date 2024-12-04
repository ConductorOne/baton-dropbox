package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/conductorone/baton-dropbox/pkg/connector"
	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

// var configure2 = flag.Bool("configure2", false, "configure the connector")

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-dropbox",
		getConnector,
		ConfigurationSchema,
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

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	configureArg := v.GetBool(ConfigureField.FieldName)
	if configureArg {
		if err := configure(ctx, v); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	if v.GetString(RefreshTokenField.FieldName) == "" {
		return nil, fmt.Errorf("refresh token is required, get it by running the connector with the --configure flag")
	}

	l := ctxzap.Extract(ctx)
	cb, err := connector.New(
		ctx,
		v.GetString(AppKey.FieldName),
		v.GetString(AppSecret.FieldName),
		v.GetString(RefreshTokenField.FieldName),
	)

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

func configure(ctx context.Context, v *viper.Viper) error {
	appKey, appSecret := v.GetString("app-key"), v.GetString("app-secret")

	if appKey == "" {
		return fmt.Errorf("app key is required")
	}
	if appSecret == "" {
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
