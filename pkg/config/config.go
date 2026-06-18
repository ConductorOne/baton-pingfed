// Package config provides the PingFederate connector configuration.
package config

import "github.com/conductorone/baton-sdk/pkg/field"

var (
	// InstanceURLField is the PingFederate instance URL configuration field.
	InstanceURLField = field.StringField(
		"instance-url",
		field.WithDescription("Your Ping Federate domain, ex: https://pingfederateserver.com"),
		field.WithDisplayName("Instance URL"),
		field.WithPlaceholder("https://pingfederateserver.com"),
		field.WithRequired(true),
	)
	// UsernameField is the PingFederate username configuration field.
	UsernameField = field.StringField(
		"username",
		field.WithDescription("Ping Federate account username"),
		field.WithDisplayName("Username"),
		field.WithPlaceholder("Your PingFederate username"),
		field.WithRequired(true),
	)
	// PasswordField is the PingFederate password configuration field.
	PasswordField = field.StringField(
		"password",
		field.WithDescription("Ping Federate account password"),
		field.WithDisplayName("Password"),
		field.WithPlaceholder("Your PingFederate password"),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen

// Config is the PingFederate connector field configuration.
var Config = field.NewConfiguration(
	[]field.SchemaField{
		InstanceURLField,
		UsernameField,
		PasswordField,
	},
	field.WithConnectorDisplayName("PingFederate"),
	field.WithIconUrl("/static/app-icons/pingfed.svg"),
	field.WithHelpUrl("/docs/baton/pingfed"),
)
