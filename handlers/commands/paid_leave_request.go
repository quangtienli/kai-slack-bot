package commands

import (
	"fmt"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/types"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	PlRequestTypeBlockID     = "pl-request-type-id"
	PlRequestTypeActionID    = "pl-request-type-action-id"
	PlRequestModalCallbackID = "pl-request-modal-callback-id"
)

func handlePaidLeaveRequestCommand(command *slack.SlashCommand, api *slack.Client, c *gin.Context) {
	mvr := buildPaidLeaveRequestModalBySDK()
	mvr.ExternalID = command.TriggerID

	_, err := api.OpenView(command.TriggerID, mvr)

	if err != nil {
		errorMessage := fmt.Sprintf("Unable to open paid leave request view: %s\nn", err.Error())
		errorMessageBlock := botutils.BuildResponseMessageBlockWithContext(errorMessage)
		botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, errorMessageBlock)
		return
	}
}

func buildPaidLeaveRequestModalBySDK() slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject(slack.PlainTextType, "Paid Leave Request", false, false)
	closeText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)
	nextText := slack.NewTextBlockObject(slack.PlainTextType, "Next", false, false)

	options := []*slack.OptionBlockObject{}
	for _, plType := range types.PaidLeaveTypes {
		option := slack.NewOptionBlockObject(
			plType,
			slack.NewTextBlockObject(slack.PlainTextType, plType, false, false),
			nil,
		)
		options = append(options, option)
	}
	optionBlockElement := slack.NewRadioButtonsBlockElement(
		PlRequestTypeActionID,
		options...,
	)
	optionBlock := slack.NewInputBlock(
		PlRequestTypeBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Select a type of paid leave to request", false, false),
		nil,
		optionBlockElement,
	)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			optionBlock,
		},
	}

	mvr := slack.ModalViewRequest{
		Type:       slack.VTModal,
		Title:      titleText,
		Submit:     nextText,
		Close:      closeText,
		Blocks:     blocks,
		CallbackID: PlRequestModalCallbackID,
	}

	return mvr
}
