package sheetutils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"
	"time"

	"github.com/slack-go/slack"
	"google.golang.org/api/sheets/v4"
)

const (
	ptoReadRange      = "PTO!A2:L"
	totalPtoReadRange = "PTO!M2:M2"

	replyReadRange      = "Replies!A2:D"
	totalReplyReadRange = "Replies!E2:E2"
)

func UpdateApprovedPtoByReplyID(replyID, managerUsername string) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	replyIdsReadRange := "PTO!B2:B"
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, replyIdsReadRange).Do()
	if err != nil {
		panic(err)
	} else {
		log.Printf("Reply ids: %s\n", utils.JSONString(resp.Values))

		for idx, row := range resp.Values {
			if row[0].(string) == replyID {
				updateStatusValueRange := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!G%d:G%d", idx+2, idx+2),
					Values: [][]interface{}{{types.Approved}},
				}
				updateApproverUsername := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!J%d:J%d", idx+2, idx+2),
					Values: [][]interface{}{{managerUsername}},
				}
				updateUpdatedAt := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!L%d:L%d", idx+2, idx+2),
					Values: [][]interface{}{{utils.ToSpreadsheetDateTime(time.Now())}},
				}
				batchResp, err := service.Spreadsheets.Values.BatchUpdate(
					spreadsheetID,
					&sheets.BatchUpdateValuesRequest{
						Data: []*sheets.ValueRange{
							updateStatusValueRange,
							updateApproverUsername,
							updateUpdatedAt,
						},
						ValueInputOption: VALUE_INPUT_OPTION_USER_ENTERED,
					},
				).Do()
				if err != nil {
					panic(err)
				} else {
					// log.Printf("Update pl status response: %s\n", utils.JSONString(updateResp))
					log.Printf("Update pl status response: %s\n", utils.JSONString(batchResp))
				}
			}
		}
	}
}

func UpdateDeniedPtoByReplyID(replyID, note, managerUsername string) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	replyIdsReadRange := "PTO!B2:B"
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, replyIdsReadRange).Do()
	if err != nil {
		panic(err)
	} else {
		log.Printf("Reply ids: %s\n", utils.JSONString(resp.Values))

		for idx, row := range resp.Values {
			if row[0].(string) == replyID {
				updateStatusValueRange := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!G%d:G%d", idx+2, idx+2),
					Values: [][]interface{}{{types.Rejected}},
				}
				updateApproverNote := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!I%d:I%d", idx+2, idx+2),
					Values: [][]interface{}{{note}},
				}
				updateApproverUsername := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!J%d:J%d", idx+2, idx+2),
					Values: [][]interface{}{{managerUsername}},
				}
				updateUpdatedAt := &sheets.ValueRange{
					Range:  fmt.Sprintf("PTO!L%d:L%d", idx+2, idx+2),
					Values: [][]interface{}{{utils.ToSpreadsheetDateTime(time.Now())}},
				}
				batchResp, err := service.Spreadsheets.Values.BatchUpdate(
					spreadsheetID,
					&sheets.BatchUpdateValuesRequest{
						Data: []*sheets.ValueRange{
							updateStatusValueRange,
							updateApproverUsername,
							updateUpdatedAt,
							updateApproverNote,
						},
						ValueInputOption: VALUE_INPUT_OPTION_USER_ENTERED,
					},
				).Do()
				if err != nil {
					panic(err)
				} else {
					// log.Printf("Update pl status response: %s\n", utils.JSONString(updateResp))
					log.Printf("Update pl status response: %s\n", utils.JSONString(batchResp))
				}

			}
		}
	}
}

func InsertPto(pto types.Pto) {
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
	).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
	if err != nil {
		panic(err)
	} else {
		totalPTOValueRange := &sheets.ValueRange{
			Values: [][]interface{}{
				{pto.ID + 1},
			},
		}
		_, err := service.Spreadsheets.Values.Update(
			spreadsheetID,
			totalPtoReadRange,
			totalPTOValueRange,
		).ValueInputOption(VALUE_INPUT_OPTION_USER_ENTERED).Do()
		if err != nil {
			panic(err)
		}
	}
}

func FindPtosByReplyID(replyID string) []types.Pto {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	// Get member's paid leave requests
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, ptoReadRange).Do()
	if err != nil {
		panic(err)
	}

	ptos := []types.Pto{}
	for _, row := range resp.Values {
		if row[types.PtoReplyIDColIdx].(string) == replyID {
			ptos = append(ptos, types.ToPtoInstace(row))
		}
	}

	return ptos
}

func GetTotalPto() int {
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

// [TODO]
func GetRemainingDaysByPLType(plType string, user *slack.User) int {
	var count int = 12
	return count
}
