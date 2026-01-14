# Migration Checklist: Drop pulumiverse dependency for talosctl

**Purpose**: Requirements-quality checklist focused on the migration guide and runtime requirements for moving to talosctl-only usage.
**Created**: 2026-01-12
**Feature**: specs/001-drop-pulumiverse/spec.md

## Requirement Completeness

- [x] CHK001 Are prerequisites for migration (supported talosctl versions, OS/arch expectations, Pulumi CLI versions) fully enumerated before steps begin? [Completeness, Spec §FR-002, Spec §Edge Cases]
- [x] CHK002 Does the guide specify a complete backup/restore path (export/import) for Pulumi state prior to migration with rollback triggers? [Completeness, Spec §FR-005]
- [x] CHK003 Are step-by-step migration actions documented from pulumiverse removal through validation and rollback, with no implicit steps? [Completeness, Spec §FR-005]
- [x] CHK004 Are validation steps after migration defined (cluster reachability, lifecycle operations) to confirm talosctl-only operation? [Completeness, Spec §FR-001, Spec §FR-003]
- [x] CHK005 Are instructions included for detecting and remediating residual pulumiverse resources in state/config before apply? [Completeness, Spec §FR-004]

## Requirement Clarity

- [x] CHK006 Is the supported talosctl version matrix clearly stated with how compatibility is determined? [Clarity, Spec §FR-002]
- [x] CHK007 Are instructions for sourcing talosctl (PATH location, validation command, required permissions) unambiguous? [Clarity, Spec §FR-002]
- [x] CHK008 Is “idempotent talosctl usage” defined with expectations for secret reuse and no unintended regeneration? [Clarity, Spec §FR-007]
- [x] CHK009 Are success/failure signals for migration and runtime logging specified without exposing secrets/PII? [Clarity, Spec §FR-006]

## Requirement Consistency

- [x] CHK010 Do migration steps avoid conflicting guidance about retaining vs removing pulumiverse packages and state? [Consistency, Spec §FR-001, Spec §FR-004]
- [x] CHK011 Are lifecycle defaults (create/update/delete/kubeconfig via talosctl) aligned across guide narrative and requirements? [Consistency, Spec §FR-003]

## Acceptance Criteria Quality

- [x] CHK012 Are measurable acceptance criteria defined for a “successful migration” (e.g., no pulumiverse downloads, talosctl validation outcome, post-migration operations)? [Acceptance Criteria, Spec §SC-001, Spec §FR-001]
- [x] CHK013 Are rollback success criteria stated (state restored, cluster unchanged) with objective checks? [Acceptance Criteria, Spec §FR-005]

## Scenario Coverage

- [x] CHK014 Are primary migration flows covered for both existing stacks and new installs using external talosctl? [Coverage, Spec §User Story 1, Spec §User Story 2]
- [x] CHK015 Are exception/recovery paths described when migration fails mid-apply (e.g., partial state changes)? [Coverage, Spec §Edge Cases, Spec §FR-005]
- [x] CHK016 Are post-migration steady-state operations (update/delete/kubeconfig) requirements captured to ensure pulumiverse-free behavior? [Coverage, Spec §FR-003]

## Edge Case Coverage

- [x] CHK017 Does the guide address behavior when talosctl version mismatches cluster requirements or unsupported architectures are present? [Edge Case, Spec §Edge Cases]
- [x] CHK018 Are scenarios with in-progress operations or mixed architectures during migration documented with safe handling steps? [Edge Case, Spec §Edge Cases]

## Non-Functional Requirements

- [x] CHK019 Are logging/observability requirements for migration and runtime actions specified with redaction rules? [Non-Functional, Spec §FR-006]
- [x] CHK020 Are operational reliability expectations defined (e.g., no downtime during migration, idempotent re-runs) and linked to talosctl-only flow? [Non-Functional, Spec §FR-007, Spec §User Story 1]

## Dependencies & Assumptions

- [x] CHK021 Are assumptions about external talosctl availability, PATH configuration, and permissions explicitly stated and validated? [Dependency, Spec §FR-002]
- [x] CHK022 Are dependencies on Pulumi state handling (export/import tooling) and backup storage documented with verification steps? [Dependency, Spec §FR-005]

## Ambiguities & Conflicts

- [x] CHK023 Are terms like “blocking error” for residual pulumiverse usage defined with exact remediation expectations? [Ambiguity, Spec §FR-004]
- [x] CHK024 Are there any conflicting instructions between narrative and requirements about secret persistence or regeneration? [Conflict, Spec §FR-007]

## Notes

- Check items off as completed: `[x]`
- Add comments or findings inline
- Link to relevant resources or documentation
- Items are numbered sequentially for easy reference
