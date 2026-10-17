# Claude Instructions

Read `AGENTS.md` before making changes in this repository. It is the canonical source for repository instructions, workflows, testing rules, Git rules, and available per-repository skills.

Per-repository skills live in `.agents/skills/`.

Per-repository MCP servers are configured in `.mcp.json` at the repository root (the project-scoped location Claude Code reads). The `pulumi-talos-cluster-mcp` runner is configured for Claude only; the Codex (`.codex/config.toml`) and Gemini (`.gemini/settings.json`) configs do not include it.
