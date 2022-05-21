package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"gitea-sonarqube-pr-bot/internal/api"
	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sonarQubeSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	"gitea-sonarqube-pr-bot/internal/settings"

	"github.com/fvbock/endless"
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
		Name:        "gitea-sonarqube-pr-bot",
		Usage:       "Improve your experience with SonarQube and Gitea",
		Description: `By default, gitea-sonarqube-pr-bot will start running the webserver if no arguments are passed.`,
		Action:      serveApi,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func serveApi(c *cli.Context) error {
	fmt.Println("Hi! I'm the Gitea-SonarQube-PR bot. At your service.")

	giteaHandler := api.NewGiteaWebhookHandler(giteaSdk.New(), sonarQubeSdk.New())
	sqHandler := api.NewSonarQubeWebhookHandler(giteaSdk.New(), sonarQubeSdk.New())
	server := api.New(giteaHandler, sqHandler)

	return endless.ListenAndServe(":3000", server.Engine)
}
