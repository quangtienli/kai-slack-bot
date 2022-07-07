package interactions

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	ReplyApproveButtonActionId = "reply-approve-button-action-id"
	ReplyDenyButtonActionID    = "deny-approve-button-action-id"

	ReplyHeaderBlockID = "reply-header-block-id"
	ReplyInfoBlockID   = "reply-info-block-id"
	ReplyActionBlockID = "reply-action-block-id"
)

func handlePaidLeaveRequestSubmissionByType(msg *slack.InteractionCallback, api *slack.Client, c *gin.Context) {
	// log.Printf("handlePaidLeaveRequestSubmissionByType - interaction callback: %s\n", utils.JSONString(msg))

	go func() {
		appChannelID := os.Getenv("APP_CHANNEL_ID")

		loadingMessage := buildLoadingMessage()
		_, _, err := api.PostMessage(appChannelID, slack.MsgOptionBlocks(loadingMessage))
		if err != nil {
			panic(err)
		}

		reqs := extractPaidLeaveRequests(msg)
		if err := validateRequests(reqs); err != nil {
			// Send error message
			failResponse := buildFailResponseMessage(err)
			_, _, err := api.PostMessage(appChannelID, slack.MsgOptionBlocks(failResponse...))
			if err != nil {
				panic(err)
			}
			return
		}
		log.Printf("Extracted paid leave requests: %s\n", utils.JSONString(reqs))

		_, userSumPto := sheetutils.FindSumPtoByUsername(msg.User.Name)

		user, err := api.GetUserInfo(msg.User.ID)
		if err != nil {
			failResponse := buildFailResponseMessage(err)
			_, _, err := api.PostMessage(appChannelID, slack.MsgOptionBlocks(failResponse...))
			if err != nil {
				panic(err)
			}
			return
		}

		// Add a reply to database
		totalReply := sheetutils.GetTotalReply()
		reply := types.NewReply(totalReply, user.Name, userSumPto.ManagerUsername, types.ReplyPending)
		sheetutils.InsertReply(reply)

		// Add pl requests to database
		for i := 0; i < len(reqs); i++ {
			reqs[i].ReplyID = reply.ID
			sheetutils.InsertPto(reqs[i])
		}

		// Send reply message to manager
		_, managerSumPto := sheetutils.FindSumPtoByUsername(userSumPto.ManagerUsername)
		manager, err := api.GetUserByEmail(managerSumPto.UserEmail)
		if err != nil {
			failResponse := buildFailResponseMessage(err)
			_, _, err := api.PostMessage(appChannelID, slack.MsgOptionBlocks(failResponse...))
			if err != nil {
				panic(err)
			}
			return
		}
		// log.Println(manager)

		// Build reply message and send it to the manager
		replyMesage := buildReplyMessageBlocks(reqs, reply)
		_, _, err = api.PostMessage(manager.ID, slack.MsgOptionBlocks(replyMesage...))
		if err != nil {
			failResponse := buildFailResponseMessage(err)
			_, _, err := api.PostMessage(appChannelID, slack.MsgOptionBlocks(failResponse...))
			if err != nil {
				panic(err)
			}
			return
		}

		successResponse := buildSuccessResponseMessage(reqs)
		_, _, err = api.PostMessage(appChannelID, slack.MsgOptionBlocks(successResponse...))
		if err != nil {
			panic(err)
		}
	}()

	c.JSON(http.StatusOK, slack.NewClearViewSubmissionResponse())
}

func buildReplyMessageBlocks(ptos []types.Pto, reply *types.Reply) []slack.Block {
	if len(ptos) == 0 {
		return nil
	}

	username := ptos[0].Username
	plType := ptos[0].Type
	note := ptos[0].Note

	var duration float64
	plInfo := ""
	for _, pto := range ptos {
		plInfo += fmt.Sprintf("• %s (%s)\n", pto.Date.Format(utils.LAYOUT_YYYYMMDD), strings.ToLower(pto.Option.Text))
		duration += pto.Option.Value
	}

	headerBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf("You have a new request to be confirmed, %s!\n*%s - Paid Leave Request*", reply.ManagerUsername, username),
			false,
			false,
		),
		nil,
		nil,
		slack.SectionBlockOptionBlockID(ReplyHeaderBlockID),
	)

	infoAccessory := slack.NewImageBlockElement(
		"https://img.freepik.com/free-vector/calendar-deadline-with-clock-flat-design_115464-601.jpg?w=2000",
		"Calendar",
	)
	infoBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf(
				"*Type*: `%s`\n*Total duration*: `%.1f days`\n*When*:\n%s*Current balance*: `%.1f days`\n*Note*: \"%s\"",
				plType,
				duration,
				plInfo,
				float64(12),
				note,
			),
			false,
			false,
		),
		nil,
		slack.NewAccessory(infoAccessory),
		slack.SectionBlockOptionBlockID(ReplyInfoBlockID),
	)
	approveButtonBlockElement := slack.NewButtonBlockElement(
		ReplyApproveButtonActionId,
		strconv.Itoa(reply.ID),
		slack.NewTextBlockObject(slack.PlainTextType, "Approve", false, false),
	)
	approveButtonBlockElement.Style = slack.StylePrimary

	denyButtonBlockElement := slack.NewButtonBlockElement(
		ReplyDenyButtonActionID,
		strconv.Itoa(reply.ID),
		slack.NewTextBlockObject(slack.PlainTextType, "Deny", false, false),
	)
	denyButtonBlockElement.Style = slack.StyleDanger

	actionBlock := slack.NewActionBlock(
		ReplyActionBlockID,
		approveButtonBlockElement,
		denyButtonBlockElement,
	)

	blocks := []slack.Block{
		headerBlock,
		slack.NewDividerBlock(),
		infoBlock,
		slack.NewDividerBlock(),
		actionBlock,
	}

	return blocks
}

func validateRequests(reqs []types.Pto) error {
	visited := make(map[string]bool, 0)
	for _, req := range reqs {
		if req.Date.Before(time.Now()) {
			return errors.New("Invalid days")
		}

		if _, ok := visited[req.Date.String()]; ok {
			return errors.New("Duplicated days")
		} else {
			visited[req.Date.String()] = true
		}

	}

	return nil
}

func extractPaidLeaveRequests(msg *slack.InteractionCallback) []types.Pto {
	viewCallbackID := msg.View.CallbackID
	plType := getPlType(viewCallbackID)

	values := msg.View.State.Values
	blocks := msg.View.Blocks.BlockSet

	note := values[plRequestByTypeNoteBlockID][plRequestByTypeNoteActionID].Value

	dayBlocks := extractDayActionBlocks(blocks)
	log.Printf("Number of day blocks in the pto request: %d\n", len(dayBlocks))

	totalPTO := sheetutils.GetTotalPto()
	id := totalPTO

	reqs := []types.Pto{}
	for _, block := range dayBlocks {
		d := values[block.BlockID][block.Elements.ElementSet[0].(*slack.DatePickerBlockElement).ActionID].SelectedDate
		o := values[block.BlockID][block.Elements.ElementSet[1].(*slack.SelectBlockElement).ActionID].SelectedOption.Value
		req := types.NewPto(
			id,
			msg.User.Name,
			utils.DatifyStringDate(d),
			plType,
			types.GetPtoOption(o),
			note,
		)
		reqs = append(reqs, *req)

		id++
	}

	return reqs
}

func extractDayActionBlocks(blocks []slack.Block) []slack.ActionBlock {
	arr := []slack.ActionBlock{}

	for _, block := range blocks {
		if actionBlock, ok := isDayActionBlock(block); ok {
			log.Println(block.(*slack.ActionBlock).BlockID)
			arr = append(arr, *actionBlock)
		}
	}

	return arr
}

func buildLoadingMessage() *slack.SectionBlock {
	loadingMessage := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"Your leave request is being processed",
			false,
			false,
		),
		nil,
		nil,
	)

	return loadingMessage
}

func buildFailResponseMessage(err error) []slack.Block {
	failResponse := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, err.Error(), false, false),
		nil,
		nil,
	)

	return []slack.Block{
		failResponse,
	}
}

func buildSuccessResponseMessage(ptos []types.Pto) []slack.Block {
	var duration float64 = 0
	for _, pto := range ptos {
		duration += pto.Option.Value
	}

	plType := ptos[0].Type
	var text string
	if duration > 1 {
		text += fmt.Sprintf("Your request for *%.1f days* of `%s` has been sent to your manager.", duration, plType)
	} else {
		text += fmt.Sprintf("Your request for *%.1f day* of `%s` has been sent to your manager.", duration, plType)
	}
	headerBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			text,
			false,
			false,
		),
		nil,
		nil,
	)
	dividerBlock := slack.NewDividerBlock()

	ctx := ""
	for _, pto := range ptos {
		ctx += fmt.Sprintf("• *%s* (%s)\n", pto.Date.Format(utils.LAYOUT_YYYYMMDD), strings.ToLower(pto.Option.Text))
	}

	contextBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			ctx,
			false,
			false,
		),
		nil,
		nil,
	)

	return []slack.Block{
		headerBlock,
		dividerBlock,
		contextBlock,
	}
}

func getPlType(callbackID string) string {
	if callbackID == AnnualLeaveModalCallbackID {
		return types.AnnualLeave
	}

	if callbackID == SickLeaveModalCallbackID {
		return types.SickLeave
	}

	if callbackID == WeddingLeaveModalCallbackID {
		return types.WeddingLeave
	}

	return types.FuneralLeave
}
