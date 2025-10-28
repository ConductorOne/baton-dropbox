package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AppKey = field.StringField(
		"app-key",
		field.WithDisplayName("App key"),
		field.WithDescription("The app key used to authenticate with Dropbox"),
		field.WithRequired(true),
	)
	AppSecret = field.StringField(
		"app-secret",
		field.WithDisplayName("App secret"),
		field.WithIsSecret(true),
		field.WithDescription("The app secret used to authenticate with Dropbox"),
		field.WithRequired(true),
	)
	RefreshTokenField = field.StringField(
		"refresh-token",
		field.WithDisplayName("OAuth refresh token"),
		field.WithIsSecret(true),
		field.WithDescription("The refresh token used to get an access token for authentication with Dropbox"),
		field.WithRequired(false),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)
	ConfigureField = field.BoolField(
		"configure",
		field.WithDisplayName("Configure"),
		field.WithDescription("Get the refresh token the first time you run the connector."),
		field.WithRequired(false),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)

	// /
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

//go:generate go run ./gen
var ConfigurationSchema = field.NewConfiguration(
	ConfigurationFields,
	field.WithConnectorDisplayName("Dropbox v2"),
	field.WithHelpUrl("/docs/baton/dropbox"),
	field.WithIconUrl("/static/app-icons/dropbox.svg"),
)
