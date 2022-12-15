// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export function getTwingateGroup(args: GetTwingateGroupArgs, opts?: pulumi.InvokeOptions): Promise<GetTwingateGroupResult> {
    if (!opts) {
        opts = {}
    }

    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
    return pulumi.runtime.invoke("twingate:index/getTwingateGroup:getTwingateGroup", {
        "id": args.id,
    }, opts);
}

/**
 * A collection of arguments for invoking getTwingateGroup.
 */
export interface GetTwingateGroupArgs {
    id: string;
}

/**
 * A collection of values returned by getTwingateGroup.
 */
export interface GetTwingateGroupResult {
    readonly id: string;
    readonly isActive: boolean;
    readonly name: string;
    readonly type: string;
}

export function getTwingateGroupOutput(args: GetTwingateGroupOutputArgs, opts?: pulumi.InvokeOptions): pulumi.Output<GetTwingateGroupResult> {
    return pulumi.output(args).apply(a => getTwingateGroup(a, opts))
}

/**
 * A collection of arguments for invoking getTwingateGroup.
 */
export interface GetTwingateGroupOutputArgs {
    id: pulumi.Input<string>;
}
