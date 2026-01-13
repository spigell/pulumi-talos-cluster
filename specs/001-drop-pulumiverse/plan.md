# Implementation Plan: Drop pulumiverse dependency for talosctl

**Branch**: `001-drop-pulumiverse` | **Date**: 2026-01-12 | **Spec**: specs/001-drop-pulumiverse/spec.md
**Input**: Feature specification from `/specs/001-drop-pulumiverse/spec.md`

**Note**: Filled per /speckit.plan workflow.

## Summary

Transition provider and workflows to rely solely on operator-supplied talosctl binaries, removing pulumiverse runtime dependencies, and deliver a migration guide with rollback and validation steps.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.21+ provider, Node/Python/TS SDK artifacts generated via Pulumi toolchain  
**Primary Dependencies**: Pulumi SDK v3, talosctl (external, operator-provided), pulumiverse removal, provider toolchain (pulumictl, golangci-lint)  
**Storage**: Pulumi state and Talos configs persisted via stack exports/imports; no app DB  
**Testing**: make unit_tests, make lint, integration tests for cluster lifecycle where needed  
**Target Platform**: Linux runners with talosctl on PATH  
**Project Type**: Pulumi provider + integration test suites  
**Performance Goals**: Zero-downtime migration expectation; talosctl latency acceptable for lifecycle operations (minutes)  
**Constraints**: No pulumiverse runtime usage; idempotent talosctl operations; no credential logging  
**Scale/Scope**: Standard integration-test scale (small/medium clusters); document supported architectures/version matrix

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Determinism: talosctl-only flows with state export/import guidance keep Pulumi/Talos configs authoritative. ✅  
- Security: No secrets or kubeconfigs committed/logged; Linux runners with external talosctl only. ✅  
- Testing: gofmt, lint, unit tests planned; add integration coverage for lifecycle sans pulumiverse. ✅  
- Observability: Require clear success/failure logging with stderr separation and artifact preservation. ✅  
- Version Discipline: Document talosctl support matrix; maintain Pulumi/Talos pins and regenerate SDKs when touched. ✅

## Project Structure

### Documentation (this feature)

```text
specs/001-drop-pulumiverse/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
provider/          # Go provider source and binaries
sdk/               # Generated SDKs (go/dotnet/nodejs/python)
integration-tests/ # Integration suites and fixtures
specs/001-drop-pulumiverse/ # Feature docs (spec, plan, research, contracts)
deploy/            # Build/deploy assets
```

**Structure Decision**: Use existing provider + SDK + integration-tests layout; feature documentation under specs/001-drop-pulumiverse.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
