// Code generated by the Pulumi Terraform Bridge (tfgen) Tool DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package twingate

import (
	"context"
	"reflect"

	"github.com/Twingate-Labs/pulumi-twingate/sdk/go/twingate/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumix"
)

func GetTwingateUsers(ctx *pulumi.Context, args *GetTwingateUsersArgs, opts ...pulumi.InvokeOption) (*GetTwingateUsersResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv GetTwingateUsersResult
	err := ctx.Invoke("twingate:index/getTwingateUsers:getTwingateUsers", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getTwingateUsers.
type GetTwingateUsersArgs struct {
	Users []GetTwingateUsersUser `pulumi:"users"`
}

// A collection of values returned by getTwingateUsers.
type GetTwingateUsersResult struct {
	// The provider-assigned unique ID for this managed resource.
	Id    string                 `pulumi:"id"`
	Users []GetTwingateUsersUser `pulumi:"users"`
}

func GetTwingateUsersOutput(ctx *pulumi.Context, args GetTwingateUsersOutputArgs, opts ...pulumi.InvokeOption) GetTwingateUsersResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (GetTwingateUsersResult, error) {
			args := v.(GetTwingateUsersArgs)
			r, err := GetTwingateUsers(ctx, &args, opts...)
			var s GetTwingateUsersResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(GetTwingateUsersResultOutput)
}

// A collection of arguments for invoking getTwingateUsers.
type GetTwingateUsersOutputArgs struct {
	Users GetTwingateUsersUserArrayInput `pulumi:"users"`
}

func (GetTwingateUsersOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*GetTwingateUsersArgs)(nil)).Elem()
}

// A collection of values returned by getTwingateUsers.
type GetTwingateUsersResultOutput struct{ *pulumi.OutputState }

func (GetTwingateUsersResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*GetTwingateUsersResult)(nil)).Elem()
}

func (o GetTwingateUsersResultOutput) ToGetTwingateUsersResultOutput() GetTwingateUsersResultOutput {
	return o
}

func (o GetTwingateUsersResultOutput) ToGetTwingateUsersResultOutputWithContext(ctx context.Context) GetTwingateUsersResultOutput {
	return o
}

func (o GetTwingateUsersResultOutput) ToOutput(ctx context.Context) pulumix.Output[GetTwingateUsersResult] {
	return pulumix.Output[GetTwingateUsersResult]{
		OutputState: o.OutputState,
	}
}

// The provider-assigned unique ID for this managed resource.
func (o GetTwingateUsersResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v GetTwingateUsersResult) string { return v.Id }).(pulumi.StringOutput)
}

func (o GetTwingateUsersResultOutput) Users() GetTwingateUsersUserArrayOutput {
	return o.ApplyT(func(v GetTwingateUsersResult) []GetTwingateUsersUser { return v.Users }).(GetTwingateUsersUserArrayOutput)
}

func init() {
	pulumi.RegisterOutputType(GetTwingateUsersResultOutput{})
}
