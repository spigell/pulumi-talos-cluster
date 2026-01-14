# Research: Proposed Flow (talosctl + pulumi.Stash)

**Context**: Replacing `pulumiverse` provider resources with `talosctl` CLI calls, utilizing `pulumi.Stash` to persist critical state (secrets, configs) to ensure idempotency and safety.

## 1. Secrets Generation (State Root)
**Goal**: Replace `machine.NewSecrets`.
*   **Current**: Generates a bundle of secrets (PKI, tokens) managed by Pulumi state.
*   **Proposed**:
    *   Define a `pulumi.Stash` resource named `ClusterSecrets`.
    *   **Generation**: Execute `talosctl gen secrets -o json` *only* if the stash is empty.
    *   **Persistence**: Save the JSON output into the stash.
    *   **Why**: Ensures secrets are generated once and reused. Using `Stash` keeps them encrypted in the stack state.

## 2. Configuration Generation
**Goal**: Replace `machine.GetConfigurationOutput`, `client.GetConfigurationOutput`, and `pulumi_cluster.NewKubeconfig`.
*   **Current**: Generates machine configs (controlplane/worker), `talosconfig`, and `kubeconfig` based on inputs and secrets.
*   **Proposed**:
    *   Define a `pulumi.Stash` resource named `ClusterConfigurations`.
    *   **Input**: Takes the output of `ClusterSecrets` (the stashed secrets) and cluster specification (name, endpoint).
    *   **Generation**: Execute `talosctl gen config --with-secrets <secrets> ... -o json` (or equivalent flags to output all components).
    *   **Persistence**: Save the resulting JSON structure (containing machine configs, client config) into the stash.
    *   **Why**: `talosctl gen config` is deterministic *if* provided with the same secrets. Stashing the output avoids re-running the binary for every read and provides a stable source for the "Apply" step.

## 3. Config Application
**Goal**: Replace `machine.NewConfigurationApply`.
*   **Current**: Applies specific configurations to nodes.
*   **Proposed**:
    *   Use `command.local.Command` (or equivalent execution resource).
    *   **Input**: `ClusterConfigurations` stash output + Node IP.
    *   **Action**: 
        *   Extract the specific machine type config (CP/Worker) from the stash.
        *   Run `talosctl apply-config --insecure --nodes <NodeIP> --config-patch <content>` (or pipe content).
    *   **Lifecycle**: Triggers on changes to the Stashed configuration or Node IP.

## 4. Bootstrap
**Goal**: Replace `machine.NewBootstrap`.
*   **Current**: Bootstraps the etcd cluster on the initial control plane node.
*   **Proposed**:
    *   Use `command.local.Command`.
    *   **Input**: `ClusterConfigurations` (for `talosconfig`), Init Node IP.
    *   **Action**: Run `talosctl bootstrap --nodes <InitNodeIP> --talosconfig <path_to_temp_talosconfig>`.
    *   **Lifecycle**: Create-only resource. Needs to handle "already bootstrapped" errors gracefully or check cluster health first.

## 5. Exports & Outputs
*   **Proposed**:
    *   `kubeconfig` and `talosconfig` are parsed directly from the `ClusterConfigurations` stash and exported as Stack Outputs, mimicking the previous provider's behavior.

## Summary of Changes
| Pulumiverse Resource | New Flow Component | State Mechanism |
| :--- | :--- | :--- |
| `machine.NewSecrets` | `talosctl gen secrets` | `pulumi.Stash` (Encrypted) |
| `machine.GetConfigurationOutput` | `talosctl gen config` | `pulumi.Stash` |
| `client.GetConfigurationOutput` | Derived from Config Stash | N/A (Derived) |
| `machine.NewConfigurationApply` | `talosctl apply-config` | `command.local.Command` |
| `machine.NewBootstrap` | `talosctl bootstrap` | `command.local.Command` |
