package commands

import (
	"log"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/slack-go/slack"
)

const (
	ID_PROJECT        = "project-select-block"
	ACTION_ID_PROJECT = "project-select-action"

	ID_START_DATE        = "start-date-picker-block"
	ACTION_ID_START_DATE = "start-date-picker-action"

	ID_START_TIME        = "start-time-picker-block"
	ACTION_ID_START_TIME = "start-time-picker-action"

	ID_END_TIME        = "end-time-picker-block"
	ACTION_ID_END_TIME = "end-time-picker-action"

	ID_NOTE        = "note-block"
	ACTION_ID_NOTE = "note-action"

	CALLBACK_ID_OT_MODAL_REQUEST = "callback-id-ot-modal-request"
)

const (
	ERROR_INVALID_DATE     = "invalid-date"
	ERROR_INVALID_END_TIME = "invalid-end-time"

	VALID = "valid"

	YYYYMMDD = "2006-01-02" // Trick
)

func handleOvertimeCommand(command slack.SlashCommand, api *slack.Client) error {
	projects, err := sheetutils.FetchProjects()
	if err != nil {
		return err
	}
	mvr := createOTModalBySDK(projects)
	resp, err := api.OpenView(command.TriggerID, *mvr)
	if err != nil {
		return err
	} else {
		// Print modal response
		log.Printf("Modal response: %s\n", utils.JSONString(resp))
	}
	return nil
}

func createOTModalBySDK(projects []types.Project) *slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject("plain_text", "kai", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Next", false, false)
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
				slack.NewTextBlockObject(slack.PlainTextType, p.Name, false, false),
			),
		)
	}
	projectSelectBlock := slack.NewInputBlock(
		ID_PROJECT,
		slack.NewTextBlockObject(slack.PlainTextType, "Project (*)", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Select the project you're currently working on.", false, false),
		slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, "Project", false, false),
			ACTION_ID_PROJECT,
			projectOptions...,
		),
	)
	// projectSelectBlock.DispatchAction = true
	startDatePickerBlock := slack.NewInputBlock(
		ID_START_DATE,
		slack.NewTextBlockObject(slack.PlainTextType, "Date", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Select a valid day on which you want to OT", false, false),
		slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    ACTION_ID_START_DATE,
			InitialDate: time.Now().Format(YYYYMMDD),
			Confirm: &slack.ConfirmationBlockObject{
				Title:   slack.NewTextBlockObject(slack.PlainTextType, "Confirm title", false, false),
				Text:    slack.NewTextBlockObject(slack.PlainTextType, "Confirm text", false, false),
				Confirm: slack.NewTextBlockObject(slack.PlainTextType, "Confirm confirm", false, false),
				Deny:    slack.NewTextBlockObject(slack.PlainTextType, "Confirm deny", false, false),
				Style:   slack.StyleDanger,
			},
		},
	)
	// startDatePickerBlock.DispatchAction = true
	// startDatePickerBlock.Optional = false
	startTimePickerBlock := slack.NewInputBlock(
		ID_START_TIME,
		slack.NewTextBlockObject(slack.PlainTextType, "Start from", false, false),
		nil,
		slack.NewTimePickerBlockElement(ACTION_ID_START_TIME),
	)
	// startTimePickerBlock.DispatchAction = true
	endTimePickerBlock := slack.NewInputBlock(
		ID_END_TIME,
		slack.NewTextBlockObject(slack.PlainTextType, "Finish at", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "AM demonstrates overnight OT", false, false),
		slack.NewTimePickerBlockElement(ACTION_ID_END_TIME),
	)
	// endTimePickerBlock.DispatchAction = true
	noteBlock := slack.NewInputBlock(
		ID_NOTE,
		slack.NewTextBlockObject(slack.PlainTextType, "Note", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Additional notes", false, false),
		slack.PlainTextInputBlockElement{
			Type:        slack.METPlainTextInput,
			ActionID:    ACTION_ID_NOTE,
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
	mvr := &slack.ModalViewRequest{
		Type:          slack.VTModal,
		Title:         titleText,
		Submit:        submitText,
		Close:         closeText,
		Blocks:        blocks,
		CallbackID:    CALLBACK_ID_OT_MODAL_REQUEST,
		ClearOnClose:  true,
		NotifyOnClose: true,
	}
	return mvr
}
