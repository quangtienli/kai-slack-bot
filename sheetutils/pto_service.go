package sheetutils

import (
	"os"
	"strconv"
	"test-go-slack-bot/types"

	"github.com/slack-go/slack"
	"google.golang.org/api/sheets/v4"
)

const (
	ptoReadRange      = "PTO!A2:K"
	totalPtoReadRange = "PTO!L2:L2"
)

func GetRemainingDaysByPLType(plType string, user *slack.User) int {
	var count int
	return count
}

func AddPTORequest(pto types.PaidLeaveRequest) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{
			pto.ToSpreadsheetObject(),
		},
	}
	_, err := service.Spreadsheets.Values.Append(
		spreadsheetID,
		ptoReadRange,
		valueRange,
	).ValueInputOption(VALUE_INPUT_OPTION_RAW).Do()
	if err != nil {
		panic(err)
	} else {
		totalPTOValueRange := &sheets.ValueRange{
			Values: [][]interface{}{
				{pto.ID + 1},
			},
		}
		_, err := service.Spreadsheets.Values.Update(spreadsheetID, totalPtoReadRange, totalPTOValueRange).ValueInputOption(VALUE_INPUT_OPTION_RAW).Do()
		if err != nil {
			panic(err)
		}
	}
}

func GetPTOTotalNumber() int {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, totalPtoReadRange).Do()
	if err != nil {
		panic(err)
	}
	N, err := strconv.Atoi(resp.Values[0][0].(string))
	if err != nil {
		panic(err)
	}
	return N
}
