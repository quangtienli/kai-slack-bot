package sheetutils

import (
	"fmt"
	"log"
	"os"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"google.golang.org/api/sheets/v4"
)

func ConfirmReplyByID(replyID string) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	// How to do an conditional update?
	replyIDsValueRange, err := service.Spreadsheets.Values.Get(spreadsheetID, replyReadRange).Do()
	if err != nil {
		panic(err)
	}

	for idx, row := range replyIDsValueRange.Values {
		// log.Println(row)
		if row[0].(string) == replyID {
			updateStatusReadRange := fmt.Sprintf("Replies!D%d:D%d", idx+2, idx+2)
			updateStatusValueRange := &sheets.ValueRange{
				Values: [][]interface{}{
					{types.ReplyCompleted},
				},
			}
			updateResp, err := service.Spreadsheets.Values.Update(
				spreadsheetID,
				updateStatusReadRange,
				updateStatusValueRange).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
			if err != nil {
				panic(err)
			} else {
				log.Printf("Update response: %s\n", utils.JSONString(updateResp))
			}
			break
		}
	}
}

func FindReplyByID(id string) *types.Reply {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, replyReadRange).Do()
	if err != nil {
		panic(err)
	}

	for _, row := range resp.Values {
		if row[types.ReplyIDColIdx].(string) == id {
			return types.ToReplyInstance(row)
		}
	}

	return nil
}

func InsertReply(reply *types.Reply) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{
			reply.ToSpreadsheetObject(),
		},
	}
	_, err := service.
		Spreadsheets.
		Values.
		Append(spreadsheetID, replyReadRange, valueRange).
		ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).
		Do()
	if err != nil {
		panic(err)
	}

	totalReply := GetTotalReply()
	totalReplyValueRange := &sheets.ValueRange{
		Values: [][]interface{}{
			{totalReply + 1},
		},
	}

	_, err = service.Spreadsheets.Values.Update(spreadsheetID, totalReplyReadRange, totalReplyValueRange).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
	if err != nil {
		panic(err)
	}

}

func GetTotalReply() int {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	totalReply, err := service.Spreadsheets.Values.Get(spreadsheetID, totalReplyReadRange).Do()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return utils.ToInt(totalReply.Values[0][0].(string))
}
