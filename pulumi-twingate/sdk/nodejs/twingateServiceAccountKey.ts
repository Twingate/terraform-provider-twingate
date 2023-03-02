// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

/**
 * A Service Key authorizes access to all Resources assigned to a Service Account.
 *
 * ## Example Usage
 *
 * ```typescript
 * import * as pulumi from "@pulumi/pulumi";
 * import * as twingate from "@twingate-labs/pulumi-twingate";
 *
 * const githubActionsProd = new twingate.TwingateServiceAccount("githubActionsProd", {name: "Github Actions PROD"});
 * const githubKey = new twingate.TwingateServiceAccountKey("githubKey", {
 *     name: "Github Actions PROD key",
 *     serviceAccountId: githubActionsProd.id,
 * });
 * ```
 */
export class TwingateServiceAccountKey extends pulumi.CustomResource {
    /**
     * Get an existing TwingateServiceAccountKey resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param state Any extra arguments used during the lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, state?: TwingateServiceAccountKeyState, opts?: pulumi.CustomResourceOptions): TwingateServiceAccountKey {
        return new TwingateServiceAccountKey(name, <any>state, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'twingate:index/twingateServiceAccountKey:TwingateServiceAccountKey';

    /**
     * Returns true if the given object is an instance of TwingateServiceAccountKey.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is TwingateServiceAccountKey {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === TwingateServiceAccountKey.__pulumiType;
    }

    /**
     * The name of the Service Key
     */
    public readonly name!: pulumi.Output<string>;
    /**
     * The id of the Service Account
     */
    public readonly serviceAccountId!: pulumi.Output<string>;
    /**
     * Autogenerated Service Key token. Used to configure a Twingate Client running in headless mode.
     */
    public /*out*/ readonly token!: pulumi.Output<string>;

    /**
     * Create a TwingateServiceAccountKey resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: TwingateServiceAccountKeyArgs, opts?: pulumi.CustomResourceOptions)
    constructor(name: string, argsOrState?: TwingateServiceAccountKeyArgs | TwingateServiceAccountKeyState, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (opts.id) {
            const state = argsOrState as TwingateServiceAccountKeyState | undefined;
            resourceInputs["name"] = state ? state.name : undefined;
            resourceInputs["serviceAccountId"] = state ? state.serviceAccountId : undefined;
            resourceInputs["token"] = state ? state.token : undefined;
        } else {
            const args = argsOrState as TwingateServiceAccountKeyArgs | undefined;
            if ((!args || args.serviceAccountId === undefined) && !opts.urn) {
                throw new Error("Missing required property 'serviceAccountId'");
            }
            resourceInputs["name"] = args ? args.name : undefined;
            resourceInputs["serviceAccountId"] = args ? args.serviceAccountId : undefined;
            resourceInputs["token"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        const secretOpts = { additionalSecretOutputs: ["token"] };
        opts = pulumi.mergeOptions(opts, secretOpts);
        super(TwingateServiceAccountKey.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * Input properties used for looking up and filtering TwingateServiceAccountKey resources.
 */
export interface TwingateServiceAccountKeyState {
    /**
     * The name of the Service Key
     */
    name?: pulumi.Input<string>;
    /**
     * The id of the Service Account
     */
    serviceAccountId?: pulumi.Input<string>;
    /**
     * Autogenerated Service Key token. Used to configure a Twingate Client running in headless mode.
     */
    token?: pulumi.Input<string>;
}

/**
 * The set of arguments for constructing a TwingateServiceAccountKey resource.
 */
export interface TwingateServiceAccountKeyArgs {
    /**
     * The name of the Service Key
     */
    name?: pulumi.Input<string>;
    /**
     * The id of the Service Account
     */
    serviceAccountId: pulumi.Input<string>;
}
