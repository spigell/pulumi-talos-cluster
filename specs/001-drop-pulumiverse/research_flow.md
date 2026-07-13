# Research: talosctl command flow

## Decision

Replace Pulumiverse Talos resources with operator-supplied `talosctl` commands managed by the Pulumi command provider. Generated secrets and configurations are command outputs tracked by Pulumi and marked secret at the component boundary.

## Resource mapping

| Previous operation | talosctl operation | Pulumi behavior |
| --- | --- | --- |
| Generate machine secrets | `talosctl gen secrets` | Command output is secret and feeds configuration generation. |
| Generate machine configuration | `talosctl gen config --with-secrets` | One command resource per machine; output changes follow its inputs. |
| Generate client configuration | `talosctl gen config --output-types talosconfig` | Raw talosconfig and parsed client credentials are secret outputs. |
| Apply initial configuration | `talosctl apply-config --insecure` | Runs only when initial application is required. |
| Apply authenticated configuration | `talosctl apply-config --talosconfig ...` | Uses generated client credentials and bounded retries. |
| Bootstrap etcd | `talosctl bootstrap` | Runs against the selected init node after configuration. |
| Retrieve kubeconfig | `talosctl kubeconfig -` | Kubeconfig is returned as a secret output. |

## Operational constraints

- `talosctl` must be installed on `PATH` on the machine running the provider.
- Temporary files are written with restrictive permissions and removed after commands finish.
- Command stdout containing secrets is not logged.
- Changes to cluster inputs can replace command resources and regenerate their outputs. Operators must export state before upgrading an existing stack and review the preview for replacements.
- Migration from Pulumiverse resources is an operator-controlled state transition; the component does not attempt automatic state conversion.
