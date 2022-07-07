package botutils

import (
	"log"
	"os"

	"github.com/slack-go/slack"
)

const (
	TEST_BOT_CHANNEL_ID = "C03L398AKV0"
)

func InitSlackBotClient() *slack.Client {
	return slack.New(
		os.Getenv("SLACK_OAUTH_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		slack.OptionLog(log.New(os.Stdout, "Slack client: ", log.Lshortfile|log.LstdFlags)),
	)
}

func BuildResponseMessageBlockWithContext(ctx string) *slack.SectionBlock {
	return slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, ctx, false, false),
		nil,
		nil,
	)
}

func SendResponseMessage(channelID string, api *slack.Client, blocks ...slack.Block) {
	_, _, err := api.PostMessage(channelID, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		panic(err)
	}
}

func SendEphemeralResponseMessage(channelID string, userID string, api *slack.Client, blocks ...slack.Block) {
	_, err := api.PostEphemeral(channelID, userID, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		panic(err)
	}
}

func UpdateModal(view slack.View, newBlocks []slack.Block) *slack.ModalViewRequest {
	return &slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           view.Title,
		Close:           view.Close,
		Submit:          view.Submit,
		CallbackID:      view.CallbackID,
		ExternalID:      view.ExternalID,
		ClearOnClose:    view.ClearOnClose,
		PrivateMetadata: view.PrivateMetadata,
		NotifyOnClose:   view.NotifyOnClose,
		Blocks: slack.Blocks{
			BlockSet: newBlocks,
		},
	}
}
