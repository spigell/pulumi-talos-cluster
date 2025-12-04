# pulumi-talos-cluster (WIP)

`pulumi-talos-cluster` is a Pulumi component designed to simplify the creation and management of Talos clusters. This component abstracts the complexities of setting up and managing Talos-based Kubernetes clusters, allowing for streamlined deployment and configuration.

This component can be used for bare-metal and cloud installations. Direct access to `apid` on nodes is required.

*Note: This project is in active development, and not all features are complete.*

## Requirements

Only Linux is supported as the runner operating system. The following tools must be available:

- `bash`
- `printf`
- `talosctl`

## Quick Start

1. Install `bash`, `printf`, and `talosctl` on a Linux machine.
2. Clone this repository.
3. Run an example program, such as those under `integration-tests/testdata`, using `pulumi up`. The provider plugin installs automatically.

## Motivation

The official Terraform (and therefore Pulumi) provider for Talos has certain limitations, particularly around upgrading and configuring clusters, as highlighted in issues like [#195](https://github.com/siderolabs/terraform-provider-talos/issues/195). This component leverages the `pulumiverse/talos` and `pulumi/command` providers to fully manage Talos clusters, overcoming these limitations.


## Development
### Go
*Note: It is recommended to use Pulumi's local storage for development, as using the Pulumi service or self-hosted S3 storage can slow down deployments.*

To build the component:
```
$ make build && make install_provider # This generates all SDKs and builds the provider
$ export PATH=$PATH:~/go/bin
```

### Preparing for release
To build the provider and all SDKs (Go, Node.js, Python, and .NET) in one step, set the desired version explicitly (for example, `v0.7.0`):

```bash
VERSION=v0.7.0 make build
git add .
git commit -m 'regenerate'
VERSION=v0.7.0 make build
git add .
git commit -m 'release'
```

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
- [ ] **`talosctl` Installation**: Automate the installation and configuration of `talosctl`.
