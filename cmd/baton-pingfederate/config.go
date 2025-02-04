package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	InstanceUrlField = field.StringField(
		"instance-url",
		field.WithDescription("Your Ping Federate domain, ex: https://pingfederateserver.com"),
		field.WithRequired(true),
	)
	UsernameField = field.StringField(
		"username",
		field.WithDescription("Ping Federate account username"),
		field.WithRequired(true),
	)
	PasswordField = field.StringField(
		"password",
		field.WithDescription("Ping Federate account password"),
		field.WithRequired(true),
	)

	configurationFields = []field.SchemaField{
		InstanceUrlField,
		UsernameField,
		PasswordField,
	}
	Configuration = field.NewConfiguration(
		configurationFields,
	)
)
