# Data Model

## Entities

### Talosctl Binary (external)
- **Fields**: path (string), version (string), architecture (string), checksum/provenance (string), permissions (executable boolean).
- **Validation**: Version must match documented support matrix; architecture must align with runner; checksum/provenance recommended for supply chain assurance.
- **State**: Provided by operator; not persisted by provider.

### Migration Guide
- **Fields**: prerequisites list, backup steps (Pulumi stack export), migration steps, validation steps, rollback steps (stack import), troubleshooting/remediation guidance for pulumiverse remnants.
- **Validation**: Steps must be sequential and idempotent; rollback path required; includes detection instructions for pulumiverse artifacts.
- **State**: Documentation asset; referenced by users executing migration.

### Cluster Spec State
- **Fields**: cluster configuration (schema-validated), secrets/configs (stable reuse), Pulumi state entries referencing cluster resources.
- **Validation**: Must remain unchanged except expected provider updates; residual pulumiverse resources must be blocked with remediation guidance.
- **State Transitions**: Pre-migration (pulumiverse present) → backup → talosctl-only apply → steady state; rollback restores pre-migration state via stack import if failure.
