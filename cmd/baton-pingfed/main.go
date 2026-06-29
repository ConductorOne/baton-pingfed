package main

import (
	"context"

	cfg "github.com/conductorone/baton-pingfed/pkg/config"
	"github.com/conductorone/baton-pingfed/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(
		ctx,
		"baton-pingfed",
		version,
		cfg.Config,
		connector.NewLambdaConnector,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(&connector.Connector{}),
	)
}
