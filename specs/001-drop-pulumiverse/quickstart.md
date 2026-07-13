# Quickstart: talosctl-only operation

1. Install `talosctl v1.12.0` or a compatible release for the runner architecture and verify `talosctl version --client`.
2. Back up an existing stack with `pulumi stack export --file pulumi-state-before-talosctl.json`.
3. Follow `contracts/migration-guide.md` for an existing Pulumiverse-backed stack. Stop if the preview proposes unexpected node replacement or deletion.
4. Install the provider and SDK, then run `pulumi preview --diff`.
5. Apply with `pulumi up` after reviewing the command-resource changes.
6. Verify that credential and configuration outputs are secret, retrieve the talosconfig explicitly for validation, and run `talosctl health`.
7. Run a second preview and confirm that it has no unexpected configuration or secret replacement.

The component uses the Pulumi command provider and external `talosctl`. It does not use Pulumiverse Talos resources or bundle a `talosctl` binary.
