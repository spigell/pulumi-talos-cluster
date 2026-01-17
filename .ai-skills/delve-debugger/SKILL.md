name: delve-debugger
description: How to attach to the remote Delve server, inspect state, set breakpoints, and continue execution.

## Overview
- Delve server runs headless at `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.
- Use local `dlv` client; commands run from repo root.
- Rebuilds restart the process; breakpoints are lost and must be re-set.

## Common Commands
- Attach + run one or more commands from a file:
  - `printf 'goroutines\nexit\n' > /tmp/dlv_cmds`
  - `dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds`
- Quick inline command (fails if file missing): `dlv connect <addr> --init='goroutines'`
- Set a breakpoint:
  - `printf 'break provider/pkg/provider/applier/talosctl_gen.go:16\nexit\n' > /tmp/dlv_cmds`
  - `dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds`
- Continue from current stop:
  - `printf 'continue\n' > /tmp/dlv_cmds`
  - `dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds`
- Inspect locals/args at the current frame:
  - `printf 'locals\nargs\n' > /tmp/dlv_cmds`
  - `dlv connect ... --init=/tmp/dlv_cmds`

## Notes
- If the server is rebuilt/restarted, reattach and reapply breakpoints.
- To avoid “no such file” errors with `--init`, always write commands to a temp file first.
- If asked to keep the server running, answer “n” when Delve prompts “Would you like to kill the headless instance?”.
