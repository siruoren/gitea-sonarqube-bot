package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitea-sonarqube-bot/internal/api"
	giteaSdk "gitea-sonarqube-bot/internal/clients/gitea"
	sonarQubeSdk "gitea-sonarqube-bot/internal/clients/sonarqube"
	"gitea-sonarqube-bot/internal/settings"

	"github.com/urfave/cli/v2"
)

var (
	HammerTime time.Duration = 15 * time.Second
)

func main() {
	app := &cli.App{
		Name:        "gitea-sonarqube-bot",
		Usage:       "Improve your experience with SonarQube and Gitea",
		Description: `Start an instance of gitea-sonarqube-bot to integrate SonarQube analysis into Gitea Pull Requests.`,
		Action:      serveApi,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Value:     "./config/config.yaml",
				Usage:     "Full path to configuration file.",
				EnvVars:   []string{"GITEA_SQ_BOT_CONFIG_PATH"},
				TakesFile: true,
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   3000,
				Usage:   "Port the bot will listen on.",
				EnvVars: []string{"GITEA_SQ_BOT_PORT"},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func serveApi(c *cli.Context) error {
	config := c.Path("config")
	settings.Load(config)

	log.Println("Hi! I'm Gitea SonarQube Bot. At your service.")
	log.Println("Config file in use:", config)

	giteaHandler := api.NewGiteaWebhookHandler(giteaSdk.New(), sonarQubeSdk.New(&settings.SonarQube))
	sqHandler := api.NewSonarQubeWebhookHandler(giteaSdk.New(), sonarQubeSdk.New(&settings.SonarQube))
	server := api.New(giteaHandler, sqHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Int("port")),
		Handler: server.Engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	log.Println("Listen on", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), HammerTime)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("[STOP - Hammer Time] Forcefully shutting down\n", err)
	}

	return nil
}
