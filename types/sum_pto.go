package types

import "test-go-slack-bot/utils"

const (
	SumPtoUsernameColIdx        = 0
	SumPtoAnnualColIdx          = 1
	SumPtoSickColIdx            = 2
	SumPtoWeddingColIdx         = 3
	SumPtoFuneralColIdx         = 4
	SumPtoAllotmentColIdx       = 5
	SumPtoUserEmailColIdx       = 6
	SumPtoManagerUsernameColIdx = 7
)

type SumPto struct {
	Username        string
	Annual          float64
	Sick            float64
	Wedding         float64
	Funeral         float64
	Allotment       float64
	UserEmail       string
	ManagerUsername string
}

func NewSumPto(
	username string,
	annual float64,
	sick float64,
	wedding float64,
	funeral float64,
	allotment float64,
	userEmail string,
	managerUsername string,
) *SumPto {
	return &SumPto{
		Username:        username,
		Annual:          annual,
		Sick:            sick,
		Wedding:         wedding,
		Funeral:         funeral,
		Allotment:       allotment,
		UserEmail:       userEmail,
		ManagerUsername: managerUsername,
	}
}

func (s *SumPto) ToSpreadsheet() []interface{} {
	return []interface{}{
		s.Username,
		s.Annual,
		s.Sick,
		s.Wedding,
		s.Funeral,
		s.Allotment,
		s.UserEmail,
		s.ManagerUsername,
	}
}

func ToSumPtoInstance(row []interface{}) *SumPto {
	return NewSumPto(
		row[SumPtoManagerUsernameColIdx].(string),
		utils.ToFloat(row[SumPtoAnnualColIdx].(string)),
		utils.ToFloat(row[SumPtoSickColIdx].(string)),
		utils.ToFloat(row[SumPtoWeddingColIdx].(string)),
		utils.ToFloat(row[SumPtoFuneralColIdx].(string)),
		utils.ToFloat(row[SumPtoAllotmentColIdx].(string)),
		row[SumPtoUserEmailColIdx].(string),
		row[SumPtoManagerUsernameColIdx].(string),
	)
}
