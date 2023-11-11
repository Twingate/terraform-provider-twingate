// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as utilities from "./utilities";

export function getTwingateSecurityPolicies(args?: GetTwingateSecurityPoliciesArgs, opts?: pulumi.InvokeOptions): Promise<GetTwingateSecurityPoliciesResult> {
    args = args || {};

    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("twingate:index/getTwingateSecurityPolicies:getTwingateSecurityPolicies", {
        "securityPolicies": args.securityPolicies,
    }, opts);
}

/**
 * A collection of arguments for invoking getTwingateSecurityPolicies.
 */
export interface GetTwingateSecurityPoliciesArgs {
    securityPolicies?: inputs.GetTwingateSecurityPoliciesSecurityPolicy[];
}

/**
 * A collection of values returned by getTwingateSecurityPolicies.
 */
export interface GetTwingateSecurityPoliciesResult {
    /**
     * The provider-assigned unique ID for this managed resource.
     */
    readonly id: string;
    readonly securityPolicies?: outputs.GetTwingateSecurityPoliciesSecurityPolicy[];
}
export function getTwingateSecurityPoliciesOutput(args?: GetTwingateSecurityPoliciesOutputArgs, opts?: pulumi.InvokeOptions): pulumi.Output<GetTwingateSecurityPoliciesResult> {
    return pulumi.output(args).apply((a: any) => getTwingateSecurityPolicies(a, opts))
}

/**
 * A collection of arguments for invoking getTwingateSecurityPolicies.
 */
export interface GetTwingateSecurityPoliciesOutputArgs {
    securityPolicies?: pulumi.Input<pulumi.Input<inputs.GetTwingateSecurityPoliciesSecurityPolicyArgs>[]>;
}
