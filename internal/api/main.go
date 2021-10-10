package api

import (
	"fmt"
	"net/http"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea_sdk"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube_sdk"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

func addPingApi(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}

type validSonarQubeEndpointHeader struct {
	SonarQubeProject string `header:"X-SonarQube-Project" binding:"required"`
}

func addSonarQubeEndpoint(r *gin.Engine) {
	webhookHandler := NewSonarQubeWebhookHandler(giteaSdk.New(), sqSdk.New())
	r.POST("/hooks/sonarqube", func(c *gin.Context) {
		h := validSonarQubeEndpointHeader{}

		if err := c.ShouldBindHeader(&h); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		webhookHandler.Handle(c.Writer, c.Request)
	})
}

func addGiteaEndpoint(r *gin.Engine) {
	webhookHandler := NewGiteaWebhookHandler(giteaSdk.New(), sqSdk.New())
	r.POST("/hooks/gitea", func(c *gin.Context) {
		webhookHandler.Handle(c.Writer, c.Request)
	})
}

func Serve(c *cli.Context) error {
	fmt.Println("Hi! I'm the Gitea-SonarQube-PR bot. At your service.")

	r := gin.Default()

	addPingApi(r)
	addSonarQubeEndpoint(r)
	addGiteaEndpoint(r)

	return endless.ListenAndServe(":3000", r)
}
