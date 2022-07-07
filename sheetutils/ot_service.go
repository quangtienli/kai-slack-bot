package sheetutils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"test-go-slack-bot/types"

	"google.golang.org/api/sheets/v4"
)

const (
	otReadRange      = "OT!A2:I"
	totalOtReadRange = "OT!J2:J2"
	projectReadRange = "Projects!A2:A"
)

func FindOtsByUsername(username string) ([]types.OT, error) {
	var ots = []types.OT{}
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, otReadRange).Do()
	if err != nil {
		return ots, fmt.Errorf("Unable to retrieve overtime history by username: %s\n", err.Error())
	}
	for _, row := range resp.Values {
		if row[types.OtUsernameColIdx].(string) == username {
			ot := types.ToOtInstance(row)
			ots = append(ots, *ot)
		}
	}

	return ots, nil
}

func InsertOt(ot *types.OT) {
	service := initSheetService()
	spreadSheetID := os.Getenv("SPREADSHEET_ID")
	valuerange := &sheets.ValueRange{
		Values: [][]interface{}{
			ot.ToSpreadsheetObject(),
		},
	}
	_, err := service.Spreadsheets.Values.Append(spreadSheetID, otReadRange, valuerange).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
	if err != nil {
		log.Printf("Error while adding a new ot request: %s\n", err.Error())
	} else {
		// ID for the latest request equals to the total number of requests + 1
		valuerange := &sheets.ValueRange{
			Values: [][]interface{}{
				{ot.ID + 1},
			},
		}
		_, err := service.Spreadsheets.Values.Update(spreadSheetID, totalOtReadRange, valuerange).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
		if err != nil {
			log.Printf("Error while incrementing total number of ots: %s\n", err.Error())
		} else {
			log.Printf("Current number of ots: %d\n", ot.ID+1)
		}
	}
}

func GetTotalOt() int {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, totalOtReadRange).Do()
	if err != nil {
		log.Printf("Error while fetching total number of OT requests: %v\n", err.Error())
	}
	N, _ := strconv.Atoi(resp.Values[0][0].(string))
	return N
}
