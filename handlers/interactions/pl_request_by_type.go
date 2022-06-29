package interactions

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/handlers/commands"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	// Can be used for all types of pto request
	ID_LEAVE_DAY           = "id-leave-day"
	ACTION_ID_LEAVE_DATE   = "action-id-leave-date"
	ACTION_ID_LEAVE_OPTION = "action-id-time-option"

	ACTION_ID_REMOVE_PAID_LEAVE = "action-id-remove-paid-leave"
	VALUE_REMOVE_PAID_LEAVE     = "value-remove-paid-leave"

	ID_LEAVE_MORE_DAY        = "id-leave-more-day"
	ACTION_ID_LEAVE_MORE_DAY = "action-id-leave-more-day"
	VALUE_LEAVE_MORE_DAY     = "value-leave-more-day"

	ID_LEAVE_NOTE        = "id-leave-note"
	ACTION_ID_LEAVE_NOTE = "action-id-leave-note"

	// Unique callback id to identify a pto modal
	CALLBACK_ID_ANNUAL_LEAVE_MODAL  = "callback-id-annual-leave-modal"
	CALLBACK_ID_SICK_LEAVE_MODAL    = "callback-id-sick-leave-modal"
	CALLBACK_ID_WEDDING_LEAVE_MODAL = "callback-id-wedding-leave-modal"
	CALLBACK_ID_FUNERAL_LEAVE_MODAL = "callback-id-funeral-leave-modal"
)

const (
	FULL_DAY = "Full day"
	HALF_DAY = "Half day"
)

var (
	plOptions = map[string]string{
		FULL_DAY: "1",
		HALF_DAY: "0.5",
	}
)

// Open paid leave modal request
func handlePaidLeaveRequest(msg slack.InteractionCallback, api *slack.Client, c *gin.Context) error {
	plType := msg.View.State.Values[commands.ID_PAID_LEAVE_TYPE][commands.ACTION_ID_PAID_LEAVE_TYPE].SelectedOption.Value
	user, err := api.GetUserInfo(msg.User.ID)
	if err != nil {
		return err
	}
	count := sheetutils.GetRemainingDaysByPLType(plType, user)
	mvr := buildPaidLeaveRequestModalByType(plType, user, count)

	// Not working, why?
	// _, err = api.PushView(msg.TriggerID, *mvr)
	// if err != nil {
	// 	return err
	// } else {
	// 	// log.Printf("Modal response: %s\n", utils.JSONString(resp))
	// }

	resp := slack.NewPushViewSubmissionResponse(mvr)
	c.JSON(http.StatusOK, resp)

	return nil
}

// Open paid leave modal request with an additional day picker
func handleMoreDayPaidLeaveRequest(msg slack.InteractionCallback, api *slack.Client, c *gin.Context) error {
	log.Printf("After click add date: %s\n", utils.JSONString(msg))
	mvr := buildMoreDayPaidLeaveRequestModalByType(&msg.View)
	_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
	if err != nil {
		return err
	}

	return nil
}

// Open paid leave modal request with less an additional day picker
func handleLessDayPaidLeaveRequest(msg slack.InteractionCallback, api *slack.Client, c *gin.Context) error {
	blocks := removeClickedDayBlock(msg)
	mvr := botutils.UpdateModal(msg.View, blocks)
	_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
	if err != nil {
		return err
	}

	return nil
}

// Remove the day action block that was clicked to be removed
func removeClickedDayBlock(msg slack.InteractionCallback) []slack.Block {
	clikedBlockID := msg.ActionCallback.BlockActions[0].BlockID
	blocks := msg.View.Blocks.BlockSet
	values := msg.View.State.Values

	// Find its index
	clickedBlockIndex := 0
	for i, block := range blocks {
		if actionBlock, ok := isDayActionBlock(block); ok {
			if actionBlock.BlockID == clikedBlockID {
				clickedBlockIndex = i
				break
			}
		}
	}

	// Remove it
	blocks = append(blocks[0:clickedBlockIndex], blocks[clickedBlockIndex+1:]...)

	// Since "blocks" only contain the UI -> update value for every other block
	for _, block := range blocks {
		if actionBlock, ok := isDayActionBlock(block); ok {
			d := actionBlock.Elements.ElementSet[0].(*slack.DatePickerBlockElement)
			date := values[actionBlock.BlockID][d.ActionID].SelectedDate
			if date != "" {
				d.InitialDate = date
			}

			o := actionBlock.Elements.ElementSet[1].(*slack.SelectBlockElement)
			option := values[actionBlock.BlockID][o.ActionID].SelectedOption
			if option.Value != "" {
				o.InitialOption = slack.NewOptionBlockObject(
					option.Value,
					option.Text,
					option.Description,
				)
			}
		}
	}

	// If after remove and there's only one block left, make it unremovable
	if count := countDayBlock(blocks); count == 1 {
		for _, block := range blocks {
			if actionBlock, ok := isDayActionBlock(block); ok {
				// Remove the button
				actionBlock.Elements.ElementSet = actionBlock.Elements.ElementSet[:2]
			}
		}
	}

	return blocks
}

func buildMoreDayPaidLeaveRequestModalByType(view *slack.View) *slack.ModalViewRequest {
	blocks := view.Blocks.BlockSet
	idxToInsert := indexToInsertDayBlock(blocks)
	serializedIdx := getSerializedIndex(blocks)

	// Create a new block with the serialized id
	dayActionBlock := createDayActionBlockBySDK(
		uniqueDateActionID(serializedIdx),
		uniqueOptionActionId(serializedIdx),
		uniqueDayBlockID(serializedIdx),
		uniqueRemoveButtonID(serializedIdx),
	)

	// If there's only 1 block at first, make it, actionBlock[0], removable, before append the new block
	if count := countDayBlock(blocks); count < 2 {
		for _, block := range blocks {
			if actionBlock, ok := isDayActionBlock(block); ok {
				arr := strings.Split(actionBlock.BlockID, "-")
				log.Println(arr)
				id, err := strconv.Atoi(arr[len(arr)-1])
				if err != nil {
					panic(err)
				}

				removeButton := &slack.ButtonBlockElement{
					Type:     slack.METButton,
					ActionID: uniqueRemoveButtonID(id),
					Value:    VALUE_REMOVE_PAID_LEAVE,
					Text:     slack.NewTextBlockObject(slack.PlainTextType, "Remove", false, false),
					Style:    slack.StyleDanger,
				}
				actionBlock.Elements.ElementSet = append(actionBlock.Elements.ElementSet, removeButton)
				break
			}
		}
	}

	blocks = AddMoreDayBlock(blocks, idxToInsert, dayActionBlock)
	mvr := botutils.UpdateModal(*view, blocks)

	return mvr
}

func getSerializedIndex(blocks []slack.Block) int {
	lastDayBlockIdx := 0
	for i := len(blocks) - 1; i >= 0; i-- {
		if _, ok := isDayActionBlock(blocks[i]); ok {
			lastDayBlockIdx = i
			break
		}
	}
	arr := strings.Split(blocks[lastDayBlockIdx].(*slack.ActionBlock).BlockID, "-")
	idx, err := strconv.Atoi(arr[len(arr)-1])
	if err != nil {
		panic(err)
	}

	return idx + 1
}

func buildPaidLeaveRequestModalByType(plType string, user *slack.User, count int) *slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject(
		slack.PlainTextType,
		fmt.Sprintf("%s Request", commands.PL_TYPES[plType]),
		false,
		false,
	)
	closeText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)
	submitText := slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)

	dividerBlock := slack.NewDividerBlock()

	countBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.PlainTextType,
			fmt.Sprintf("Hi %s, your remaining annual leaves are %d", user.Name, count),
			false,
			false,
		),
		nil,
		nil,
	)

	options := []*slack.OptionBlockObject{}
	for text, value := range plOptions {
		option := slack.NewOptionBlockObject(
			value,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			nil,
		)
		options = append(options, option)
	}
	optionPickerBlockElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, "Option", false, false),
		uniqueOptionActionId(0),
		options...,
	)
	optionPickerBlockElement.InitialOption = slack.NewOptionBlockObject(
		plOptions[FULL_DAY],
		slack.NewTextBlockObject(slack.PlainTextType, FULL_DAY, false, false),
		nil,
	)
	dayPickerBlockElement := slack.NewDatePickerBlockElement(uniqueDateActionID(0))
	dayBlock := slack.NewActionBlock(
		uniqueDayBlockID(0),
		dayPickerBlockElement,
		optionPickerBlockElement,
	)

	moreDayBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.PlainTextType,
			"Request an additional day",
			false,
			false,
		),
		nil,
		slack.NewAccessory(
			slack.NewButtonBlockElement(
				ACTION_ID_LEAVE_MORE_DAY,
				VALUE_LEAVE_MORE_DAY,
				slack.NewTextBlockObject(slack.PlainTextType, "Add a date", false, false),
			),
		),
		slack.SectionBlockOptionBlockID(ID_LEAVE_MORE_DAY),
	)

	noteBlock := slack.NewInputBlock(
		ID_LEAVE_NOTE,
		slack.NewTextBlockObject(slack.PlainTextType, "Note", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Additional notes", false, false),
		slack.PlainTextInputBlockElement{
			Type:        slack.METPlainTextInput,
			ActionID:    ACTION_ID_LEAVE_NOTE,
			Multiline:   true,
			Placeholder: slack.NewTextBlockObject(slack.PlainTextType, "Notes", false, false),
		},
	)
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			countBlock,
			dividerBlock,
			dayBlock,
			dividerBlock,
			moreDayBlock,
			dividerBlock,
			noteBlock,
			dividerBlock,
		},
	}

	mvr := &slack.ModalViewRequest{
		Type:          slack.VTModal,
		CallbackID:    getModalCallBackIDByType(plType),
		Title:         titleText,
		Close:         closeText,
		Submit:        submitText,
		Blocks:        blocks,
		ClearOnClose:  true,
		NotifyOnClose: true,
	}

	return mvr
}

func createDayActionBlockBySDK(datePickerActionID, optionSelectActionID, dayBlockID, removeButtonActionID string) *slack.ActionBlock {
	options := []*slack.OptionBlockObject{}
	for text, value := range plOptions {
		option := slack.NewOptionBlockObject(
			value,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			nil,
		)
		options = append(options, option)
	}
	optionPickerBlockElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, "Option", false, false),
		optionSelectActionID,
		options...,
	)
	optionPickerBlockElement.InitialOption = slack.NewOptionBlockObject(
		plOptions[FULL_DAY],
		slack.NewTextBlockObject(slack.PlainTextType, FULL_DAY, false, false),
		nil,
	)
	dayPickerBlockElement := slack.NewDatePickerBlockElement(datePickerActionID)
	removeButtonElement := &slack.ButtonBlockElement{
		Type:     slack.METButton,
		ActionID: ACTION_ID_REMOVE_PAID_LEAVE,
		Value:    VALUE_REMOVE_PAID_LEAVE,
		Text:     slack.NewTextBlockObject(slack.PlainTextType, "Remove", false, false),
		Style:    slack.StyleDanger,
	}
	dayBlock := slack.NewActionBlock(
		dayBlockID,
		dayPickerBlockElement,
		optionPickerBlockElement,
		removeButtonElement,
	)

	return dayBlock
}

func countDayBlock(blocks []slack.Block) int {
	count := 0
	for _, block := range blocks {
		if _, ok := isDayActionBlock(block); ok {
			count++
		}
	}

	return count
}

func indexToInsertDayBlock(blocks []slack.Block) int {
	for i := len(blocks) - 1; i >= 0; i-- {
		if _, ok := isDayActionBlock(blocks[i]); ok {
			return i + 1
		}
	}

	return -1
}

func AddMoreDayBlock(blocks []slack.Block, idx int, block slack.Block) []slack.Block {
	if len(blocks) == idx {
		return append(blocks, block)
	}

	blocks = append(blocks[:idx+1], blocks[idx:]...)
	blocks[idx] = block
	return blocks
}

func uniqueDateActionID(idx int) string {
	return fmt.Sprintf("%s-%d", ACTION_ID_LEAVE_DATE, idx)
}

func uniqueOptionActionId(idx int) string {
	return fmt.Sprintf("%s-%d", ACTION_ID_LEAVE_OPTION, idx)
}

func uniqueDayBlockID(idx int) string {
	return fmt.Sprintf("%s-%d", ID_LEAVE_DAY, idx)
}

func uniqueRemoveButtonID(idx int) string {
	return fmt.Sprintf("%s-%d", ACTION_ID_REMOVE_PAID_LEAVE, idx)
}

func getModalCallBackIDByType(plType string) string {
	switch plType {
	case commands.ANNUAL_PL:
		return CALLBACK_ID_ANNUAL_LEAVE_MODAL
	case commands.SICK_PL:
		return CALLBACK_ID_SICK_LEAVE_MODAL
	case commands.WEDDING_PL:
		return CALLBACK_ID_WEDDING_LEAVE_MODAL
	case commands.FUNERAL_PL:
		return CALLBACK_ID_FUNERAL_LEAVE_MODAL
	}

	return CALLBACK_ID_UNKNOWN
}

// For debugging
func viewDayBlockInfo(msg string, blocks []slack.Block) {
	dayBlocks := []slack.Block{}

	for _, block := range blocks {
		if _, ok := isDayActionBlock(block); ok {
			dayBlocks = append(dayBlocks, block)
		}
	}

	log.Printf("%s: %s\n", msg, utils.JSONString(dayBlocks))
}

func isDayActionBlock(block slack.Block) (*slack.ActionBlock, bool) {
	if block.BlockType() == slack.MBTAction && strings.Contains(block.(*slack.ActionBlock).BlockID, ID_LEAVE_DAY) {
		return block.(*slack.ActionBlock), true
	}

	return nil, false
}
