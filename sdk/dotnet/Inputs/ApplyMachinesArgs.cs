// *** WARNING: this file was generated by pulumi-language-dotnet. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.TalosCluster.Inputs
{

    public sealed class ApplyMachinesArgs : global::Pulumi.ResourceArgs
    {
        [Input("controlplane")]
        private InputList<Inputs.MachineInfoArgs>? _controlplane;
        public InputList<Inputs.MachineInfoArgs> Controlplane
        {
            get => _controlplane ?? (_controlplane = new InputList<Inputs.MachineInfoArgs>());
            set => _controlplane = value;
        }

        [Input("init", required: true)]
        private InputList<Inputs.MachineInfoArgs>? _init;
        public InputList<Inputs.MachineInfoArgs> Init
        {
            get => _init ?? (_init = new InputList<Inputs.MachineInfoArgs>());
            set => _init = value;
        }

        [Input("worker")]
        private InputList<Inputs.MachineInfoArgs>? _worker;
        public InputList<Inputs.MachineInfoArgs> Worker
        {
            get => _worker ?? (_worker = new InputList<Inputs.MachineInfoArgs>());
            set => _worker = value;
        }

        public ApplyMachinesArgs()
        {
        }
        public static new ApplyMachinesArgs Empty => new ApplyMachinesArgs();
    }
}
