package sheetutils

import (
	"fmt"
	"log"
	"os"
	"test-go-slack-bot/types"
	"test-go-slack-bot/utils"

	"google.golang.org/api/sheets/v4"
)

const (
	sumPtoReadRange = "Sum PTO!A2:H"
)

func FindSumPtoByUsername(username string) (int, *types.SumPto) {
	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")

	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, sumPtoReadRange).Do()
	if err != nil {
		panic(err)
	}

	for idx, row := range resp.Values {
		if row[types.SumPtoUsernameColIdx].(string) == username {
			return idx + 2, types.ToSumPtoInstance(row)
		}
	}

	return -1, nil
}

func UpdateSumPtoDayTypeByReplyID(replyID string) {
	ptos := FindPtosByReplyID(replyID)
	log.Printf("Number of pto requests: %d\n", len(ptos))
	log.Printf("Pto requests: %s\n", utils.JSONString(ptos))
	if len(ptos) == 0 {
		return
	}

	var duration float64 = 0
	for _, pto := range ptos {
		duration += pto.Option.Value
	}
	log.Printf("Duration: %1.f\n", duration)
	rowIdx, sumPto := FindSumPtoByUsername(ptos[0].Username)
	if sumPto == nil {
		return
	}

	service := initSheetService()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	ptoType := ptos[0].Type
	var batchUpdateValueRequest = &sheets.BatchUpdateValuesRequest{
		ValueInputOption: VALUE_INPUT_OPTION_USER_ENTERED,
	}

	if ptoType == types.SickLeave {
		if sumPto.Sick+duration <= 2 {
			sickValueRange := &sheets.ValueRange{
				Range: PtoTypeReadRange(types.SickLeave, rowIdx),
				Values: [][]interface{}{
					{sumPto.Sick + duration},
				},
			}
			batchUpdateValueRequest.Data = append(batchUpdateValueRequest.Data, sickValueRange)
		} else {
			maxSick := 2
			sickValueRange := &sheets.ValueRange{
				Range: PtoTypeReadRange(types.SickLeave, rowIdx),
				Values: [][]interface{}{
					{maxSick},
				},
			}
			annualValueRange := &sheets.ValueRange{
				Range: PtoTypeReadRange(types.AnnualLeave, rowIdx),
				Values: [][]interface{}{
					{sumPto.Annual + (duration - (float64(maxSick) - sumPto.Sick))},
				},
			}
			log.Printf("Annual duration: %1.f\n", sumPto.Annual+(duration-(float64(maxSick)-sumPto.Sick)))
			batchUpdateValueRequest.Data = append(batchUpdateValueRequest.Data, sickValueRange, annualValueRange)
		}
	}

	if ptoType == types.WeddingLeave {
		weddingValueRange := &sheets.ValueRange{
			Range: PtoTypeReadRange(types.WeddingLeave, rowIdx),
			Values: [][]interface{}{
				{sumPto.Wedding + duration},
			},
		}
		batchUpdateValueRequest.Data = append(batchUpdateValueRequest.Data, weddingValueRange)
	}

	if ptoType == types.FuneralLeave {
		funeralValueRange := &sheets.ValueRange{
			Range: PtoTypeReadRange(types.FuneralLeave, rowIdx),
			Values: [][]interface{}{
				{sumPto.Funeral + duration},
			},
		}
		batchUpdateValueRequest.Data = append(batchUpdateValueRequest.Data, funeralValueRange)
	}

	if ptoType == types.AnnualLeave {
		annualValueRange := &sheets.ValueRange{
			Range: PtoTypeReadRange(ptoType, rowIdx),
			Values: [][]interface{}{
				{sumPto.Annual + duration},
			},
		}
		batchUpdateValueRequest.Data = append(batchUpdateValueRequest.Data, annualValueRange)
	}

	resp, err := service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateValueRequest).Do()
	if err != nil {
		panic(err)
	}

	log.Printf("UpdatePtoDayTypeByReplyID - response: %s\n", utils.JSONString(resp))
}

func PtoTypeReadRange(ptoType string, rowIdx int) string {
	if ptoType == types.SickLeave {
		return fmt.Sprintf("Sum PTO!C%d:C%d", rowIdx, rowIdx)

	}

	if ptoType == types.WeddingLeave {
		return fmt.Sprintf("Sum PTO!D%d:D%d", rowIdx, rowIdx)

	}

	if ptoType == types.FuneralLeave {
		return fmt.Sprintf("Sum PTO!E%d:E%d", rowIdx, rowIdx)

	}

	return fmt.Sprintf("Sum PTO!B%d:B%d", rowIdx, rowIdx)
}
