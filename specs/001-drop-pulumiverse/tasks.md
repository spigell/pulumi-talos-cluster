# Tasks: Drop Pulumiverse dependency for talosctl

## Provider implementation

- [x] Generate Talos secrets with external `talosctl`.
- [x] Generate machine configurations and talosconfig with external `talosctl`.
- [x] Apply initial and authenticated machine configuration with `talosctl`.
- [x] Bootstrap the init node and retrieve kubeconfig with `talosctl`.
- [x] Use bounded retries and restricted temporary working directories.
- [x] Keep sensitive stdout out of provider logs.
- [x] Mark generated configurations and credentials as Pulumi secret outputs.

## Migration and documentation

- [x] Document prerequisites and state backup.
- [x] Document detection of remaining Pulumiverse resources and blocking preview conditions.
- [x] Document the operator-controlled migration sequence.
- [x] Document validation, rollback, and troubleshooting.
- [x] Update quickstart and README for the external `talosctl` requirement.
- [x] Update the cluster workflow diagram for the CLI resource flow.

## Verification

- [x] Add real `talosctl` secret/config generation tests.
- [x] Add provider command-generation unit coverage.
- [x] Run provider and integration linters.
- [x] Run provider and integration unit tests.
- [x] Run the scoped `TestHcloudClusterGo` lifecycle test successfully in GitHub Actions.
