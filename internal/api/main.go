package api

import (
	"fmt"
	"net/http"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"

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

type validGiteaEndpointHeader struct {
	GiteaEvent string `header:"X-Gitea-Event" binding:"required"`
}

func addGiteaEndpoint(r *gin.Engine) {
	webhookHandler := NewGiteaWebhookHandler(giteaSdk.New(), sqSdk.New())
	r.POST("/hooks/gitea", func(c *gin.Context) {
		h := validGiteaEndpointHeader{}

		if err := c.ShouldBindHeader(&h); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		switch h.GiteaEvent {
		case "pull_request":
			webhookHandler.HandleSynchronize(c.Writer, c.Request)
		case "issue_comment":
			webhookHandler.HandleComment(c.Writer, c.Request)
		default:
			c.JSON(http.StatusOK, gin.H{
				"message": "ignore unknown event",
			})
		}
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
