package main

import (
	"os"
	"path"

	"gitea-sonarqube-pr-bot/internal/settings"
)

func getConfigLocation() string {
	configPath := path.Join("config")
	if customConfigPath, ok := os.LookupEnv("PRBOT_CONFIG_PATH"); ok {
		configPath = customConfigPath
	}

	return configPath
}

func main() {
	settings.Load(getConfigLocation())
}
