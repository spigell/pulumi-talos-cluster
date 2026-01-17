# Quickstart: talosctl-only migration

1) Install supported `talosctl` for your runner architecture, place it on PATH, and verify with `talosctl version` (use the documented matrix as guidance; provider does not enforce).
2) Back up state: `pulumi stack export > backup.json` (store securely and confirm file integrity).
3) Scan for pulumiverse usage (state, provider config) and remediate per migration guide before proceeding.
4) Apply the talosctl-only stack changes (secrets/config generation via Stash, command provider apply/bootstrap).
5) Validate lifecycle: create/update/delete and kubeconfig retrieval complete without pulumiverse downloads; inspect command logs for clear success signals.
6) On failure, restore with `pulumi stack import < backup.json`, address root cause (version/arch mismatch, in-progress ops), and retry.
