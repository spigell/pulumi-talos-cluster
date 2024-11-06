// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as inputs from "./types/input";
import * as outputs from "./types/output";
import * as enums from "./types/enums";
import * as utilities from "./utilities";

/**
 * Apply config: creates etcd cluster
 */
export class Apply extends pulumi.ComponentResource {
    /** @internal */
    public static readonly __pulumiType = 'talos-cluster:index:Apply';

    /**
     * Returns true if the given object is an instance of Apply.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Apply {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Apply.__pulumiType;
    }


    /**
     * Create a Apply resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ApplyArgs, opts?: pulumi.ComponentResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.applyMachines === undefined) && !opts.urn) {
                throw new Error("Missing required property 'applyMachines'");
            }
            if ((!args || args.clientConfiguration === undefined) && !opts.urn) {
                throw new Error("Missing required property 'clientConfiguration'");
            }
            resourceInputs["applyMachines"] = args ? args.applyMachines : undefined;
            resourceInputs["clientConfiguration"] = args ? args.clientConfiguration : undefined;
        } else {
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(Apply.__pulumiType, name, resourceInputs, opts, true /*remote*/);
    }
}

/**
 * The set of arguments for constructing a Apply resource.
 */
export interface ApplyArgs {
    /**
     * The machine configurations for apply.
     */
    applyMachines: pulumi.Input<pulumi.Input<inputs.ApplyMachinesArgs>[]>;
    /**
     * The client configuration. Can be used for bootstraping and apply
     */
    clientConfiguration: pulumi.Input<inputs.ClientConfigurationArgs>;
}
