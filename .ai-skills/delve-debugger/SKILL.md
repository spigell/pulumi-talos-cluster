---
name: delve-debugger
description: Practical playbook for fully debugging the talos-cluster provider with Delve and pdebug.
---

## Overview
- Goal: attach to the remote Delve server, trigger `pdebug.sh <command>` to reproduce pulumi run, and collect backtraces/locals without letting the session drop.
- The flow: It should be accomplished with 2 parallel workers: debugger and pulumi-runner

## Environment
- Use local `dlv` client; commands run from repo root. Use a temp init file for commands to avoid `--init` file-missing issues.
- Delve server: `pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345`.

- Before starting a session with rebuild, try rebuilding the provider binary manually (e.g., `make build_provider`) via mcp env and resolve any errors; only then attach and issue `rebuild`/`continue` in Delve.
- If asked to keep the server running, answer “n” when Delve prompts “Would you like to kill the headless instance?”.
- **Breakpoints:** Do not use the `break` command. Insert `runtime.Breakpoint()` in your Go code and rebuild.

## Quick Attach (rebuild + continue)
```bash
printf 'rebuild\ncontinue\n' > /tmp/dlv_cmds_full

dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds_full
```
- If Delve asks about killing the headless instance, answer `n` to keep it running.

## Catching a Panic
1) Attach and leave Delve waiting (do **not** quit).
2) In another terminal, run `pdebug.sh pre` via pulumi mcp env from the target program dir, e.g.:
`{dir: integration-tests/testdata/programs/<program>}` with mcp call
```bash
/project/deploy/workbench/pdebug.sh pre
```
3) When Delve stops on panic, collect:
```bash
bt
goroutines
goroutine <id> bt   # pick the crashing goroutine id
locals
args
```
4) If needed, inspect variables:
```bash
print <expr>
whatis <expr>
```
5) To resume after inspection:
```bash
continue
```

## Keeping the Session Alive
- Avoid `quit`; if Delve drops (`Unknown process id`), reconnect with the quick-attach snippet.
- If `dlv` is missing from PATH, locate/install it first (`dlv version` should work).

## Rebuild Only
If you just need to rebuild the provider in Delve without continuing:
```bash
printf 'rebuild\n' > /tmp/dlv_cmds_full
dlv connect pulumi-talos-cluster-runner-delve.pulumi-talos-cluster-workbench:2345 --init=/tmp/dlv_cmds_full
```

## Common Symptoms & Actions
- **`plugin ... did not begin responding to RPC connections`**: Provider is not started. 
- **`EOF` / resource monitor shut down**: Provider crashed or some other error.
- **`--init` file missing**: Always create the temp file under `/tmp` before `dlv connect`.
