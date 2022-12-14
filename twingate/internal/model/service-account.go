package model

type ServiceAccount struct {
	ID   string
	Name string
}

func (s ServiceAccount) GetName() string {
	return s.Name
}

func (s ServiceAccount) GetID() string {
	return s.ID
}
