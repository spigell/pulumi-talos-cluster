import * as pulumi from "@pulumi/pulumi";
import type { Cluster as ClusterSpec } from "pulumi-talos-cluster-integration-tests-infra/pkg/cluster/typescript/spec.js";

export type DeployedServer = {
  id: string;
  ip: pulumi.Output<string>;
};

export type Cluster = ClusterSpec;
