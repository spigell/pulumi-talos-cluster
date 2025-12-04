import * as pulumi from "@pulumi/pulumi";
import * as talos from "@spigell/pulumi-talos-cluster";
import { Hetzner } from "./hetzner.js";
import { load } from "pulumi-talos-cluster-integration-tests-infra/pkg/cluster/typescript/spec.js";
import path from "node:path";
import { fileURLToPath } from "node:url";

const cluster = load(path.resolve("cluster.yaml"));

const servers = Hetzner(cluster);
const machines: pulumi.Input<
  pulumi.Input<talos.types.input.ClusterMachinesArgs>[]
> = [];

cluster.machines.forEach((v: typeof cluster.machines[number]) =>
  machines.push({
    machineId: v.id,
    nodeIp: servers.find((m) => v.id == m.id)?.ip as pulumi.Input<string>,
    machineType: v.type as talos.MachineTypes,
    configPatches: v.configPatches,
  }),
);

const clu = new talos.Cluster(cluster.name, {
  kubernetesVersion: cluster.kubernetesVersion,
  clusterEndpoint: pulumi.interpolate`https://${servers[0].ip}:6443`,
  clusterName: cluster.name,
  clusterMachines: machines,
});

export const apply = new talos.Apply(cluster.name, {
  clientConfiguration: clu.clientConfiguration,
  applyMachines: clu.machines,
});
