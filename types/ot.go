package types

import (
	"test-go-slack-bot/utils"
	"time"
)

const (
	OtIDColIdx        = 0
	OtUsernameColIdx  = 1
	OtProjectColIdx   = 2
	OtDateColIdx      = 3
	OtStartTimeColIdx = 4
	OtEndTimeColIdx   = 5
	OtDurationColIdx  = 6
	OtCreatedAtColIdx = 7
	OtNoteColIdx      = 8
)

type OT struct {
	ID        int
	Username  string
	Project   string
	Date      string
	StartAt   time.Time
	EndAt     time.Time
	Duration  float64
	CreatedAt time.Time
	Note      string
}

func NewOT(ID int, username, project, date, note string, startAt, endAt time.Time, duration float64) *OT {
	return &OT{
		ID:        ID,
		Username:  username,
		Project:   project,
		Date:      date,
		StartAt:   startAt,
		EndAt:     endAt,
		Duration:  duration,
		CreatedAt: time.Now(),
		Note:      note,
	}
}

func (ot *OT) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		ot.ID,
		ot.Username,
		ot.Project,
		ot.Date,
		utils.ToSpreadsheetDateTime(ot.StartAt),
		utils.ToSpreadsheetDateTime(ot.EndAt),
		ot.Duration,
		utils.ToSpreadsheetDateTime(ot.CreatedAt),
		ot.Note,
	}
}

func ToOtInstance(row []interface{}) *OT {
	// "note" exists
	var note string
	if len(row) > 8 {
		note = row[OtNoteColIdx].(string)
	}

	return NewOT(
		utils.ToInt(row[OtIDColIdx].(string)),
		row[OtUsernameColIdx].(string),
		row[OtProjectColIdx].(string),
		row[OtDateColIdx].(string),
		note,
		utils.FromSpreadsheetDateTime(row[OtStartTimeColIdx].(string)),
		utils.FromSpreadsheetDateTime(row[OtEndTimeColIdx].(string)),
		utils.ToFloat(row[OtDurationColIdx].(string)),
	)
}
