// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export class TwingateConnectorTokens extends pulumi.CustomResource {
    /**
     * Get an existing TwingateConnectorTokens resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param state Any extra arguments used during the lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, state?: TwingateConnectorTokensState, opts?: pulumi.CustomResourceOptions): TwingateConnectorTokens {
        return new TwingateConnectorTokens(name, <any>state, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'twingate:index/twingateConnectorTokens:TwingateConnectorTokens';

    /**
     * Returns true if the given object is an instance of TwingateConnectorTokens.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is TwingateConnectorTokens {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === TwingateConnectorTokens.__pulumiType;
    }

    /**
     * The Access Token of the parent Connector
     */
    public /*out*/ readonly accessToken!: pulumi.Output<string>;
    /**
     * The ID of the parent Connector
     */
    public readonly connectorId!: pulumi.Output<string>;
    /**
     * Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
     * Connector tokens on a schedule.
     */
    public readonly keepers!: pulumi.Output<{[key: string]: any} | undefined>;
    /**
     * The Refresh Token of the parent Connector
     */
    public /*out*/ readonly refreshToken!: pulumi.Output<string>;

    /**
     * Create a TwingateConnectorTokens resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: TwingateConnectorTokensArgs, opts?: pulumi.CustomResourceOptions)
    constructor(name: string, argsOrState?: TwingateConnectorTokensArgs | TwingateConnectorTokensState, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (opts.id) {
            const state = argsOrState as TwingateConnectorTokensState | undefined;
            resourceInputs["accessToken"] = state ? state.accessToken : undefined;
            resourceInputs["connectorId"] = state ? state.connectorId : undefined;
            resourceInputs["keepers"] = state ? state.keepers : undefined;
            resourceInputs["refreshToken"] = state ? state.refreshToken : undefined;
        } else {
            const args = argsOrState as TwingateConnectorTokensArgs | undefined;
            if ((!args || args.connectorId === undefined) && !opts.urn) {
                throw new Error("Missing required property 'connectorId'");
            }
            resourceInputs["connectorId"] = args ? args.connectorId : undefined;
            resourceInputs["keepers"] = args ? args.keepers : undefined;
            resourceInputs["accessToken"] = undefined /*out*/;
            resourceInputs["refreshToken"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        const secretOpts = { additionalSecretOutputs: ["accessToken", "refreshToken"] };
        opts = pulumi.mergeOptions(opts, secretOpts);
        super(TwingateConnectorTokens.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * Input properties used for looking up and filtering TwingateConnectorTokens resources.
 */
export interface TwingateConnectorTokensState {
    /**
     * The Access Token of the parent Connector
     */
    accessToken?: pulumi.Input<string>;
    /**
     * The ID of the parent Connector
     */
    connectorId?: pulumi.Input<string>;
    /**
     * Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
     * Connector tokens on a schedule.
     */
    keepers?: pulumi.Input<{[key: string]: any}>;
    /**
     * The Refresh Token of the parent Connector
     */
    refreshToken?: pulumi.Input<string>;
}

/**
 * The set of arguments for constructing a TwingateConnectorTokens resource.
 */
export interface TwingateConnectorTokensArgs {
    /**
     * The ID of the parent Connector
     */
    connectorId: pulumi.Input<string>;
    /**
     * Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
     * Connector tokens on a schedule.
     */
    keepers?: pulumi.Input<{[key: string]: any}>;
}
