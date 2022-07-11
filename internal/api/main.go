package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type validSonarQubeEndpointHeader struct {
	SonarQubeProject string `header:"X-SonarQube-Project" binding:"required"`
}

type validGiteaEndpointHeader struct {
	GiteaEvent string `header:"X-Gitea-Event" binding:"required"`
}

type ApiServer struct {
	Engine                  *gin.Engine
	sonarQubeWebhookHandler SonarQubeWebhookHandlerInferface
	giteaWebhookHandler     GiteaWebhookHandlerInferface
}

func (s *ApiServer) setup() {
	s.Engine.Use(gin.Recovery())
	s.Engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping", "/favicon.ico"},
	}))

	s.Engine.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	}).GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	}).POST("/hooks/sonarqube", func(c *gin.Context) {
		h := validSonarQubeEndpointHeader{}

		if err := c.ShouldBindHeader(&h); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		status, response := s.sonarQubeWebhookHandler.Handle(c.Request)
		c.JSON(status, gin.H{
			"message": response,
		})
	}).POST("/hooks/gitea", func(c *gin.Context) {
		h := validGiteaEndpointHeader{}

		if err := c.ShouldBindHeader(&h); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		var status int
		var response string

		switch h.GiteaEvent {
		case "pull_request":
			status, response = s.giteaWebhookHandler.HandleSynchronize(c.Request)
		case "issue_comment":
			status, response = s.giteaWebhookHandler.HandleComment(c.Request)
		default:
			status = http.StatusOK
			response = "ignore unknown event"
		}

		c.JSON(status, gin.H{
			"message": response,
		})
	})
}

func New(giteaHandler GiteaWebhookHandlerInferface, sonarQubeHandler SonarQubeWebhookHandlerInferface) *ApiServer {
	s := &ApiServer{
		Engine:                  gin.New(),
		giteaWebhookHandler:     giteaHandler,
		sonarQubeWebhookHandler: sonarQubeHandler,
	}

	s.setup()

	return s
}
