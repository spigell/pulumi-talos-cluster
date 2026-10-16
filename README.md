# pulumi-talos-cluster (WIP)

`pulumi-talos-cluster` is a Pulumi component designed to simplify the creation and management of Talos clusters. This component abstracts the complexities of setting up and managing Talos-based Kubernetes clusters, allowing for streamlined deployment and configuration.

This component can be used for bare-metal and cloud installations. Direct access to `apid` on nodes is required.

*Note: This project is in active development, and not all features are complete.*

## Requirements

Only Linux is supported as the runner operating system. The following tools must be available:

- `bash` (must be available at `/bin/bash`)
- `printf`
- `talosctl` (version `v1.12.0` or compatible recommended to match the default Talos image version)

## Quick Start

1. Install `bash` (at `/bin/bash`), `printf`, and `talosctl` on a Linux machine.
2. Clone this repository.
3. Run an example program, such as those under `integration-tests/testdata`, using `pulumi up`. The provider plugin installs automatically.

Existing stacks that use the Pulumiverse Talos provider must follow the [migration guide](specs/001-drop-pulumiverse/contracts/migration-guide.md) before applying this version.

Generated machine configurations, client credentials, talosconfig, and kubeconfig are Pulumi secret outputs. Use `pulumi stack output --show-secrets <name>` only in a controlled environment.

## Cluster Lifecycle

```mermaid
flowchart TD
    Start[Cluster inputs] --> Secrets[talosctl gen secrets]
    Secrets --> MachineConfig[talosctl gen config: machine configs]
    Secrets --> Talosconfig[talosctl gen config: talosconfig]
    MachineConfig --> InitialApply{skip initial apply?}
    InitialApply -- no --> Insecure[talosctl apply-config --insecure]
    InitialApply -- yes --> Secure
    Insecure --> Secure[talosctl apply-config with talosconfig]
    Talosconfig --> Secure
    Secure --> Bootstrap[talosctl bootstrap]
    Bootstrap --> Upgrade[talosctl upgrade-k8s]
    Upgrade --> Kubeconfig[talosctl kubeconfig]
    Kubeconfig --> Ready[Cluster ready]
```

The Pulumi command provider executes each CLI operation. Temporary credentials are written with restrictive permissions, sensitive stdout is suppressed, and working directories are removed after execution.

## Talos Image Upgrades

Changing a machine's `talosImage` upgrades the node to that image: `talosctl upgrade` uses the machine's `talosImage` directly and the following `apply-config` sets `machine.install.image` to the same value. (Up to 0.4.1, the upgrade read the install image from a stale generated configuration, so image changes were not applied.)

`Apply` accepts an optional `machineTopology` input (plain machine IDs grouped by type). Pass `Cluster.machineTopology` so that per-node resources can be planned before `applyMachines` is known. When it is omitted, as in programs written for 0.4.1, the topology is derived from `applyMachines`, which keeps the 0.4.1 behaviour.

The CLI stages now re-run (update) when their inputs change, not only when a few trigger values change. They keep the same `triggers` as 0.4.1 (`talosImage` for `cli-upgrade`, the config patches and cluster endpoint for `cli-apply-config`, the Kubernetes version for `cli-upgrade-k8s`). A change to a trigger still replaces the command, and a replacement runs the create command. To avoid needless reboots, the create and update commands run the same guard:

- `talosctl upgrade` is skipped when the node already runs the target image. This means the installed config's `machine.install.image` equals `talosImage`, `talosctl version` reports the image tag and, for Image Factory images, the running `schematic` extension matches the schematic in the image path. If any of these cannot be read or does not match, the upgrade runs. Images without a tag are always upgraded. On a new node the upgrade runs as before, unless the node already runs the target image (for example, because its initial install used that image).
- `talosctl apply-config` is skipped when `apply-config --dry-run` reports `No changes`. Otherwise it applies in the default `auto` mode.
- `talosctl upgrade-k8s` is not guarded. It never reboots nodes, and it skips components that already run the target version, but it re-syncs the bootstrap manifests.

### First `pulumi up` after upgrading from 0.4.1

The preview shows updates, not replaces, for `gen-machine-config`, `cli-upgrade`, `cli-apply-config`, `cli-upgrade-k8s` and `kubeconfig`. The command resources gain an `update` command and a content hash, and their `triggers` stay unchanged. Expect the following:

- Nodes that already run their `talosImage` skip `talosctl upgrade` and are not rebooted.
- Nodes whose `talosImage` was changed under 0.4.1 still run the old image, because 0.4.1 ignored the change. These nodes are upgraded and rebooted once. Control planes are upgraded one after another, each gated by the etcd health check. Workers are upgraded in parallel.
- `apply-config` applies only real differences. One example is the de-duplication of list entries that 0.4.1 produced by merging the config patches twice.
- `upgrade-k8s` runs once and re-syncs the bootstrap manifests.

The skip check depends on the `talosctl` output formats described above. To be safe, run `pulumi preview` first and roll out one stack at a time.

## Motivation

The official Terraform (and therefore Pulumi) provider for Talos has certain limitations, particularly around upgrading and configuring clusters, as highlighted in issues like [#195](https://github.com/siderolabs/terraform-provider-talos/issues/195). This component directly orchestrates `talosctl` and utilizes the `pulumi/command` provider to fully manage Talos clusters, overcoming these limitations and providing a more native experience.

## Development
### Go
*Note: It is recommended to use Pulumi's local storage for development, as using the Pulumi service or self-hosted S3 storage can slow down deployments.*

To build the component:
```
$ make build && make install_provider # This generates all SDKs and builds the provider
$ export PATH=$PATH:~/go/bin
```

### Preparing for release
Run these commands in the release branch's local checkout. The worker needs Go, a Pulumi CLI matching `.pulumi.version`, Node.js and Yarn, .NET, Python 3 with `venv`, and Make on its `PATH`. On Linux x86_64, install a missing Pulumi CLI outside the checkout with:

```bash
PULUMI_TOOLS=$(mktemp -d)
curl -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v$(cat .pulumi.version)-linux-x64.tar.gz" -o "$PULUMI_TOOLS/pulumi.tar.gz"
tar -xzf "$PULUMI_TOOLS/pulumi.tar.gz" -C "$PULUMI_TOOLS"
export PATH="$PULUMI_TOOLS/pulumi:$PATH"
```

The `pulumi-deployer` MCP may report its Pulumi or `pulumictl` version, but its working directory is separate from this checkout; it cannot write the SDKs here. An explicit `VERSION` makes `pulumictl` unnecessary for the build.

```bash
export VERSION=v0.5.1 # Replace with the intended release tag.
make generate_schema
make generate # Includes the Go SDK.
make build    # Builds the provider, .NET, Node.js, and Python SDKs.

node -e 'const v = require("./sdk/nodejs/bin/package.json").version; if (v !== process.env.VERSION || v.includes("+dirty")) process.exit(1)'
test "$(cat sdk/dotnet/version.txt)" = "${VERSION#v}"
grep -Fx "  version = \"$(python3 scripts/python-version.py "${VERSION#v}")\"" sdk/python/pyproject.toml
go version -m bin/pulumi-resource-talos-cluster | grep -F "version.Version=$VERSION"
git diff --check
git status --short
```

Confirm that the provider binary records the requested `VERSION`, then review the generated diff before the release worker commits it. Go SDK consumers take their version from the Go module tag. When `VERSION` is unset, the Makefile retains the automatic `pulumictl` version lookup used by CI.

## Example
Refer to the `integration-tests/testdata` directory for sample Pulumi programs using the `pulumi-talos-cluster` component.

## Roadmap

### Current focus
- Refactor tests.
- Add more cloud providers to tests.
- Include all supported languages in tests.

### Planned enhancements
- [x] **Kubernetes Version Configuration**: Allow setting the Kubernetes version directly via the CLI.
- [ ] **Tests and Continuous Integration**: Implement tests and CI/CD pipelines to ensure code quality and stability.
- [ ] **Multi-language Examples**: Provide usage examples in four languages (TypeScript, Python, Go, and .NET).
- [ ] **Comprehensive Documentation**: Enhance documentation with detailed setup, customization, and troubleshooting guides.
