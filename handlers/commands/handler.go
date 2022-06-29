package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

const (
	OVERTIME_REQUEST   = "/kai-ot"
	PAID_LEAVE_REQUEST = "/kai-pl-request"
	OVERTIME_HISOTRY   = "/kai-ot-history"
)

func HandleCommandRequest(c *gin.Context, api *slack.Client) {
	command, err := verify(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"text": err.Error()})
	}
	// log.Printf("Command request: %s\n", utils.JSONString(command))
	switch command.Command {
	case OVERTIME_REQUEST:
		go func() {
			err := handleOvertimeCommand(*command, api)
			if err != nil {
				// log.Fatalf("Unable to handle overtime command: %s\n", err.Error())
				log.Printf("Unable to handle overtime command: %s\n", err.Error())
			}
		}()
		// Loading animation
		msg := fmt.Sprintf("Wait for a sec, I'm loading your request.")
		c.JSON(http.StatusOK, gin.H{"text": msg})

	case OVERTIME_HISOTRY:
		go func() {
			err := handleOTHistoryCommand(*command, api)
			if err != nil {
				// c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
				log.Printf("Unable to find overtime history: %s\n", err.Error())
			}
		}()
		// Loading animation
		msg := fmt.Sprintf("Wait for a sec :blobdance: I'm gathering your ot history.")
		c.JSON(http.StatusOK, gin.H{"text": msg})

	case PAID_LEAVE_REQUEST:
		log.Println("Paid leave request")
		if err := handlePaidLeaveCommand(*command, api, c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"text": err.Error()})
		}

	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unsupported command"})
	}
}

// Verify command requests from Slack
func verify(c *gin.Context) (*slack.SlashCommand, error) {
	// log.Printf("ENV: %s\n", os.Getenv("SLACK_SIGNING_SECRET"))
	// log.Printf("Secret: %s\n", secret)
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
