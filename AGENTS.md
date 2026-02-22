# AGENTS

Scope: applies to the entire repository.

## Instructions
- Prefer minimal, behavior-preserving refactors.
- Keep changes small and focused on the current task.
- After modifying Go files, run `gofmt` on touched files.
- Run relevant validation (`gofmt`, `go test ./...`, and `gofmt` + `go test` output checks) before marking refactor items complete.
- Avoid large API changes unless required by lint/plan items.
- Keep existing style and avoid unnecessary abstractions.
- Update documentation checklists/files only when task requires it.
