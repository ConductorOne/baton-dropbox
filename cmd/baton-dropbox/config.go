package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AppKey = field.StringField(
		"app-key",
		field.WithDescription("The app key used to authenticate with Dropbox"),
		field.WithRequired(true),
	)
	AppSecret = field.StringField(
		"app-secret",
		field.WithDescription("The app secret used to authenticate with Dropbox"),
		field.WithRequired(true),
	)
	RefreshTokenField = field.StringField(
		"refresh-token",
		field.WithDescription("The refresh token used to get an access token for authentication with Dropbox"),
		field.WithRequired(false),
	)
	ConfigureField = field.BoolField(
		"configure",
		field.WithDescription("Get the refresh token the first time you run the connector."),
		field.WithRequired(false),
	)
	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{
		AppKey,
		AppSecret,
		RefreshTokenField,
		ConfigureField,
	}
	ConfigurationSchema = field.NewConfiguration(ConfigurationFields)
)
