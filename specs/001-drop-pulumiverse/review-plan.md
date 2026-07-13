# Implementation Plan Review

## Final decision

The implementation replaces Pulumiverse Talos resources with external `talosctl` commands managed by the Pulumi command provider.

## Required properties

- Keep the public `Cluster` and `Apply` component contracts compatible where possible.
- Run `talosctl` supplied by the operator; do not bundle or download it from provider code.
- Mark generated configurations, client credentials, talosconfig, and kubeconfig as secret outputs.
- Suppress command stdout that may contain secret material and clean restricted temporary files.
- Require an explicit, reviewed state migration for existing Pulumiverse-backed stacks.
- Test real `talosctl` generation and a complete basic-cluster lifecycle in CI.

## Verification

- Provider and integration unit tests pass.
- Provider and integration linters pass.
- `TestHcloudClusterGo` completes on the GitHub integration runner.
- A second preview does not propose unexpected replacements.
- The migration guide includes backup, blocking conditions, validation, and rollback.
