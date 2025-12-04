import AjvModule, { type ErrorObject } from "ajv";

import clusterSchema from "../schema.json" with { type: "json" }; // ESM requires JSON imports to declare type

const Ajv = AjvModule.default; // NodeNext: Ajv constructor lives on default export
const ajv = new Ajv({ allErrors: true, strict: true });
const validateSchema = ajv.compile(clusterSchema);

export function validateCluster(spec: Record<string, unknown>): void {
  if (!validateSchema(spec)) {
    throw new Error(formatAjvError(validateSchema.errors?.[0]));
  }

  const usePrivateNetwork = Boolean((spec as { usePrivateNetwork?: boolean }).usePrivateNetwork);
  let validationCidr = "";

  const machineSpecs = spec["machines"] as Record<string, unknown>[];

  if (usePrivateNetwork) {
    const privateNetwork =
      typeof spec["privateNetwork"] === "string" ? spec["privateNetwork"].trim() : "";
    const privateSubnetwork =
      typeof spec["privateSubnetwork"] === "string" ? spec["privateSubnetwork"].trim() : "";

    if (!privateNetwork || !privateSubnetwork) {
      throw new Error(
        "When 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required"
      );
    }

    validationCidr = privateSubnetwork;
  }

  (machineSpecs as Record<string, unknown>[]).forEach((raw) =>
    validateMachine(raw, validationCidr)
  );
}

function validateMachine(
  raw: Record<string, unknown>,
  validationCidr: string
): void {
  const id = raw["id"] as string;
  const privateIP = raw["privateIP"] as string | undefined;

  if (validationCidr !== "") {
    if (!privateIP) {
      throw new Error(
        `Invalid cluster spec: machine '${id}' must define privateIP when usePrivateNetwork is true`
      );
    }
    assertIpWithinNetwork(privateIP, validationCidr, id);
  }
}

function assertIpWithinNetwork(ip: string, cidr: string, machineId: string): void {
  const { network, prefix } = parseCIDR(cidr);
  const ipValue = parseIPv4(ip);
  const size = 2 ** (32 - prefix);
  const start = network;
  const end = network + size - 1;

  if (ipValue < start || ipValue > end) {
    throw new Error(
      `Invalid cluster spec: machine '${machineId}' privateIP '${ip}' must be inside '${cidr}'`
    );
  }
}

// Parse dotted decimal IPv4 string into a numeric representation for comparisons.
function parseIPv4(address: string): number {
  const octets = address.split(".").map((part) => Number(part));
  if (
    octets.length !== 4 ||
    octets.some(
      (part) => Number.isNaN(part) || !Number.isInteger(part) || part < 0 || part > 255
    )
  ) {
    throw new Error(`Invalid cluster spec: '${address}' is not a valid IPv4 address`);
  }
  return (
    ((octets[0] << 24) | (octets[1] << 16) | (octets[2] << 8) | octets[3]) >>> 0
  );
}

function parseCIDR(cidr: string): { network: number; prefix: number } {
  const [address, prefixText] = cidr.split("/");
  if (!address || prefixText === undefined) {
    throw new Error(`Invalid cluster spec: '${cidr}' is not a valid CIDR`);
  }

  const prefix = Number(prefixText);
  if (!Number.isInteger(prefix) || prefix < 0 || prefix > 32) {
    throw new Error(`Invalid cluster spec: '${cidr}' is not a valid CIDR`);
  }

  const ip = parseIPv4(address);
  const mask = prefix === 0 ? 0 : (~((1 << (32 - prefix)) - 1)) >>> 0;

  return {
    network: ip & mask,
    prefix,
  };
}


// Normalize Ajv errors into readable, domain-specific messages.
function formatAjvError(error?: ErrorObject): string {
  if (!error) {
    return "Invalid cluster spec: unknown validation error";
  }

  if (error.keyword === "required") {
    const missing = (error.params as { missingProperty: string }).missingProperty;
    if (missing === "machines") {
      return "Invalid cluster spec: 'machines' must be a non-empty array";
    }
    const path = formatPath(error.instancePath, missing);
    return `Invalid cluster spec: '${path}' is a required string`;
  }

  if (error.keyword === "additionalProperties") {
    const prop = (error.params as { additionalProperty: string }).additionalProperty;
    const path = formatPath(error.instancePath, prop);
    return `Invalid cluster spec: unknown field '${path}' is not allowed`;
  }

  if (error.keyword === "minItems" && error.instancePath === "/machines") {
    return "Invalid cluster spec: 'machines' must be a non-empty array";
  }

  if (error.keyword === "enum" && error.instancePath.endsWith("/platform")) {
    const path = formatPath(error.instancePath);
    return `Invalid cluster spec: '${path}' must be 'hcloud'`;
  }

  const path = formatPath(error.instancePath);
  return `Invalid cluster spec: ${path ? `'${path}' ` : ""}${error.message ?? "is invalid"}`;
}

function formatPath(instancePath: string, missing?: string): string {
  const segments = instancePath.split("/").filter(Boolean);
  if (missing) {
    segments.push(missing);
  }

  return segments
    .map((segment, idx) => {
      const asNumber = Number(segment);
      if (!Number.isNaN(asNumber) && Number.isInteger(asNumber)) {
        return `[${segment}]`;
      }
      return idx === 0 ? segment : `.${segment}`;
    })
    .join("");
}
