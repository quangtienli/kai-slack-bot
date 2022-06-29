package main

import (
	"log"
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
	// testScript()
	PORT := os.Getenv("PORT")
	// var signingSecret string
	// flag.StringVar(&signingSecret, "secret", os.Getenv("SLACK_SIGNING_SECRET"), "Slack app's signing secret")
	// flag.Parse()
	// log.Printf("Secret by flag: %s\n", signingSecret)
	log.Printf("Secret by env: %s\n", os.Getenv("SLACK_SIGNING_SECRET"))
	api := botutils.InitSlackBotClient()
	router := gin.Default()
	{
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "App is running.")
		})
		router.POST("/json", func(c *gin.Context) {
		})
	}
	v1 := router.Group("/slack")
	{
		v1.POST("/events-endpoint", func(c *gin.Context) {
			events.HandleEventRequest(c, api)
		})
		v1.POST("/commands-endpoint", func(c *gin.Context) {
			commands.HandleCommandRequest(c, api)
		})
		v1.POST("/interactive-endpoint", func(c *gin.Context) {
			interactions.HandleInteractionRequest(c, api)
		})
	}
	router.Run(":" + PORT)
}
