package model

type Group struct {
	ID       string
	Name     string
	Type     string
	IsActive bool
}

func (g Group) GetName() string {
	return g.Name
}

func (g Group) GetID() string {
	return g.ID
}
