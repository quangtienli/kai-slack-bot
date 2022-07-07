package commands

import (
	"fmt"
	"sort"
	"time"

	"test-go-slack-bot/botutils"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func handleOverTimeHistoryCommand(command *slack.SlashCommand, api *slack.Client, c *gin.Context) {
	loadingMsg := fmt.Sprintf("Wait for me a second :blobdance: I'm gathering them for you.")
	loadingMsgBlock := botutils.BuildResponseMessageBlockWithContext(loadingMsg)
	botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, loadingMsgBlock)

	go func() {
		ots, err := sheetutils.FindOtsByUsername(command.UserName)
		if err != nil {
			errorMessage := fmt.Sprintf("Unable to retrieve ot history: %s.\n", err.Error())
			errorMessageBlock := botutils.BuildResponseMessageBlockWithContext(errorMessage)
			botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, errorMessageBlock)
			return
		}
		// No ot request is recorded
		if len(ots) == 0 {
			emptyMessage := "You have not made any overtime request yet."
			emptyMessageBlock := botutils.BuildResponseMessageBlockWithContext(emptyMessage)
			botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, emptyMessageBlock)
			return
		}
		// Send ephemeral response message
		otHistoryMessageBlocks := buildOverTimeHistoryResponseMessageBlocks(ots)
		botutils.SendEphemeralResponseMessage(command.ChannelID, command.UserID, api, otHistoryMessageBlocks...)
	}()
}

func buildOverTimeHistoryResponseMessageBlocks(ots []types.OT) []slack.Block {
	// Group overtime history by project
	groups := make(map[string][]types.OT)
	for _, ot := range ots {
		if _, ok := groups[ot.Project]; !ok {
			groups[ot.Project] = []types.OT{ot}
		} else {
			groups[ot.Project] = append(groups[ot.Project], ot)
		}
	}
	// Sort each group by date in asc order
	for _, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			return utils.CompareStringDates(group[i].Date, group[j].Date) == -1
		})
	}

	blocks := []slack.Block{}

	headerBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, "*Your OT History:*", false, false),
		nil,
		nil,
	)
	dividerBlock := slack.NewDividerBlock()
	blocks = append(blocks, headerBlock, dividerBlock)

	for key, group := range groups {
		groupHeaderBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*%s*", key), false, false),
			nil,
			nil,
		)
		blocks = append(blocks, groupHeaderBlock)

		for _, ot := range group {
			otBlock := slack.NewSectionBlock(
				slack.NewTextBlockObject(
					slack.MarkdownType,
					fmt.Sprintf(
						":calendar: %10s | :clock3: %-7s - %7s | :hourglass_flowing_sand: %.2f hours",
						ot.Date,
						ot.StartAt.Format(time.Kitchen),
						ot.EndAt.Format(time.Kitchen),
						ot.Duration,
					),
					false,
					false,
				),
				nil,
				nil,
			)
			blocks = append(blocks, otBlock)
		}
	}

	return blocks
}
