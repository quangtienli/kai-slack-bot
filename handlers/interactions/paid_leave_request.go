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
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	// Can be used for all types of pto request
	plRequestByTypeDayBlockID  = "pl-request-by-type-day-block-id"
	plRequestByTypeDayActionID = "pl-request-by-type-day-action-id"
	plRequestByTypeOptionID    = "pl-request-by-type-option-id"

	plRequestByTypeRemoveDayActionID = "pl-request-by-type-remove-day-action-id"
	plRequestByTypeRemoveDayValue    = "pl-request-by-type-remove-day-value"

	plRequestByTypeMoreDayBlockID  = "pl-request-by-type-more-day-block-id"
	plRequestByTypeMoreDayActionID = "pl-request-by-type-more-day-action-id"
	plRequestByTypeMoreDayValue    = "pl-request-by-type-more-day-value"

	plRequestByTypeNoteBlockID  = "pl-request-by-type-note-block-id"
	plRequestByTypeNoteActionID = "pl-request-by-type-note-action-id"

	// Unique callback id to identify a pto modal
	AnnualLeaveModalCallbackID  = "annual-leave-modal-callback-id"
	SickLeaveModalCallbackID    = "sick-leave-modal-callback-id"
	WeddingLeaveModalCallbackID = "wedding-leave-modal-callback-id"
	FuneralLeaveCallbackID      = "funeral-le ave-modal-callback-id"
)

// Open modal view request for paid leave request by a pre-selected type
func handlePaidLeaveRequest(msg *slack.InteractionCallback, api *slack.Client, c *gin.Context) {
	plTypeText := msg.View.State.Values[commands.PlRequestTypeID][commands.PlRequestTypeActionID].SelectedOption.Text.Text
	user, err := api.GetUserInfo(msg.User.ID)
	if err != nil {
		panic(err)
	}

	// Get allowable remaining days according to the type of the day
	count := sheetutils.GetRemainingDaysByPLType(plTypeText, user)

	// Build modal view request and push it
	mvr := buildPaidLeaveRequestModalByType(plTypeText, user, count)
	resp := slack.NewPushViewSubmissionResponse(mvr)
	c.JSON(http.StatusOK, resp)
}

// Build the paid leave modal by a type that members has chosen
func buildPaidLeaveRequestModalByType(plTypeText string, user *slack.User, count int) *slack.ModalViewRequest {
	titleText := slack.NewTextBlockObject(
		slack.PlainTextType,
		fmt.Sprintf("%s Request", plTypeText),
		false,
		false,
	)
	closeText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)
	submitText := slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)

	dividerBlock := slack.NewDividerBlock()

	countBlock := new(slack.SectionBlock)
	if plTypeText == types.AnnualLeave {
		countBlock = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf("Hi <@%s>, your remaining annual leaves are %d days", user.ID, count),
				false,
				false,
			),
			nil,
			nil,
		)
	} else if plTypeText == types.SickLeave {
		countBlock = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf("Hi <@%s>, you can request up to *%d days* for sick leaves.\n_Requested leaves on weekend will also be paid._", user.ID, 2),
				false,
				false,
			),
			nil,
			nil,
		)
	} else if plTypeText == types.WeddingLeave {
		countBlock = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf("Hi <@%s>, you can request up to *%d days* for wedding leaves.\n_Requested leaves on weekend will also be paid._", user.ID, 3),
				false,
				false,
			),
			nil,
			nil,
		)
	} else if plTypeText == types.FuneralLeave {
		countBlock = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf("Hi <@%s>, you can request up to *%d days* for funeral leaves.\n_Requested leaves on weekend will also be paid._", user.ID, 2),
				false,
				false,
			),
			nil,
			nil,
		)
	}

	options := []*slack.OptionBlockObject{}
	for _, option := range types.PaidLeaveOptions {
		o := slack.NewOptionBlockObject(
			fmt.Sprintf("%.1f", option.Value),
			slack.NewTextBlockObject(slack.PlainTextType, option.Text, false, false),
			nil,
		)
		options = append(options, o)
	}
	optionPickerBlockElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, "Option", false, false),
		uniqueOptionActionId(0),
		options...,
	)
	optionPickerBlockElement.InitialOption = slack.NewOptionBlockObject(
		fmt.Sprintf("%.1f", types.FullDay.Value),
		slack.NewTextBlockObject(slack.PlainTextType, types.FullDay.Text, false, false),
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
				plRequestByTypeMoreDayActionID,
				plRequestByTypeMoreDayValue,
				slack.NewTextBlockObject(slack.PlainTextType, "Add a date", false, false),
			),
		),
		slack.SectionBlockOptionBlockID(plRequestByTypeMoreDayBlockID),
	)

	noteBlock := slack.NewInputBlock(
		plRequestByTypeNoteBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Note", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Additional notes", false, false),
		slack.PlainTextInputBlockElement{
			Type:        slack.METPlainTextInput,
			ActionID:    plRequestByTypeNoteActionID,
			Multiline:   true,
			Placeholder: slack.NewTextBlockObject(slack.PlainTextType, "Notes", false, false),
		},
	)
	noteBlock.Optional = false

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
		CallbackID:    getModalCallBackIDByType(plTypeText),
		Title:         titleText,
		Close:         closeText,
		Submit:        submitText,
		Blocks:        blocks,
		ClearOnClose:  true,
		NotifyOnClose: true,
	}

	return mvr
}

// Update the current modal view request with an additional day block
func handleMoreDayPaidLeaveRequest(msg *slack.InteractionCallback, api *slack.Client) {
	mvr := buildMoreDayPaidLeaveRequestModalByType(msg)
	_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
	if err != nil {
		panic(err)
	}
}

// Build the paid leave modal with an additional day block, by a type that member choose
func buildMoreDayPaidLeaveRequestModalByType(msg *slack.InteractionCallback) *slack.ModalViewRequest {
	blocks := msg.View.Blocks.BlockSet

	// Create a new block with the serialized id
	dayActionBlock := createDayActionBlockBySDK(getSerializedIndex(blocks))

	// If there's only 1 block at first, make it, actionBlock[0], removable, before append the new block
	if count := countDayBlock(blocks); count < 2 {
		for _, block := range blocks {
			if actionBlock, ok := isDayActionBlock(block); ok {
				arr := strings.Split(actionBlock.BlockID, "-")
				id, err := strconv.Atoi(arr[len(arr)-1])
				if err != nil {
					panic(err)
				}

				removeButton := &slack.ButtonBlockElement{
					Type:     slack.METButton,
					ActionID: uniqueRemoveButtonID(id),
					Value:    plRequestByTypeRemoveDayValue,
					Text:     slack.NewTextBlockObject(slack.PlainTextType, "Remove", false, false),
					Style:    slack.StyleDanger,
				}
				actionBlock.Elements.ElementSet = append(actionBlock.Elements.ElementSet, removeButton)
				break
			}
		}
	}

	blocks = insertDayBlock(blocks, getIndexToInsertDayBlock(blocks), dayActionBlock)
	mvr := botutils.UpdateModal(msg.View, blocks)

	return mvr
}

// Update the current modal view request by removing the clicked day block
func handleLessDayPaidLeaveRequest(msg *slack.InteractionCallback, api *slack.Client) {
	mvr := buildLessDayPaidLeaveRequestModalByType(msg)
	_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
	if err != nil {
		panic(err)
	}
}

// A function to remove the clicked day block
func buildLessDayPaidLeaveRequestModalByType(msg *slack.InteractionCallback) *slack.ModalViewRequest {
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

	// Since "blocks" only contain the UI implementation -> move selected value to every block
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

	updatedView := botutils.UpdateModal(msg.View, blocks)
	return updatedView
}

// Create a unique day block which can be identified from others
func createDayActionBlockBySDK(idx int) *slack.ActionBlock {
	options := []*slack.OptionBlockObject{}
	for _, option := range types.PaidLeaveOptions {
		o := slack.NewOptionBlockObject(
			fmt.Sprintf("%.1f", option.Value),
			slack.NewTextBlockObject(slack.PlainTextType, option.Text, false, false),
			nil,
		)
		options = append(options, o)
	}
	optionPickerBlockElement := slack.NewOptionsSelectBlockElement(
		slack.OptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, "Option", false, false),
		uniqueOptionActionId(idx),
		options...,
	)
	optionPickerBlockElement.InitialOption = slack.NewOptionBlockObject(
		fmt.Sprintf("%.1f", types.FullDay.Value),
		slack.NewTextBlockObject(slack.PlainTextType, types.FullDay.Text, false, false),
		nil,
	)

	dayPickerBlockElement := slack.NewDatePickerBlockElement(uniqueDateActionID(idx))

	removeButtonElement := &slack.ButtonBlockElement{
		Type:     slack.METButton,
		ActionID: uniqueRemoveButtonID(idx),
		Value:    plRequestByTypeRemoveDayValue,
		Text:     slack.NewTextBlockObject(slack.PlainTextType, "Remove", false, false),
		Style:    slack.StyleDanger,
	}
	dayBlock := slack.NewActionBlock(
		uniqueDayBlockID(idx),
		dayPickerBlockElement,
		optionPickerBlockElement,
		removeButtonElement,
	)

	return dayBlock
}

// Count current day blocks in the modal view request
func countDayBlock(blocks []slack.Block) int {
	count := 0
	for _, block := range blocks {
		if _, ok := isDayActionBlock(block); ok {
			count++
		}
	}

	return count
}

// Get serial index for day blocks
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

// Find the correct index to insert day block
func getIndexToInsertDayBlock(blocks []slack.Block) int {
	for i := len(blocks) - 1; i >= 0; i-- {
		if _, ok := isDayActionBlock(blocks[i]); ok {
			return i + 1
		}
	}

	return -1
}

// Insert the day block by an index
func insertDayBlock(blocks []slack.Block, idx int, block slack.Block) []slack.Block {
	if len(blocks) == idx {
		return append(blocks, block)
	}

	blocks = append(blocks[:idx+1], blocks[idx:]...)
	blocks[idx] = block
	return blocks
}

// Check if the current block is the day block
func isDayActionBlock(block slack.Block) (*slack.ActionBlock, bool) {
	if block.BlockType() == slack.MBTAction && strings.Contains(block.(*slack.ActionBlock).BlockID, plRequestByTypeDayBlockID) {
		return block.(*slack.ActionBlock), true
	}

	return nil, false
}

// Unique id generators
func uniqueDateActionID(idx int) string {
	return fmt.Sprintf("%s-%d", plRequestByTypeDayActionID, idx)
}

func uniqueOptionActionId(idx int) string {
	return fmt.Sprintf("%s-%d", plRequestByTypeOptionID, idx)
}

func uniqueDayBlockID(idx int) string {
	return fmt.Sprintf("%s-%d", plRequestByTypeDayBlockID, idx)
}

func uniqueRemoveButtonID(idx int) string {
	return fmt.Sprintf("%s-%d", plRequestByTypeRemoveDayActionID, idx)
}

func getModalCallBackIDByType(plTypeText string) string {
	if plTypeText == types.AnnualLeave {
		return AnnualLeaveModalCallbackID
	}

	if plTypeText == types.SickLeave {
		return SickLeaveModalCallbackID
	}

	if plTypeText == types.WeddingLeave {
		return WeddingLeaveModalCallbackID
	}

	return FuneralLeaveCallbackID
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
