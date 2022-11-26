package model

type RemoteNetwork struct {
	ID   string
	Name string
}

func (n RemoteNetwork) GetName() string {
	return n.Name
}

func (n RemoteNetwork) GetID() string {
	return n.ID
}
