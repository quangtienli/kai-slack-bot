package commands

import (
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	ID_PAID_LEAVE_TYPE        = "id-paid-leave-type"
	ACTION_ID_PAID_LEAVE_TYPE = "action-id-paid-leave-type"

	CALLBACK_ID_PAID_LEAVE_MODAL = "callback-id-paid-leave-modal"

	ANNUAL_PL  = "Annual Leaved"
	SICK_PL    = "Sick Leaved"
	WEDDING_PL = "Wedding"
	FUNERAL_PL = "Funeral"
)

var (
	PL_TYPES = map[string]string{
		ANNUAL_PL:  ANNUAL_PL,
		SICK_PL:    SICK_PL,
		WEDDING_PL: WEDDING_PL,
		FUNERAL_PL: FUNERAL_PL,
	}
)

func handlePaidLeaveCommand(command slack.SlashCommand, api *slack.Client, c *gin.Context) error {
	mvr := buildPaidLeaveRequestModalBySDK()
	_, err := api.OpenView(command.TriggerID, mvr)
	if err != nil {
		return err
	}

	return nil
}

func buildPaidLeaveRequestModalBySDK() slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject(slack.PlainTextType, "Paid Leave Request", false, false)
	closeText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)
	nextText := slack.NewTextBlockObject(slack.PlainTextType, "Next", false, false)

	options := []*slack.OptionBlockObject{}
	for value, text := range PL_TYPES {
		option := slack.NewOptionBlockObject(
			value,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			nil,
		)
		options = append(options, option)
	}
	optionBlockElement := slack.NewRadioButtonsBlockElement(
		ACTION_ID_PAID_LEAVE_TYPE,
		options...,
	)
	optionBlock := slack.NewInputBlock(
		ID_PAID_LEAVE_TYPE,
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
		Type:          slack.VTModal,
		Title:         titleText,
		Submit:        nextText,
		Close:         closeText,
		Blocks:        blocks,
		CallbackID:    CALLBACK_ID_PAID_LEAVE_MODAL,
		NotifyOnClose: true,
		ClearOnClose:  true,
	}

	return mvr
}
