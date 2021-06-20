package main

import (
	"path"

	"github.com/justusbunsi/gitea-sonarqube-pr-bot/internal/settings"
)

func main() {
	configPath := path.Join("config")
	settings.Load(configPath)
}
