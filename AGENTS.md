# AGENTS

Scope: applies to the entire repository.

## Goals
- Keep changes minimal and behavior-preserving by default.
- Prefer small, reviewable diffs over large rewrites.
- Maintain cross-platform build boundaries (Darwin vs non-Darwin stubs).

## Workflow (Go)
- Always run `gofmt` on touched Go files.
- Validate before considering work “done”:
  - `go test ./...`
  - `golangci-lint run` (or `make lint` if available)
- If you split/move code, keep public APIs stable unless explicitly required by a plan item or lint finding.

## Refactors
- Fix root causes; avoid “style-only” churn.
- Prefer file-level splits over large internal abstractions.
- Keep error wrapping consistent (`fmt.Errorf("...: %w", err)`) at package boundaries.

## Docs / Checklists
- Update checklists only for work actually completed in code.
- Keep checkbox state accurate; don’t pre-check planned work.

## Commits
- Split commits by concern when practical (e.g., code refactor vs docs/checklist updates).
