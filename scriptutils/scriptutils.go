package scriptutils

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/script/v1"
)

func AddUser(ctx context.Context, httpClient *http.Client) error {
	service, err := script.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("Unable to retrieve script client: %v\n", err.Error())
	}
	scriptID := "1zFlUruqA4xtAzUR1NC4_opnsW1lNVUYBYAHbPVkAq9NRzfuBOGccautf"
	function := "addUser"
	request := &script.ExecutionRequest{
		Function: function,
		DevMode:  true,
	}
	_, err = service.Scripts.Run(scriptID, request).Do()
	if err != nil {
		return fmt.Errorf("Unable to retrieve script client: %v\n", err.Error())
	}
	return nil
}

// [TODO]: Cannot reload
func ReloadSpreadSheet(ctx context.Context, httpClient *http.Client) {
	service, err := script.NewService(ctx, option.WithHTTPClient(httpClient))
	// srv, err := script.New(httpClient)
	if err != nil {
		log.Fatalf("Unable to retrieve Script client: %v", err)
	}
	updateScriptId := "1zFlUruqA4xtAzUR1NC4_opnsW1lNVUYBYAHbPVkAq9NRzfuBOGccautf"
	function := "refreshOTSheet"
	req := script.ExecutionRequest{Function: function, DevMode: true}
	// log.Println(service)
	// log.Println(updateScriptId)
	// log.Printf("Execution request: %v\n", req)
	// service.Scripts.Run()
	_, err = service.Scripts.Run(updateScriptId, &req).Do()
	// log.Printf("Script executed successfully?: %v\n", ope.Done)
	if err != nil {
		log.Fatalf("Unable to run Script with error: %v", err.Error())
	}
}
