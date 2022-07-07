package types

import (
	"log"
	"test-go-slack-bot/utils"
	"time"
)

const (
	Pending   = "Pending"
	Approved  = "Approved"
	Rejected  = "Rejected"
	Cancelled = "Cancelled"
)

const (
	AnnualLeave  = "Annual leave"
	SickLeave    = "Sick leave"
	WeddingLeave = "Wedding leave"
	FuneralLeave = "Funeral leave"
)

var (
	PaidLeaveTypes = []string{AnnualLeave, SickLeave, WeddingLeave, FuneralLeave}
)

type PaidLeaveOption struct {
	Text  string
	Value float64
}

var (
	FullDay = PaidLeaveOption{
		Text:  "Full day",
		Value: 1.0,
	}
	HalfDay = PaidLeaveOption{
		Text:  "Half day",
		Value: 0.5,
	}
	PaidLeaveOptions = []PaidLeaveOption{FullDay, HalfDay}
)

func GetPtoOption(value string) PaidLeaveOption {
	log.Printf("Pto option: %s\n", value)
	float := utils.ToFloat(value)

	if float == float64(1) {
		return FullDay
	}

	return HalfDay
}

const (
	PtoIDColIdx              = 0
	PtoReplyIDColIdx         = 1
	PtoMemberUsernameColIdx  = 2
	PtoDateColIdx            = 3
	PtoTypeColIdx            = 4
	PtoOptionColIdx          = 5
	PtoStatusColIdx          = 6
	PtoMemberNoteColIdx      = 7
	PtoManagerNoteColIdx     = 8
	PtoManagerUsernameColIdx = 9
	PtoCreatedAtColIdx       = 10
	PtoUpdatedAtColIdx       = 11
)

type Pto struct {
	ID               int
	Username         string
	Date             time.Time
	Type             string
	Option           PaidLeaveOption
	Status           string
	Note             string
	ApproverNote     string
	ApproverUsername string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ReplyID          int
}

func NewPto(
	id int,
	username string,
	date time.Time,
	plType string,
	option PaidLeaveOption,
	note string,
) *Pto {
	return &Pto{
		ID:               id,
		Username:         username,
		Date:             date,
		Type:             plType,
		Option:           option,
		Status:           Pending,
		Note:             note,
		ApproverNote:     "",
		ApproverUsername: "",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (pto *Pto) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		pto.ID,
		pto.ReplyID,
		pto.Username,
		utils.ToSpreadsheetDateTime(pto.Date),
		pto.Type,
		pto.Option.Value,
		pto.Status,
		pto.Note,
		pto.ApproverNote,
		pto.ApproverUsername,
		utils.ToSpreadsheetDateTime(pto.CreatedAt),
		utils.ToSpreadsheetDateTime(pto.UpdatedAt),
	}
}

func ToPtoInstace(row []interface{}) Pto {
	return Pto{
		ID:               utils.ToInt(row[PtoIDColIdx].(string)),
		ReplyID:          utils.ToInt(row[PtoReplyIDColIdx].(string)),
		Username:         row[PtoMemberUsernameColIdx].(string),
		Date:             utils.FromSpreadsheetDateTime(row[PtoDateColIdx].(string)),
		Type:             row[PtoTypeColIdx].(string),
		Option:           GetPtoOption(row[PtoOptionColIdx].(string)),
		Status:           row[PtoStatusColIdx].(string),
		Note:             row[PtoMemberNoteColIdx].(string),
		ApproverNote:     row[PtoManagerNoteColIdx].(string),
		ApproverUsername: row[PtoManagerUsernameColIdx].(string),
		CreatedAt:        utils.FromSpreadsheetDateTime(row[PtoCreatedAtColIdx].(string)),
		UpdatedAt:        utils.FromSpreadsheetDateTime(row[PtoUpdatedAtColIdx].(string)),
	}
}
