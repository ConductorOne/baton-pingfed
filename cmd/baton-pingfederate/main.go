package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-pingfed/pkg/config"
	"github.com/conductorone/baton-pingfed/pkg/connector"
	configSchema "github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := configSchema.DefineConfiguration(
		ctx,
		"baton-pingfederate",
		getConnector,
		config.Configuration,
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

func getConnector(ctx context.Context, cfg *config.Pingfed) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	cb, err := connector.New(
		ctx,
		cfg.InstanceUrl,
		cfg.Username,
		cfg.Password,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return c, nil
}
