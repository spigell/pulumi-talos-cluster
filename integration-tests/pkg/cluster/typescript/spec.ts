import { readFileSync } from "fs";

import { parse } from "yaml";

import { validateCluster } from "./validation.js";

import type { Cluster, Machine } from "./cluster.js";

// These specs are internal
type MachineSpec = Omit<Machine, "configPatches" | "applyConfigViaUserdata"> & {
  "apply-config-via-userdata"?: boolean;
  configPatches?: string[];
};

type ClusterSpec = Omit<Cluster, "machines"> & {
  machines: MachineSpec[];
  kubernetesVersion?: string;
  skipInitApply?: boolean;
};

export function load(path: string): Cluster {
  const data = readFileSync(path, "utf8");
  const spec = (parse(data) ?? {}) as Record<string, unknown>;

  validateCluster(spec);
  const typed = spec as ClusterSpec;

  const machines = typed.machines.map((raw) => normalizeMachine(raw));

  return {
    name: typed.name,
    kubernetesVersion: typed.kubernetesVersion ?? "",
    privateNetwork: typed.privateNetwork ?? "",
    privateSubnetwork: typed.privateSubnetwork ?? "",
    skipInitApply: typed.skipInitApply ?? false,
    machines,
  };
}

function normalizeMachine(raw: MachineSpec): Machine {
  const {
    configPatches,
    userdata,
    "apply-config-via-userdata": applyConfigViaUserdataDashed,
    variant,
    ...rest
  } = raw ?? {};

  const applyConfigViaUserdata =
    typeof applyConfigViaUserdataDashed === "boolean"
      ? applyConfigViaUserdataDashed
      : false;

  return {
    ...rest,
    id: raw.id,
    type: raw.type,
    platform: raw.platform,
    talosInitialVersion: raw.talosInitialVersion,
    talosImage: raw.talosImage,
    variant: (variant as Machine["variant"]) || "metal",
    privateIP: raw.privateIP,
    configPatches: Array.isArray(configPatches) ? configPatches : [],
    userdata,
    applyConfigViaUserdata,
    hcloud: raw.hcloud,
  };
}
