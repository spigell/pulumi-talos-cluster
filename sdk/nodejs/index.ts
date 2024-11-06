// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

// Export members:
export { ApplyArgs } from "./apply";
export type Apply = import("./apply").Apply;
export const Apply: typeof import("./apply").Apply = null as any;
utilities.lazyLoad(exports, ["Apply"], () => require("./apply"));

export { BootstrapArgs } from "./bootstrap";
export type Bootstrap = import("./bootstrap").Bootstrap;
export const Bootstrap: typeof import("./bootstrap").Bootstrap = null as any;
utilities.lazyLoad(exports, ["Bootstrap"], () => require("./bootstrap"));

export { ClusterArgs } from "./cluster";
export type Cluster = import("./cluster").Cluster;
export const Cluster: typeof import("./cluster").Cluster = null as any;
utilities.lazyLoad(exports, ["Cluster"], () => require("./cluster"));

export { ProviderArgs } from "./provider";
export type Provider = import("./provider").Provider;
export const Provider: typeof import("./provider").Provider = null as any;
utilities.lazyLoad(exports, ["Provider"], () => require("./provider"));


// Export enums:
export * from "./types/enums";

// Export sub-modules:
import * as types from "./types";

export {
    types,
};

const _module = {
    version: utilities.getVersion(),
    construct: (name: string, type: string, urn: string): pulumi.Resource => {
        switch (type) {
            case "talos-cluster:index:Apply":
                return new Apply(name, <any>undefined, { urn })
            case "talos-cluster:index:Bootstrap":
                return new Bootstrap(name, <any>undefined, { urn })
            case "talos-cluster:index:Cluster":
                return new Cluster(name, <any>undefined, { urn })
            default:
                throw new Error(`unknown resource type ${type}`);
        }
    },
};
pulumi.runtime.registerResourceModule("talos-cluster", "index", _module)
pulumi.runtime.registerResourcePackage("talos-cluster", {
    version: utilities.getVersion(),
    constructProvider: (name: string, type: string, urn: string): pulumi.ProviderResource => {
        if (type !== "pulumi:providers:talos-cluster") {
            throw new Error(`unknown provider type ${type}`);
        }
        return new Provider(name, <any>undefined, { urn });
    },
});
