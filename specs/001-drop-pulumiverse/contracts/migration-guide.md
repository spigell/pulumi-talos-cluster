# Migration Guide Contract

## Purpose
Required sections and acceptance expectations for the migration guide that moves users from pulumiverse dependency to talosctl-only operation.

## Sections
1. **Prerequisites**: Supported talosctl versions/architectures, PATH placement, checksum/provenance guidance, required Pulumi CLI version, backup location requirements.
2. **Backup**: Mandatory `pulumi stack export` with storage guidance and verification of export success.
3. **Detection**: Steps to identify pulumiverse resources/config in state; blocking criteria and remediation instructions prior to apply.
4. **Migration Steps**: Ordered, idempotent actions to remove pulumiverse usage and switch to talosctl-only; include environment prep and provider install notes.
5. **Validation**: Post-migration checks for lifecycle operations (create/update/delete/kubeconfig) confirming no pulumiverse downloads and talosctl-only execution.
6. **Rollback**: Use `pulumi stack import` to restore pre-migration state; describe triggers for rollback and validation after restore.
7. **Troubleshooting**: Common failures (version mismatch, mixed architectures, in-progress operations) with expected outcomes and remediation.
8. **Logging & Observability**: Expectations for success/failure signals and secret redaction; where logs live.

## Acceptance Criteria
- Each section is present and complete with measurable outcomes (e.g., validation confirms zero pulumiverse downloads, talosctl version matches matrix).
- Steps are executable on Linux with operator-supplied talosctl and do not require bundled binaries.
- Rollback instructions are explicit and testable.
- Support matrix and architecture guidance are documented.
