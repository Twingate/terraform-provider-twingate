// Code generated by the Pulumi Terraform Bridge (tfgen) Tool DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package twingate

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).
//
// ## Example Usage
//
// ```go
// package main
//
// import (
//
//	"github.com/Twingate-Labs/pulumi-twingate/sdk/go/twingate"
//	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
//
// )
//
//	func main() {
//		pulumi.Run(func(ctx *pulumi.Context) error {
//			_, err := twingate.GetTwingateGroups(ctx, &twingate.GetTwingateGroupsArgs{
//				Name: pulumi.StringRef("<your group's name>"),
//			}, nil)
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//
// ```
func GetTwingateGroups(ctx *pulumi.Context, args *GetTwingateGroupsArgs, opts ...pulumi.InvokeOption) (*GetTwingateGroupsResult, error) {
	opts = pkgInvokeDefaultOpts(opts)
	var rv GetTwingateGroupsResult
	err := ctx.Invoke("twingate:index/getTwingateGroups:getTwingateGroups", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getTwingateGroups.
type GetTwingateGroupsArgs struct {
	// List of Groups
	Groups []GetTwingateGroupsGroup `pulumi:"groups"`
	// Returns only Groups matching the specified state.
	IsActive *bool `pulumi:"isActive"`
	// Returns only Groups that exactly match this name.
	Name *string `pulumi:"name"`
	// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
	Type *string `pulumi:"type"`
}

// A collection of values returned by getTwingateGroups.
type GetTwingateGroupsResult struct {
	// List of Groups
	Groups []GetTwingateGroupsGroup `pulumi:"groups"`
	// The provider-assigned unique ID for this managed resource.
	Id string `pulumi:"id"`
	// Returns only Groups matching the specified state.
	IsActive *bool `pulumi:"isActive"`
	// Returns only Groups that exactly match this name.
	Name *string `pulumi:"name"`
	// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
	Type *string `pulumi:"type"`
}

func GetTwingateGroupsOutput(ctx *pulumi.Context, args GetTwingateGroupsOutputArgs, opts ...pulumi.InvokeOption) GetTwingateGroupsResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (GetTwingateGroupsResult, error) {
			args := v.(GetTwingateGroupsArgs)
			r, err := GetTwingateGroups(ctx, &args, opts...)
			var s GetTwingateGroupsResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(GetTwingateGroupsResultOutput)
}

// A collection of arguments for invoking getTwingateGroups.
type GetTwingateGroupsOutputArgs struct {
	// List of Groups
	Groups GetTwingateGroupsGroupArrayInput `pulumi:"groups"`
	// Returns only Groups matching the specified state.
	IsActive pulumi.BoolPtrInput `pulumi:"isActive"`
	// Returns only Groups that exactly match this name.
	Name pulumi.StringPtrInput `pulumi:"name"`
	// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
	Type pulumi.StringPtrInput `pulumi:"type"`
}

func (GetTwingateGroupsOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetTwingateGroupsArgs)(nil)).Elem()
}

// A collection of values returned by getTwingateGroups.
type GetTwingateGroupsResultOutput struct{ *pulumi.OutputState }

func (GetTwingateGroupsResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetTwingateGroupsResult)(nil)).Elem()
}

func (o GetTwingateGroupsResultOutput) ToGetTwingateGroupsResultOutput() GetTwingateGroupsResultOutput {
	return o
}

func (o GetTwingateGroupsResultOutput) ToGetTwingateGroupsResultOutputWithContext(ctx context.Context) GetTwingateGroupsResultOutput {
	return o
}

// List of Groups
func (o GetTwingateGroupsResultOutput) Groups() GetTwingateGroupsGroupArrayOutput {
	return o.ApplyT(func(v GetTwingateGroupsResult) []GetTwingateGroupsGroup { return v.Groups }).(GetTwingateGroupsGroupArrayOutput)
}

// The provider-assigned unique ID for this managed resource.
func (o GetTwingateGroupsResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetTwingateGroupsResult) string { return v.Id }).(pulumi.StringOutput)
}

// Returns only Groups matching the specified state.
func (o GetTwingateGroupsResultOutput) IsActive() pulumi.BoolPtrOutput {
	return o.ApplyT(func(v GetTwingateGroupsResult) *bool { return v.IsActive }).(pulumi.BoolPtrOutput)
}

// Returns only Groups that exactly match this name.
func (o GetTwingateGroupsResultOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v GetTwingateGroupsResult) *string { return v.Name }).(pulumi.StringPtrOutput)
}

// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
func (o GetTwingateGroupsResultOutput) Type() pulumi.StringPtrOutput {
	return o.ApplyT(func(v GetTwingateGroupsResult) *string { return v.Type }).(pulumi.StringPtrOutput)
}

func init() {
	pulumi.RegisterOutputType(GetTwingateGroupsResultOutput{})
}
