# Repository Guidelines

## Project Structure & Modules
- `provider/`: Pulumi component source and CLI binaries (`pulumi-resource-talos-cluster`, codegen under `cmd/`, shared packages in `pkg/`).
- `sdk/`: Generated language SDKs (`go`, `dotnet`, `nodejs`, `python`); rebuilt via Make targets.
- `integration-tests/`: End-to-end and fixture programs; `testdata/programs` holds example Pulumi apps for clouds (e.g., `hcloud-go`, `hcloud-ha-go`).
- `deploy/`: Environment presets and helper assets; `bin/` holds built binaries.

## Build, Test, and Development Commands
- `make build && make install_provider`: Generate schema, build provider, and install it into `GOPATH/bin`.
- `make generate`: Regenerate all SDKs from the current schema.
- `make lint`: Run `golangci-lint` across provider and integration helpers.
- `make unit_tests`: Run Go unit tests (excludes generated/CRD packages).
- `make -C integration-tests integration_tests`: Build provider and SDKs, then run long E2E suites (set `TEST=<regex>` to scope, e.g., `TEST=TestHcloud make -C integration-tests integration_tests_go`).
- Node SDK iteration: `make build_nodejs_sdk` then `yarn link --cwd sdk/nodejs/bin` for local use.

## Coding Style & Naming Conventions
- Go code must be `gofmt`-clean and pass `golangci-lint`; prefer idiomatic interfaces and explicit error wrapping (`github.com/pkg/errors`).
- Keep package names lower_snake for generated SDKs and lowerCamel for Go vars/functions unless exported.
- Use concise, imperative names for Make targets and tests (e.g., `TestApplyClusterSpec`).
- Avoid editing generated SDK output directly; regenerate via Make if schema changes.
- TypeScript: prefer `type` aliases over `interface` unless interface merging is required.
- Schema validation: when a JSON Schema validates inputs, avoid repeating the same checks in code unless absolutely necessary; rely on the validated shape for type assertions.
- Avoid generic coercion helpers (e.g., `toString(any) string`); prefer explicit typed access after validation or straightforward type assertions.
- Avoid tiny pass-through helpers; if a helper just wraps a single function call without adding clarity, inline the call instead.

## Versions sync
- Pulumi upgrades:
  1) Update Go modules: `go get github.com/pulumi/pulumi/sdk/v3@<version>` in `provider/`, `integration-tests/`, `sdk/`, and Go test programs (e.g., `integration-tests/testdata/programs/hcloud-go`, `hcloud-ha-go`), then `go mod tidy` in each. Also bump `github.com/pulumi/pulumi/pkg/v3@<version>` in those same modules to avoid mismatch errors.
  2) Update Node dependencies: bump `@pulumi/pulumi` in `integration-tests/package.json` and JS test programs (e.g., `integration-tests/testdata/programs/hcloud-js/package.json`), then run `yarn install` to refresh locks.
  3) Update Python requirements: set `pulumi==<version>` in `integration-tests/pyproject.toml` and Python test programs (e.g., `integration-tests/testdata/programs/hcloud-ha-py/requirements.txt`).
  4) Update `.pulumi.version` to the same version you just bumped.
  5) Tidy modules/checksums after bumps: `go mod tidy` in `provider/` and `integration-tests/`, then `go work sync` at repo root to refresh `go.sum`/`go.work.sum`. If stale Pulumi versions linger in `go.work.sum`, delete the file and run `go list -m all` (then `go work sync`) to regenerate it.
  6) After version bumps, regenerate SDKs with `make generate_schema && make generate && make build` (requires `pulumictl` and `pulumi` on PATH).
- Pulumi/command provider:
  1) pulumi/command should be pinned across schema generation and provider usage; regenerate SDKs via the standard schema/generate/build pipeline when upgrading.
- Talos SDK:
  1) Bump `github.com/siderolabs/talos/pkg/machinery` to the target Talos release in `provider/` and tidy.
  2) Packer: when upgrading Talos, bump versions in `integration-tests/packer/hcloud-talos.pkr.hcl`.

## Testing Guidelines
- **Framework**: The testing framework uses Go's standard `testing` package with `stretchr/testify` helpers. The integration tests, located in `integration-tests/`, are written in Go and orchestrate deployments of Pulumi programs written in various languages (Go, Python, Node.js).
- **Unit Tests**:
  - Run with: `make unit_tests`
  - These are standard Go tests located in the `provider/` directory. They focus on testing individual components of the provider and exclude generated code.
- **Remote runner**: Integration tests can be executed via the remote `pulumi-talos-cluster-mcp` server using the `shell_execute` tool; provide `command`, `timeout`, and `directory` fields.
- **Integration Tests**:
  - Run all: `make -C integration-tests integration_tests`
  - These are end-to-end tests that deploy real infrastructure. They can be time-consuming (default timeout is 25m).
  - You can run tests for a specific language SDK, for example: `make -C integration-tests integration_tests_go` or `make -C integration-tests integration_tests_nodejs`.
  - To run a specific test case, use the `TEST` variable: `TEST=TestHcloud make -C integration-tests integration_tests_go`.
  - Test programs (the Pulumi apps being deployed) are located in `integration-tests/testdata/programs/`.
  - Shared test helper code is in `integration-tests/pkg/`.
**Note**: Only scoped test are allowed to run via agents.

- **Prerequisites**: Integration tests require credentials for the target cloud provider and a working `talosctl` configuration.
- **Layout**: The test helpers contain spec files for different languages (`spec.go`, `spec.py`, `spec.ts`). These implementations share a common `schema.json` source of truth to ensure consistent validation logic across all supported languages.
- **Architecture**:
  - The `cluster` package serves as the main entrypoint for tests. It orchestrates calls to the `cloud` and `talos` packages.
  - The `cluster` validation logic only covers the basic and common cluster specification.
  - Default values and specific validation for `talos` and `cloud` configurations are handled within their respective packages.
- **Python helper modules**: Shared Python helpers now live under `integration-tests/pkg/cluster/python` as packages (see `__init__.py`). New helper directories (e.g., `integration-tests/pkg/talos/python`) should also be proper packages (`__init__.py`) and added to the codebase; Pyright resolves them because `pyrightconfig.json` points `extraPaths` at `integration-tests/pkg`. Keep that root path in `extraPaths` so new packages continue to resolve.
- **Talos provider pinning**: The schema generator pins `pulumiverse-talos` for Python to `==0.6.1` (aligned with Talos 1.11.5). When bumping Talos, change the version in `provider/cmd/pulumi-gen-talos-cluster/main.go` (language.python.requires), then regenerate schema/SDKs so `provider/cmd/pulumi-resource-talos-cluster/schema.json` and `sdk/python/pyproject.toml` pick up the new pin.

### Configuration File Validation (`cluster.yaml`)

When loading and validating configuration files like the cluster specification YAML, a consistent, multi-language pattern is enforced to ensure robustness and maintainability. This pattern is implemented for TypeScript, Go, and Python in `integration-tests/pkg/cluster/`.

**Key Principles:**

*   **Single Source of Truth**: A canonical `schema.json` file, located in the `integration-tests/pkg/cluster/` directory, serves as the single source of truth for the configuration's structure, types, and constraints.
*   **Strict Schema**: The schema should be "strict." This means `additionalProperties` should be set to `false` to prevent unknown fields, and all properties should be explicitly listed as `required` where appropriate.
*   **Two-Phase Validation**: A two-phase validation approach is used:
    1.  **Schema Validation**: Raw data parsed from YAML is first validated against the `schema.json` using a language-specific JSON Schema validator (`ajv` for TS, `jsonschema` for Python). This catches all structural, type, and enum violations.
    2.  **Custom Logic Validation**: After schema validation passes, programmatic checks are run to enforce complex rules that cannot be expressed in JSON Schema (e.g., validating that a machine's IP address falls within a given network CIDR).
*   **Implementation**:
    *   Validation logic is isolated in a dedicated module (e.g., `validation.ts`).
    *   Validation functions should throw an exception (or return an error in Go) on failure and return nothing on success.
    *   The loader/spec file (e.g., `spec.ts`) orchestrates the process:
        1. Parse the YAML file into a generic object (e.g., `Record<string, unknown>` or `dict`).
        2. Call the validation function with the generic object.
        3. If validation succeeds, normalize the validated object into strongly-typed language-specific structs or classes.

This pattern ensures that any data used in the integration tests is guaranteed to be structurally and logically correct.

## Commit & Pull Request Guidelines
- Commit messages follow short, imperative style with optional scope (`chore: gofmt integration test helpers`, `refactor hcloud spec`); do not use emoji. The emoji is used by developers only.
- PRs should explain intent, link issues, and note any infra prerequisites. Include run results for `make lint`, `make unit_tests`, and any scoped integration tests.
- For schema or SDK changes, mention regenerated artifacts and the `VERSION` used; prefer separate commits for regeneration vs. release steps.

## Security & Configuration Tips
- Development assumes Linux with `bash`, `printf`, `talosctl`, Go toolchain, Pulumi CLI, and `pulumictl`.
- Avoid committing cloud credentials; rely on environment variables or profiles consumed by integration tests.
- When debugging provider binaries, `make start_delve` launches `dlv` headless on port 2345 for remote attach. You can also connect directly with `PAGER=cat dlv connect pulumi-workbench-delve:2345`; prepare a dlv init script to automate any interactive commands.
- Do not add `pulumi-talos-cluster-integration-tests-infra` to `package.json`; Pulumi installs projects in a temp directory with relative paths and the dependency fails to resolve there.

## Provider Bootstrap Process
The following diagram illustrates the sequential bootstrap process for the `Apply` resource when `skipInitApply` is false.

```mermaid
flowchart TD
    Start["Apply Resource Create/Update"] --> LeaderStart["Bootstrap Leader"]

    subgraph LeaderPhase["Leader Control Plane (init)"]
        InitCP["Select First/Leader CP Node"]
        ApplyLeader["talosctl apply-config --insecure (leader)"]
        RebootLeader["Leader Installs & Reboots"]

        LeaderStart --> InitCP --> ApplyLeader --> RebootLeader
    end

    RebootLeader --> CheckSkip{"skipInitApply?"}

    CheckSkip -- "False" --> BootstrapStart[Bootstrap Remaining Nodes]
    CheckSkip -- "True" --> Done

    subgraph BootstrapLoop["Bootstrap Loop (Sequential)"]
        direction TB

        subgraph CPPhase["1. Apply to Other Control Planes"]
            IterateCP["For Each Remaining CP"]
            ApplyConfigCP["talosctl apply-config --insecure"]
            RebootCP["Node Installs & Reboots"]

            IterateCP --> ApplyConfigCP --> RebootCP
            RebootCP -- "Next" --> IterateCP
        end

        subgraph WorkerPhase["2. Apply to Workers"]
            IterateWork["For Each Worker"]
            ApplyConfigWork["talosctl apply-config --insecure"]
            RebootWork["Node Installs & Reboots"]

            IterateWork --> ApplyConfigWork --> RebootWork
            RebootWork -- "Next" --> IterateWork
        end

        BootstrapStart --> IterateCP
        RebootCP --> IterateWork
    end

    IterateWork --> Kubeconfig["Retrieve Kubeconfig"]
    Kubeconfig --> Done["Resource Ready"]

    classDef dangerous fill:#f96,stroke:#333,stroke-width:2px;
    class RebootLeader,RebootCP,RebootWork dangerous;
```

## MCP Shell Server & Tooling

This environment provides a specialized MCP server (`pulumi-talos-cluster-mcp`) with the `shell_execute` tool for performing safe, remote operations within the workspace.

**Configuration:**
- **Tool:** `shell_execute`
- **Default Directory:** `/project/workspace-pulumi/pulumi-talos-cluster` (Always use absolute paths)
- **Timeouts:** Mandatory. Use reasonable limits (e.g., 300s for quick checks, 1800s for integration tests).

**Allowed Commands:**
`cat`, `find`, `go`, `grep`, `ls`, `make`, `pulumi`, `pwd`, `talosctl`, `touch`, `wc`.

**Capabilities & Workflows:**

1.  **Build & Test:**
    - **Provider Build:** `["make", "build"]`
    - **Full SDK regen/install pipeline**: run `make generate_schema`, `make generate`, `make build`, then install the SDK as needed (e.g., `make install_nodejs_sdk`).
    - **Unit Tests:** `["make", "unit_tests"]`
    - **Integration Tests:** Run via make targets. Only scoped mode is allowed.
      - Scoped: `["make", "-C", "integration-tests", "integration_tests_go", "TEST=<TestName>"]`.
    - There are a skill called pulumi-integration-tests-runner with additional context. You should ask to enable it if you think should use the skill.

2.  **Talosctl Operations:**
    - The environment in the MCP shell server has `talosctl` configured.
    - Automation logic resides in the skill called talosctl-operator. You should ask if you think you need the skill.

3.  **Debugging (Delve):**
    - A headless Delve server runs at `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.
    - Connect using a local `dlv` client from the repo root.
    - *Note:* The `dlv` command is run locally (not via `shell_execute`), but interacts with the provider process managed within this environment.
    - Always use the skill delve-debuger which holds context about the operations.
