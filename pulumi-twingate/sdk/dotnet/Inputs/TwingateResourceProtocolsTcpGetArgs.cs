// *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace TwingateLabs.Twingate.Inputs
{

    public sealed class TwingateResourceProtocolsTcpGetArgs : global::Pulumi.ResourceArgs
    {
        [Input("policy", required: true)]
        public Input<string> Policy { get; set; } = null!;

        [Input("ports")]
        private InputList<string>? _ports;
        public InputList<string> Ports
        {
            get => _ports ?? (_ports = new InputList<string>());
            set => _ports = value;
        }

        public TwingateResourceProtocolsTcpGetArgs()
        {
        }
        public static new TwingateResourceProtocolsTcpGetArgs Empty => new TwingateResourceProtocolsTcpGetArgs();
    }
}
