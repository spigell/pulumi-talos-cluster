<!--
Sync Impact Report
- Version change: N/A → 1.0.0
- Modified principles: Initial set defined
- Added sections: Platform Constraints, Delivery Workflow
- Removed sections: None
- Templates requiring updates: ✅ .specify/templates/plan-template.md, ✅ .specify/templates/tasks-template.md, ⚠️ .specify/templates/commands (directory not present to review)
- Follow-up TODOs: None
-->

# pulumi-talos-cluster Constitution

## Core Principles

### I. Deterministic Cluster Sources
All Talos cluster definitions MUST be expressed as declarative specs validated against the canonical schema; manual node drift and ad-hoc changes are prohibited. Pulumi state, Talos configs, and generated assets remain the single source of truth and must be reproducible from versioned code.

### II. Security and Access Hygiene
Credentials, kubeconfigs, and Talos secrets MUST never be committed or logged; ephemeral access and least privilege are mandatory. Only Linux runners with required tooling (bash, printf, talosctl) may execute operations, and cloud or node access must be auditable.

### III. Testing and Validation First
Changes MUST pass gofmt, lint, and unit tests; schema validation is required before processing cluster specs. Integration tests are mandatory for cluster lifecycle changes, provider interactions, and contract updates; add scoped integration coverage when touching Talos workflows or cloud hooks.

### IV. Observability and Operability
Provider and test tooling MUST emit structured, text-friendly logs that differentiate stdout and stderr. Operational flows (apply, upgrade, validation) must surface actionable diagnostics, and failure handling must preserve artifacts under the temp Talos workdir for inspection.

### V. Version Discipline and Simplicity
Dependencies and SDKs MUST stay aligned with pinned Pulumi, pulumiverse/talos, and Talos machinery versions; upgrades follow documented sequences. Prefer the simplest viable implementation that maintains clarity and reduces breakage risk; justify added complexity explicitly.

## Platform Constraints

- Supported execution environment is Linux only with bash, printf, and talosctl available on PATH.  
- Generated SDKs MUST NOT be edited manually; regenerate via make targets when schemas change.  
- Do not commit cloud credentials; rely on environment-based auth for integration workflows.  
- Talos client state resides under the stack-scoped temp workdir and must not be reused across stacks.

## Delivery Workflow

- Feature work starts from a written spec and plan in `/specs/[feature]/`, reflecting the constitution gates.  
- Pull requests MUST show lint, gofmt, unit test results, and required integration coverage for cluster-affecting changes.  
- Code review must verify schema alignment, version pin consistency, and adherence to observability/logging expectations.  
- Release preparation requires regenerating SDKs via `make build && make install_provider` with explicit VERSION tagging.

## Governance

- This constitution supersedes conflicting practices; exceptions require explicit, documented approval in the relevant PR.  
- Amendments demand a recorded rationale, version bump per semantic rules, and updates to dependent templates and docs.  
- Compliance is reviewed during planning (Constitution Check), code review, and before releases; non-compliance blocks merge.  
- Version bumps: MAJOR for breaking governance changes or principle removal; MINOR for new principles or material expansions; PATCH for clarifications.

**Version**: 1.0.0 | **Ratified**: 2026-01-12 | **Last Amended**: 2026-01-12
