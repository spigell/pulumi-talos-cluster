// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as enums from "./types/enums";
import * as utilities from "./utilities";

/**
 * Initialize a new Talos cluster:
 * - Creates secrets
 * - Generates machine configurations for all nodes
 */
export class Cluster extends pulumi.ComponentResource {
    /** @internal */
    public static readonly __pulumiType = 'talos-cluster:index:Cluster';

    /**
     * Returns true if the given object is an instance of Cluster.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Cluster {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Cluster.__pulumiType;
    }

    /**
     * Client configuration for bootstrapping and applying resources.
     */
    public /*out*/ readonly clientConfiguration!: pulumi.Output<outputs.ClientConfiguration>;
    /**
     * TO DO
     */
    public /*out*/ readonly generatedConfigurations!: pulumi.Output<{[key: string]: string}>;
    /**
     * TO DO
     */
    public /*out*/ readonly machines!: pulumi.Output<outputs.ApplyMachines>;

    /**
     * Create a Cluster resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ClusterArgs, opts?: pulumi.ComponentResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.clusterEndpoint === undefined) && !opts.urn) {
                throw new Error("Missing required property 'clusterEndpoint'");
            }
            if ((!args || args.clusterMachines === undefined) && !opts.urn) {
                throw new Error("Missing required property 'clusterMachines'");
            }
            if ((!args || args.clusterName === undefined) && !opts.urn) {
                throw new Error("Missing required property 'clusterName'");
            }
            resourceInputs["clusterEndpoint"] = args ? args.clusterEndpoint : undefined;
            resourceInputs["clusterMachines"] = args ? args.clusterMachines : undefined;
            resourceInputs["clusterName"] = args ? args.clusterName : undefined;
            resourceInputs["kubernetesVersion"] = (args ? args.kubernetesVersion : undefined) ?? "v1.31.0";
            resourceInputs["talosVersionContract"] = (args ? args.talosVersionContract : undefined) ?? "v1.8.2";
            resourceInputs["clientConfiguration"] = undefined /*out*/;
            resourceInputs["generatedConfigurations"] = undefined /*out*/;
            resourceInputs["machines"] = undefined /*out*/;
        } else {
            resourceInputs["clientConfiguration"] = undefined /*out*/;
            resourceInputs["generatedConfigurations"] = undefined /*out*/;
            resourceInputs["machines"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(Cluster.__pulumiType, name, resourceInputs, opts, true /*remote*/);
    }
}

/**
 * The set of arguments for constructing a Cluster resource.
 */
export interface ClusterArgs {
    /**
     * Cluster endpoint, the Kubernetes API endpoint accessible by all nodes
     */
    clusterEndpoint: pulumi.Input<string>;
    /**
     * Configuration settings for machines
     */
    clusterMachines: pulumi.Input<pulumi.Input<inputs.ClusterMachinesArgs>[]>;
    /**
     * Name of the cluster
     */
    clusterName: string;
    /**
     * Kubernetes version to install. 
     * Default is v1.31.0.
     */
    kubernetesVersion?: pulumi.Input<string>;
    /**
     * Version of Talos features used for configuration generation. 
     * Do not confuse this with the talosImage property. 
     * Used in NewSecrets() and GetConfigurationOutput() resources. 
     * This property is immutable to prevent version conflicts across provider updates. 
     * See issue: https://github.com/siderolabs/terraform-provider-talos/issues/168 
     * The default value is based on gendata.VersionTag, current: v1.8.2.
     */
    talosVersionContract?: pulumi.Input<string>;
}
