package main

import (
	"fmt"
	"net/http"

	"github.com/eDyrr/url-shortener/handler"
	"github.com/eDyrr/url-shortener/store"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hey Go URL shortener",
		})
	})

	r.POST("/create-short-url", func(c *gin.Context) {
		handler.CreateShortUrl(c)
	})

	r.GET("/:shortUrl", func(c *gin.Context) {
		handler.HandleShortUrlRedirects(c)
	})

	store.InitializeStore()
	err := r.Run(":9808")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error %v", err))
	}
}
