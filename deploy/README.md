# Workbench Deployment

This directory contains Kustomize configurations for running the `pulumi-talos-cluster` development workbench inside a Kubernetes cluster.

## Prerequisites
- A Kubernetes cluster **with user namespace support**.
  - On clusters without user namespaces, you must enable `hostUsers` in the Pod spec. An example patch is provided in [`develop/overrides/patch-nfs-and-user.yaml`](develop/overrides/patch-nfs-and-user.yaml).

## Usage
To deploy using the default configuration (user namespace supported):

```bash
kubectl apply -k develop/base
```

For clusters without user namespace support, apply the overrides (enables `hostUsers` and additional configuration):

```bash
kubectl apply -k develop/overrides
```

This will create the `pulumi-talos-cluster-workbench` pod in the `pulumi-talos-cluster-dev` namespace with all required development tools.

