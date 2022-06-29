package types

import "time"

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

func NewOT(ID int, username, project, date, note string, startAt, endAt, createdAt time.Time, duration float64) *OT {
	return &OT{
		ID:        ID,
		Username:  username,
		Project:   project,
		Date:      date,
		StartAt:   startAt,
		EndAt:     endAt,
		Duration:  duration,
		CreatedAt: createdAt,
		Note:      note,
	}
}

func (ot *OT) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		ot.ID,
		ot.Username,
		ot.Project,
		ot.Date,
		ot.StartAt,
		ot.EndAt,
		ot.Duration,
		ot.CreatedAt,
		ot.Note,
	}
}
