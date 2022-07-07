package events

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func HandleEventRequest(c *gin.Context, api *slack.Client) {
	eventsAPIEvent, err := verify(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"text": err.Error()})
	}
	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			go func() {
				handleAppMentionEvent(event, api, c)
			}()
		}
	}
}

func verify(c *gin.Context) (*slackevents.EventsAPIEvent, error) {
	verifier, err := slack.NewSecretsVerifier(c.Request.Header, os.Getenv("SLACK_SIGNING_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("Unable to verify header event request: %s\n", err.Error())
	}
	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read event request body: %s\n", err.Error())
	}
	if _, err := verifier.Write(bytes); err != nil {
		return nil, err
	}
	eventsAPIEvent, err := slackevents.ParseEvent(bytes, slackevents.OptionNoVerifyToken())
	if err != nil {
		return nil, fmt.Errorf("Unable to parse event request: %s\n", err.Error())
	}
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var challengeResp *slackevents.ChallengeResponse
		if err := json.Unmarshal(bytes, &challengeResp); err != nil {
			return nil, fmt.Errorf("Unable to unmarshal challenge response: %s\n", err.Error())
		}
		c.Request.Header.Set("Content-Type", "text")
		c.String(http.StatusOK, challengeResp.Challenge)
	}
	return &eventsAPIEvent, nil
}
