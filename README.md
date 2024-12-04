![Baton Logo](./docs/images/baton-logo.png)

# `baton-dropbox` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-dropbox.svg)](https://pkg.go.dev/github.com/conductorone/baton-dropbox) ![main ci](https://github.com/conductorone/baton-dropbox/actions/workflows/main.yaml/badge.svg)

`baton-dropbox` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## Prerequisites

You need to pass an application key, secret, and refresh token to the connector. You can get these by following these steps:
1. Create a Dropbox app. You can follow [this Dropbox quickstart guide](https://www.dropbox.com/developers/reference/getting-started) and click on the "App Console" link.
2. You need to set the "Team Scopes" in the Permissions tab.
3. Copy the application's key and secret.
4. Get a refresh token, you can get one by using the --configure flag.

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-dropbox
baton-dropbox
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-dropbox:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-dropbox/cmd/baton-dropbox@main

baton-dropbox

baton resources
```

# Data Model

`baton-dropbox` will pull down information about the following resources:
- Users
- Roles
- Groups

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-dropbox` Command Line Usage

```
baton-dropbox

Usage:
  baton-dropbox [flags]
  baton-dropbox [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --configure bool               Get the refresh token the first time you run the connector.
      --refresh-token string         The refresh token used to get an access token for authentication with Dropbox ($BATON_REFRESH_TOKEN)
      --app-key string               The app key used to authenticate with Dropbox ($BATON_APP_KEY)
      --app-secret string            The app secret used to authenticate with Dropbox ($BATON_APP_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-dropbox
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-dropbox

Use "baton-dropbox [command] --help" for more information about a command.
```
