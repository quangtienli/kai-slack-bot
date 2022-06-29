package interactions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"test-go-slack-bot/handlers/commands"
	"test-go-slack-bot/utils"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"golang.org/x/exp/slices"
)

const (
	CALLBACK_ID_UNKNOWN = "callback-id-unknown"

	MORE_ANNUAL_LEAVE = "more-annual-leave"
)

const (
	BUTTON_MORE_DAY_PAID_LEAVE   = "button-more-day-paid-leave"
	BUTTON_LESS_DAY_PAID_LEAVE   = "button-less-day-paid-leave"
	PAID_LEAVE_MODAL             = "paid-leave-modal"
	OT_MODAL                     = "ot-modal"
	PAID_LEAVE_MODAL_SUBMISSIONN = "paid-leave-modal-submission"
)

func HandleInteractionRequest(c *gin.Context, api *slack.Client) {
	msg, err := verify(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"text": err.Error()})
	}
	log.Printf("Interaction message: %s\n", utils.JSONString(*msg))

	switch identifyRequestType(msg) {
	case OT_MODAL:
		go func() {
			err := handleOTRequestModalSubmission(*msg, api, c)
			if err != nil {
				log.Printf("Unable to save overtime ot submission: %s\n", err.Error())
			}
		}()
	case PAID_LEAVE_MODAL:
		handlePaidLeaveRequest(*msg, api, c)
	case BUTTON_MORE_DAY_PAID_LEAVE:
		// log.Println("Add more date!")
		handleMoreDayPaidLeaveRequest(*msg, api, c)
	case BUTTON_LESS_DAY_PAID_LEAVE:
		// log.Println("Remove a date")
		handleLessDayPaidLeaveRequest(*msg, api, c)
	case PAID_LEAVE_MODAL_SUBMISSIONN:
		go func() {
			handlePaidLeaveRequestSubmissionByType(msg, api)
		}()
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unsupported interactions"})
	}
}

// Identify type of interaction
func identifyRequestType(msg *slack.InteractionCallback) string {
	if msg.Type == slack.InteractionTypeViewSubmission && msg.View.CallbackID == commands.CALLBACK_ID_OT_MODAL_REQUEST {
		return OT_MODAL
	}

	if msg.Type == slack.InteractionTypeViewSubmission && msg.View.CallbackID == commands.CALLBACK_ID_PAID_LEAVE_MODAL {
		return PAID_LEAVE_MODAL
	}

	if msg.Type == slack.InteractionTypeViewSubmission && slices.Index([]string{CALLBACK_ID_ANNUAL_LEAVE_MODAL, CALLBACK_ID_FUNERAL_LEAVE_MODAL, CALLBACK_ID_SICK_LEAVE_MODAL, CALLBACK_ID_WEDDING_LEAVE_MODAL}, msg.View.CallbackID) > -1 {
		return PAID_LEAVE_MODAL_SUBMISSIONN
	}

	if msg.Type == slack.InteractionTypeBlockActions && msg.BlockID == commands.ID_START_DATE {
		return commands.ID_START_DATE
	}

	if msg.Type == slack.InteractionTypeBlockActions && msg.BlockID == "" {
		actionCallback := msg.ActionCallback.BlockActions[0]
		// log.Printf("%s\n", utils.JSONString(actionCallback))
		// // User add more day in paid leave request
		if actionCallback.BlockID == ID_LEAVE_MORE_DAY && actionCallback.ActionID == ACTION_ID_LEAVE_MORE_DAY {
			return BUTTON_MORE_DAY_PAID_LEAVE
		}

		// User remove a selected day in paid leave request
		if strings.Contains(actionCallback.BlockID, ID_LEAVE_DAY) && strings.Contains(actionCallback.ActionID, ACTION_ID_REMOVE_PAID_LEAVE) {
			// log.Printf("Response after clicking onto remove the date: %s\n", utils.JSONString(actionCallback))
			return BUTTON_LESS_DAY_PAID_LEAVE
		}
	}

	return CALLBACK_ID_UNKNOWN
}

// Verify interaction requests from Slack
func verify(c *gin.Context) (*slack.InteractionCallback, error) {
	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read body request: %s\n", err.Error())
	}
	str, err := url.QueryUnescape(string(bytes)[8:])
	if err != nil {
		return nil, fmt.Errorf("Unable to query unescape: %s\n", err.Error())
	}
	var msg slack.InteractionCallback
	if err := json.Unmarshal([]byte(str), &msg); err != nil {
		return nil, fmt.Errorf("Failed to decode json message from Slack: %v\n", err.Error())
	}
	return &msg, nil
}
