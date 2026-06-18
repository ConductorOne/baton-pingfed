// Package connector implements the PingFederate Baton connector.
package connector

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/conductorone/baton-pingfed/pkg/connector/client"
	cfg "github.com/conductorone/baton-pingfed/pkg/config"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// Connector is the PingFederate connector implementation.
type Connector struct {
	ctx         context.Context
	instanceURL string
	client      *client.PingFederateClient
}

func fallBackToHTTPS(domain string) (string, error) {
	parsed, err := url.Parse(domain)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" {
		parsed, err = url.Parse(fmt.Sprintf("https://%s", domain))
		if err != nil {
			return "", err
		}
	}
	return parsed.String(), nil
}

// ResourceSyncers returns a ResourceSyncerV2 for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(d.client),
		newRoleBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(_ context.Context, _ *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Ping Federate",
		Description: "Connector syncing  PingFederate users",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(_ context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(
	ctx context.Context,
	instanceURL string,
	username string,
	password string,
) (*Connector, error) {
	logger := ctxzap.Extract(ctx)
	instanceURL, err := fallBackToHTTPS(instanceURL)
	if err != nil {
		return nil, err
	}

	logger.Debug(
		"New PingFederate connector",
		zap.String("instanceURL", instanceURL),
		zap.String("username", username),
		zap.Bool("password?", password != ""),
	)

	PingFederateClient, err := client.New(
		ctx,
		instanceURL,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}

	connector := Connector{
		client:      PingFederateClient,
		ctx:         ctx,
		instanceURL: instanceURL,
	}
	return &connector, nil
}

// NewLambdaConnector returns a new ConnectorBuilderV2 for use with lambda/containerized deployment.
func NewLambdaConnector(ctx context.Context, ac *cfg.Pingfed, _ *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	cb, err := New(ctx, ac.InstanceUrl, ac.Username, ac.Password)
	if err != nil {
		return nil, nil, err
	}
	return cb, nil, nil
}
