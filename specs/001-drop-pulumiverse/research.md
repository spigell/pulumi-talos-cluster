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
- **Decision**: Require operator-supplied talosctl on PATH with checksum/source provenance and version validation against a documented matrix. 
- **Rationale**: Removes pulumiverse downloads while keeping deterministic tooling; aligns with FR-002 and constitution security/version discipline. 
- **Alternatives considered**: Bundling talosctl — rejected as against feature goal and increases supply-chain risk.

### Pulumi State Safety
- **Decision**: Migration guide will mandate `pulumi stack export` backup before changes and describe restore on failure; block apply if pulumiverse resources detected in state. 
- **Rationale**: FR-004/FR-005 require safe migration and rollback; preserves determinism. 
- **Alternatives considered**: Optional backups — rejected due to risk of unintended deletes.
