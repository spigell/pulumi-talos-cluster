// *** WARNING: this file was generated by pulumi-language-dotnet. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.TalosCluster.Inputs
{

    public sealed class ClientConfigurationArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// The Certificate Authority (CA) certificate used to verify connections to the Talos API server.
        /// </summary>
        [Input("caCertificate")]
        public Input<string>? CaCertificate { get; set; }

        /// <summary>
        /// The client certificate used to authenticate to the Talos API server.
        /// </summary>
        [Input("clientCertificate")]
        public Input<string>? ClientCertificate { get; set; }

        /// <summary>
        /// The private key for the client certificate, used for authenticating the client to the Talos API server.
        /// </summary>
        [Input("clientKey")]
        public Input<string>? ClientKey { get; set; }

        public ClientConfigurationArgs()
        {
        }
        public static new ClientConfigurationArgs Empty => new ClientConfigurationArgs();
    }
}
