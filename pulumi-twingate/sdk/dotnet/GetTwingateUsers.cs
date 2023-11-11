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
    public static class GetTwingateUsers
    {
        public static Task<GetTwingateUsersResult> InvokeAsync(GetTwingateUsersArgs? args = null, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.InvokeAsync<GetTwingateUsersResult>("twingate:index/getTwingateUsers:getTwingateUsers", args ?? new GetTwingateUsersArgs(), options.WithDefaults());

        public static Output<GetTwingateUsersResult> Invoke(GetTwingateUsersInvokeArgs? args = null, InvokeOptions? options = null)
            => global::Pulumi.Deployment.Instance.Invoke<GetTwingateUsersResult>("twingate:index/getTwingateUsers:getTwingateUsers", args ?? new GetTwingateUsersInvokeArgs(), options.WithDefaults());
    }


    public sealed class GetTwingateUsersArgs : global::Pulumi.InvokeArgs
    {
        [Input("users")]
        private List<Inputs.GetTwingateUsersUserArgs>? _users;
        public List<Inputs.GetTwingateUsersUserArgs> Users
        {
            get => _users ?? (_users = new List<Inputs.GetTwingateUsersUserArgs>());
            set => _users = value;
        }

        public GetTwingateUsersArgs()
        {
        }
        public static new GetTwingateUsersArgs Empty => new GetTwingateUsersArgs();
    }

    public sealed class GetTwingateUsersInvokeArgs : global::Pulumi.InvokeArgs
    {
        [Input("users")]
        private InputList<Inputs.GetTwingateUsersUserInputArgs>? _users;
        public InputList<Inputs.GetTwingateUsersUserInputArgs> Users
        {
            get => _users ?? (_users = new InputList<Inputs.GetTwingateUsersUserInputArgs>());
            set => _users = value;
        }

        public GetTwingateUsersInvokeArgs()
        {
        }
        public static new GetTwingateUsersInvokeArgs Empty => new GetTwingateUsersInvokeArgs();
    }


    [OutputType]
    public sealed class GetTwingateUsersResult
    {
        /// <summary>
        /// The provider-assigned unique ID for this managed resource.
        /// </summary>
        public readonly string Id;
        public readonly ImmutableArray<Outputs.GetTwingateUsersUserResult> Users;

        [OutputConstructor]
        private GetTwingateUsersResult(
            string id,

            ImmutableArray<Outputs.GetTwingateUsersUserResult> users)
        {
            Id = id;
            Users = users;
        }
    }
}
