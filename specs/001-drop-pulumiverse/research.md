# Phase 0 Research

## Decisions

### Performance Goals
- **Decision**: Target zero-downtime migration for existing stacks; accept talosctl command latency typical for cluster lifecycle (minutes) without additional performance SLOs. 
- **Rationale**: Feature scope is dependency removal; user stories emphasize reliability over speed. Zero-downtime and no pulumiverse downloads align with SC-001 and migration safety. 
- **Alternatives considered**: Define explicit p95 timings for talosctl commands — rejected due to dependence on cluster size/network and not called out in spec.

### Scale/Scope
- **Decision**: Plan for standard Talos cluster sizes used in existing integration tests (small to medium, single architecture per run) with guidance to document multi-arch expectations and version matrix in migration guide. 
- **Rationale**: Spec lists architecture/version edge cases; no explicit node-count limits. Integration fixtures set practical bounds; documenting support matrix keeps expectations explicit. 
- **Alternatives considered**: Declare unlimited scale — rejected as unvalidated; would require dedicated performance testing not in scope.

### Talosctl Dependency Practices
- **Decision**: Require operator-supplied talosctl on PATH with checksum/source provenance; document a recommended version/arch matrix for operators but do not enforce automated version validation. 
- **Rationale**: Removes pulumiverse downloads while keeping deterministic tooling; trusts operator choice while still guiding compatibility; aligns with FR-002 and constitution security/version discipline without adding validation coupling. 
- **Alternatives considered**: Bundling talosctl — rejected as against feature goal and increases supply-chain risk.

### Pulumi State Safety
- **Decision**: Migration guide will mandate `pulumi stack export` backup before changes and describe restore on failure; block apply if pulumiverse resources detected in state. 
- **Rationale**: FR-004/FR-005 require safe migration and rollback; preserves determinism. 
- **Alternatives considered**: Optional backups — rejected due to risk of unintended deletes.

## Alignment guidance (talosctl vs Terraform provider)

- First apply often needs `--insecure` until talosconfig/CA align; expect to use insecure mode initially when hitting fresh nodes.
- Generate secrets/config to stdout (avoid intermediate files) and stash via `pulumi.Stash` instead of writing to disk.
- Follow Terraform provider semantics for secrets generation (no custom doc/example flow): mirror `talos_machine_secrets` + `machine_configuration_apply` calls rather than ad-hoc gen/apply sequences.
- Stash usage reference: see `specs/001-drop-pulumiverse/research_stash.md` for how to persist generated `talosconfig`/kubeconfig/secrets in stack state (helps avoid regeneration and aligns with FR-007 idempotency).

### Pulumiverse resources currently in use (to replace with talosctl flow)

- `machine.NewSecrets` (`github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine`) – generates cluster secrets/client config.
- `machine.GetConfigurationOutput` – renders per-machine Talos configs.
- `machine.NewConfigurationApply` – applies initial configs (init/worker/CP).
- `machine.NewBootstrap` – etcd bootstrap on init node.
- `client.GetConfigurationOutput` – builds talosconfig (CA/cert/key + endpoints/nodes).
- `pulumi_cluster.NewKubeconfig` (`github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster`) – emits kubeconfig.
