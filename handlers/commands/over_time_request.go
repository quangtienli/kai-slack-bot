package commands

import (
	"fmt"
	"time"

	"test-go-slack-bot/botutils"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"github.com/slack-go/slack"
)

const (
	OtRequestProjectBlockID  = "ot-request-project-block-id"
	OtRequestProjectActionID = "ot-request-project-action-id"

	OtRequestStartDayBlockID  = "ot-request-start-day-block-id"
	OtRequestStartDayActionID = "ot-request-start-day-action-id"

	OtRequestStartTimeBlockID  = "ot-request-start-time-block-id"
	OtRequestStartTimeActionID = "ot-request-start-time-action-id"

	OtRequestEndTimeBlockID  = "ot-request-end-time-block-id"
	OtRequestEndTimeActionID = "ot-request-end-time-action-id"

	OtRequestNoteBlockID  = "ot-request-note-block-id"
	OtRequestNoteActionID = "ot-request-note-action-id"

	OtRequestModalCallbackID = "ot-request-modal-callback-id"

	// OtRequestLoadingMessageBlockID = "ot-request-loading-message-block-id"
)

func handleOvertTimeRequestCommand(command *slack.SlashCommand, api *slack.Client) {
	// log.Printf("Overtime command: %s\n", utils.JSONString(command))

	loadingMessageBlock := botutils.BuildResponseMessageBlockWithContext("Wait for me a second :blobdance: I'm loading the request for you.")
	botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, loadingMessageBlock)

	go func() {
		projects, err := sheetutils.FindAllProjects()
		if err != nil {
			errorMessage := fmt.Sprintf("Error while fetching projects: %s, please try again.\n", err.Error())
			errorMessageBlock := botutils.BuildResponseMessageBlockWithContext(errorMessage)
			botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, errorMessageBlock)
			return
		}

		mvr := buildOverTimeRequestModalBySDK(projects)
		_, err = api.OpenView(command.TriggerID, mvr)
		if err != nil {
			errorMessage := fmt.Sprintf("Error while opening ot request modal: %s, please try again.\n", err.Error())
			errorMessageBlock := botutils.BuildResponseMessageBlockWithContext(errorMessage)
			botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, errorMessageBlock)
			return
		}
	}()

}

func buildOverTimeRequestModalBySDK(projects []types.Project) slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject("plain_text", "kai", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Cancel", false, false)

	headerBlock := slack.NewHeaderBlock(
		slack.NewTextBlockObject(slack.PlainTextType, "OT Tracking Form", false, false),
	)

	var projectOptions []*slack.OptionBlockObject = []*slack.OptionBlockObject{}
	for _, p := range projects {
		projectOptions = append(
			projectOptions,
			slack.NewOptionBlockObject(
				p.Name,
				slack.NewTextBlockObject(slack.PlainTextType, p.Name, false, false),
				nil,
			),
		)
	}
	projectSelectBlock := slack.NewInputBlock(
		OtRequestProjectBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Project", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Select the project you're currently working on.", false, false),
		slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, "Project", false, false),
			OtRequestProjectActionID,
			projectOptions...,
		),
	)
	projectSelectBlock.Optional = false

	startDatePickerBlock := slack.NewInputBlock(
		OtRequestStartDayBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Date", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Select a valid day on which you want to OT", false, false),
		slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    OtRequestStartDayActionID,
			InitialDate: time.Now().Format(utils.LAYOUT_YYYYMMDD),
		},
	)
	startDatePickerBlock.Optional = false

	startTimePickerBlock := slack.NewInputBlock(
		OtRequestStartTimeBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Start at", false, false),
		nil,
		slack.NewTimePickerBlockElement(OtRequestStartTimeActionID),
	)

	endTimePickerBlock := slack.NewInputBlock(
		OtRequestEndTimeBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Finish at", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "End time should be within the day", false, false),
		slack.NewTimePickerBlockElement(OtRequestEndTimeActionID),
	)

	noteBlock := slack.NewInputBlock(
		OtRequestNoteBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Note", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Additional notes", false, false),
		slack.PlainTextInputBlockElement{
			Type:        slack.METPlainTextInput,
			ActionID:    OtRequestNoteActionID,
			Multiline:   true,
			Placeholder: slack.NewTextBlockObject(slack.PlainTextType, "Notes", false, false),
		},
	)
	noteBlock.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerBlock,
			projectSelectBlock,
			startDatePickerBlock,
			startTimePickerBlock,
			endTimePickerBlock,
			noteBlock,
		},
	}

	mvr := slack.ModalViewRequest{
		Type:          slack.VTModal,
		Title:         titleText,
		Submit:        submitText,
		Close:         closeText,
		Blocks:        blocks,
		CallbackID:    OtRequestModalCallbackID,
		ClearOnClose:  true,
		NotifyOnClose: true,
	}

	return mvr
}
