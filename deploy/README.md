# Workbench Deployment

This directory contains Kubernetes manifests for running the `pulumi-talos-cluster` development workbench inside a cluster.

## Pod Contents
The workbench pod includes two containers:
- `plugin`—an init container that fetches the repository, builds the provider plugin, and starts a Delve debugger.
- `workbench`—the main container with Pulumi, Go, Talos, and other development tools.

## Prerequisites
- `kubectl` installed locally.
- A Kubernetes cluster **with user namespace support**.
  - On clusters without user namespaces, you must enable `hostUsers` in the Pod spec. An example patch is provided in [`develop/overrides/patch-nfs-and-user.yaml`](develop/overrides/patch-nfs-and-user.yaml).

## Usage
Before applying, verify the pod spec in `deploy/develop/base/develop.yaml` meets your cluster constraints (for example, toggle `hostUsers` if your nodes cannot run user namespaces).

Apply the manifest:

```bash
kubectl apply -f deploy/develop/base/develop.yaml
```

This creates the `pulumi-talos-cluster-workbench` pod in the `pulumi-talos-cluster-dev` namespace with all required development tools.
