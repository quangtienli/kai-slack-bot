package interactions

import (
	"encoding/json"
	"fmt"
	"log"
	"test-go-slack-bot/botutils"
	"test-go-slack-bot/sheetutils"
	"test-go-slack-bot/utils"

	"github.com/slack-go/slack"
)

const (
	ReplyLoadingBlockID = "reply-loading-block-id"

	ReplyApproveMessageBlockID = "reply-approve-message-block-id"
	ReplyDenyMessageBlockID    = "reply-deny-message-block-id"

	ReplyDenyNoteBlockID       = "reply-deny-note-block-id"
	ReplyDenyNoteBlockActionID = "reply-deny-note-block-action-id"

	ReplyDenyModalCallbackID = "reply-deny-modal-callback-id"
)

type DenyReplyPrivateMetadata struct {
	Message slack.Message         `json:"message"`
	Channel slack.Channel         `json:"channel"`
	Actions slack.ActionCallbacks `json:"actions"`
}

// Handle interaction when manager clicks "Approve" on member's paid leave request
func handleAcceptReply(msg *slack.InteractionCallback, api *slack.Client) {
	replyID := msg.ActionCallback.BlockActions[0].Value
	log.Printf("Reply id: %s\n", replyID)
	log.Printf("handleAcceptReply - interaction callback: %s\n", utils.JSONString(msg))

	go func() {
		sendLoadingMessage(msg.Channel.ID, msg.Message, api)
		sheetutils.ConfirmReplyByID(replyID)
		sheetutils.UpdateApprovedPtoByReplyID(replyID, msg.User.Name)
		sheetutils.UpdateSumPtoDayTypeByReplyID(replyID)
		sendApproveMessage(msg, api)

		reply := sheetutils.FindReplyByID(replyID)
		_, sumPto := sheetutils.FindSumPtoByUsername(reply.MemberUsername)

		_, managerSumPto := sheetutils.FindSumPtoByUsername(sumPto.ManagerUsername)
		manager, err := api.GetUserByEmail(managerSumPto.UserEmail)
		if err != nil {
			panic(err)
		}
		acceptedResponseMessageBlocks := buildAcceptedReplyResponseMessageBlocksBySDK(msg.Message.Blocks.BlockSet, manager)
		botutils.SendResponseMessage(msg.Channel.ID, api, acceptedResponseMessageBlocks...)
	}()
}

func buildAcceptedReplyResponseMessageBlocksBySDK(oldMsgBlocks []slack.Block, manager *slack.User) []slack.Block {
	approverInfoBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(":white_check_mark: <@%s> has approved your request.", manager.ID), false, false),
		nil,
		nil,
	)

	return append([]slack.Block{approverInfoBlock}, oldMsgBlocks[1:len(oldMsgBlocks)-2]...)
}

// Handle interaction when manager clicks "Deny" on member's paid leave request
func handleDenyReply(msg *slack.InteractionCallback, api *slack.Client) {
	log.Printf("handleDenyReply - interaction callback: %s\n", utils.JSONString(msg))

	mvr := buildDenyReplyModalBySDK(msg, api)
	viewResp, err := api.OpenView(msg.TriggerID, *mvr)
	if err != nil {
		panic(err)
	} else {
		log.Printf("Deny modal view response: %s\n", utils.JSONString(viewResp))
	}
}

// Handle interaction when manager submit view modal for denying paid leave requests
func handleDenyReplySubmission(msg *slack.InteractionCallback, api *slack.Client) {
	// Unmarshal private metadata
	// log.Printf("handleDenyReplySubmission - interaction callback: %s\n", utils.JSONString(msg))
	denyReplyPrivateMetadata := DenyReplyPrivateMetadata{}
	err := json.Unmarshal([]byte(msg.View.PrivateMetadata), &denyReplyPrivateMetadata)
	if err != nil {
		panic(err)
	}

	replyID := denyReplyPrivateMetadata.Actions.BlockActions[0].Value
	note := msg.View.State.Values[ReplyDenyNoteBlockID][ReplyDenyNoteBlockActionID].Value
	managerUsername := msg.User.Name

	go func() {
		sendLoadingMessage(denyReplyPrivateMetadata.Channel.ID, denyReplyPrivateMetadata.Message, api)
		sheetutils.ConfirmReplyByID(replyID)
		sheetutils.UpdateDeniedPtoByReplyID(replyID, note, managerUsername)
		sendDenyMessage(denyReplyPrivateMetadata.Channel.ID, denyReplyPrivateMetadata.Actions, denyReplyPrivateMetadata.Message, api)

		reply := sheetutils.FindReplyByID(replyID)
		_, sumPto := sheetutils.FindSumPtoByUsername(reply.MemberUsername)
		_, managerSumPto := sheetutils.FindSumPtoByUsername(sumPto.ManagerUsername)
		manager, err := api.GetUserByEmail(managerSumPto.UserEmail)
		if err != nil {
			panic(err)
		}

		deniedResponseMessageBlocks := buildDeniedReplyResponseMessageBlocksBySDK(denyReplyPrivateMetadata.Message.Blocks.BlockSet, manager, note)
		botutils.SendResponseMessage(denyReplyPrivateMetadata.Channel.ID, api, deniedResponseMessageBlocks...)
	}()
}

func buildDeniedReplyResponseMessageBlocksBySDK(oldMsgBlocks []slack.Block, manager *slack.User, managerNote string) []slack.Block {
	deniedInfoBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(":cross-mark-button_274e: <@%s> has denied your request.", manager.ID), false, false),
		nil,
		nil,
	)
	if len(managerNote) > 0 {
		managerInfoBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*:memo: Manager's note: \"%s\"*", managerNote), false, false),
			nil,
			nil,
		)

		return append([]slack.Block{deniedInfoBlock, managerInfoBlock}, oldMsgBlocks[1:len(oldMsgBlocks)-2]...)
	}

	return append([]slack.Block{deniedInfoBlock}, oldMsgBlocks[1:len(oldMsgBlocks)-2]...)
}

// Open view modal request for denying member's paid leave request
func buildDenyReplyModalBySDK(msg *slack.InteractionCallback, api *slack.Client) *slack.ModalViewRequest {
	replyID := msg.ActionCallback.BlockActions[0].Value

	titleText := slack.NewTextBlockObject(slack.PlainTextType, "Note", false, false)
	submitText := slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	closeText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)

	noteBlockElement := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, "Reason", false, false),
		ReplyDenyNoteBlockActionID,
	)
	noteBlockElement.Multiline = true
	noteBlockElement.MinLength = 1

	noteBlock := slack.NewInputBlock(
		ReplyDenyNoteBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, "Reason why rejecting paid leave request", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "This note is required", false, false),
		noteBlockElement,
	)
	noteBlock.Optional = true

	blocks := []slack.Block{
		noteBlock,
	}

	mvr := &slack.ModalViewRequest{}
	mvr.CallbackID = uniqueDenyModalCallbackID(replyID)
	mvr.Type = slack.VTModal
	mvr.Title = titleText
	mvr.Submit = submitText
	mvr.Close = closeText
	mvr.Blocks = slack.Blocks{
		BlockSet: blocks,
	}
	// mvr.ExternalID = ""
	mvr.ClearOnClose = true
	mvr.NotifyOnClose = true

	// Include this message interaction callback in field private metadata
	denyReplyPrivateMetadata := DenyReplyPrivateMetadata{
		Message: msg.Message,
		Channel: msg.Channel,
		Actions: msg.ActionCallback,
	}
	bytes, err := json.Marshal(denyReplyPrivateMetadata)
	if err != nil {
		panic(err)
	}
	mvr.PrivateMetadata = fmt.Sprintf(`%s`, string(bytes))

	return mvr
}

// Update reply in loading state
func sendLoadingMessage(channelID string, slackMessage slack.Message, api *slack.Client) {
	blocks := slackMessage.Blocks.BlockSet
	blocks = blocks[:len(blocks)-1]

	loadingMessageBlock := slack.NewContextBlock(
		ReplyLoadingBlockID,
		slack.NewImageBlockElement("https://i.stack.imgur.com/kOnzy.gif", "Loading"),
	)
	blocks = append(blocks, loadingMessageBlock)
	_, _, _, err := api.UpdateMessage(channelID, slackMessage.Timestamp, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		panic(err)
	}
}

// Update reply in approved state
func sendApproveMessage(msg *slack.InteractionCallback, api *slack.Client) {
	replyId := msg.ActionCallback.BlockActions[0].Value

	blocks := msg.Message.Blocks.BlockSet
	blocks = blocks[:len(blocks)-1]

	successMessageBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, "*:white_check_mark: You have approved this request.*", false, false),
		nil,
		nil,
		slack.SectionBlockOptionBlockID(uniqueSuccessMessageBlockID(replyId)),
	)
	blocks = append(blocks, successMessageBlock)
	_, _, _, err := api.UpdateMessage(msg.Channel.ID, msg.Message.Timestamp, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		panic(err)
	}
}

// Update reply in denied state
func sendDenyMessage(channelID string, actions slack.ActionCallbacks, slackMessage slack.Message, api *slack.Client) {
	replyID := actions.BlockActions[0].Value

	blocks := slackMessage.Blocks.BlockSet
	blocks = blocks[:len(blocks)-1]

	denyMessageBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, "*:cross-mark-button_274e: You have denied this request.*", false, false),
		nil,
		nil,
		slack.SectionBlockOptionBlockID(replyID),
	)
	blocks = append(blocks, denyMessageBlock)

	_, _, _, err := api.UpdateMessage(channelID, slackMessage.Timestamp, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		panic(err)
	}
}

// Send pto request result to member
func sendPtoResult() {

}

// Generate unique block ids for every paid leave request message
func uniqueSuccessMessageBlockID(replyID string) string {
	return fmt.Sprintf("%s-%s", ReplyApproveMessageBlockID, replyID)
}

func uniqueDenyMessageBlockID(replyID string) string {
	return fmt.Sprintf("%s-%s", ReplyDenyMessageBlockID, replyID)
}

func uniqueDenyModalCallbackID(replyID string) string {
	return fmt.Sprintf("%s-%s", ReplyDenyModalCallbackID, replyID)
}
