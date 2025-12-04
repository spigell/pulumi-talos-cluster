import { readFileSync } from "fs";
import { join } from "path";
import { fileURLToPath } from "url";
import { parse } from "yaml";
import { describe, expect, test } from "vitest";

import { validateCluster } from "./validation.js";

const fixturesDir = join(
  fileURLToPath(new URL(".", import.meta.url)),
  "..",
  "fixtures"
);
const fixture = (name: string) => join(fixturesDir, name);
const loadFixtureObject = (name: string) =>
  (parse(readFileSync(fixture(name), "utf8")) ?? {}) as Record<string, unknown>;

describe("validateCluster", () => {
  test("loads valid fixture without errors", () => {
    expect(() => validateCluster(loadFixtureObject("load-valid.yaml"))).not.toThrowError();
  });

  test("allows anchors field for templates", () => {
    expect(() => validateCluster(loadFixtureObject("validation-anchors.yaml"))).not.toThrowError();
  });

  test("loads minimal fixture without errors", () => {
    expect(() => validateCluster(loadFixtureObject("load-minimal.yaml"))).not.toThrowError();
  });
  
  test("throws when required top-level fields are missing", () => {
    expect(() => validateCluster(loadFixtureObject("validation-missing-name.yaml"))).toThrowError(
      "Invalid cluster spec: 'name' is a required string"
    );
  });

  test("throws when machines is missing or empty", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-missing-machines.yaml"))
    ).toThrowError("Invalid cluster spec: 'machines' must be a non-empty array");
    expect(() =>
      validateCluster(loadFixtureObject("validation-empty-machines.yaml"))
    ).toThrowError(
      "Invalid cluster spec: 'machines' must be a non-empty array"
    );
  });

  test("throws when machine id or type is missing", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-missing-id.yaml"))
    ).toThrowError("Invalid cluster spec: 'machines[0].id' is a required string");

    expect(() =>
      validateCluster(loadFixtureObject("validation-missing-type.yaml"))
    ).toThrowError("Invalid cluster spec: 'machines[0].type' is a required string");
  });

  test("throws when platform is missing or unsupported", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-missing-platform.yaml"))
    ).toThrowError("Invalid cluster spec: 'machines[0].platform' is a required string");

    expect(() =>
      validateCluster(loadFixtureObject("validation-unsupported-platform.yaml"))
    ).toThrowError("Invalid cluster spec: 'machines[0].platform' must be 'hcloud'");
  });

  test("throws when networks are missing or only one is provided", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-missing-networks.yaml"))
    ).toThrowError(
      "When 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"
    );
    expect(() =>
      validateCluster(loadFixtureObject("validation-single-network.yaml"))
    ).toThrowError(
      "When 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"
    );
  });

  test("throws when machine ip is outside the provided network (both present)", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-ip-outside.yaml"))
    ).toThrowError(
      "Invalid cluster spec: machine 'worker-1' privateIP '10.0.1.10' must be inside '10.0.0.0/24'"
    );
  });

  test("passes when both networks are present and IPs are in range", () => {
    expect(() => validateCluster(loadFixtureObject("validation-networks-present.yaml"))).not.toThrow();
  });

  test("throws for unknown top-level or machine fields", () => {
    expect(() =>
      validateCluster(loadFixtureObject("validation-unknown-top.yaml"))
    ).toThrowError("Invalid cluster spec: unknown field 'extra' is not allowed");

    expect(() =>
      validateCluster(loadFixtureObject("validation-unknown-machine.yaml"))
    ).toThrowError("Invalid cluster spec: unknown field 'machines[0].unknown' is not allowed");
  });
});
