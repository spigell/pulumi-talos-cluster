# Node-Level Experiments and Talosctl Flow

## Hetzner nodes-only bring-up (talos maintenance mode)

- Commands (via pulumi-talos-cluster-mcp):
  - Created scratch Pulumi project/stack `pulumi-talos-hcloud-go` with stack `dev` for node-only testing.
  - `pulumi stack init dev` (integration-tests/testdata/programs/hcloud-go) to create a scratch stack.
  - Temporarily adjusted `main.go` to call `hcloud.NewWithIPS(...).Up()` directly (skip talos cluster Apply); reverted after run.
  - `pulumi up --stack dev --yes` to provision only infrastructure (network, subnet, SSH key, primary IPs, servers). Outputs: `controlplane-1=88.198.164.253`, `worker-1=49.13.18.214`.
  - `talosctl -n 88.198.164.253 version --insecure` to verify maintenance-mode behavior.
  - Cleanup: `pulumi destroy --stack dev --yes` and `pulumi stack rm dev --yes`.
- Findings:
  - Hetzner nodes provision successfully using existing hcloud-go stack infra without Talos Apply; nodes boot into Talos maintenance mode (version RPC unimplemented).
  - talosctl client v1.12.0 used; maintenance-mode response confirms need for explicit config application when replacing the provider with talosctl flows.

## TF provider mapping → talosctl flow (nodes-only project `integration-tests/testdata/programs/hcloud-nodes-go`)

- Terraform provider parity:
  - `talos_machine_secrets` resource generates cluster secrets/client config. Mapped talosctl: `talosctl gen secrets --output-file integration-tests/testdata/programs/hcloud-nodes-go/tmp/secrets.yaml --force`.
  - `talos_machine_configuration`/`talos_machine_bootstrap` use generated secrets/client config to render/apply machine configs. Mapped talosctl: `talosctl gen config hcloud-nodes-go https://49.13.18.214:6443 --with-secrets integration-tests/testdata/programs/hcloud-nodes-go/tmp/secrets.yaml --output-dir integration-tests/testdata/programs/hcloud-nodes-go/tmp/config --force`.
  - Apply/bootstrap equivalents:
    - CP apply: `talosctl --talosconfig .../tmp/config/talosconfig --endpoints 49.13.18.214 apply-config --nodes 49.13.18.214 -f .../controlplane.yaml`.
    - Worker apply: `talosctl --talosconfig .../tmp/config/talosconfig --endpoints 88.198.164.253 apply-config --nodes 88.198.164.253 -f .../worker.yaml`.
    - Bootstrap attempt: `talosctl --talosconfig .../tmp/config/talosconfig --endpoints 49.13.18.214 bootstrap --nodes 49.13.18.214` (failed previously when etcd data present).
- Stack execution notes:
  - `pulumi stack init dev` / `pulumi up --stack dev --yes` for `pulumi-talos-hcloud-nodes-go` produced nodes: cp `49.13.18.214`, worker `88.198.164.253` (no Talos Apply/userdata).
  - Maintenance check pre-apply: `talosctl -n 49.13.18.214 version --insecure` → API unimplemented.
  - Configs/talosconfig stored under `integration-tests/testdata/programs/hcloud-nodes-go/tmp/config/`; talosconfig kept for reference.
  - Cleanup: `pulumi destroy --stack dev --yes` and `pulumi stack rm dev --yes`.

### Bootstrap challenges

- Attempts to bootstrap hit `AlreadyExists desc = etcd data directory is not empty` when nodes retained etcd data; even with STATE/EPHEMERAL wipe, the image appears to retain etcd.
- Recommendation: use fresh images or broader wipe (include etcd partition) before bootstrap when using talosctl-only flow.

## Cloudflared extension patch + talosctl apply (no reboot)

- Added `ExtensionServiceConfig` for cloudflared to generated configs at `.../tmp/config/controlplane.yaml` and `.../tmp/config/worker.yaml`:
  ```
  ---
  apiVersion: v1alpha1
  kind: ExtensionServiceConfig
  name: cloudflared
  environment:
    - TUNNEL_TOKEN=CHANGE_ME_AGAIN
    - TUNNEL_METRICS=localhost:2001
    - TUNNEL_EDGE_IP_VERSION=auto
  ```
- Applies using talosconfig and control-plane endpoint:
  - CP: `talosctl --talosconfig .../tmp/config/talosconfig --endpoints 49.13.18.214:50000 apply-config --nodes 49.13.18.214 -f .../controlplane.yaml`
  - Worker: `talosctl --talosconfig .../tmp/config/talosconfig --endpoints 49.13.18.214:50000 apply-config --nodes 88.198.164.253 -f .../worker.yaml`
  - Result: “Applied configuration without a reboot.” Replace TUNNEL_TOKEN with real value; align image if baking the patch.
