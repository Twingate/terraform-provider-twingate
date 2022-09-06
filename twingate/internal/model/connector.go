package model

type Connector struct {
	ID        string
	Name      string
	NetworkID string
}

func (c Connector) GetName() string {
	return c.Name
}

func (c Connector) GetID() string {
	return c.ID
}
