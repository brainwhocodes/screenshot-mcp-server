# Refactor TODO (Open Items)

- [ ] Split oversized files (`internal/mcpserver/server.go`, `internal/window/window.go`, `internal/mcpserver/tools_utility.go`).
- [ ] Clean up style/consistency backlog from secondary linters (`revive`, additional naming/comment polish).
- [ ] Reduce cyclomatic complexity in core handlers (especially `internal/mcpserver/server.go` registration flow).
- [ ] Split long functions (`internal/mcpserver/server.go` `NewServer`, `internal/mcpserver/image_compare.go` comparison helpers, key `internal/window/window.go` helper functions).
- [ ] Reduce cognitive complexity (`internal/mcpserver/server.go` `NewServer`).
- [ ] Split `internal/mcpserver/server.go` into per-tool registration files.
- [ ] Split `internal/window/window.go` into focused input/list/focus/screenshot/permission files.
- [ ] Move long helper implementations out of MCP wiring code to keep `NewServer` mostly `register` calls.
- [ ] Standardize error shapes (error codes + actionable messages) for reliable agent behavior.
- [ ] Standardize naming/style polish called out by `revive` where worthwhile.
- [ ] Normalize style issues (non-blocking refactors for readability).
