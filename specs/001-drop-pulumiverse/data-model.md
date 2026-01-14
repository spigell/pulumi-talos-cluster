# Data Model

## Entities

### Talosctl Binary (external)
- **Fields**: path (string), version (string), architecture (string), checksum/provenance (string), permissions (executable boolean).
- **Validation**: Version must match documented support matrix; architecture must align with runner; checksum/provenance recommended for supply chain assurance.
- **State**: Provided by operator; not persisted by provider.

### ClusterSecrets Stash
- **Fields**: secretsJSON (secret object from `talosctl gen secrets`), generatedAt (timestamp), talosctlVersion (string).
- **Validation**: Present only once per stack unless replacement is explicitly triggered; secrets must include required bundles for Talos config generation.
- **State**: Stored via `pulumi.Stash` to keep secrets encrypted and stable across runs.

### ClusterConfig Stash
- **Fields**: controlPlaneConfig (YAML/JSON), workerConfig (YAML/JSON), talosconfig (YAML), kubeconfig (YAML), talosctlVersion (string), clusterName (string), endpoint (string).
- **Validation**: Derived from ClusterSecrets; regeneration should not occur unless secrets change; clusterName/endpoint must align with spec input.
- **State**: Stored via `pulumi.Stash` to enable deterministic apply/bootstrap and exports without rerunning talosctl.

### ApplyCommand
- **Fields**: nodeIP (string), role (enum: controlplane|worker), configPayload (YAML from ClusterConfig), insecureFlag (bool for first apply), talosconfigPath (temp path), retries/backoff (int/duration).
- **Validation**: nodeIP must match cluster spec; configPayload sourced from stashed config; retries/backoff bounded to avoid runaway loops.
- **State**: Ephemeral execution via command provider; logs captured per node.

### BootstrapCommand
- **Fields**: initNodeIP (string), talosconfigPath (temp path), commandFlags (list), tolerateAlreadyBootstrapped (bool).
- **Validation**: Runs once per cluster; must short-circuit if etcd already bootstrapped; uses stashed talosconfig.
- **State**: Ephemeral execution; success recorded via Pulumi resource status.

### Migration Guide
- **Fields**: prerequisites list, backup steps (Pulumi stack export), migration steps, validation steps, rollback steps (stack import), troubleshooting/remediation guidance for pulumiverse remnants.
- **Validation**: Steps must be sequential and idempotent; rollback path required; includes detection instructions for pulumiverse artifacts.
- **State**: Documentation asset; referenced by users executing migration.

### Cluster Spec State
- **Fields**: cluster configuration (schema-validated), secrets/configs (stable reuse), Pulumi state entries referencing cluster resources.
- **Validation**: Must remain unchanged except expected provider updates; residual pulumiverse resources must be blocked with remediation guidance.
- **State Transitions**: Pre-migration (pulumiverse present) → backup → talosctl-only apply → steady state; rollback restores pre-migration state via stack import if failure.
