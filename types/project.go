package types

type Project struct {
	Name string
}

const (
	ProjectNameColIdx = 0
)

func NewProject(name string) *Project {
	return &Project{
		Name: name,
	}
}

func (p *Project) ToSpreadsheetObject() []interface{} {
	return []interface{}{
		p.Name,
	}
}

func ToProjectInstance(row []interface{}) *Project {
	return NewProject(
		row[ProjectNameColIdx].(string),
	)
}
