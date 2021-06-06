package main

import (
  "net/http"
  "github.com/gin-gonic/gin"
)

func main() {
//  gin.SetMode(gin.ReleaseMode)

  server := gin.Default()

  server.GET("/", func(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, gin.H{"data": "Hi! I'm the Gitea-SonarQube-PR bot. At your service."})
  })

  server.Run()
}
