# Changelog

## Unreleased

### Fixed

- `talosctl upgrade` now uses the machine's `talosImage`. Before, it read the install image from a stale generated configuration, so image changes never reached the nodes. The secure `apply-config` also sets `machine.install.image` to `talosImage`, and it no longer merges the config patches a second time, which duplicated list entries.
- Extension documents (`apiVersion`/`kind`) in config patches are passed to `talosctl gen config` as a separate patch, so they are kept.

### Added

- `Cluster.machineTopology` output and optional `Apply.machineTopology` input. They hold plain machine IDs grouped by type, so per-node resources can be planned before `applyMachines` is known. When `machineTopology` is omitted, it is derived from `applyMachines`. Programs written for 0.4.1 work unchanged.
- Re-runs of `talosctl upgrade` are skipped when the node already runs the target image (installed image, Talos tag and Image Factory schematic match). Re-runs of `apply-config` are skipped when the dry run reports no changes.

### Changed

- CLI stages use an `update` command and re-run when their inputs or files change, instead of relying on `triggers`.

### Upgrade notes

- On the first `pulumi up` after upgrading from 0.4.1, every node shows updates for its CLI stages. Nodes that already run their `talosImage` are not upgraded or rebooted. Nodes whose `talosImage` was changed under 0.4.1 are upgraded and rebooted once. `upgrade-k8s` runs once without rebooting nodes. See "Talos Image Upgrades" in the README.
