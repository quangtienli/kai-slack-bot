package events

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleAppMentionEvent(event *slackevents.AppMentionEvent, api *slack.Client) error {
	return nil
}
