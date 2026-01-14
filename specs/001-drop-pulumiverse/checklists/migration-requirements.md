# Requirements Quality Checklist: Drop pulumiverse dependency for talosctl

**Purpose**: Unit-test the migration and install requirements for clarity, completeness, and measurability before implementation.
**Created**: 2026-01-14
**Feature**: specs/001-drop-pulumiverse/spec.md

## Requirement Completeness

- [x] CHK001 Are prerequisites for operator-run migration (talosctl on PATH, supported arch/matrix, Pulumi version) fully enumerated in the migration guide? [Completeness, Spec §FR-002, FR-005]
- [x] CHK002 Does the migration guide specify all mandatory backups (stack export location/validation) before changes? [Completeness, Spec §FR-005]
- [x] CHK003 Are steps to remove/disable pulumiverse dependencies documented end-to-end (state, provider config, code references)? [Completeness, Spec §FR-001, FR-004]

## Requirement Clarity

- [x] CHK004 Are manual talosctl commands in the guide specified with exact flags/inputs for secrets/config generation and apply/bootstrap? [Clarity, Spec §FR-003, FR-005]
- [x] CHK005 Is “blocking error” for residual pulumiverse usage defined with explicit detection criteria and required operator action? [Clarity, Spec §FR-004]
- [x] CHK006 Is guidance on using recommended talosctl versions/architectures written as non-enforcing operator instructions with measurable expectations? [Clarity, Spec §FR-002, Plan Constraints]

## Requirement Consistency

- [x] CHK007 Are migration guide instructions consistent with success criteria SC-001/SC-002 regarding “no pulumiverse downloads” and verified rollback? [Consistency, Spec §SC-001/SC-002]
- [x] CHK008 Does quickstart guidance align with migration steps (backups, detection, talosctl use) without conflicting prerequisites? [Consistency, Quickstart]

## Acceptance Criteria Quality

- [x] CHK009 Are independent test/validation steps in the migration guide measurable (e.g., observable logs, absence of pulumiverse artifacts, kubeconfig/talosconfig preserved)? [Acceptance Criteria, Spec §FR-005, SC-002]
- [x] CHK010 Is rollback success defined in terms of restored state and service availability after `pulumi stack import`? [Acceptance Criteria, Spec §FR-005]

## Scenario Coverage

- [x] CHK011 Are in-progress operations and partial migration scenarios addressed with guidance (pause/retry/abort paths)? [Coverage, Edge Case]
- [x] CHK012 Are post-migration lifecycle actions (create/update/delete/kubeconfig retrieval) covered in manual validation steps to ensure talosctl-only flows? [Coverage, Spec §FR-003]

## Edge Case Coverage

- [x] CHK013 Are version/arch mismatches, mixed-arch runners, and missing talosctl binaries explicitly handled with operator guidance? [Edge Case, Spec Edge Cases, Plan Constraints]
- [x] CHK014 Are handling instructions provided for residual pulumiverse resources in state to prevent accidental deletes? [Edge Case, Spec §FR-004]
- [x] CHK015 Is secret/config persistence (avoid regeneration) addressed with Stash or equivalent persistence guidance during manual steps? [Edge Case, Spec §FR-007]

## Non-Functional Requirements

- [x] CHK016 Are logging/observability expectations (stdout/stderr separation, redaction, success/failure signals) documented for operator runs? [Non-Functional, Spec §FR-006, Plan Observability]
- [x] CHK017 Are Linux-only and tooling assumptions (bash, printf, talosctl availability) explicitly stated for the migration process? [Non-Functional, Plan Constraints]

## Dependencies & Assumptions

- [x] CHK018 Are external dependencies (Pulumi CLI version, talosctl binary source/provenance, cloud access) and their validation steps documented? [Dependencies, Spec §FR-002, FR-005]
- [x] CHK019 Are assumptions about existing cluster state (schema-valid config, prior pulumiverse resources) called out with required pre-checks? [Assumption, Spec Edge Cases]

## Ambiguities & Conflicts

- [x] CHK020 Is the non-enforcing talosctl version guidance clearly distinguished from any prior version validation statements to avoid conflict? [Ambiguity, Plan Constraints, Spec §FR-002]
- [x] CHK021 Are terms like “zero-downtime migration” scoped (what downtime is acceptable, if any) or flagged as needing explicit definition? [Ambiguity, Spec Motivation]

## Notes

- Check items off as completed: `[x]`
- Add comments or findings inline
- Each checklist run creates a new file; retain for traceability
