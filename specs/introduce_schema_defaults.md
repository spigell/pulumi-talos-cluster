# Introduction of Schema Defaults

## Rationale
Currently, the integration test schema (`integration-tests/pkg/cluster/schema.json`) does not define default values for optional fields. This forces the loading logic in each language (Go, Python, TypeScript) to manually implement default fallback values, leading to potential inconsistencies and code duplication. We will keep defaults narrowly scoped to version pins and the base hcloud shape to avoid expanding the implicit contract.

By introducing `default` fields directly into the JSON schema, we establish a single source of truth for these values.

## Proposed Changes

### 1. Update `integration-tests/pkg/cluster/schema.json`
Add tightly scoped defaults:
*   `kubernetesVersion`: `"v1.33.0"`
*   `machines[].talosImage`: `"ghcr.io/siderolabs/talos:v1.11.5"`
*   `machineDefaults.hcloud.serverType`: `"cx21"`
*   `machineDefaults.hcloud.datacenter`: `"nbg1-dc3"`

```json
{
  ...
  "properties": {
    ...
    "machineDefaults": {
      "type": "object",
      "properties": {
        "hcloud": {
          "type": "object",
          "properties": {
            "serverType": { "type": "string", "default": "cx21" },
            "datacenter": { "type": "string", "default": "nbg1-dc3" }
          }
        }
      }
    },
    "kubernetesVersion": { "type": "string", "default": "v1.33.0" },
    "machines": {
      "items": {
        "properties": {
          "talosImage": { "type": "string", "default": "ghcr.io/siderolabs/talos:v1.11.5" }
        }
      }
    },
    ...
  }
}
```

### 2. Update Loading Logic

#### TypeScript
Update `integration-tests/pkg/cluster/typescript/validation.ts` to enable `useDefaults: true` in the Ajv configuration. This allows Ajv to automatically inject the default values into the data object during validation.

#### Python
Extend the `Draft7Validator` to automatically inject defaults during validation, similar to how Ajv works in TypeScript.

```python
from jsonschema import Draft7Validator, validators

def extend_with_default(validator_class):
    validate_properties = validator_class.VALIDATORS["properties"]

    def set_defaults(validator, properties, instance, schema):
        for property, subschema in properties.items():
            if "default" in subschema:
                instance.setdefault(property, subschema["default"])

        for error in validate_properties(
            validator, properties, instance, schema,
        ):
            yield error

    return validators.extend(
        validator_class, {"properties": set_defaults},
    )

DefaultValidatingDraft7Validator = extend_with_default(Draft7Validator)
```

#### Go
The `santhosh-tekuri/jsonschema` library does not automatically inject defaults during validation. However, it exposes the `Default` field in the compiled `Schema` struct. We will implement a helper function to manually inject these defaults into the map before validation.

```go
func applyDefaults(s *jsonschema.Schema, m map[string]any) {
	for name, prop := range s.Properties {
		if prop.Default != nil {
			if _, ok := m[name]; !ok {
				m[name] = prop.Default
			}
		}
	}
}
```

This function will be called in `validateCluster` immediately after loading the schema.

## Benefits
*   **Single Source of Truth**: Defaults are defined in one place (the schema).
*   **Documentation**: The schema clearly indicates default behaviors to users/developers.
*   **Consistency**: Reduces the risk of different languages using different default values.
