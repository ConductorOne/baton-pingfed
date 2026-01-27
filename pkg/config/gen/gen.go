package main

import (
	cfg "github.com/conductorone/baton-pingfed/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("pingfed", cfg.Configuration)
}
