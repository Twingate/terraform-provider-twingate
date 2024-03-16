package resource

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var schemaResourceV0 = &schema.Schema{
	Attributes: map[string]schema.Attribute{
		attr.ID: schema.StringAttribute{
			Computed: true,
		},
		attr.Name: schema.StringAttribute{
			Required: true,
		},
		attr.Address: schema.StringAttribute{
			Required: true,
		},
		attr.RemoteNetworkID: schema.StringAttribute{
			Required: true,
		},
		attr.IsActive: schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		attr.IsAuthoritative: schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		attr.Alias: schema.StringAttribute{
			Optional: true,
		},
		attr.SecurityPolicyID: schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		attr.IsVisible: schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
	},

	Blocks: map[string]schema.Block{
		attr.Access: schema.ListNestedBlock{
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					attr.GroupIDs: schema.SetAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
					attr.ServiceAccountIDs: schema.SetAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
			},
		},
		attr.Protocols: schema.ListNestedBlock{
			Validators: []validator.List{
				listvalidator.SizeAtMost(1),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					attr.AllowIcmp: schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
				Blocks: map[string]schema.Block{
					attr.UDP: schema.ListNestedBlock{
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								attr.Policy: schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
								attr.Ports: schema.SetAttribute{
									Optional:    true,
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
					attr.TCP: schema.ListNestedBlock{
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								attr.Policy: schema.StringAttribute{
									Optional: true,
									Computed: true,
								},
								attr.Ports: schema.SetAttribute{
									Optional:    true,
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	},
}
