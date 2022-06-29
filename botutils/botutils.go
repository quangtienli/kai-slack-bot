package botutils

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
)

// Initialize a new Slack App Client
func InitSlackBotClient() *slack.Client {
	return slack.New(
		os.Getenv("SLACK_OAUTH_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		slack.OptionLog(log.New(os.Stdout, "Slack client: ", log.Lshortfile|log.LstdFlags)),
	)
}

// Error: Can not send response message
func SendOTMessageResponse(userID string, block slack.Block, api *slack.Client) error {
	_, _, err := api.PostMessage(
		"C03L398AKV0",
		slack.MsgOptionBlocks(block),
		slack.MsgOptionPostEphemeral(userID),
	)
	if err != nil {
		return fmt.Errorf("Unable to send ot history response: %s\n", err.Error())
	}
	return nil
}

func SendOTHistoryResponse(userID string, api *slack.Client, blocks ...slack.Block) error {
	_, _, err := api.PostMessage(
		"C03L398AKV0",
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionPostEphemeral(userID),
	)
	if err != nil {
		return fmt.Errorf("Unable to send ot history response: %s\n", err.Error())
	}
	return nil
}

func UpdateModal(view slack.View, newBlocks []slack.Block) *slack.ModalViewRequest {
	mvr := &slack.ModalViewRequest{
		Type:   slack.VTModal,
		Title:  view.Title,
		Close:  view.Close,
		Submit: view.Submit,
		Blocks: slack.Blocks{
			BlockSet: newBlocks,
		},
		CallbackID:      view.CallbackID,
		ExternalID:      view.ExternalID,
		ClearOnClose:    view.ClearOnClose,
		PrivateMetadata: view.PrivateMetadata,
		NotifyOnClose:   view.NotifyOnClose,
	}
	return mvr
}

func SendFailMessage(err error, command slack.SlashCommand, api *slack.Client) {
}
