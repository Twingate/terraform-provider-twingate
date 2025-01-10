package model

type DLPPolicy struct {
	ID   string
	Name string
}

func (p DLPPolicy) GetName() string {
	return p.Name
}

func (p DLPPolicy) GetID() string {
	return p.ID
}
