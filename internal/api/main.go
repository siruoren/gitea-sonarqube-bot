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

		s.sonarQubeWebhookHandler.Handle(c.Writer, c.Request)
	}).POST("/hooks/gitea", func(c *gin.Context) {
		h := validGiteaEndpointHeader{}

		if err := c.ShouldBindHeader(&h); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		switch h.GiteaEvent {
		case "pull_request":
			s.giteaWebhookHandler.HandleSynchronize(c.Writer, c.Request)
		case "issue_comment":
			s.giteaWebhookHandler.HandleComment(c.Writer, c.Request)
		default:
			c.JSON(http.StatusOK, gin.H{
				"message": "ignore unknown event",
			})
		}
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
