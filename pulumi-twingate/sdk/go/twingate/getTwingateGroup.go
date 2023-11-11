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

func LookupTwingateGroup(ctx *pulumi.Context, args *LookupTwingateGroupArgs, opts ...pulumi.InvokeOption) (*LookupTwingateGroupResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv LookupTwingateGroupResult
	err := ctx.Invoke("twingate:index/getTwingateGroup:getTwingateGroup", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// A collection of arguments for invoking getTwingateGroup.
type LookupTwingateGroupArgs struct {
	Id string `pulumi:"id"`
}

// A collection of values returned by getTwingateGroup.
type LookupTwingateGroupResult struct {
	Id               string `pulumi:"id"`
	IsActive         bool   `pulumi:"isActive"`
	Name             string `pulumi:"name"`
	SecurityPolicyId string `pulumi:"securityPolicyId"`
	Type             string `pulumi:"type"`
}

func LookupTwingateGroupOutput(ctx *pulumi.Context, args LookupTwingateGroupOutputArgs, opts ...pulumi.InvokeOption) LookupTwingateGroupResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (LookupTwingateGroupResult, error) {
			args := v.(LookupTwingateGroupArgs)
			r, err := LookupTwingateGroup(ctx, &args, opts...)
			var s LookupTwingateGroupResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(LookupTwingateGroupResultOutput)
}

// A collection of arguments for invoking getTwingateGroup.
type LookupTwingateGroupOutputArgs struct {
	Id pulumi.StringInput `pulumi:"id"`
}

func (LookupTwingateGroupOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupTwingateGroupArgs)(nil)).Elem()
}

// A collection of values returned by getTwingateGroup.
type LookupTwingateGroupResultOutput struct{ *pulumi.OutputState }

func (LookupTwingateGroupResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*LookupTwingateGroupResult)(nil)).Elem()
}

func (o LookupTwingateGroupResultOutput) ToLookupTwingateGroupResultOutput() LookupTwingateGroupResultOutput {
	return o
}

func (o LookupTwingateGroupResultOutput) ToLookupTwingateGroupResultOutputWithContext(ctx context.Context) LookupTwingateGroupResultOutput {
	return o
}

func (o LookupTwingateGroupResultOutput) ToOutput(ctx context.Context) pulumix.Output[LookupTwingateGroupResult] {
	return pulumix.Output[LookupTwingateGroupResult]{
		OutputState: o.OutputState,
	}
}

func (o LookupTwingateGroupResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v LookupTwingateGroupResult) string { return v.Id }).(pulumi.StringOutput)
}

func (o LookupTwingateGroupResultOutput) IsActive() pulumi.BoolOutput {
	return o.ApplyT(func(v LookupTwingateGroupResult) bool { return v.IsActive }).(pulumi.BoolOutput)
}

func (o LookupTwingateGroupResultOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v LookupTwingateGroupResult) string { return v.Name }).(pulumi.StringOutput)
}

func (o LookupTwingateGroupResultOutput) SecurityPolicyId() pulumi.StringOutput {
	return o.ApplyT(func(v LookupTwingateGroupResult) string { return v.SecurityPolicyId }).(pulumi.StringOutput)
}

func (o LookupTwingateGroupResultOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v LookupTwingateGroupResult) string { return v.Type }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(LookupTwingateGroupResultOutput{})
}
