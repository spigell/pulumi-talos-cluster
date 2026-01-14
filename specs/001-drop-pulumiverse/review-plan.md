# Review: Implementation Plan (Drop pulumiverse)

**Reviewer**: Gemini CLI (Self-Review)
**Date**: 2026-01-13
**Subject**: `specs/001-drop-pulumiverse/plan.md`
**Reference**: `specs/001-drop-pulumiverse/spec.md`, `specs/001-drop-pulumiverse/research_flow.md`, `specs/001-drop-pulumiverse/research_stash.md`

## Summary
The implementation plan accurately reflects the goal of replacing the Pulumiverse provider with a `talosctl` + `pulumi.Stash` based workflow. It correctly identifies the technical context, dependencies, and constraints.

## Findings

### 1. Alignment with Research
*   **Observation**: The plan's "Summary" and "Technical Context" sections explicitly mention "Pulumi Stash" and "command provider resources," which aligns perfectly with the findings in `specs/001-drop-pulumiverse/research_flow.md`.
*   **Verdict**: **Aligned**. The plan effectively operationalizes the research decisions.

### 2. Dependency Management
*   **Observation**: The plan lists "Pulumi Stash" as a primary dependency.
*   **Verification Needed**: Does the Pulumi Go SDK (`pulumi/sdk/v3`) expose `Stash` directly, or is it a resource within a specific provider (like `command`) or a core primitive that needs specific invocation?
    *   *Self-Correction/Note*: `Stash` is a concept/resource likely found in newer Pulumi versions or specific patterns. The plan acknowledges `command` provider as well. The implementation phase will need to confirm the exact Go package path for `Stash` or its functional equivalent if it's a new feature.
*   **Verdict**: **Acceptable**, but requires implementation-time verification of `Stash` availability in the pinned SDK version.

### 3. Testing Strategy
*   **Observation**: The plan lists `make unit_tests`, `make lint`, and `integration tests`.
*   **Gap**: The plan mentions "migration guide validation." It should explicitly state that *new* integration tests or modified existing tests will be required to verify the *migration path* itself (e.g., a test that starts with the old provider and upgrades to the new one).
*   **Recommendation**: Ensure the task list (next step) includes a specific task for "Create/Modify integration test for migration scenario."

### 4. Constitution Check
*   **Observation**: The Constitution Check section is marked with green checks (✅).
*   **Validity**:
    *   *Determinism*: Moving to `talosctl` + `Stash` maintains determinism by persisting the generated state. Correct.
    *   *Security*: `talosctl` on PATH prevents checking in binaries. `Stash` encrypts secrets. Correct.
    *   *Version Discipline*: The plan mandates validation of the external `talosctl` version. **Correction Required**: The user has requested to *skip* validation against the CLI version matrix. The plan should be updated to reflect that while we require `talosctl`, we will not enforce strict version checks in code.
*   **Verdict**: **Passes (with modification)**.

### 7. Code Contract & Unit Testing
*   **Observation**: The provider implements Component Resources (`Cluster` and `Apply`) in Go. The migration must replace the internal implementation (switching from Pulumiverse to `talosctl`) while maintaining the existing Go API contract (`NewCluster` and `NewApply` functions).
*   **Requirement**: Unit tests are strictly required in `provider/pkg/provider` to verify that:
    1.  The public API signatures remain backward compatible.
    2.  The new implementation correctly orchestrates `talosctl` commands and `Stash` resources.
    3.  Command generation logic (flags, arguments) is correct.
*   **Action**: Add a task to implement/update unit tests for `NewCluster` and `NewApply` ensuring contract adherence.

### 5. Project Structure
*   **Observation**: The structure correctly identifies the key directories (`provider/`, `sdk/`, `integration-tests/`, `specs/`).
*   **Verdict**: **Correct**.

## Conclusion
The plan is solid and ready for task decomposition. It correctly incorporates the architectural shift to `talosctl` and `Stash` while respecting the project's constraints.

## Next Steps
*   Proceed to `/speckit.tasks` to break down the work items.
*   Ensure tasks specifically cover:
    1.  Prototyping the `Stash` + `talosctl` interaction.
    2.  Implementing the `ClusterSecrets` and `ClusterConfigurations` logic (as per `research_flow.md`).
    3.  Updating integration tests to remove Pulumiverse references and validate the new flow.
    4.  Writing the Migration Guide.
