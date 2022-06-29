package interactions

import (
	"log"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"github.com/slack-go/slack"
)

var (
	plTypes = map[string]types.PaidLeaveType{
		CALLBACK_ID_ANNUAL_LEAVE_MODAL:  types.ANNUAL,
		CALLBACK_ID_SICK_LEAVE_MODAL:    types.SICK,
		CALLBACK_ID_WEDDING_LEAVE_MODAL: types.WEDDING,
		CALLBACK_ID_FUNERAL_LEAVE_MODAL: types.FUNERAL,
	}
)

func handlePaidLeaveRequestSubmissionByType(msg *slack.InteractionCallback, api *slack.Client) error {
	viewCallbackID := msg.View.CallbackID
	plType := plTypes[viewCallbackID]

	log.Println("Annual leave request")
	values := msg.View.State.Values
	blocks := msg.View.Blocks.BlockSet

	note := values[ID_LEAVE_NOTE][ACTION_ID_LEAVE_NOTE].Value

	dayBlocks := extractDayActionBlocks(blocks)
	totalPTO := sheetutils.GetPTOTotalNumber()
	id := totalPTO

	user, err := api.GetUserInfo(msg.User.ID)
	if err != nil {
		return err
	}

	reqs := []types.PaidLeaveRequest{}
	for _, block := range dayBlocks {
		d := values[block.BlockID][block.Elements.ElementSet[0].(*slack.DatePickerBlockElement).ActionID].SelectedDate
		o := values[block.BlockID][block.Elements.ElementSet[1].(*slack.SelectBlockElement).ActionID].SelectedOption.Value
		req := types.NewPaidLeaveRequest(
			id,
			user.Name,
			utils.DatifyStringDate(d),
			types.PaidLeaveType(plType),
			types.PaidLeaveOption(utils.ToFloat(o)),
			note,
		)
		reqs = append(reqs, *req)

		id++
	}
	for i := len(reqs) - 1; i >= 0; i-- {
		defer sheetutils.AddPTORequest(reqs[i])
	}

	return nil
}

func extractDayActionBlocks(blocks []slack.Block) []slack.ActionBlock {
	arr := []slack.ActionBlock{}

	for _, block := range blocks {
		if actionBlock, ok := isDayActionBlock(block); ok {
			arr = append(arr, *actionBlock)
		}
	}

	return arr
}
