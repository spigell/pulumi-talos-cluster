# Migration Guide: Pulumiverse to talosctl

This procedure migrates an existing stack to the talosctl-based component. It does not convert Pulumiverse resource state automatically.

## Prerequisites

- Linux with `bash`, `printf`, and `talosctl` on `PATH`.
- `talosctl v1.12.0` is recommended for the current provider release; the provider does not enforce the client version.
- Pulumi CLI compatible with the version in `.pulumi.version`.
- Access to the stack backend and its secrets provider.
- A maintenance window and a secure location for an encrypted stack export.

Verify the tools before changing state:

```bash
pulumi version
talosctl version --client
command -v talosctl
```

## Back Up State

Select the stack and export it before installing the new component:

```bash
pulumi stack select <stack>
pulumi stack export --file pulumi-state-before-talosctl.json
pulumi preview --diff
```

The export contains encrypted secrets but remains sensitive. Store it with access controls and verify that it is nonempty.

## Detect Pulumiverse Resources

Inspect the export and project dependencies:

```bash
grep -n 'pulumiverse\|talos:index' pulumi-state-before-talosctl.json
grep -R -n 'pulumiverse\|pulumi-talos' Pulumi.yaml package.json go.mod requirements.txt 2>/dev/null
```

Do not apply the new component while Pulumiverse-managed Talos resources are still scheduled for deletion or replacement. Review `pulumi preview --diff`; unexpected machine configuration, bootstrap, or secrets deletion is a blocking condition.

## Migrate

1. Update the project to the talosctl-based `pulumi-talos-cluster` SDK and remove its Pulumiverse Talos dependency.
2. Ensure the existing cluster name, endpoint, machine IDs, machine types, Talos image, Kubernetes version, and configuration patches are unchanged.
3. Remove obsolete Pulumiverse resource registrations from the program and state using the state procedure already approved for the stack. Do not delete the live Talos nodes.
4. Install the updated provider and SDK.
5. Run a detailed preview:

```bash
pulumi preview --diff --save-plan talosctl-migration.plan
```

6. Confirm that the preview creates command resources for secrets/config generation and Talos operations, does not replace cloud servers unexpectedly, and does not expose credentials in plaintext.
7. Apply the reviewed plan:

```bash
pulumi up --plan talosctl-migration.plan
```

Generated machine configurations, client credentials, talosconfig, and kubeconfig are Pulumi secret outputs.

## Validate

```bash
pulumi stack output
pulumi stack output --show-secrets talosconfig > /tmp/talosconfig
chmod 600 /tmp/talosconfig
talosctl --talosconfig /tmp/talosconfig health
talosctl --talosconfig /tmp/talosconfig kubeconfig /tmp/kubeconfig
kubectl --kubeconfig /tmp/kubeconfig get nodes
```

Successful migration means:

- the update completes without downloading or invoking a Pulumiverse Talos provider;
- all expected nodes are healthy and Ready;
- a second `pulumi preview` has no unexpected secret/configuration replacement;
- normal update and destroy previews contain only intended changes.

Remove `/tmp/talosconfig` and `/tmp/kubeconfig` after validation.

## Roll Back

If the update fails before changing live nodes, restore the program revision and import the backup:

```bash
pulumi stack import --file pulumi-state-before-talosctl.json
pulumi preview --diff
```

If any Talos command reached a live node, capture its logs and current machine state before importing old state. State rollback does not undo configuration already applied to a node.

## Troubleshooting

- `talosctl: command not found`: install the expected binary on the provider runner and verify `PATH`.
- Authentication failures: verify the cluster endpoint and that the client configuration belongs to the existing cluster.
- Unexpected replacements: stop, compare cluster inputs with the backup, and correct IDs/versions before applying.
- Architecture errors: install the `talosctl` binary matching the runner architecture.
- Secret values in logs: stop the run, rotate exposed credentials, and report the command path; stdout containing generated credentials must remain suppressed.
