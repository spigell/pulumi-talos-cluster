# Tasks: Drop pulumiverse dependency for talosctl

**Input**: Design documents from `/specs/001-drop-pulumiverse/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Integration coverage is REQUIRED for cluster lifecycle/provider contract changes; include unit/lint tasks by default. Additional tests may be added when explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and shared helpers for talosctl/Stash flows

- [ ] T001 [P] Scaffold talosctl executor wrapper for command provider invocations in `provider/pkg/provider/applier/talosctl_exec.go`
- [ ] T002 [P] Add ClusterSecrets/ClusterConfig stash helper types in `provider/pkg/provider/types/stash.go`
- [ ] T003 Add logging adapter to separate stdout/stderr for talosctl commands in `provider/pkg/provider/applier/logging.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core plumbing required before user stories

- [ ] T004 Wire provider config to accept operator-supplied talosctl path (no version enforcement) with PATH fallback in `provider/pkg/provider/provider.go`
- [ ] T005 Add talosctl path/arch detection and guidance emission in `provider/pkg/provider/applier/talosctl_exec.go`
- [ ] T006 Update talosctl guidance and recommended version/arch matrix (non-enforcing) in `specs/001-drop-pulumiverse/contracts/migration-guide.md`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Migrate existing stacks off pulumiverse (Priority: P1) 🎯 MVP

**Goal**: Existing stacks migrate from pulumiverse resources to talosctl-only flows without downtime or unintended deletes.

**Independent Test**: Starting from a stack containing pulumiverse resources, follow migration guide to switch to talosctl-only apply/bootstrap; stack operations succeed, no pulumiverse downloads, kubeconfig/talosconfig preserved.

### Tests for User Story 1

- [ ] T007 [P] [US1] Add provider unit tests for API signature stability, talosctl command generation, and Stash wiring in `provider/pkg/provider/applier/apply_test.go`
- [ ] T008 [P] [US1] Add migration integration test (pulumiverse state → talosctl-only) in `integration-tests/migration/migration_test.go` with fixture `integration-tests/testdata/programs/migration-talosctl-go/`

### Implementation for User Story 1

- [ ] T009 [US1] Implement pulumiverse state detection and blocking with remediation messaging in `provider/pkg/provider/apply.go`
- [ ] T010 [US1] Implement talosctl secrets/config generation using ClusterSecrets/ClusterConfig Stash in `provider/pkg/provider/applier/apply.go`
- [ ] T011 [US1] Implement talosctl apply/bootstrap execution replacing pulumiverse resources with structured logging in `provider/pkg/provider/applier/apply.go`
- [ ] T012 [US1] Export talosconfig/kubeconfig from stashed config to maintain provider contract in `provider/pkg/provider/cluster.go`
- [ ] T013 [US1] Expand migration guide with prerequisites, detection/remediation steps, rollback, and validation flows in `specs/001-drop-pulumiverse/contracts/migration-guide.md`

**Checkpoint**: User Story 1 independently delivers migration without pulumiverse usage.

---

## Phase 4: User Story 2 - New installs rely on external talosctl (Priority: P2)

**Goal**: New users install provider and run lifecycle using operator-supplied talosctl on PATH with no pulumiverse downloads.

**Independent Test**: On a clean environment with talosctl on PATH, create/update/delete cluster and retrieve kubeconfig via talosctl-only flows; no pulumiverse artifacts are fetched.

### Tests for User Story 2

- [ ] T014 [P] [US2] Add integration test for clean install using external talosctl binary in `integration-tests/install/install_test.go` with fixture `integration-tests/testdata/programs/install-talosctl-go/`
- [ ] T015 [P] [US2] Add unit test ensuring default provider config prefers operator talosctl and avoids pulumiverse downloads in `provider/pkg/provider/provider_test.go`

### Implementation for User Story 2

- [ ] T016 [US2] Ensure create/update/delete flows use talosctl executor with PATH resolution and arch guidance in `provider/pkg/provider/apply.go`
- [ ] T017 [US2] Update quickstart with external talosctl onboarding and validation steps in `specs/001-drop-pulumiverse/quickstart.md`
- [ ] T018 [US2] Add logging/diagnostics for talosctl availability and arch mismatch guidance in `provider/pkg/provider/applier/talosctl_exec.go`

**Checkpoint**: User Story 2 independently delivers talosctl-only onboarding for new installs.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Repo-wide quality and documentation alignment

- [ ] T019 [P] Run lint/format/unit suite and rebuild provider artifacts as needed via `make lint` and `make unit_tests` from repo root
- [ ] T020 [P] Align documentation references (plan.md, research.md, migration-guide.md) to reflect non-enforcing talosctl guidance in `specs/001-drop-pulumiverse/`
- [ ] T021 Validate quickstart steps against a sample stack and record notes in `specs/001-drop-pulumiverse/review.md`

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → User stories (Phase 3 then Phase 4) → Polish.
- User Story order: US1 (migration) must complete before US2 to reuse stable talosctl/Stash plumbing and migration learnings; US2 can start after foundational work but should not block US1 validation.

### User Story Dependencies

- US1 depends on Phase 2 completion; no dependency on other stories.
- US2 depends on Phase 2 completion; should follow US1 to ensure migration paths are stable before onboarding net-new installs.

### Within Each User Story

- Tests first (fail), then implementation tasks.
- Models/helpers before command wiring; logging/diagnostics last.

### Parallel Opportunities

- Setup tasks T001–T002 can run in parallel.
- Tests T007/T008 (US1) and T014/T015 (US2) are parallelizable once prerequisites land.
- Documentation tasks (T013, T017, T020–T021) can run in parallel with code once respective flows are defined.

---

## Parallel Example: User Story 1

```bash
# In parallel after foundational work:
# 1) Add migration integration test while implementing apply/bootstrap logic.
# 2) Write unit tests for talosctl command generation + Stash wiring.
# 3) Expand migration guide with detection/rollback details.
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Implement User Story 1 (migration) and validate integration test passes without pulumiverse downloads.
3. Pause for review/demo; migration guide becomes deliverable MVP.

### Incremental Delivery

1. Deliver US1 migration path (P1).
2. Deliver US2 new-install path (P2) once migration is stable.
3. Finish polish tasks and doc alignment.

