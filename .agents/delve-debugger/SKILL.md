---
name: delve-debugger
description: Practical playbook for debugging the talos-cluster provider with Delve using the background terminal and pdebug to capture panics.
---

## Overview
- Goal: attach to the remote Delve server, trigger `pdebug.sh <command>` to reproduce a Pulumi run, and collect backtraces/locals without dropping the session.
- Use the background terminal for `dlv` attachment; orchestrator can trigger `pdebug.sh` via MCP from the target program directory.

## Environment
- Use the local `dlv` client; run commands from repo root. Always create an init file under `/tmp` before `dlv connect` to avoid `--init` missing errors.
- Delve server: `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.
- If prompted “Would you like to kill the headless instance?”, answer `n` to keep it running.
- Breakpoints: do not use `break`; add `runtime.Breakpoint()` in code and rebuild if a breakpoint is required.

## Quick Attach (background terminal)
1) Prepare init file (rebuild + continue):
```bash
printf 'rebuild\ncontinue\n' > /tmp/dlv_cmds_full
```
2) Attach in a TTY background terminal:
```bash
dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds_full
```
3) Stay attached; do not quit. If `Unknown process id` appears, recreate the init file and reconnect.

## Catching a Panic
1) Attach and leave Delve waiting (do **not** quit) using the quick-attach steps.
2) From the target program directory, trigger the provider via MCP:
```bash
/spigell-reforge-ai/deploy/workbench/pdebug.sh pre
```
   Example directory: `integration-tests/testdata/programs/<program>`.
3) When Delve stops on panic, collect:
```bash
bt
goroutines
goroutine <id> bt   # crashing goroutine
goroutine <id> locals
goroutine <id> args
```
4) Optional inspections:
```bash
print <expr>
whatis <expr>
```
5) Resume after inspection:
```bash
continue
```

## Debugging an Active State (background terminal)
- If the process is already running and you need to inspect its current state:
  1) Create an empty init file: `printf '' > /tmp/dlv_cmds_capture`.
  2) Attach in a TTY background terminal:  
     `dlv connect <server> --init=/tmp/dlv_cmds_capture`
  3) Inspect the state without restarting:
     - `threads` (or `goroutines`) to find the interesting goroutine.
     - `bt` or `goroutine <id> bt` for stack.
     - `goroutine <id> locals` / `args` or `print <expr>` for values.
  4) Use `continue` only when you’re ready to let the process run again.
  5) If the session drops (`Unknown process id`), recreate the init file and reconnect.

## Common Symptoms & Actions
- `plugin ... did not begin responding to RPC connections`: ensure the `dlv` session issued `rebuild/continue`, then rerun `pdebug.sh`.
- `EOF` / resource monitor shut down: reattach with the init file and rerun `pdebug.sh`.
- `--init file missing`: create the file under `/tmp` before connecting.
