# Quickstart: talosctl-only migration

1) Place supported `talosctl` on PATH (per version/arch matrix) and verify with `talosctl version`.
2) Export Pulumi state backup: `pulumi stack export > backup.json` and store securely.
3) Remove pulumiverse dependencies from environment/config as documented.
4) Apply migration steps to switch provider to talosctl-only flows.
5) Validate lifecycle actions (create/update/delete/kubeconfig) run without pulumiverse downloads.
6) If failures occur, restore state with `pulumi stack import < backup.json` and retry after remediation.
