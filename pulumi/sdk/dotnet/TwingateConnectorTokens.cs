// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace TwingateLabs.Twingate
{
    [TwingateResourceType("twingate:index/twingateConnectorTokens:TwingateConnectorTokens")]
    public partial class TwingateConnectorTokens : global::Pulumi.CustomResource
    {
        /// <summary>
        /// The Access Token of the parent Connector
        /// </summary>
        [Output("accessToken")]
        public Output<string> AccessToken { get; private set; } = null!;

        /// <summary>
        /// The ID of the parent Connector
        /// </summary>
        [Output("connectorId")]
        public Output<string> ConnectorId { get; private set; } = null!;

        /// <summary>
        /// Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
        /// Connector tokens on a schedule.
        /// </summary>
        [Output("keepers")]
        public Output<ImmutableDictionary<string, object>?> Keepers { get; private set; } = null!;

        /// <summary>
        /// The Refresh Token of the parent Connector
        /// </summary>
        [Output("refreshToken")]
        public Output<string> RefreshToken { get; private set; } = null!;


        /// <summary>
        /// Create a TwingateConnectorTokens resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public TwingateConnectorTokens(string name, TwingateConnectorTokensArgs args, CustomResourceOptions? options = null)
            : base("twingate:index/twingateConnectorTokens:TwingateConnectorTokens", name, args ?? new TwingateConnectorTokensArgs(), MakeResourceOptions(options, ""))
        {
        }

        private TwingateConnectorTokens(string name, Input<string> id, TwingateConnectorTokensState? state = null, CustomResourceOptions? options = null)
            : base("twingate:index/twingateConnectorTokens:TwingateConnectorTokens", name, state, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                PluginDownloadURL = "https://github.com/Twingate-Labs/pulumi-twingate/releases/download/v${VERSION}",
                AdditionalSecretOutputs =
                {
                    "accessToken",
                    "refreshToken",
                },
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing TwingateConnectorTokens resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="state">Any extra arguments used during the lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static TwingateConnectorTokens Get(string name, Input<string> id, TwingateConnectorTokensState? state = null, CustomResourceOptions? options = null)
        {
            return new TwingateConnectorTokens(name, id, state, options);
        }
    }

    public sealed class TwingateConnectorTokensArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The ID of the parent Connector
        /// </summary>
        [Input("connectorId", required: true)]
        public Input<string> ConnectorId { get; set; } = null!;

        [Input("keepers")]
        private InputMap<object>? _keepers;

        /// <summary>
        /// Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
        /// Connector tokens on a schedule.
        /// </summary>
        public InputMap<object> Keepers
        {
            get => _keepers ?? (_keepers = new InputMap<object>());
            set => _keepers = value;
        }

        public TwingateConnectorTokensArgs()
        {
        }
        public static new TwingateConnectorTokensArgs Empty => new TwingateConnectorTokensArgs();
    }

    public sealed class TwingateConnectorTokensState : global::Pulumi.ResourceArgs
    {
        [Input("accessToken")]
        private Input<string>? _accessToken;

        /// <summary>
        /// The Access Token of the parent Connector
        /// </summary>
        public Input<string>? AccessToken
        {
            get => _accessToken;
            set
            {
                var emptySecret = Output.CreateSecret(0);
                _accessToken = Output.Tuple<Input<string>?, int>(value, emptySecret).Apply(t => t.Item1);
            }
        }

        /// <summary>
        /// The ID of the parent Connector
        /// </summary>
        [Input("connectorId")]
        public Input<string>? ConnectorId { get; set; }

        [Input("keepers")]
        private InputMap<object>? _keepers;

        /// <summary>
        /// Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate
        /// Connector tokens on a schedule.
        /// </summary>
        public InputMap<object> Keepers
        {
            get => _keepers ?? (_keepers = new InputMap<object>());
            set => _keepers = value;
        }

        [Input("refreshToken")]
        private Input<string>? _refreshToken;

        /// <summary>
        /// The Refresh Token of the parent Connector
        /// </summary>
        public Input<string>? RefreshToken
        {
            get => _refreshToken;
            set
            {
                var emptySecret = Output.CreateSecret(0);
                _refreshToken = Output.Tuple<Input<string>?, int>(value, emptySecret).Apply(t => t.Item1);
            }
        }

        public TwingateConnectorTokensState()
        {
        }
        public static new TwingateConnectorTokensState Empty => new TwingateConnectorTokensState();
    }
}
