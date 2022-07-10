package events

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleAppMentionEvent(event *slackevents.AppMentionEvent, api *slack.Client, c *gin.Context) {
	msgBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf("Hello <@%s>, I'm `Kai`.\n", event.User),
			false,
			false,
		),
		nil,
		nil,
		slack.SectionBlockOptionBlockID("event-app-mention"),
	)

	_, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(msgBlock))
	if err != nil {
		panic(err)
	}
}
