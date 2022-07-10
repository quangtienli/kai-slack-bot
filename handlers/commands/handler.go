package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	OtRequest = "/kai-ot-request"
	PlRequest = "/kai-pl-request"
	OtHistory = "/kai-ot-history"
)

func HandleCommandRequest(c *gin.Context, api *slack.Client) {
	command, err := verify(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"text": err.Error()})
	}

	switch command.Command {
	case OtRequest:
		handleOvertTimeRequestCommand(command, api)
	case OtHistory:
		handleOverTimeHistoryCommand(command, api, c)
	case PlRequest:
		handlePaidLeaveRequestCommand(command, api, c)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unsupported command"})
	}
}

// Verify command requests from Slack
func verify(c *gin.Context) (*slack.SlashCommand, error) {
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("Unable to verify interaction request: %s\n", err.Error())
	}

	c.Request.Body = ioutil.NopCloser(io.TeeReader(c.Request.Body, &verifier))
	command, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request to slash command: %s\n", err.Error())
	}

	if err = verifier.Ensure(); err != nil {
		return nil, fmt.Errorf("Unauthorized interaction request: %s\n", err.Error())
	}

	return &command, nil
}
