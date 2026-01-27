package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	InstanceUrl = field.StringField(
		"instance-url",
		field.WithDescription("Your Ping Federate domain, ex: https://pingfederateserver.com"),
		field.WithRequired(true),
		field.WithDisplayName("Instance URL"),
	)
	Username = field.StringField(
		"username",
		field.WithDescription("Ping Federate account username"),
		field.WithRequired(true),
		field.WithDisplayName("Username"),
	)
	Password = field.StringField(
		"password",
		field.WithDescription("Ping Federate account password"),
		field.WithRequired(true),
		field.WithIsSecret(true),
		field.WithDisplayName("Password"),
	)

	// FieldRelationships defines relationships between the fields.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Configuration = field.NewConfiguration([]field.SchemaField{
	InstanceUrl,
	Username,
	Password,
}, field.WithConstraints(FieldRelationships...))
