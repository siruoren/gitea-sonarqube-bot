package main

import (
	"log"
	"os"
	"path"

	"gitea-sonarqube-pr-bot/internal/settings"
	handler "gitea-sonarqube-pr-bot/internal/webhook_handler"
	"github.com/urfave/cli/v2"
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

	app := &cli.App{
		Name: "gitea-sonarqube-pr-bot",
		Usage: "Improve your experience with SonarQube and Gitea",
		Description: `By default, gitea-sonarqube-pr-bot will start running the webserver if no arguments are passed.`,
		Action: handler.Serve,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
