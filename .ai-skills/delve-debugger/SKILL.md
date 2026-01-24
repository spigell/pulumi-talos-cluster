---
name: delve-debugger
description: Practical playbook for fully debugging the talos-cluster provider with Delve and pdebug, including a multi-agent workflow.
---

## Overview
- Goal: attach to the remote Delve server, trigger `pdebug.sh <command>` to reproduce pulumi run, and collect backtraces/locals without letting the session drop.
- Multi-agent flow: one worker in coordination and you (Orcestrator)
  - Agent 1 (debugger): attach to Delve, run `rebuild`/`continue`, stay attached, capture panic data.
  - Orcestrator: run `pdebug.sh pre` (or other pdebug command) via pulumi-talos-cluster-mcp with dir to trigger the crash while Agent 1 is attached.

## Environment
- Use local `dlv` client; commands run from repo root. Use a temp init file for commands to avoid `--init` file-missing issues.
- Delve server: `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.

## Limitations
- DO NOT USE BACKGROUND TERMINAL
- Before starting a session with rebuild, try rebuilding the provider binary manually (e.g., `make build_provider`) via mcp env and resolve any errors; only then attach and issue `rebuild`/`continue` in Delve.
- If asked to keep the server running, answer “n” when Delve prompts “Would you like to kill the headless instance?”.
- If Delve asks about killing the headless instance, answer `n` to keep it running.
- **Breakpoints:** Do not use the `break` command. Insert `runtime.Breakpoint()` in your Go code and rebuild.


## Catching a Panic
1) Agent 1: attach and leave Delve waiting (do **not** quit).
```bash
printf 'rebuild\ncontinue\n' > /tmp/dlv_cmds_agent1
dlv connect <server> --init=/tmp/dlv_cmds_agent1
```
2) Orcestrator: run `/project/deploy/workbench/pdebug.sh pre` via pulumi-talos-cluster-mcp from the target program dir, e.g. `{dir: integration-tests/testdata/programs/<program>}`:
```bash
/project/deploy/workbench/pdebug.sh pre
```
3) When Delve stops on panic (Agent 1), collect:
```bash
bt
goroutines
goroutine <id> bt   # pick the crashing goroutine id
locals
args
```
4) If needed, inspect variables (Agent 1):
```bash
print <expr>
whatis <expr>
```
5) To resume after inspection (Agent 1):
```bash
continue
```

## Common Symptoms & Actions
- **`plugin ... did not begin responding to RPC connections`**: Provider is not started. Ensure Agent 1 is attached and has issued `rebuild/continue`; rerun Agent 2.
- **`EOF` / resource monitor shut down**: Provider crashed; reattach Agent 1 with rebuild and rerun Agent 2 to capture the panic backtrace.
- **`--init` file missing**: Always create the temp file under `/tmp` before `dlv connect`.
