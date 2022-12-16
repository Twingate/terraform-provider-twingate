package model

const (
	StatusActive  = "ACTIVE"
	StatusRevoked = "REVOKED"
)

type ServiceKey struct {
	ID             string
	Name           string
	Status         string
	Service        string
	ExpirationTime int
	Token          string
}

func (s ServiceKey) GetName() string {
	return s.Name
}

func (s ServiceKey) GetID() string {
	return s.ID
}

func (s ServiceKey) IsActive() bool {
	return s.Status == StatusActive
}
