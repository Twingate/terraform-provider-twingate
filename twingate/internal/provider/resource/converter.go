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

	return &model.Resource{
		Name:            data.Get("name").(string),
		RemoteNetworkID: data.Get("remote_network_id").(string),
		Address:         data.Get("address").(string),
		Protocols:       protocols,
		Groups:          convertGroups(data),
		ServiceAccounts: convertServiceAccounts(data),
	}, nil
}

func convertGroups(data *schema.ResourceData) []string {
	if groupIDs, ok := data.GetOk("group_ids"); ok {
		return convertIDs(groupIDs)
	}

	groups, _ := convertAccess(data)

	return groups
}

func convertIDs(data interface{}) []string {
	return utils.Map[interface{}, string](
		data.(*schema.Set).List(),
		func(elem interface{}) string {
			return elem.(string)
		},
	)
}

func convertAccess(data *schema.ResourceData) ([]string, []string) {
	rawList := data.Get("access").([]interface{})
	if len(rawList) == 0 {
		return nil, nil
	}

	if rawList[0] == nil {
		return nil, nil
	}

	rawMap := rawList[0].(map[string]interface{})

	return convertIDs(rawMap["group_ids"]), convertIDs(rawMap["service_account_ids"])
}

func convertServiceAccounts(data *schema.ResourceData) []string {
	_, serviceAccounts := convertAccess(data)

	return serviceAccounts
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
