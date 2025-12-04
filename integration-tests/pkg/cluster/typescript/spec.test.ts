import { join } from "path";
import { fileURLToPath } from "url";
import { describe, test, expect } from "vitest";

import { load } from "./spec.js";

describe("load", () => {
  const fixturesDir = join(
    fileURLToPath(new URL(".", import.meta.url)),
    "..",
    "fixtures"
  );
  const fixture = (name: string) => join(fixturesDir, name);

  test("throws error for non-existent file", () => {
    expect(() => load("non-existent-file.yaml")).toThrowError(/ENOENT/);
  });

  test("throws error for malformed YAML", () => {
    expect(() => load(fixture("load-malformed.yaml"))).toThrowError();
  });
});
