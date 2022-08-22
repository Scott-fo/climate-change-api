package main

import (
	"net/http"

	"github.com/Scott-fo/climate-change-api/service"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/news", service.GetNews)
	router.GET("/news/:source", service.GetNewsBySource)
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "View /news and /news/source to see articles related to Climate Change")
	})

	router.Run(":3000")
}
