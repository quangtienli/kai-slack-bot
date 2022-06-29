package types

import "time"

type LeaveType string

const (
	ANNUAL_LEAVE  = LeaveType("annual")
	SICK_LEAVE    = LeaveType("sick")
	WEDDING_LEAVE = LeaveType("wedding")
	FUNERAL_LEAVE = LeaveType("funeral")
)

type LeaveRequest struct {
	ID           uint8
	User         string
	Date         time.Time
	Type         LeaveType
	Duration     float32
	Status       string
	MemberNote   string
	ApproverNote string
	Approver     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
