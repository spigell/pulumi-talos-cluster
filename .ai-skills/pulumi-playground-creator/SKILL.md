---
name: pulumi-playground-creator
description: How to use the pulumi-talos-cluster MCP server to run Pulumi/talosctl commands in the playground.
---

Use the `pulumi-talos-cluster-mcp` server via `shell_execute`.

Allowed commands (per MCP whitelist): `wc`, `touch`, `pwd`, `find`, `go`, `grep`, `pulumi`, `ls`, `cat`, `talosctl`. Always set `directory` and `timeout`; pass the command as an array, e.g. `["pulumi","up","--stack","dev","--yes"]`.

Typical flow:
1) Inspect: `pwd`, `ls`, `find`.
2) Pulumi: `pulumi stack init|select|rm`, `pulumi preview|up|destroy --stack <name>`.
3) Talos: `talosctl version`, `talosctl gen secrets|config`, `talosctl apply-config|bootstrap|reset` with `--talosconfig`, `--endpoints`, `--nodes`, and `--insecure` when needed.

Tips: keep `timeout` reasonable (e.g., 300–1800s for up/destroy), use absolute paths, and read `stdout`/`err` from tool responses. Report blocked commands or errors succinctly.***
