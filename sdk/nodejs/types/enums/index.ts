// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***


export const MachineTypes = {
    Controlplane: "controlplane",
    Worker: "worker",
} as const;

/**
 * Allowed types for machines: controlplane or worker
 */
export type MachineTypes = (typeof MachineTypes)[keyof typeof MachineTypes];
