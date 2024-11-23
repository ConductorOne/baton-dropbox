package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AppKey = field.StringField(
		"app-key",
		field.WithDescription("Your app's key"),
		field.WithRequired(true),
	)
	AppSecret = field.StringField(
		"app-secret",
		field.WithDescription("Your app's secret"),
		field.WithRequired(true),
	)
	RefreshTokenField = field.StringField(
		"refresh-token",
		field.WithDescription("The refresh token used to get a new access token"),
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
)
