package events

import (
	"log"
	"test-go-slack-bot/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleAppMentionEvent(event *slackevents.AppMentionEvent, api *slack.Client, c *gin.Context) error {
	msgBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.PlainTextType, "I have sent you a direct message", false, false),
		nil,
		nil,
	)
	s1, s2, err := api.PostMessage("U03HNSQ9R0C", slack.MsgOptionBlocks(msgBlock))
	if err != nil {
		panic(err)
	}
	log.Printf("s1: %s\n", utils.JSONString(s1))
	log.Printf("s2: %s\n", utils.JSONString(s2))

	return nil
}
