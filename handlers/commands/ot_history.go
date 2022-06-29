package commands

import (
	"fmt"
	"sort"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/slack-go/slack"
)

func handleOTHistoryCommand(command slack.SlashCommand, api *slack.Client) error {
	groups := make(map[string][]types.OT)
	ots, err := sheetutils.GetOTHistoryByUsername(command.UserName)
	if err != nil {
		return err
	}
	if len(ots) == 0 {
		block := buildEmptyOTHistoryResponse()
		err := botutils.SendOTHistoryResponse(command.UserID, api, block)
		if err != nil {
			return err
		}
	} else {
		// Group ots by project
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
		// Build GUI response
		blocks := buildOTHistoryResponse(groups)
		// Send response
		err = botutils.SendOTHistoryResponse(command.UserID, api, blocks...)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildEmptyOTHistoryResponse() slack.Block {
	block := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, "No overtime requests recorded.", false, false),
		nil,
		nil,
	)
	return block
}

func buildOTHistoryResponse(groups map[string][]types.OT) []slack.Block {
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
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(":dart: *%s*", key), false, false),
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
