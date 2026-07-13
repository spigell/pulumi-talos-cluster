# Implementation Plan: Drop pulumiverse dependency for talosctl

**Branch**: `001-drop-pulumiverse` | **Date**: 2026-01-12 | **Spec**: specs/001-drop-pulumiverse/spec.md
**Input**: Feature specification from `/specs/001-drop-pulumiverse/spec.md`

## Summary

Remove Pulumiverse provider usage in favor of talosctl-only flows: use operator-supplied talosctl binaries plus command-provider resources to generate and apply Talos configs/secrets, expose sensitive values as Pulumi secrets, and deliver a migration guide with backup and rollback paths.

## Technical Context

**Language/Version**: Go 1.21+ provider; TypeScript/Python SDK artifacts (generated)  
**Primary Dependencies**: Pulumi SDK v3, Pulumi command provider, external talosctl binary (operator-supplied), Talos machinery libraries  
**Storage**: Pulumi state only; sensitive command and component outputs are marked secret  
**Testing**: gofmt, go unit tests, golangci-lint, integration tests for cluster lifecycle/migration flows  
**Target Platform**: Linux runners with talosctl on PATH; supports x86_64/arm64 runners aligned to talosctl binary  
**Project Type**: Provider + SDK/tooling repo (no frontend/mobile)  
**Performance Goals**: Zero-downtime migration; no additional latency SLOs beyond talosctl command expectations  
**Constraints**: No pulumiverse provider usage; operator supplies talosctl (document recommended matrix but do not enforce validation); architecture alignment between runner and binary; no bundled binaries; secrets/kubeconfigs must remain redacted and persisted deterministically  
**Scale/Scope**: Small/medium Talos clusters as covered by existing integration suites; document multi-arch/version expectations in migration guide

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Determinism: talosctl gen/apply flows remain driven by declared Pulumi inputs; migration blocks on unexpected replacements or Pulumiverse residue.  
- Security: No credentials or kubeconfigs are checked in; sensitive outputs are Pulumi secrets and command stdout is suppressed.  
- Testing: gofmt, golangci-lint, unit tests, and scoped integration tests for talosctl-only lifecycle and migration guide validation are required before release.  
- Observability: Command executions keep stdout/stderr separated; failure artifacts (talos workdir) preserved under stack-scoped temp dirs; clear success/failure signals in logs.  
- Version Discipline: talosctl version guidance documented (operator enforced); Pulumi/Talos libs stay pinned with regenerate commands (`make build && make install_provider`) when schema changes.

## Project Structure

### Documentation (this feature)

```text
specs/001-drop-pulumiverse/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md (created by /speckit.tasks)
```

### Source Code (repository root)

```text
provider/                # Pulumi provider source (Go), codegen, binaries
sdk/                     # Generated SDKs (go/dotnet/nodejs/python)
integration-tests/       # E2E and fixtures using talosctl flows
deploy/                  # Environment presets/assets
specs/                   # Feature specs, plans, research artifacts
Makefile                 # build/lint/test targets (make build, make install_provider, make generate, etc.)
```

**Structure Decision**: Provider-centric repo with generated SDKs and integration fixtures; feature work touches provider/, integration-tests/, sdk/ regeneration, and specs/ for docs.

## Testing Commitments

- Unit tests in `provider/pkg/provider` keep public API signatures stable and validate talosctl command generation, flags, and arguments.
- Integration test covering migration path: start from pulumiverse-backed state, apply talosctl-only migration steps, and validate lifecycle/kubeconfig without pulumiverse artifacts.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
