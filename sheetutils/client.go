package sheetutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"test-go-slack-bot/scriptutils"

	_ "golang.org/x/exp/slices"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	GOOGLE_APIS_AUTH_SPREADSHEETS_CURRENTONLY = "https://www.googleapis.com/auth/spreadsheets.currentonly"
	GOOGLE_APIS_AUTH_SPREADSHEETS             = "https://www.googleapis.com/auth/spreadsheets"
	GOOGLE_APIS_AUTH_SPREADSHEETS_READONLY    = "https://www.googleapis.com/auth/spreadsheets.readonly"
	GOOGLE_APIS_AUTH_SCRIPT_PROJECTS          = "https://www.googleapis.com/auth/script.projects"
	GOOGLE_APIS_AUTH_SCRIPT_APP               = "https://www.googleapis.com/auth/script.scriptapp"
	GOOGLE_APIS_AUTH_DRIVE_FILE               = "https://www.googleapis.com/auth/drive.file"
)

const (
	TOKEN_FILE_PATH       = "token.json"
	CREDENTIALS_FILE_PATH = "credentials.json"
)

const (
	VALUE_INPUT_OPTION_RAW          = "RAW"
	VALUE_INPUT_OPTION_USER_ENTERED = "USER_ENTERED"
)

func tokenFromFilePath(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}
	return token, nil
}

func tokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in the browser then type the authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to scan authorization code: %v\n", err)
	}
	ctx := context.TODO()
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v\n", err)
	}
	return token
}

func saveTokenToFile(file string, token *oauth2.Token) {
	log.Printf("Saving credentials file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v\n", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getHttpClient(config *oauth2.Config) *http.Client {
	token, err := tokenFromFilePath(TOKEN_FILE_PATH)
	if err != nil {
		token = tokenFromWeb(config)
		saveTokenToFile(TOKEN_FILE_PATH, token)
	}
	ctx := context.Background()
	return config.Client(ctx, token)
}

func initSheetService() *sheets.Service {
	b, err := ioutil.ReadFile(CREDENTIALS_FILE_PATH)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v\n", err)
	}
	config, err := google.ConfigFromJSON(
		b,
		GOOGLE_APIS_AUTH_SPREADSHEETS_READONLY,
		GOOGLE_APIS_AUTH_SPREADSHEETS_CURRENTONLY,
		GOOGLE_APIS_AUTH_SPREADSHEETS,
		GOOGLE_APIS_AUTH_SCRIPT_APP,
		GOOGLE_APIS_AUTH_SCRIPT_PROJECTS,
		GOOGLE_APIS_AUTH_DRIVE_FILE,
	)
	if err != nil {
		log.Fatalf("Unable to parse cliennt secret file to config: %v\n", err)
	}
	ctx := context.Background()
	httpClient := getHttpClient(config)
	scriptutils.ReloadSpreadSheet(ctx, httpClient)
	service, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Google Sheets client: %v\n", err)
	}
	return service
}
