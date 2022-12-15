// Code generated by the Pulumi Terraform Bridge (tfgen) Tool DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package twingate

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TwingateResource struct {
	pulumi.CustomResourceState

	// The Resource's IP/CIDR or FQDN/DNS zone
	Address pulumi.StringOutput `pulumi:"address"`
	// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
	// from the Twingate Admin Console or API
	GroupIds pulumi.StringArrayOutput `pulumi:"groupIds"`
	// The name of the Resource
	Name pulumi.StringOutput `pulumi:"name"`
	// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
	// restriction, and all protocols and ports are allowed.
	Protocols TwingateResourceProtocolsPtrOutput `pulumi:"protocols"`
	// Remote Network ID where the Resource lives
	RemoteNetworkId pulumi.StringOutput `pulumi:"remoteNetworkId"`
}

// NewTwingateResource registers a new resource with the given unique name, arguments, and options.
func NewTwingateResource(ctx *pulumi.Context,
	name string, args *TwingateResourceArgs, opts ...pulumi.ResourceOption) (*TwingateResource, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Address == nil {
		return nil, errors.New("invalid value for required argument 'Address'")
	}
	if args.Name == nil {
		return nil, errors.New("invalid value for required argument 'Name'")
	}
	if args.RemoteNetworkId == nil {
		return nil, errors.New("invalid value for required argument 'RemoteNetworkId'")
	}
	opts = pkgResourceDefaultOpts(opts)
	var resource TwingateResource
	err := ctx.RegisterResource("twingate:index/twingateResource:TwingateResource", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetTwingateResource gets an existing TwingateResource resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetTwingateResource(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *TwingateResourceState, opts ...pulumi.ResourceOption) (*TwingateResource, error) {
	var resource TwingateResource
	err := ctx.ReadResource("twingate:index/twingateResource:TwingateResource", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering TwingateResource resources.
type twingateResourceState struct {
	// The Resource's IP/CIDR or FQDN/DNS zone
	Address *string `pulumi:"address"`
	// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
	// from the Twingate Admin Console or API
	GroupIds []string `pulumi:"groupIds"`
	// The name of the Resource
	Name *string `pulumi:"name"`
	// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
	// restriction, and all protocols and ports are allowed.
	Protocols *TwingateResourceProtocols `pulumi:"protocols"`
	// Remote Network ID where the Resource lives
	RemoteNetworkId *string `pulumi:"remoteNetworkId"`
}

type TwingateResourceState struct {
	// The Resource's IP/CIDR or FQDN/DNS zone
	Address pulumi.StringPtrInput
	// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
	// from the Twingate Admin Console or API
	GroupIds pulumi.StringArrayInput
	// The name of the Resource
	Name pulumi.StringPtrInput
	// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
	// restriction, and all protocols and ports are allowed.
	Protocols TwingateResourceProtocolsPtrInput
	// Remote Network ID where the Resource lives
	RemoteNetworkId pulumi.StringPtrInput
}

func (TwingateResourceState) ElementType() reflect.Type {
	return reflect.TypeOf((*twingateResourceState)(nil)).Elem()
}

type twingateResourceArgs struct {
	// The Resource's IP/CIDR or FQDN/DNS zone
	Address string `pulumi:"address"`
	// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
	// from the Twingate Admin Console or API
	GroupIds []string `pulumi:"groupIds"`
	// The name of the Resource
	Name string `pulumi:"name"`
	// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
	// restriction, and all protocols and ports are allowed.
	Protocols *TwingateResourceProtocols `pulumi:"protocols"`
	// Remote Network ID where the Resource lives
	RemoteNetworkId string `pulumi:"remoteNetworkId"`
}

// The set of arguments for constructing a TwingateResource resource.
type TwingateResourceArgs struct {
	// The Resource's IP/CIDR or FQDN/DNS zone
	Address pulumi.StringInput
	// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
	// from the Twingate Admin Console or API
	GroupIds pulumi.StringArrayInput
	// The name of the Resource
	Name pulumi.StringInput
	// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
	// restriction, and all protocols and ports are allowed.
	Protocols TwingateResourceProtocolsPtrInput
	// Remote Network ID where the Resource lives
	RemoteNetworkId pulumi.StringInput
}

func (TwingateResourceArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*twingateResourceArgs)(nil)).Elem()
}

type TwingateResourceInput interface {
	pulumi.Input

	ToTwingateResourceOutput() TwingateResourceOutput
	ToTwingateResourceOutputWithContext(ctx context.Context) TwingateResourceOutput
}

func (*TwingateResource) ElementType() reflect.Type {
	return reflect.TypeOf((**TwingateResource)(nil)).Elem()
}

func (i *TwingateResource) ToTwingateResourceOutput() TwingateResourceOutput {
	return i.ToTwingateResourceOutputWithContext(context.Background())
}

func (i *TwingateResource) ToTwingateResourceOutputWithContext(ctx context.Context) TwingateResourceOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TwingateResourceOutput)
}

// TwingateResourceArrayInput is an input type that accepts TwingateResourceArray and TwingateResourceArrayOutput values.
// You can construct a concrete instance of `TwingateResourceArrayInput` via:
//
//	TwingateResourceArray{ TwingateResourceArgs{...} }
type TwingateResourceArrayInput interface {
	pulumi.Input

	ToTwingateResourceArrayOutput() TwingateResourceArrayOutput
	ToTwingateResourceArrayOutputWithContext(context.Context) TwingateResourceArrayOutput
}

type TwingateResourceArray []TwingateResourceInput

func (TwingateResourceArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TwingateResource)(nil)).Elem()
}

func (i TwingateResourceArray) ToTwingateResourceArrayOutput() TwingateResourceArrayOutput {
	return i.ToTwingateResourceArrayOutputWithContext(context.Background())
}

func (i TwingateResourceArray) ToTwingateResourceArrayOutputWithContext(ctx context.Context) TwingateResourceArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TwingateResourceArrayOutput)
}

// TwingateResourceMapInput is an input type that accepts TwingateResourceMap and TwingateResourceMapOutput values.
// You can construct a concrete instance of `TwingateResourceMapInput` via:
//
//	TwingateResourceMap{ "key": TwingateResourceArgs{...} }
type TwingateResourceMapInput interface {
	pulumi.Input

	ToTwingateResourceMapOutput() TwingateResourceMapOutput
	ToTwingateResourceMapOutputWithContext(context.Context) TwingateResourceMapOutput
}

type TwingateResourceMap map[string]TwingateResourceInput

func (TwingateResourceMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TwingateResource)(nil)).Elem()
}

func (i TwingateResourceMap) ToTwingateResourceMapOutput() TwingateResourceMapOutput {
	return i.ToTwingateResourceMapOutputWithContext(context.Background())
}

func (i TwingateResourceMap) ToTwingateResourceMapOutputWithContext(ctx context.Context) TwingateResourceMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TwingateResourceMapOutput)
}

type TwingateResourceOutput struct{ *pulumi.OutputState }

func (TwingateResourceOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**TwingateResource)(nil)).Elem()
}

func (o TwingateResourceOutput) ToTwingateResourceOutput() TwingateResourceOutput {
	return o
}

func (o TwingateResourceOutput) ToTwingateResourceOutputWithContext(ctx context.Context) TwingateResourceOutput {
	return o
}

// The Resource's IP/CIDR or FQDN/DNS zone
func (o TwingateResourceOutput) Address() pulumi.StringOutput {
	return o.ApplyT(func(v *TwingateResource) pulumi.StringOutput { return v.Address }).(pulumi.StringOutput)
}

// List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved
// from the Twingate Admin Console or API
func (o TwingateResourceOutput) GroupIds() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *TwingateResource) pulumi.StringArrayOutput { return v.GroupIds }).(pulumi.StringArrayOutput)
}

// The name of the Resource
func (o TwingateResourceOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *TwingateResource) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

// Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
// restriction, and all protocols and ports are allowed.
func (o TwingateResourceOutput) Protocols() TwingateResourceProtocolsPtrOutput {
	return o.ApplyT(func(v *TwingateResource) TwingateResourceProtocolsPtrOutput { return v.Protocols }).(TwingateResourceProtocolsPtrOutput)
}

// Remote Network ID where the Resource lives
func (o TwingateResourceOutput) RemoteNetworkId() pulumi.StringOutput {
	return o.ApplyT(func(v *TwingateResource) pulumi.StringOutput { return v.RemoteNetworkId }).(pulumi.StringOutput)
}

type TwingateResourceArrayOutput struct{ *pulumi.OutputState }

func (TwingateResourceArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*TwingateResource)(nil)).Elem()
}

func (o TwingateResourceArrayOutput) ToTwingateResourceArrayOutput() TwingateResourceArrayOutput {
	return o
}

func (o TwingateResourceArrayOutput) ToTwingateResourceArrayOutputWithContext(ctx context.Context) TwingateResourceArrayOutput {
	return o
}

func (o TwingateResourceArrayOutput) Index(i pulumi.IntInput) TwingateResourceOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *TwingateResource {
		return vs[0].([]*TwingateResource)[vs[1].(int)]
	}).(TwingateResourceOutput)
}

type TwingateResourceMapOutput struct{ *pulumi.OutputState }

func (TwingateResourceMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*TwingateResource)(nil)).Elem()
}

func (o TwingateResourceMapOutput) ToTwingateResourceMapOutput() TwingateResourceMapOutput {
	return o
}

func (o TwingateResourceMapOutput) ToTwingateResourceMapOutputWithContext(ctx context.Context) TwingateResourceMapOutput {
	return o
}

func (o TwingateResourceMapOutput) MapIndex(k pulumi.StringInput) TwingateResourceOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *TwingateResource {
		return vs[0].(map[string]*TwingateResource)[vs[1].(string)]
	}).(TwingateResourceOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*TwingateResourceInput)(nil)).Elem(), &TwingateResource{})
	pulumi.RegisterInputType(reflect.TypeOf((*TwingateResourceArrayInput)(nil)).Elem(), TwingateResourceArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*TwingateResourceMapInput)(nil)).Elem(), TwingateResourceMap{})
	pulumi.RegisterOutputType(TwingateResourceOutput{})
	pulumi.RegisterOutputType(TwingateResourceArrayOutput{})
	pulumi.RegisterOutputType(TwingateResourceMapOutput{})
}
