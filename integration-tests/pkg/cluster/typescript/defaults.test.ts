import { readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

import { describe, expect, test } from "vitest";
import { parse } from "yaml";

import clusterSchema from "../schema.json" with { type: "json" };

import { validateCluster } from "./validation.js";

type SchemaNode = { default?: unknown; properties?: Record<string, SchemaNode> } & Record<
  string,
  unknown
>;

function loadSchemaDefaults() {
  const props = (clusterSchema as { properties?: Record<string, SchemaNode> }).properties ?? {};
  const machineDefaults =
    props.machineDefaults?.properties?.hcloud?.properties ?? ({} as Record<string, SchemaNode>);
  const machineProps =
    (props.machines?.items as { properties?: Record<string, SchemaNode> } | undefined)
      ?.properties ?? {};

  return {
    kubernetesVersion: props.kubernetesVersion?.default,
    talosImage: machineProps.talosImage?.default,
    hcloudServerType: machineDefaults.serverType?.default,
    hcloudDatacenter: machineDefaults.datacenter?.default,
  };
}

const __dirname = dirname(fileURLToPath(import.meta.url));
const fixturesDir = join(__dirname, "..", "fixtures");

const loadFixtureObject = (name: string) =>
  (parse(readFileSync(join(fixturesDir, name), "utf8")) ?? {}) as Record<string, unknown>;

describe("defaults", () => {
  test("schema defaults are present", () => {
    const defaults = loadSchemaDefaults();

    expect(defaults.kubernetesVersion).toBeDefined();
    expect(defaults.talosImage).toBeDefined();
    expect(defaults.hcloudServerType).toBeDefined();
    expect(defaults.hcloudDatacenter).toBeDefined();
  });

  test("applies schema defaults when optional fields are omitted", () => {
    const spec = loadFixtureObject("load-defaults.yaml");

    const defaults = loadSchemaDefaults();

    expect(() => validateCluster(spec)).not.toThrowError();
    expect(spec).toMatchObject({
      kubernetesVersion: defaults.kubernetesVersion,
    });

    const machine = (spec["machines"] as Record<string, unknown>[])[0];
    expect(machine).toMatchObject({
      talosImage: defaults.talosImage,
      hcloud: {
        serverType: defaults.hcloudServerType,
        datacenter: defaults.hcloudDatacenter,
      },
    });
  });
});
