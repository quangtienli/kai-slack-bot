package sheetutils

import (
	"os"
	"test-go-slack-bot/types"
)

func FindAllProjects() ([]types.Project, error) {
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
		project := types.ToProjectInstance(row)
		projects = append(projects, *project)
	}

	return projects, nil
}
