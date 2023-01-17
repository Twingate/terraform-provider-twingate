package model

import (
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const (
	portRangeSeparator    = "-"
	expectedPortsRangeLen = 2

	PolicyRestricted = "RESTRICTED"
	PolicyAllowAll   = "ALLOW_ALL"
	PolicyDenyAll    = "DENY_ALL"
)

//nolint:gochecknoglobals
var Policies = []string{PolicyRestricted, PolicyAllowAll, PolicyDenyAll}

type Resource struct {
	ID                       string
	RemoteNetworkID          string
	Address                  string
	Name                     string
	Groups                   []string
	Protocols                *Protocols
	IsActive                 bool
	IsVisible                *bool
	IsBrowserShortcutEnabled *bool
}

func (r Resource) GetID() string {
	return r.ID
}

func (r Resource) GetName() string {
	return r.Name
}

func (r Resource) ToTerraform() interface{} {
	return map[string]interface{}{
		"id":                r.ID,
		"name":              r.Name,
		"address":           r.Address,
		"remote_network_id": r.RemoteNetworkID,
		"protocols":         r.Protocols.ToTerraform(),
	}
}

type PortRange struct {
	Start int32
	End   int32
}

func (p PortRange) String() string {
	if p.Start == p.End {
		return fmt.Sprintf("%d", p.Start)
	}

	return fmt.Sprintf("%d-%d", p.Start, p.End)
}

func NewPortRange(str string) (*PortRange, error) {
	var (
		portRange *PortRange
		err       error
	)

	if strings.Contains(str, portRangeSeparator) {
		portRange, err = newPortRange(str)
	} else {
		portRange, err = newSinglePort(str)
	}

	if err != nil {
		return nil, ErrInvalidPortRange(str, err)
	}

	return portRange, nil
}

func newSinglePort(str string) (*PortRange, error) {
	port, err := validatePort(str)
	if err != nil {
		return nil, err
	}

	return &PortRange{Start: port, End: port}, nil
}

func newPortRange(str string) (*PortRange, error) {
	ports := strings.Split(str, portRangeSeparator)
	if len(ports) != expectedPortsRangeLen {
		return nil, ErrInvalidPortRangeLen
	}

	start, err := validatePort(ports[0])
	if err != nil {
		return nil, err
	}

	end, err := validatePort(ports[1])
	if err != nil {
		return nil, err
	}

	if end < start {
		return nil, NewPortRangeNotRisingSequenceError(start, end)
	}

	return &PortRange{
		Start: start,
		End:   end,
	}, nil
}

type Protocol struct {
	Ports  []*PortRange
	Policy string
}

func (p *Protocol) PortsToString() []string {
	if len(p.Ports) == 0 {
		return nil
	}

	return utils.Map[*PortRange, string](p.Ports, func(port *PortRange) string {
		return port.String()
	})
}

func NewProtocol(policy string, ports []*PortRange) *Protocol {
	switch policy {
	case PolicyAllowAll:
		return &Protocol{Policy: PolicyAllowAll}
	case PolicyDenyAll:
		return &Protocol{Policy: PolicyRestricted}
	default:
		return &Protocol{Policy: policy, Ports: ports}
	}
}

func DefaultProtocol() *Protocol {
	return &Protocol{
		Policy: PolicyAllowAll,
	}
}

type Protocols struct {
	UDP       *Protocol
	TCP       *Protocol
	AllowIcmp bool
}

func DefaultProtocols() *Protocols {
	return &Protocols{
		UDP:       DefaultProtocol(),
		TCP:       DefaultProtocol(),
		AllowIcmp: true,
	}
}

func (p *Protocols) ToTerraform() []interface{} {
	if p == nil {
		return nil
	}

	rawMap := make(map[string]interface{})
	rawMap["allow_icmp"] = p.AllowIcmp

	if p.TCP != nil {
		rawMap["tcp"] = p.TCP.ToTerraform()
	}

	if p.UDP != nil {
		rawMap["udp"] = p.UDP.ToTerraform()
	}

	return []interface{}{rawMap}
}

func (p *Protocol) ToTerraform() []interface{} {
	if p == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"policy": p.Policy,
			"ports":  p.PortsToString(),
		},
	}
}
