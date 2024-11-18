// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.TalosCluster.Outputs
{

    [OutputType]
    public sealed class Credentials
    {
        /// <summary>
        /// The Kubeconfig for cluster
        /// </summary>
        public readonly string Kubeconfig;
        /// <summary>
        /// The talosconfig with all nodes and controlplanes as endpoints
        /// </summary>
        public readonly string Talosconfig;

        [OutputConstructor]
        private Credentials(
            string kubeconfig,

            string talosconfig)
        {
            Kubeconfig = kubeconfig;
            Talosconfig = talosconfig;
        }
    }
}