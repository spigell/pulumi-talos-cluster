# pulumi-talos-cluster (WIP)

`pulumi-talos-cluster` is a Pulumi component designed to simplify the creation and management of Talos clusters. This component abstracts the complexities of setting up and managing Talos-based Kubernetes clusters, allowing for streamlined deployment and configuration.

This component can be used for bare-metal and cloud installation. The direct access to apid in nodes is required.

*Note: This project is in active development, and not everything is completed.*

## Motivation

The official Terraform (and therefore Pulumi) provider for Talos has certain limitations, particularly around upgrading and configuring clusters, as highlighted in issues like [#195](https://github.com/siderolabs/terraform-provider-talos/issues/195). This component leverages the `pulumiverse/talos` and `pulumi/command` providers to fully manage Talos clusters, overcoming these limitations.
This component is also planned for use in the [pulumi-hcloud-kube-hetzner](https://github.com/spigell/pulumi-hcloud-kube-hetzner) project, which provides a comprehensive solution for managing Kubernetes clusters on Hetzner Cloud with Pulumi.


## Development
### GO
*Note: It is recommended to use Pulumi local storage for development, as using the Pulumi service or self-hosted S3 storage can impact the speed of deployments.*

For component building:
```
$ make build && make install_provider # It generates all SDKs and build providers
$ export PATH=$PATH:~/go/bin
```

### Example

Refer to the `examples/` directory for sample Pulumi programs using the `pulumi-talos-cluster` component.

## Roadmap

The following features and improvements are planned for future releases:

- [x] **Kubernetes Version Configuration**: Allow setting the Kubernetes version directly via CLI

- [ ] **Tests and Continuous Integration**: Implement tests and CI/CD pipelines to ensure code quality and stability.
- [ ] **Multi-language Examples**: Provide usage examples in four languages (TypeScript, Python, Go, and .NET).
- [ ] **Comprehensive Documentation**: Enhance documentation with detailed setup, customization, and troubleshooting guides.
- [ ] **`talosctl` Installation**: Automate the installation and configuration of `talosctl`.
