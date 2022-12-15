// Code generated by the Pulumi Terraform Bridge (tfgen) Tool DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package twingate

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func LookupTwingateRemoteNetwork(ctx *pulumi.Context, args *LookupTwingateRemoteNetworkArgs, opts ...pulumi.InvokeOption) (*LookupTwingateRemoteNetworkResult, error) {
	opts = pkgInvokeDefaultOpts(opts)
	var rv LookupTwingateRemoteNetworkResult
	err := ctx.Invoke("twingate:index/getTwingateRemoteNetwork:getTwingateRemoteNetwork", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getTwingateRemoteNetwork.
type LookupTwingateRemoteNetworkArgs struct {
	Id   *string `pulumi:"id"`
	Name *string `pulumi:"name"`
}

// A collection of values returned by getTwingateRemoteNetwork.
type LookupTwingateRemoteNetworkResult struct {
	Id       *string `pulumi:"id"`
	Location string  `pulumi:"location"`
	Name     *string `pulumi:"name"`
}

func LookupTwingateRemoteNetworkOutput(ctx *pulumi.Context, args LookupTwingateRemoteNetworkOutputArgs, opts ...pulumi.InvokeOption) LookupTwingateRemoteNetworkResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (LookupTwingateRemoteNetworkResult, error) {
			args := v.(LookupTwingateRemoteNetworkArgs)
			r, err := LookupTwingateRemoteNetwork(ctx, &args, opts...)
			var s LookupTwingateRemoteNetworkResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(LookupTwingateRemoteNetworkResultOutput)
}

// A collection of arguments for invoking getTwingateRemoteNetwork.
type LookupTwingateRemoteNetworkOutputArgs struct {
	Id   pulumi.StringPtrInput `pulumi:"id"`
	Name pulumi.StringPtrInput `pulumi:"name"`
}

func (LookupTwingateRemoteNetworkOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupTwingateRemoteNetworkArgs)(nil)).Elem()
}

// A collection of values returned by getTwingateRemoteNetwork.
type LookupTwingateRemoteNetworkResultOutput struct{ *pulumi.OutputState }

func (LookupTwingateRemoteNetworkResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupTwingateRemoteNetworkResult)(nil)).Elem()
}

func (o LookupTwingateRemoteNetworkResultOutput) ToLookupTwingateRemoteNetworkResultOutput() LookupTwingateRemoteNetworkResultOutput {
	return o
}

func (o LookupTwingateRemoteNetworkResultOutput) ToLookupTwingateRemoteNetworkResultOutputWithContext(ctx context.Context) LookupTwingateRemoteNetworkResultOutput {
	return o
}

func (o LookupTwingateRemoteNetworkResultOutput) Id() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupTwingateRemoteNetworkResult) *string { return v.Id }).(pulumi.StringPtrOutput)
}

func (o LookupTwingateRemoteNetworkResultOutput) Location() pulumi.StringOutput {
	return o.ApplyT(func(v LookupTwingateRemoteNetworkResult) string { return v.Location }).(pulumi.StringOutput)
}

func (o LookupTwingateRemoteNetworkResultOutput) Name() pulumi.StringPtrOutput {
	return o.ApplyT(func(v LookupTwingateRemoteNetworkResult) *string { return v.Name }).(pulumi.StringPtrOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupTwingateRemoteNetworkResultOutput{})
}
