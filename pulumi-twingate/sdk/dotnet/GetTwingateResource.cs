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
    public static class GetTwingateResource
    {
        /// <summary>
        /// Resources in Twingate represent any network destination address that you wish to provide private access to for users authorized via the Twingate Client application. Resources can be defined by either IP or DNS address, and all private DNS addresses will be automatically resolved with no client configuration changes. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).
        /// 
        /// {{% examples %}}
        /// ## Example Usage
        /// {{% example %}}
        /// 
        /// ```csharp
        /// using System.Collections.Generic;
        /// using Pulumi;
        /// using Twingate = Pulumi.Twingate;
        /// 
        /// return await Deployment.RunAsync(() =&gt; 
        /// {
        ///     var foo = Twingate.GetTwingateResource.Invoke(new()
        ///     {
        ///         Id = "&lt;your resource's id&gt;",
        ///     });
        /// 
        /// });
        /// ```
        /// {{% /example %}}
        /// {{% /examples %}}
        /// </summary>
        public static Task<GetTwingateResourceResult> InvokeAsync(GetTwingateResourceArgs args, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.InvokeAsync<GetTwingateResourceResult>("twingate:index/getTwingateResource:getTwingateResource", args ?? new GetTwingateResourceArgs(), options.WithDefaults());

        /// <summary>
        /// Resources in Twingate represent any network destination address that you wish to provide private access to for users authorized via the Twingate Client application. Resources can be defined by either IP or DNS address, and all private DNS addresses will be automatically resolved with no client configuration changes. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).
        /// 
        /// {{% examples %}}
        /// ## Example Usage
        /// {{% example %}}
        /// 
        /// ```csharp
        /// using System.Collections.Generic;
        /// using Pulumi;
        /// using Twingate = Pulumi.Twingate;
        /// 
        /// return await Deployment.RunAsync(() =&gt; 
        /// {
        ///     var foo = Twingate.GetTwingateResource.Invoke(new()
        ///     {
        ///         Id = "&lt;your resource's id&gt;",
        ///     });
        /// 
        /// });
        /// ```
        /// {{% /example %}}
        /// {{% /examples %}}
        /// </summary>
        public static Output<GetTwingateResourceResult> Invoke(GetTwingateResourceInvokeArgs args, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.Invoke<GetTwingateResourceResult>("twingate:index/getTwingateResource:getTwingateResource", args ?? new GetTwingateResourceInvokeArgs(), options.WithDefaults());
    }


    public sealed class GetTwingateResourceArgs : global::Pulumi.InvokeArgs
    {
        /// <summary>
        /// The ID of the Resource. The ID for the Resource must be obtained from the Admin API.
        /// </summary>
        [Input("id", required: true)]
        public string Id { get; set; } = null!;

        [Input("protocols")]
        private List<Inputs.GetTwingateResourceProtocolArgs>? _protocols;

        /// <summary>
        /// By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.
        /// </summary>
        public List<Inputs.GetTwingateResourceProtocolArgs> Protocols
        {
            get => _protocols ?? (_protocols = new List<Inputs.GetTwingateResourceProtocolArgs>());
            set => _protocols = value;
        }

        public GetTwingateResourceArgs()
        {
        }
        public static new GetTwingateResourceArgs Empty => new GetTwingateResourceArgs();
    }

    public sealed class GetTwingateResourceInvokeArgs : global::Pulumi.InvokeArgs
    {
        /// <summary>
        /// The ID of the Resource. The ID for the Resource must be obtained from the Admin API.
        /// </summary>
        [Input("id", required: true)]
        public Input<string> Id { get; set; } = null!;

        [Input("protocols")]
        private InputList<Inputs.GetTwingateResourceProtocolInputArgs>? _protocols;

        /// <summary>
        /// By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.
        /// </summary>
        public InputList<Inputs.GetTwingateResourceProtocolInputArgs> Protocols
        {
            get => _protocols ?? (_protocols = new InputList<Inputs.GetTwingateResourceProtocolInputArgs>());
            set => _protocols = value;
        }

        public GetTwingateResourceInvokeArgs()
        {
        }
        public static new GetTwingateResourceInvokeArgs Empty => new GetTwingateResourceInvokeArgs();
    }


    [OutputType]
    public sealed class GetTwingateResourceResult
    {
        /// <summary>
        /// The Resource's address, which may be an IP address, CIDR range, or DNS address
        /// </summary>
        public readonly string Address;
        /// <summary>
        /// The ID of the Resource. The ID for the Resource must be obtained from the Admin API.
        /// </summary>
        public readonly string Id;
        /// <summary>
        /// The name of the Resource
        /// </summary>
        public readonly string Name;
        /// <summary>
        /// By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.
        /// </summary>
        public readonly ImmutableArray<Outputs.GetTwingateResourceProtocolResult> Protocols;
        /// <summary>
        /// The Remote Network ID that the Resource is associated with. Resources may only be associated with a single Remote Network.
        /// </summary>
        public readonly string RemoteNetworkId;

        [OutputConstructor]
        private GetTwingateResourceResult(
            string address,

            string id,

            string name,

            ImmutableArray<Outputs.GetTwingateResourceProtocolResult> protocols,

            string remoteNetworkId)
        {
            Address = address;
            Id = id;
            Name = name;
            Protocols = protocols;
            RemoteNetworkId = remoteNetworkId;
        }
    }
}
