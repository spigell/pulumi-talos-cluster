# Review: Drop pulumiverse dependency for talosctl

**Reviewer**: Gemini CLI (Pair Review)
**Date**: 2026-01-13
**Subject**: `specs/001-drop-pulumiverse/spec.md`

## Summary
The specification correctly identifies the need to move away from Pulumiverse for reliability. However, several critical technical risks and requirement misalignments were identified during the pair review.

## Findings

### 1. External Dependency vs. Bundling (User Story 2 & FR-002)
- **Issue**: The original spec (FR-002) suggested "packaging" or "bundling" `talosctl`. 
- **Correction**: The provider should **not** install or manage `talosctl`. It must rely on an externally provided binary (e.g., on the `$PATH`) and only perform version validation.
- **Impact**: Reduces provider complexity and respects user environment management.

### 2. State Migration & Resource Destruction Risk (User Story 1)
- **Issue**: Removing Pulumiverse resources from Pulumi code will trigger a `Delete` operation in the engine. Without careful migration (e.g., state surgery or `retainOnDelete`), this will destroy the actual cluster during migration.
- **Recommendation**: The Migration Guide (FR-005) must explicitly address how to safely remove Pulumiverse resources from the state without affecting the live infrastructure.

### 3. Configuration Idempotency
- **Issue**: Switching to `talosctl` binary calls for config generation (e.g., `talosctl gen config`) may lead to secret/certificate regeneration on every `pulumi up` if not handled correctly.
- **Recommendation**: Add a requirement to ensure that cluster secrets and configurations are persisted in the Pulumi state or a secure location to prevent unnecessary cluster reboots/updates.

### 4. Architecture Resolution
- **Issue**: `talosctl` execution depends on the host OS/Arch of the Pulumi runner.
- **Recommendation**: Add an edge case for handling environments where the runner's architecture differs from the target nodes or where multiple architectures are used in CI/CD.

### 5. Remove User Story 3 (Troubleshooting & Rollback)
- **Finding**: While important, documented troubleshooting is a standard expectation of FR-005 (Migration Guide) and doesn't require a standalone P3 User Story for the feature MVP.
- **Recommendation**: Remove User Story 3 to focus the specification on the core migration and operational changes.

### 6. FR-005 Rollback Specificity
- **Issue**: "Rollback paths" is vague.
- **Recommendation**: Explicitly define the rollback path in the Migration Guide requirement as "restoring a backup of the Pulumi state" (e.g., `pulumi stack import`), as reverting code changes alone may not suffice after state mutations.

### 7. Remove FR-007 (Automated Checks)
- **Finding**: Requirements for automated checks (lint/unit/integration) are standard engineering practices or implementation details, not functional requirements of the user-facing feature.
- **Recommendation**: Remove FR-007 from the functional requirements list.

### 8. Remove FR-008 (Checksum Publication)
- **Finding**: Since the decision was made (see Finding #1) to rely on an external `talosctl` binary rather than bundling it, the provider is not responsible for publishing checksums or signatures for that binary.
- **Recommendation**: Remove FR-008.

### 9. Adjust Success Criteria
- **Finding**: Existing criteria are either overly specific (SC-002 timing), redundant with standard practices (SC-003), or vague (SC-004 SLOs).
- **Recommendation**: 
    - Remove SC-002, SC-003, and SC-004.
    - Add a new Success Criterion: "A verified migration guide is available covering simple usage scenarios."

## Subject: `specs/001-drop-pulumiverse/checklists/requirements.md`

### 10. Checklist Accuracy
- **Finding**: Several items are marked as complete (`[x]`) but contradict findings in the `spec.md` review.
    - **"No implementation details"**: FR-007 and FR-008 were flagged as implementation details.
    - **"Requirements are testable and unambiguous"**: FR-005 was found to be vague regarding rollback.
    - **"Edge cases are identified"**: Critical edge cases (Arch mismatch, Idempotency) were missing.
- **Recommendation**: Update the checklist to `[ ]` for these items until the `spec.md` is updated to resolve the review findings.

## Constitutional Impact
- **Finding**: Core Principle V currently mandates alignment with `pulumiverse/talos`.
- **Action Required**: Upon ratification of Feature 001, the Constitution MUST be amended to remove this reference. This will trigger a Constitution version bump (likely MINOR or PATCH depending on interpretation).

## Missing Sections
- **Motivation**: The "Why" is currently only in the "Input" field. A dedicated Motivation section would clarify the specific pain points (reliability, installation friction, execution control).

## Next Steps
- Update `spec.md` based on these findings (to be done by the user or in a subsequent step).
- Refine the Migration Guide requirements (FR-005).
