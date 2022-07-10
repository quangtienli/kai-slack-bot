package main

import (
	"net/http"
	"os"

	"test-go-slack-bot/botutils"

	// "test-go-slack-bot/botutils/handlers/commands"
	"test-go-slack-bot/handlers/commands"
	"test-go-slack-bot/handlers/events"
	"test-go-slack-bot/handlers/interactions"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	PORT := os.Getenv("PORT")
	api := botutils.InitSlackBotClient()
	router := gin.Default()
	router.Use(gin.Logger())
	{
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "App is running.")
		})
	}
	v1 := router.Group("/slack")
	{
		v1.POST("/event", func(c *gin.Context) {
			events.HandleEventRequest(c, api)
		})
		v1.POST("/command", func(c *gin.Context) {
			commands.HandleCommandRequest(c, api)
		})
		v1.POST("/interaction", func(c *gin.Context) {
			interactions.HandleInteractionRequest(c, api)
		})
	}
	router.Run(":" + PORT)
}
