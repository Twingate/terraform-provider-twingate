package customvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Attribute names mirroring the in_cluster/address/bearer_token_file attributes of
// twingate_kubernetes_resource, the first user of the bool-sibling validators.
const (
	testInCluster       = "in_cluster"
	testAddress         = "address"
	testBearerTokenFile = "bearer_token_file"
	testDefaultAddress  = "kubernetes.default.svc.cluster.local"
	testCustomAddress   = "k8s-api.example.com"
)

var (
	siblingTrue    = tftypes.NewValue(tftypes.Bool, true)
	siblingFalse   = tftypes.NewValue(tftypes.Bool, false)
	siblingNull    = tftypes.NewValue(tftypes.Bool, nil)
	siblingUnknown = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
)

func boolSiblingSchema(boolName, stringName string) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			boolName:   schema.BoolAttribute{Optional: true, Computed: true},
			stringName: schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func boolSiblingType(boolName, stringName string) tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			boolName:   tftypes.Bool,
			stringName: tftypes.String,
		},
	}
}

// boolSiblingConfig builds a config holding one bool and one string attribute with
// the given raw values.
func boolSiblingConfig(boolName, stringName string, sibling, value tftypes.Value) tfsdk.Config {
	return tfsdk.Config{
		Schema: boolSiblingSchema(boolName, stringName),
		Raw: tftypes.NewValue(boolSiblingType(boolName, stringName), map[string]tftypes.Value{
			boolName:   sibling,
			stringName: value,
		}),
	}
}

// boolSiblingNullConfig builds the null config Terraform hands validators on destroy.
func boolSiblingNullConfig(boolName, stringName string) tfsdk.Config {
	return tfsdk.Config{
		Schema: boolSiblingSchema(boolName, stringName),
		Raw:    tftypes.NewValue(boolSiblingType(boolName, stringName), nil),
	}
}
