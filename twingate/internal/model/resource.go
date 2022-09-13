package model

import (
	"fmt"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
)

const (
	portRangeSeparator = "-"

	PolicyRestricted = "RESTRICTED"
	PolicyAllowAll   = "ALLOW_ALL"
	PolicyDenyAll    = "DENY_ALL"
)

var Policies = []string{PolicyRestricted, PolicyAllowAll, PolicyDenyAll}

type Resource struct {
	ID              string
	RemoteNetworkID string
	Address         string
	Name            string
	Groups          []string
	Protocols       *Protocols
	IsActive        bool
}

func (r Resource) GetID() string {
	return r.ID
}

func (r Resource) GetName() string {
	return r.Name
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
	ports := strings.SplitN(str, portRangeSeparator, 2)

	switch len(ports) {
	case 1:
		port, err := validatePort(ports[0])
		if err != nil {
			return nil, ErrInvalidPortRange(str, err)
		}

		return &PortRange{Start: port, End: port}, nil

	default:
		start, err := validatePort(ports[0])
		if err != nil {
			return nil, ErrInvalidPortRange(ports[0], err)
		}

		end, err := validatePort(ports[1])
		if err != nil {
			return nil, ErrInvalidPortRange(ports[1], err)
		}

		if end < start {
			return nil, ErrInvalidPortRange(str, NewPortRangeNotRisingSequenceError(start, end))
		}

		return &PortRange{
			Start: start,
			End:   end,
		}, nil
	}
}

type Protocol struct {
	Ports  []*PortRange
	Policy string
}

func (p Protocol) PortsToString() []string {
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
