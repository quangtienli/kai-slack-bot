package types

import "test-go-slack-bot/utils"

type Reply struct {
	ID              int
	MemberUsername  string
	ManagerUsername string
	Status          string
}

const (
	ReplyPending   = "Pending"
	ReplyCompleted = "Completed"
)

const (
	ReplyIDColIdx              = 0
	ReplyMemberUsernameColIdx  = 1
	ReplyManagerUsernameColIdx = 2
	ReplyStatusColIdx          = 3
)

func NewReply(id int, memberUsername, managerUsername, status string) *Reply {
	return &Reply{
		ID:              id,
		MemberUsername:  memberUsername,
		ManagerUsername: managerUsername,
		Status:          status,
	}
}

func (r *Reply) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		r.ID,
		r.MemberUsername,
		r.ManagerUsername,
		r.Status,
	}
}

func ToReplyInstance(row []interface{}) *Reply {
	return NewReply(
		utils.ToInt(row[ReplyIDColIdx].(string)),
		row[ReplyMemberUsernameColIdx].(string),
		row[ReplyManagerUsernameColIdx].(string),
		row[ReplyStatusColIdx].(string),
	)
}
