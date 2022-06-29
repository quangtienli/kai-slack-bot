package types

import "time"

type PaidLeaveStatus string

const (
	PENDING  PaidLeaveStatus = "pending"
	APPROVED PaidLeaveStatus = "approved"
	REJECTED PaidLeaveStatus = "rejected"
	CANCELED PaidLeaveStatus = "canceled"
)

type PaidLeaveOption float64

const (
	FULL PaidLeaveOption = 1
	HALF PaidLeaveOption = 0.5
)

type PaidLeaveType string

const (
	ANNUAL  PaidLeaveType = "annual"
	SICK    PaidLeaveType = "sick"
	WEDDING PaidLeaveType = "wedding"
	FUNERAL PaidLeaveType = "funeral"
)

type PaidLeaveRequest struct {
	ID               int
	Username         string
	Date             time.Time
	Type             PaidLeaveType
	Option           PaidLeaveOption
	Status           PaidLeaveStatus
	Note             string
	ApproverNote     string
	ApproverUsername string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewPaidLeaveRequest(
	id int,
	username string,
	date time.Time,
	plType PaidLeaveType,
	option PaidLeaveOption,
	note string,
) *PaidLeaveRequest {
	return &PaidLeaveRequest{
		ID:               id,
		Username:         username,
		Date:             date,
		Type:             plType,
		Option:           option,
		Status:           PENDING,
		Note:             note,
		ApproverNote:     "",
		ApproverUsername: "",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func (pto *PaidLeaveRequest) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		pto.ID,
		pto.Username,
		pto.Date,
		pto.Type,
		pto.Option,
		pto.Status,
		pto.Note,
		pto.ApproverNote,
		pto.ApproverUsername,
		pto.CreatedAt,
		pto.UpdatedAt,
	}
}
