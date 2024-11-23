package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-dropbox/pkg/connector"
	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

// var configure2 = flag.Bool("configure2", false, "configure the connector")

func main() {
	ctx := context.Background()

	// viper, cmd, err := config.DefineConfiguration(
	v, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-dropbox",
		getConnector,
		field.Configuration{
			Fields: ConfigurationFields,
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	configureArg := v.GetBool(ConfigureField.FieldName)
	if configureArg {
		fmt.Println("configure")
		client, err := dropbox.NewClient(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		appKey, appSecret := v.GetString(AppKey.FieldName), v.GetString(AppSecret.FieldName)
		code, err := client.Authorize(ctx, appKey, appSecret)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		_, _, refreshToken, err := client.RequestAccessToken(ctx, appKey, appSecret, code)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println("refresh token: ", refreshToken)
		os.Exit(0)
	}

	refreshTokenArg := v.GetString(RefreshTokenField.FieldName)
	if refreshTokenArg == "" {
		fmt.Fprintln(os.Stderr, "refresh-token is required")
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
