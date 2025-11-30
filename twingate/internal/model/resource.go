package model

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

const (
	portRangeSeparator    = "-"
	expectedPortsRangeLen = 2

	PolicyRestricted = "RESTRICTED"
	PolicyAllowAll   = "ALLOW_ALL"
	PolicyDenyAll    = "DENY_ALL"

	ApprovalModeAutomatic = "AUTOMATIC"
	ApprovalModeManual    = "MANUAL"

	AccessPolicyModeManual        = "MANUAL"
	AccessPolicyModeAutoLock      = "AUTO_LOCK"
	AccessPolicyModeAccessRequest = "ACCESS_REQUEST"
)

//nolint:gochecknoglobals
var Policies = []string{PolicyRestricted, PolicyAllowAll, PolicyDenyAll}

type AccessPolicy struct {
	Mode         *string
	Duration     *time.Duration
	ApprovalMode *string
}

func (p *AccessPolicy) Validate() error {
	if p.Mode != nil {
		switch *p.Mode {
		case AccessPolicyModeManual, AccessPolicyModeAutoLock, AccessPolicyModeAccessRequest:
			break
		default:
			return fmt.Errorf("invalid mode: %s", *p.Mode)
		}
	}

	if p.ApprovalMode != nil {
		switch *p.ApprovalMode {
		case ApprovalModeAutomatic, ApprovalModeManual:
			break
		default:
			return fmt.Errorf("invalid approval_mode: %s", *p.Mode)
		}
	}

	if (p.Duration != nil || p.ApprovalMode != nil) && p.Mode == nil {
		return errors.New("mode is required")
	}

	if p.Mode != nil {
		switch *p.Mode {
		case AccessPolicyModeManual:
			break
		case AccessPolicyModeAutoLock:
			if p.Duration == nil || p.ApprovalMode == nil {
				return errors.New("duration and approval_mode are required")
			}

			if *p.Duration < time.Hour*24 {
				return errors.New("minimum duration is 1 day")
			}
			break
		case AccessPolicyModeAccessRequest:
			if p.ApprovalMode == nil {
				return errors.New("approval_mode is required")
			}

			if p.Duration != nil && *p.Duration < time.Hour {
				return errors.New("minimum duration is 1 hour")
			}
			break
		}
	}

	return nil
}

type AccessGroup struct {
	GroupID            string
	SecurityPolicyID   *string
	UsageBasedDuration *int64
	ApprovalMode       *string
	AccessPolicy       *AccessPolicy
}

func (g AccessGroup) Equals(another AccessGroup) bool {
	if g.GroupID == another.GroupID &&
		equalsOptionalString(g.SecurityPolicyID, another.SecurityPolicyID) &&
		equalsOptionalInt64(g.UsageBasedDuration, another.UsageBasedDuration) &&
		equalsOptionalString(g.ApprovalMode, another.ApprovalMode) {
		return true
	}

	return false
}

func equalsOptionalString(s1, s2 *string) bool {
	return s1 == nil && s2 == nil || s1 != nil && s2 != nil && strings.EqualFold(*s1, *s2)
}

func equalsOptionalInt64(i1, i2 *int64) bool {
	return i1 == nil && i2 == nil || i1 != nil && i2 != nil && *i1 == *i2
}

type Resource struct {
	ID                             string
	RemoteNetworkID                string
	Address                        string
	Name                           string
	Protocols                      *Protocols
	IsActive                       bool
	AccessPolicy                   *AccessPolicy
	GroupsAccess                   []AccessGroup
	ServiceAccounts                []string
	IsAuthoritative                bool
	IsVisible                      *bool
	IsBrowserShortcutEnabled       *bool
	Alias                          *string
	SecurityPolicyID               *string
	ApprovalMode                   string
	Tags                           map[string]string
	UsageBasedAutolockDurationDays *int64
}

func (r Resource) AccessToTerraform() []interface{} {
	rawMap := make(map[string]interface{})
	if len(r.GroupsAccess) != 0 {
		rawMap[attr.GroupIDs] = utils.Map(r.GroupsAccess, func(item AccessGroup) string {
			return item.GroupID
		})
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

type ResourceFilter interface {
	GetName() string
	IsNil() bool
	HasNotSupportedFilters() bool
	GetFilterBy() string
	GetTypes() []string
	GetIsActive() *bool
	GetTags() map[string]string
}

func (r Resource) Match(filter ResourceFilter) bool {
	if filter.IsNil() {
		// matches all resources
		return true
	}

	if filter.HasNotSupportedFilters() {
		// for not supported filters we delegate fetching data from API
		return false
	}

	// filter by tags
	for k, v := range filter.GetTags() {
		if r.Tags[k] != v {
			return false
		}
	}

	// filter by name
	if name := filter.GetName(); name != "" {
		switch filter.GetFilterBy() {
		case "":
			if r.Name != name {
				return false
			}

		case attr.FilterByContains:
			if !strings.Contains(r.Name, name) {
				return false
			}

		case attr.FilterByExclude:
			if strings.Contains(r.Name, name) {
				return false
			}

		case attr.FilterByPrefix:
			if !strings.HasPrefix(r.Name, name) {
				return false
			}

		case attr.FilterBySuffix:
			if !strings.HasSuffix(r.Name, name) {
				return false
			}

		case attr.FilterByRegexp:
			matched, err := regexp.MatchString(name, r.Name)
			if err != nil || !matched {
				return false
			}
		}
	}

	return true
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

type ResourcesFilter struct {
	Name       *string
	NameFilter string
	Tags       map[string]string
}

func (f *ResourcesFilter) HasName() bool {
	return f != nil && f.Name != nil && *f.Name != ""
}

func (f *ResourcesFilter) GetName() string {
	if f.HasName() {
		return *f.Name
	}

	return ""
}

func (f *ResourcesFilter) GetFilterBy() string {
	return f.NameFilter
}

func (f *ResourcesFilter) GetTypes() []string {
	// not supported
	return nil
}

func (f *ResourcesFilter) GetIsActive() *bool {
	// not supported
	return nil
}

func (f *ResourcesFilter) IsNil() bool {
	return f == nil
}

func (f *ResourcesFilter) HasNotSupportedFilters() bool {
	return f != nil && !slices.Contains([]string{"", attr.FilterByRegexp, attr.FilterByContains, attr.FilterByExclude, attr.FilterByPrefix, attr.FilterBySuffix}, f.NameFilter)
}

func (f *ResourcesFilter) GetTags() map[string]string {
	return f.Tags
}
