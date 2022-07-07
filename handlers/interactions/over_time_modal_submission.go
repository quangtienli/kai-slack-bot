package interactions

import (
	"fmt"
	"net/http"
	"os"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/handlers/commands"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// Add ot request into spreadsheet andn return response message
func handleOtModalSubmission(msg *slack.InteractionCallback, api *slack.Client, c *gin.Context) {
	// log.Printf("handleOtModalSubmission - message interaction callback: %s\n", utils.JSONString(msg))
	blockActions := msg.View.State.Values

	project := blockActions[commands.OtRequestProjectBlockID][commands.OtRequestProjectActionID].SelectedOption.Value
	startDay := blockActions[commands.OtRequestStartDayBlockID][commands.OtRequestStartDayActionID].SelectedDate
	startTime := blockActions[commands.OtRequestStartTimeBlockID][commands.OtRequestStartTimeActionID].SelectedTime
	endTime := blockActions[commands.OtRequestEndTimeBlockID][commands.OtRequestEndTimeActionID].SelectedTime
	note := blockActions[commands.OtRequestNoteBlockID][commands.OtRequestNoteActionID].Value

	startHour, startMinute := utils.GetHourAndMinute(startTime)
	endHour, endMinute := utils.GetHourAndMinute(endTime)
	year, month, day := utils.GetYearAndMonthAndDay(startDay)

	start := time.Date(year, time.Month(month), int(day), int(startHour), int(startMinute), 0, 0, time.UTC)
	end := time.Date(year, time.Month(month), int(day), int(endHour), int(endMinute), 0, 0, time.UTC)

	// Validating case: invalid end time
	if endHour < startHour || (endMinute < startMinute && endHour == startHour) {
		errors := map[string]string{
			commands.OtRequestEndTimeBlockID: "This request should be valid only within a day",
		}
		errResp := slack.NewErrorsViewSubmissionResponse(errors)
		c.JSON(http.StatusOK, errResp)
		return
	}

	go func() {
		channelID := os.Getenv("TEST_BOT_CHANNEL_ID")
		user, _ := api.GetUserInfo(msg.User.ID)
		totalOT := sheetutils.GetTotalOt()

		duration := utils.RoundFloat(float64(endHour)+float64(endMinute)/60-float64(startHour)-float64(startMinute)/60, 2)
		ot := types.NewOT(
			totalOT,
			fmt.Sprintf("%s", user.Name),
			project,
			startDay,
			note,
			start,
			end,
			duration,
		)
		sheetutils.InsertOt(ot)

		user, err := api.GetUserInfo(msg.User.ID)
		if err != nil {
			panic(err)
		}

		var responseMessage string
		if len(ot.Note) > 0 {
			responseMessage = fmt.Sprintf(
				"*%s* has recorded *%.2f OT hours* for project *%s* on *%s*\nNote: %s",
				user.Name,
				ot.Duration,
				ot.Project,
				ot.Date,
				ot.Note,
			)
		} else {
			responseMessage = fmt.Sprintf(
				"*%s* has recorded *%.2f OT hours* for project *%s* on *%s*",
				user.Name,
				ot.Duration,
				ot.Project,
				ot.Date,
			)
		}

		respMsgBlock := botutils.BuildResponseMessageBlockWithContext(responseMessage)
		botutils.SendResponseMessage(channelID, api, respMsgBlock)
	}()
}

// Ways to validate requests

// 1. Invalid date [Not necessary]
// if start.Before(now) {
// 	errors := map[string]string{
// 		commands.OtRequestStartDayBlockID:  "You may not request a due date in the past",
// 		commands.OtRequestStartTimeBlockID: "You may not request a due date in the past",
// 	}
// 	resp := slack.NewErrorsViewSubmissionResponse(errors)
// 	c.JSON(http.StatusOK, resp)

// 	return
// }

// 2. Invalid end time

// 3. Overlapping previous ot history requests [Not necessary]
// ots, err := sheetutils.GetOTHistoryByUsername(msg.User.Name)
// if err != nil {
// 	panic(err)
// }
// for _, ot := range ots {
// 	if (start.After(ot.StartAt) && start.Before(ot.EndAt)) ||
// 		(end.After(ot.StartAt) && end.Before(ot.EndAt)) {
// 		log.Println("Overlapping")
// 		errors := map[string]string{
// 			commands.OtRequestStartDayBlockID:  "Request overlaps ot history",
// 			commands.OtRequestStartTimeBlockID: "Request overlaps ot history",
// 			commands.OtRequestEndTimeBlockID:   "Request overlaps ot history",
// 		}
// 		resp := slack.NewErrorsViewSubmissionResponse(errors)
// 		c.JSON(http.StatusOK, resp)

// 		return
// 	}
// }
