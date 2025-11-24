package model

import "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"

const (
	LocationAWS         = "AWS"
	LocationAzure       = "AZURE"
	LocationGoogleCloud = "GOOGLE_CLOUD"
	LocationOnPremise   = "ON_PREMISE"
	LocationOther       = "OTHER"

	NetworkTypeRegular = "REGULAR"
	NetworkTypeExit    = "EXIT"
)

var Locations = []string{LocationAWS, LocationAzure, LocationGoogleCloud, LocationOnPremise, LocationOther} //nolint

type RemoteNetwork struct {
	ID       string
	Name     string
	Location string
	Type     string
}

func (n RemoteNetwork) GetName() string {
	return n.Name
}

func (n RemoteNetwork) GetID() string {
	return n.ID
}

func (n RemoteNetwork) ToTerraform() any {
	return map[string]any{
		attr.ID:       n.ID,
		attr.Name:     n.Name,
		attr.Location: n.Location,
	}
}
