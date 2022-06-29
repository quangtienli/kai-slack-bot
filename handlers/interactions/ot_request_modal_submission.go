package interactions

import (
	"errors"
	"fmt"
	"log"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/handlers/commands"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func handleValidateOTRequestModalSubmission(msg slack.InteractionCallback, api *slack.Client, c *gin.Context) error {
	blockActions := msg.View.State.Values

	// Selected values
	// project := blockActions[commands.ID_PROJECT][commands.ACTION_ID_PROJECT].SelectedOption.Value
	startDate := blockActions[commands.ID_START_DATE][commands.ACTION_ID_START_DATE].SelectedDate
	startTime := blockActions[commands.ID_START_TIME][commands.ACTION_ID_START_TIME].SelectedTime
	// endTime := blockActions[commands.ID_END_TIME][commands.ACTION_ID_END_TIME].SelectedTime
	// note := blockActions[commands.ID_NOTE][commands.ACTION_ID_NOTE].Value

	startHour, startMinute := utils.GetHourAndMinute(startTime)
	// endHour, endMinute := utils.GetHourAndMinute(endTime)
	year, month, day := utils.GetYearAndMonthAndDay(startDate)

	start := time.Date(year, time.Month(month), int(day), int(startHour), int(startMinute), 0, 0, time.UTC)
	now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.UTC)

	blocks := msg.View.Blocks.BlockSet
	if start.Before(now) {
		for _, block := range blocks {
			if block.BlockType() == slack.MBTInput {
				inputBlock := block.(*slack.InputBlock)
				log.Println(inputBlock.BlockID)

				if inputBlock.BlockID == commands.ID_START_DATE {
					warningMsg := slack.NewTextBlockObject(
						slack.PlainTextType,
						"Invalid past day, plase select another day.",
						false,
						false,
					)
					inputBlock.Hint = warningMsg
					break
				}
			}
		}
		mvr := botutils.UpdateModal(msg.View, blocks)
		_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
		if err != nil {
			return err
		}
		return errors.New("Invalid day")
	} else if start.Equal(now) {
		if startHour < uint8(now.Hour()) {
			// Update view with warning message: invalid start hour
			for _, block := range blocks {
				if block.BlockType() == slack.MBTInput {
					inputBlock := block.(*slack.InputBlock)

					if inputBlock.BlockID == commands.ID_START_TIME {
						warningMsg := slack.NewTextBlockObject(
							slack.PlainTextType,
							"Invalid past hour, please select another time.",
							false,
							false,
						)
						inputBlock.Hint = warningMsg
						break
					}
				}
			}
			mvr := botutils.UpdateModal(msg.View, blocks)
			_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
			if err != nil {
				return err
			}
			return errors.New("Invalid hour")
		} else if startHour == uint8(now.Hour()) && startMinute < uint8(now.Minute()) {
			// Update view with warning message: invalid start minute
			for _, block := range blocks {
				if block.BlockType() == slack.MBTInput {
					inputBlock := block.(*slack.InputBlock)

					if inputBlock.BlockID == commands.ID_START_TIME {
						warningMsg := slack.NewTextBlockObject(
							slack.PlainTextType,
							"Invalid past minute, please select another time.",
							false,
							false,
						)
						inputBlock.Hint = warningMsg
						break
					}
				}
			}
			mvr := botutils.UpdateModal(msg.View, blocks)
			_, err := api.UpdateView(*mvr, msg.View.ExternalID, msg.View.Hash, msg.View.ID)
			if err != nil {
				return err
			}
			return errors.New("Invalid minute")
		}
	}

	return nil
}

func handleOTRequestModalSubmission(msg slack.InteractionCallback, api *slack.Client, c *gin.Context) error {
	blockActions := msg.View.State.Values

	// Selected values
	project := blockActions[commands.ID_PROJECT][commands.ACTION_ID_PROJECT].SelectedOption.Value
	startDate := blockActions[commands.ID_START_DATE][commands.ACTION_ID_START_DATE].SelectedDate
	startTime := blockActions[commands.ID_START_TIME][commands.ACTION_ID_START_TIME].SelectedTime
	endTime := blockActions[commands.ID_END_TIME][commands.ACTION_ID_END_TIME].SelectedTime
	note := blockActions[commands.ID_NOTE][commands.ACTION_ID_NOTE].Value

	startHour, startMinute := utils.GetHourAndMinute(startTime)
	endHour, endMinute := utils.GetHourAndMinute(endTime)
	year, month, day := utils.GetYearAndMonthAndDay(startDate)

	start := time.Date(year, time.Month(month), int(day), int(startHour), int(startMinute), 0, 0, time.UTC)
	now := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.UTC)

	user, _ := api.GetUserInfo(msg.User.ID)
	totalOT := sheetutils.GetOTTotalNumber()

	fromDate := start
	toDate := time.Date(year, time.Month(month), int(day), int(endHour), int(endMinute), 0, 0, time.UTC)
	duration := utils.RoundFloat(float64(endHour)+float64(endMinute)/60-float64(startHour)+float64(startMinute)/60, 2)
	ot := types.NewOT(
		totalOT,
		fmt.Sprintf("%s", user.Name),
		project,
		startDate,
		note,
		fromDate,
		toDate,
		now,
		duration,
	)
	// Insert the ot request into spreadsheet
	sheetutils.AddOTRequest(ot)
	// Build response message
	msgRespBlock, err := buildOTMessageResponse(ot, msg, api)
	if err != nil {
		return err
	}
	// Send response message
	err = botutils.SendOTMessageResponse(msg.User.ID, msgRespBlock, api)
	if err != nil {
		return err
	}

	return nil
}

func buildOTMessageResponse(ot *types.OT, msg slack.InteractionCallback, api *slack.Client) (*slack.SectionBlock, error) {
	user, err := api.GetUserInfo(msg.User.ID)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve user info: %v\n", err.Error())
	}
	block := slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf(
				"*%s* has recorded *%.2f OT hours* for project *%s* on *%s*",
				user.Name,
				ot.Duration,
				ot.Project,
				ot.Date,
			),
			false,
			false,
		),
		nil,
		nil,
	)

	return block, nil
}
