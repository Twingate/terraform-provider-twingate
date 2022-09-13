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
		Groups:          convertGroups(data),
		Protocols:       protocols,
	}, nil
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
		return nil, nil
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
		if port == nil {
			continue
		}

		str := port.(string)
		if str == "" {
			continue
		}

		portRange, err := model.NewPortRange(str)
		if err != nil {
			return nil, err
		}

		ports = append(ports, portRange)
	}

	if cap(ports) > len(ports) {
		ports = ports[:len(ports):len(ports)]
	}

	return ports, nil
}

//
//func (pi *Protocol) BuildPortsRange() ([]string, string) {
//	var ports []string
//
//	for _, port := range pi.Ports {
//		if port.Start == port.End {
//			ports = append(ports, strconv.Itoa(int(port.Start)))
//		} else {
//			ports = append(ports, strconv.Itoa(int(port.Start))+"-"+strconv.Itoa(int(port.End)))
//		}
//	}
//
//	return ports, string(pi.Policy)
//}
