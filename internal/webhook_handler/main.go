package webhook_handler

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
)

func Serve(c *cli.Context) error {
	fmt.Println("Hi! I'm the Gitea-SonarQube-PR bot. At your service.")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second * 15, "the duration for which the server gracefully wait for existing connections to finish")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/hooks/sonarqube", NewSonarQubeWebhookHandler().Handle).Methods("POST").Headers("X-SonarQube-Project", "")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout: time.Second * 15,
		IdleTimeout: time.Second * 60,
		Handler: r,
	}

	go func() {
		log.Println("Listen on :8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until we receive our signal.
	<-ch

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Shutting down webhook server")
	os.Exit(0)

	return nil
}
