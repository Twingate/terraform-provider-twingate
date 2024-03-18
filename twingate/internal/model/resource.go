package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
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
	Protocols                *Protocols
	IsActive                 bool
	Groups                   []string
	ServiceAccounts          []string
	IsAuthoritative          bool
	IsVisible                *bool
	IsBrowserShortcutEnabled *bool
	Alias                    *string
	SecurityPolicyID         *string
}

func (r Resource) AccessToTerraform() []interface{} {
	rawMap := make(map[string]interface{})
	if len(r.Groups) != 0 {
		rawMap[attr.GroupIDs] = r.Groups
	}

	if len(r.ServiceAccounts) != 0 {
		rawMap[attr.ServiceAccountIDs] = r.ServiceAccounts
	}

	if len(rawMap) == 0 {
		return nil
	}

	return []interface{}{rawMap}
}

func (r Resource) GetID() string {
	return r.ID
}

func (r Resource) GetName() string {
	return r.Name
}

func (r Resource) ToTerraform() interface{} {
	return map[string]interface{}{
		attr.ID:              r.ID,
		attr.Name:            r.Name,
		attr.Address:         r.Address,
		attr.RemoteNetworkID: r.RemoteNetworkID,
		attr.Protocols:       r.Protocols.ToTerraform(),
	}
}

type PortRange struct {
	Start int
	End   int
}

func (p PortRange) String() string {
	if p.Start == p.End {
		return strconv.Itoa(p.Start)
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
	return &Protocol{Policy: policy, Ports: ports}
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
	rawMap[attr.AllowIcmp] = p.AllowIcmp

	if p.TCP != nil {
		rawMap[attr.TCP] = p.TCP.ToTerraform()
	}

	if p.UDP != nil {
		rawMap[attr.UDP] = p.UDP.ToTerraform()
	}

	return []interface{}{rawMap}
}

func (p *Protocol) ToTerraform() []interface{} {
	if p == nil {
		return nil
	}

	policy := p.Policy
	if p.Policy == PolicyRestricted && len(p.Ports) == 0 {
		policy = PolicyDenyAll
	}

	return []interface{}{
		map[string]interface{}{
			attr.Policy: policy,
			attr.Ports:  p.PortsToString(),
		},
	}
}
