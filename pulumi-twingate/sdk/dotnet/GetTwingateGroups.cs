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
    public static class GetTwingateGroups
    {
        /// <summary>
        /// Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).
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
        ///     var foo = Twingate.GetTwingateGroups.Invoke(new()
        ///     {
        ///         Name = "&lt;your group's name&gt;",
        ///     });
        /// 
        /// });
        /// ```
        /// {{% /example %}}
        /// {{% /examples %}}
        /// </summary>
        public static Task<GetTwingateGroupsResult> InvokeAsync(GetTwingateGroupsArgs? args = null, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.InvokeAsync<GetTwingateGroupsResult>("twingate:index/getTwingateGroups:getTwingateGroups", args ?? new GetTwingateGroupsArgs(), options.WithDefaults());

        /// <summary>
        /// Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).
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
        ///     var foo = Twingate.GetTwingateGroups.Invoke(new()
        ///     {
        ///         Name = "&lt;your group's name&gt;",
        ///     });
        /// 
        /// });
        /// ```
        /// {{% /example %}}
        /// {{% /examples %}}
        /// </summary>
        public static Output<GetTwingateGroupsResult> Invoke(GetTwingateGroupsInvokeArgs? args = null, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.Invoke<GetTwingateGroupsResult>("twingate:index/getTwingateGroups:getTwingateGroups", args ?? new GetTwingateGroupsInvokeArgs(), options.WithDefaults());
    }


    public sealed class GetTwingateGroupsArgs : global::Pulumi.InvokeArgs
    {
        [Input("groups")]
        private List<Inputs.GetTwingateGroupsGroupArgs>? _groups;

        /// <summary>
        /// List of Groups
        /// </summary>
        public List<Inputs.GetTwingateGroupsGroupArgs> Groups
        {
            get => _groups ?? (_groups = new List<Inputs.GetTwingateGroupsGroupArgs>());
            set => _groups = value;
        }

        /// <summary>
        /// Returns only Groups matching the specified state.
        /// </summary>
        [Input("isActive")]
        public bool? IsActive { get; set; }

        /// <summary>
        /// Returns only Groups that exactly match this name.
        /// </summary>
        [Input("name")]
        public string? Name { get; set; }

        /// <summary>
        /// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
        /// </summary>
        [Input("type")]
        public string? Type { get; set; }

        public GetTwingateGroupsArgs()
        {
        }
        public static new GetTwingateGroupsArgs Empty => new GetTwingateGroupsArgs();
    }

    public sealed class GetTwingateGroupsInvokeArgs : global::Pulumi.InvokeArgs
    {
        [Input("groups")]
        private InputList<Inputs.GetTwingateGroupsGroupInputArgs>? _groups;

        /// <summary>
        /// List of Groups
        /// </summary>
        public InputList<Inputs.GetTwingateGroupsGroupInputArgs> Groups
        {
            get => _groups ?? (_groups = new InputList<Inputs.GetTwingateGroupsGroupInputArgs>());
            set => _groups = value;
        }

        /// <summary>
        /// Returns only Groups matching the specified state.
        /// </summary>
        [Input("isActive")]
        public Input<bool>? IsActive { get; set; }

        /// <summary>
        /// Returns only Groups that exactly match this name.
        /// </summary>
        [Input("name")]
        public Input<string>? Name { get; set; }

        /// <summary>
        /// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
        /// </summary>
        [Input("type")]
        public Input<string>? Type { get; set; }

        public GetTwingateGroupsInvokeArgs()
        {
        }
        public static new GetTwingateGroupsInvokeArgs Empty => new GetTwingateGroupsInvokeArgs();
    }


    [OutputType]
    public sealed class GetTwingateGroupsResult
    {
        /// <summary>
        /// List of Groups
        /// </summary>
        public readonly ImmutableArray<Outputs.GetTwingateGroupsGroupResult> Groups;
        /// <summary>
        /// The provider-assigned unique ID for this managed resource.
        /// </summary>
        public readonly string Id;
        /// <summary>
        /// Returns only Groups matching the specified state.
        /// </summary>
        public readonly bool? IsActive;
        /// <summary>
        /// Returns only Groups that exactly match this name.
        /// </summary>
        public readonly string? Name;
        /// <summary>
        /// Returns only Groups of the specified type (valid: `MANUAL`, `SYNCED`, `SYSTEM`).
        /// </summary>
        public readonly string? Type;

        [OutputConstructor]
        private GetTwingateGroupsResult(
            ImmutableArray<Outputs.GetTwingateGroupsGroupResult> groups,

            string id,

            bool? isActive,

            string? name,

            string? type)
        {
            Groups = groups;
            Id = id;
            IsActive = isActive;
            Name = name;
            Type = type;
        }
    }
}
