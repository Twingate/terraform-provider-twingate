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
    public sealed class GetTwingateUsersUserResult
    {
        /// <summary>
        /// The email address of the User
        /// </summary>
        public readonly string Email;
        /// <summary>
        /// The first name of the User
        /// </summary>
        public readonly string FirstName;
        /// <summary>
        /// The ID of the User
        /// </summary>
        public readonly string Id;
        /// <summary>
        /// Indicates whether the User is an admin
        /// </summary>
        public readonly bool IsAdmin;
        /// <summary>
        /// The last name of the User
        /// </summary>
        public readonly string LastName;
        /// <summary>
        /// Indicates the User's role. Either ADMIN, DEVOPS, SUPPORT, or MEMBER.
        /// </summary>
        public readonly string Role;

        [OutputConstructor]
        private GetTwingateUsersUserResult(
            string email,

            string firstName,

            string id,

            bool isAdmin,

            string lastName,

            string role)
        {
            Email = email;
            FirstName = firstName;
            Id = id;
            IsAdmin = isAdmin;
            LastName = lastName;
            Role = role;
        }
    }
}
