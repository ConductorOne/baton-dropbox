package main

import (
	cfg "github.com/conductorone/baton-dropbox/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("dropbox", cfg.ConfigurationSchema)
}
