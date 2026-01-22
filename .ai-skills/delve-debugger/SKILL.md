---
name: delve-debugger
description: How to attach to the remote Delve server, inspect state, and control execution flow.
---

## Overview
- Delve server runs headless at `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.
- Use local `dlv` client; commands run from repo root.
- **Breakpoints:** Do not use the `break` command. Insert `runtime.Breakpoint()` in your Go code and rebuild.

## Common Commands
- Attach + run one or more commands from a file:
  - `printf 'goroutines\nexit\n' > /tmp/dlv_cmds`
  - `dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds`
- Quick inline command (fails if file missing): `dlv connect <addr> --init='goroutines'`
- Continue from current stop:
  - `printf 'rebuild\ncontinue\n' > /tmp/dlv_cmds`
  - `dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds`
- Inspect locals/args at the current frame:
  - `printf 'locals\nargs\n' > /tmp/dlv_cmds`
  - `dlv connect ... --init=/tmp/dlv_cmds`

## Notes
- To stop execution at a specific point, add `runtime.Breakpoint()` to the code and rebuild the provider.
- To avoid “no such file” errors with `--init`, always write commands to a temp file first.
- If asked to keep the server running, answer “n” when Delve prompts “Would you like to kill the headless instance?”.
- Common flow with `pulumi pre`: the user runs `pulumi pre`, which stops the provider in Delve. When asked to start a provider session, send `rebuild` then `continue` and wait indefinitely for exit; if Delve drops you back, inspect and explain the current step (locals/args/backtrace) before proceeding.
- Before starting a session, try rebuilding the provider binary manually (e.g., `make build_provider` or the project’s build step) and resolve any errors; only then attach and issue `rebuild`/`continue` in Delve.
