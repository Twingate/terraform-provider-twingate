package resource

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func convertResource(data *schema.ResourceData) (*model.Resource, error) {
	protocols, err := convertProtocols(data)
	if err != nil {
		return nil, err
	}

	res := &model.Resource{
		Name:            data.Get("name").(string),
		RemoteNetworkID: data.Get("remote_network_id").(string),
		Address:         data.Get("address").(string),
		Groups:          convertGroups(data),
		Protocols:       protocols,
	}

	isVisible, ok := data.GetOkExists("is_visible") //nolint
	if val := isVisible.(bool); ok {
		res.IsVisible = &val
	}

	isBrowserShortcutEnabled, ok := data.GetOkExists("is_browser_shortcut_enabled") //nolint
	if val := isBrowserShortcutEnabled.(bool); ok {
		res.IsBrowserShortcutEnabled = &val
	}

	return res, nil
}

func convertGroup(resourceData *schema.ResourceData) *model.Group {
	return &model.Group{
		ID:               resourceData.Id(),
		Name:             resourceData.Get("name").(string),
		SecurityPolicyID: resourceData.Get("security_policy_id").(string),
	}
}

func convertGroups(data *schema.ResourceData) []string {
	return utils.Map[interface{}, string](
		data.Get("group_ids").(*schema.Set).List(),
		func(elem interface{}) string {
			return elem.(string)
		},
	)
}

func convertProtocols(data *schema.ResourceData) (*model.Protocols, error) {
	rawList := data.Get("protocols").([]interface{})
	if len(rawList) == 0 {
		return model.DefaultProtocols(), nil
	}

	rawMap := rawList[0].(map[string]interface{})

	udp, err := convertProtocol(rawMap["udp"].([]interface{}))
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocol(rawMap["tcp"].([]interface{}))
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		UDP:       udp,
		TCP:       tcp,
		AllowIcmp: rawMap["allow_icmp"].(bool),
	}, nil
}

func convertProtocol(rawList []interface{}) (*model.Protocol, error) {
	if len(rawList) == 0 {
		return nil, nil //nolint:nilnil
	}

	rawMap := rawList[0].(map[string]interface{})
	policy := rawMap["policy"].(string)

	ports, err := convertPorts(rawMap["ports"].([]interface{}))
	if err != nil {
		return nil, err
	}

	return model.NewProtocol(policy, ports), nil
}

func convertPorts(rawList []interface{}) ([]*model.PortRange, error) {
	var ports = make([]*model.PortRange, 0, len(rawList))

	for _, port := range rawList {
		var str string
		if port != nil {
			str = port.(string)
		}

		portRange, err := model.NewPortRange(str)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		ports = append(ports, portRange)
	}

	return ports, nil
}
