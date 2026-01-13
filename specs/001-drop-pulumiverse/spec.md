# Feature Specification: Drop pulumiverse dependency for talosctl

**Feature Branch**: `001-drop-pulumiverse`  
**Created**: 2026-01-12  
**Status**: Draft  
**Input**: User description: "I want to drop the dependenices of pulumiverse because it it too unreliable. I want to switch completelly in favour talosctl binary. We also need a migration guide."

## Motivation

Reliance on pulumiverse introduces reliability and installation risks. Switching to talosctl-only execution gives operators direct control over binaries, reduces external dependency failures, and simplifies troubleshooting.

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Migrate existing stacks off pulumiverse (Priority: P1)

Operators can move existing clusters from pulumiverse-based provider usage to talosctl binary driven flows without downtime.

**Why this priority**: Existing users need a reliable path away from flaky pulumiverse dependencies to maintain stability.

**Independent Test**: Follow the migration guide on a sample stack and confirm cluster remains reachable and manageable using only talosctl.

**Acceptance Scenarios**:

1. **Given** a stack using pulumiverse provider, **When** the operator follows the migration guide, **Then** the stack updates to talosctl-only operations without failed applies.
2. **Given** a migrated stack, **When** running standard lifecycle (create/update/delete), **Then** operations complete without pulumiverse provider present.

---

### User Story 2 - New installs rely on external talosctl (Priority: P2)

New users can install the provider and use an externally managed talosctl binary on the PATH without pulling pulumiverse artifacts.

**Why this priority**: Simplifies onboarding while keeping binary management under operator control.

**Independent Test**: Install provider on a clean environment, place a supported talosctl on PATH, verify cluster actions run successfully without pulumiverse packages or bundled binaries.

**Acceptance Scenarios**:

1. **Given** a clean environment with required tooling, **When** a supported talosctl is available on PATH, **Then** cluster operations run using that binary with no pulumiverse downloads.

---

[Add more user stories as needed, each with an assigned priority]

### Edge Cases

- What happens when talosctl version on PATH mismatches cluster version requirements?  
- How are architectures handled when the runner differs from target nodes or when CI uses multiple architectures?  
- How is migration handled if pulumiverse resources remain in state files to avoid unintended deletes?  
- How are secrets/configs kept stable to prevent regeneration or cluster restarts during repeated applies?  
- What occurs when existing clusters have in-progress operations during migration?

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: System MUST remove runtime dependency on pulumiverse provider packages for all cluster operations.  
- **FR-002**: System MUST rely on an externally provided talosctl binary (e.g., on PATH) and validate its version against a documented support matrix; the provider must not bundle or install talosctl.  
- **FR-003**: System MUST default all cluster lifecycle actions (create, update, delete, kubeconfig retrieval) to use talosctl only.  
- **FR-004**: System MUST detect any remaining pulumiverse usage in state or configuration and surface a blocking error with remediation steps to prevent accidental resource deletions.  
- **FR-005**: System MUST provide a migration guide covering prerequisites, step-by-step migration, validation, and rollback via Pulumi state backup/restore (e.g., `pulumi stack export/import`).  
- **FR-006**: System MUST log migration and runtime actions with clear success/failure signals without exposing credentials or secrets.  
- **FR-007**: System MUST persist or reuse generated cluster secrets/configurations to avoid unintended regeneration on repeat applies (idempotent talosctl usage).

### Key Entities *(include if feature involves data)*

- **Talosctl Binary (external)**: Operator-provided executable with validated version compatibility.  
- **Migration Guide**: Document detailing prerequisites, migration steps, validation, troubleshooting, and rollback via state backup/restore.  
- **Cluster Spec State**: Declarative cluster configuration and Pulumi state that must remain consistent through migration without unintended deletes or secret churn.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 100% of cluster lifecycle commands execute without pulling pulumiverse provider artifacts.  
- **SC-002**: A verified migration guide exists covering a reference stack end-to-end, including rollback via state backup/restore.  
