// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace TwingateLabs.Twingate.Outputs
{

    [OutputType]
    public sealed class GetTwingateServiceAccountsServiceAccountResult
    {
        /// <summary>
        /// ID of the Service Account resource
        /// </summary>
        public readonly string Id;
        /// <summary>
        /// List of twingate*service*account_key IDs that are assigned to the Service Account.
        /// </summary>
        public readonly ImmutableArray<string> KeyIds;
        /// <summary>
        /// Name of the Service Account
        /// </summary>
        public readonly string Name;
        /// <summary>
        /// List of twingate.TwingateResource IDs that the Service Account is assigned to.
        /// </summary>
        public readonly ImmutableArray<string> ResourceIds;

        [OutputConstructor]
        private GetTwingateServiceAccountsServiceAccountResult(
            string id,

            ImmutableArray<string> keyIds,

            string name,

            ImmutableArray<string> resourceIds)
        {
            Id = id;
            KeyIds = keyIds;
            Name = name;
            ResourceIds = resourceIds;
        }
    }
}
