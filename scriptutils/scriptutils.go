package scriptutils

import (
	"context"
	"log"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/script/v1"
)

// Reload current active spreadsheet
func ReloadSpreadSheet(ctx context.Context, httpClient *http.Client) {
	service, err := script.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Script client: %v", err)
	}
	updateScriptId := "1zFlUruqA4xtAzUR1NC4_opnsW1lNVUYBYAHbPVkAq9NRzfuBOGccautf"
	function := "refreshOTSheet"
	req := script.ExecutionRequest{Function: function, DevMode: true}
	_, err = service.Scripts.Run(updateScriptId, &req).Do()
	if err != nil {
		log.Fatalf("Unable to run Script with error: %v", err.Error())
	}
}
