package events

import (
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleAppMentionEvent(event *slackevents.AppMentionEvent, api *slack.Client, c *gin.Context) error {
	// msgBlock := slack.NewSectionBlock(
	// 	slack.NewTextBlockObject(slack.MarkdownType, "*Li Quang Tien*", false, false),
	// 	nil,
	// 	nil,
	// )

	// mvr := &slack.ModalViewRequest{
	// 	Type: slack.VTModal,
	// 	Title: slack.NewTextBlockObject(slack.PlainTextType, "Title of the modal", false, false),
	// 	Blocks: slack.Blocks{
	// 		BlockSet: []slack.Block{
	// 			msgBlock,
	// 		},
	// 	},
	// 	Close: nil,
	// 	Submit: nil,
	// 	ClearOnClose: false,
	// 	NotifyOnClose: false,
	// }

	return nil
}
