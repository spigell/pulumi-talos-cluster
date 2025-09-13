# Workbench Deployment

This directory contains Kustomize configurations for running the `pulumi-talos-cluster` development workbench inside a Kubernetes cluster.

## Pod Contents
The workbench pod includes two containers:
- `plugin`—an init container that fetches the repository, builds the provider plugin, and starts a Delve debugger.
- `workbench`—the main container with Pulumi, Go, Talos, and other development tools.

## Prerequisites
- `kubectl` installed locally.
- A Kubernetes cluster **with user namespace support**.
  - On clusters without user namespaces, you must enable `hostUsers` in the Pod spec. An example patch is provided in [`develop/overrides/patch-nfs-and-user.yaml`](develop/overrides/patch-nfs-and-user.yaml).

## Usage
To deploy using the default configuration (user namespace supported):

```bash
kubectl apply -k develop/base
```

For clusters without user namespace support, apply the overrides (enabling `hostUsers` and additional configuration):

```bash
kubectl apply -k develop/overrides
```

The `overrides` kustomization is provided as an example; adjust it to match your environment or create your own patches.

This creates the `pulumi-talos-cluster-workbench` pod in the `pulumi-talos-cluster-dev` namespace with all required development tools.
