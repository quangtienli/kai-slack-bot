package interactions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"test-go-slack-bot/handlers/commands"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"golang.org/x/exp/slices"
)

const (
	ButtonMoreDayPaidLeave = "button-more-day-paid-leave"
	ButtonLessDayPaidLeave = "button-less-day-paid-leave"

	PaidLeaveRequest           = "paid-leave-request"
	PaidLeaveRequestSubmission = "paid-leave-request-submission"
	OtModalSubmission          = "ot-request"

	ManagerAcceptReply          = "manager-accept-reply"
	ManagerDenyReply            = "manager-deny-reply"
	ManagerSubmitDenyReplyModal = "manager-submit-deny-reply-modal"

	SelectPaidLeaveRequestType = "select-paid-leave-request-type"

	UnknownRequestType = "unknown-request-type"
)

func HandleInteractionRequest(c *gin.Context, api *slack.Client) {
	msg, err := verify(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"text": err.Error()})
	}

	// log.Printf("HandleInteractionRequest: %s\n", utils.JSONString(msg))
	switch identifyRequestType(msg) {
	case OtModalSubmission:
		handleOtModalSubmission(msg, api, c)

	case PaidLeaveRequest:
		handlePaidLeaveRequest(msg, api, c)

	case ButtonMoreDayPaidLeave:
		handleMoreDayPaidLeaveRequest(msg, api)

	case ButtonLessDayPaidLeave:
		handleLessDayPaidLeaveRequest(msg, api)

	case PaidLeaveRequestSubmission:
		handlePaidLeaveRequestSubmissionByType(msg, api, c)

	case ManagerAcceptReply:
		handleAcceptReply(msg, api)

	case ManagerDenyReply:
		handleDenyReply(msg, api)

	case ManagerSubmitDenyReplyModal:
		handleDenyReplySubmission(msg, api)

	default:
		c.JSON(http.StatusOK, gin.H{"message": "Unsupported interactions"})
	}
}

// Identify type of interaction
func identifyRequestType(msg *slack.InteractionCallback) string {
	if msg.Type == slack.InteractionTypeViewSubmission && msg.View.CallbackID == commands.OtRequestModalCallbackID {
		return OtModalSubmission
	}

	if msg.Type == slack.InteractionTypeViewSubmission && msg.View.CallbackID == commands.PlRequestModalCallbackID {
		return PaidLeaveRequest
	}

	if msg.Type == slack.InteractionTypeViewSubmission &&
		slices.Index([]string{AnnualLeaveModalCallbackID, FuneralLeaveCallbackID, SickLeaveModalCallbackID, WeddingLeaveModalCallbackID}, msg.View.CallbackID) > -1 {
		return PaidLeaveRequestSubmission
	}

	if msg.Type == slack.InteractionTypeViewSubmission && strings.Contains(msg.View.CallbackID, ReplyDenyModalCallbackID) {
		return ManagerSubmitDenyReplyModal
	}

	if msg.Type == slack.InteractionTypeBlockActions && msg.BlockID == "" {
		actionCallback := msg.ActionCallback.BlockActions[0]

		if actionCallback.BlockID == plRequestByTypeMoreDayBlockID && actionCallback.ActionID == plRequestByTypeMoreDayActionID {
			return ButtonMoreDayPaidLeave
		}

		if strings.Contains(actionCallback.BlockID, plRequestByTypeDayBlockID) && strings.Contains(actionCallback.ActionID, plRequestByTypeRemoveDayActionID) {
			return ButtonLessDayPaidLeave
		}

		if actionCallback.BlockID == ReplyActionBlockID && actionCallback.ActionID == ReplyApproveButtonActionId {
			return ManagerAcceptReply
		}

		if actionCallback.BlockID == ReplyActionBlockID && actionCallback.ActionID == ReplyDenyButtonActionID {
			return ManagerDenyReply
		}

	}

	return UnknownRequestType
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
