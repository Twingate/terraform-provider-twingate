package model

const (
	StatusActive  = "ACTIVE"
	StatusRevoked = "REVOKED"
)

type ServiceAccountKey struct {
	ID               string
	Name             string
	Status           string
	ServiceAccountID string
	ExpirationTime   int
}

func (s ServiceAccountKey) GetName() string {
	return s.Name
}

func (s ServiceAccountKey) GetID() string {
	return s.ID
}

func (s ServiceAccountKey) IsActive() bool {
	return s.Status == StatusActive
}
