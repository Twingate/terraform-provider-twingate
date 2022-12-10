package model

const (
	LocationAWS         = "AWS"
	LocationAzure       = "AZURE"
	LocationGoogleCloud = "GOOGLE_CLOUD"
	LocationOnPremise   = "ON_PREMISE"
	LocationOther       = "OTHER"
)

var Locations = []string{LocationAWS, LocationAzure, LocationGoogleCloud, LocationOnPremise, LocationOther} //nolint

type RemoteNetwork struct {
	ID       string
	Name     string
	Location string
}

func (n RemoteNetwork) GetName() string {
	return n.Name
}

func (n RemoteNetwork) GetID() string {
	return n.ID
}
