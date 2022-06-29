package sheetutils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"google.golang.org/api/sheets/v4"
)

const (
	idColIdx        = 0
	usernameColIdx  = 1
	projectColIdx   = 2
	dateColIdx      = 3
	startTimeColIdx = 4
	endTimeColIdx   = 5
	durationColIdx  = 6
	createdColIdx   = 7
	noteColIdx      = 8
)

const (
	otReadRange      = "OT!A2:I"
	totalOtReadRange = "OT!J2:J2"
	projectReadRange = "Projects!A1:A"
)

// func GetOTRequestsByUsername(username string) []types.OT {
// 	service := initSheetService()
// 	spreadsheetID := os.Getenv("SPREADSHEET_ID")
// 	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, otReadRange).Do()
// 	if err != nil {
// 		log.Printf("Unable to retrieve ot requests by username: %s\n", err.Error())
// 	}
// 	var ots []types.OT
// 	for _, row := range resp.Values {
// 		if row[usernameColIdx].(string) == username {
// 			ot := types.NewOT(
// 				utils.ToInt(row[idColIdx].(string)),
// 				row[usernameColIdx].(string),
// 				row[projectColIdx].(string),
// 				row[dateColIdx].(string),
// 				row[noteColIdx].(string),
// 				utils.ToDate(row[startTimeColIdx].(string)),
// 				utils.ToDate(row[endTimeColIdx].(string)),
// 				utils.ToDate(row[createdColIdx].(string)),
// 				utils.ToFloat(row[durationColIdx].(string)),
// 			)
// 			ots = append(ots, ot)
// 		}
// 	}

// 	return ots
// }

func GetOTHistoryByUsername(username string) ([]types.OT, error) {
	var ots = []types.OT{}
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, otReadRange).Do()
	if err != nil {
		return ots, fmt.Errorf("Unable to retrieve overtime history by username: %s\n", err.Error())
	}
	for _, row := range resp.Values {
		if row[usernameColIdx].(string) == username {
			ot := types.NewOT(
				utils.ToInt(row[idColIdx].(string)),
				row[usernameColIdx].(string),
				row[projectColIdx].(string),
				row[dateColIdx].(string),
				row[noteColIdx].(string),
				utils.ToDate(row[startTimeColIdx].(string)),
				utils.ToDate(row[endTimeColIdx].(string)),
				utils.ToDate(row[createdColIdx].(string)),
				utils.ToFloat(row[durationColIdx].(string)),
			)
			ots = append(ots, *ot)
		}
	}

	return ots, nil
}

func GetOTTotalNumber() int {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, totalOtReadRange).Do()
	if err != nil {
		log.Printf("Error while fetching total number of OT requests: %v\n", err.Error())
	}
	N, _ := strconv.Atoi(resp.Values[0][0].(string))
	return N
}

func AddOTRequest(ot *types.OT) {
	service := initSheetService()
	spreadSheetID := os.Getenv("SPREADSHEET_ID")
	valuerange := &sheets.ValueRange{
		Values: [][]interface{}{
			ot.ToSpreadsheetObject(),
		},
	}
	_, err := service.Spreadsheets.Values.Append(spreadSheetID, otReadRange, valuerange).ValueInputOption(VALUE_INPUT_OPTION_RAW).Do()
	if err != nil {
		log.Printf("Error while adding a new ot request: %s\n", err.Error())
	} else {
		// ID for the latest request equals to the total number of requests + 1
		valuerange := &sheets.ValueRange{
			Values: [][]interface{}{
				{ot.ID + 1},
			},
		}
		_, err := service.Spreadsheets.Values.Update(spreadSheetID, totalOtReadRange, valuerange).ValueInputOption(VALUE_INPUT_OPTION_RAW).Do()
		if err != nil {
			log.Printf("Error while incrementing total number of ots: %s\n", err.Error())
		} else {
			log.Printf("Current number of ots: %d\n", ot.ID+1)
		}
	}
}

func FetchProjects() ([]types.Project, error) {
	service := initSheetService()
	spreadSheetID := os.Getenv("SPREADSHEET_ID")
	resp, err := service.Spreadsheets.Values.Get(spreadSheetID, projectReadRange).Do()
	if err != nil {
		return nil, err
	}
	projects := []types.Project{}
	if len(resp.Values) == 0 {
		return projects, nil
	}
	for _, row := range resp.Values {
		project := types.Project{Name: row[0].(string)}
		projects = append(projects, project)
	}

	return projects, nil
}
