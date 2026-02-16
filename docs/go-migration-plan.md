# Python to Go Migration Plan (Atomic Commits)

This plan converts the current Python MCP server/client to Go while keeping behavior stable.

## Current Behavior To Preserve

- Tool name: `take_screenshot`
- Tool output: JPEG image bytes returned as MCP image content
- Default SSE port: `3001`
- Transports: `stdio` and `sse`
- Client behavior: call `take_screenshot` and write image bytes to a user-provided file path
- Existing command names to preserve:
  - `screenshot_mcp_server-server`
  - `screenshot_mcp_server-server-sse`
  - `screenshot_mcp_server-client`

## Proposed Go Layout

- `cmd/server/main.go` (stdio server entrypoint)
- `cmd/server-sse/main.go` (SSE server entrypoint)
- `cmd/client/main.go` (client CLI)
- `internal/mcp/server.go` (tool registration + transport wiring)
- `internal/tools/screenshot.go` (`take_screenshot` tool handler)
- `internal/screenshot/capture_*.go` (platform capture adapters)
- `internal/image/jpeg.go` (JPEG encoding and size controls)
- `tests/contract/` (black-box parity tests)

## Commit-By-Commit Plan

1. `docs: add migration contract and file mapping`
- Add `docs/migration-contract.md` with exact parity requirements from current Python implementation.
- Add a Python->Go file mapping table.
- No functional changes.
- Verify: docs reviewed.

2. `test: add black-box contract tests for current python behavior`
- Add tests that execute the Python server/client over stdio and SSE.
- Assert: tool exists, output is decodable JPEG, client writes a valid JPEG file.
- Verify: `uv run pytest tests/contract -q`

3. `chore(go): initialize go module and basic tooling`
- Add `go.mod`, `go.sum`, `Makefile` targets (`build`, `test`, `lint`).
- Update `.gitignore` for Go artifacts (`bin/`, coverage files).
- Add empty `cmd/server`, `cmd/server-sse`, and `cmd/client` mains that print help.
- Verify: `go test ./...` and `go build ./cmd/...`

4. `docs(adr): choose Go MCP SDK and screenshot library`
- Add ADR for selected MCP SDK and screenshot capture package.
- Record rationale and rejected alternatives.
- No runtime behavior changes.
- Verify: docs reviewed.

5. `feat(go): add screenshot capture abstraction`
- Implement `internal/screenshot` with a small interface and platform-specific capture code.
- Return clear errors for unsupported/headless environments.
- Add unit tests for interface-level behavior with fakes.
- Verify: `go test ./internal/screenshot/...`

6. `feat(go): add JPEG encoder matching python defaults`
- Implement JPEG encoding with default quality `60`.
- Add optional size guard (target near 1MB) with quality fallback loop.
- Add unit tests for valid JPEG output and size-control behavior.
- Verify: `go test ./internal/image/...`

7. `feat(go): implement take_screenshot tool handler`
- Implement Go equivalent of `screenshot_mcp_server/tools/screenshot.py`.
- Wire capture + JPEG encoding into MCP image response.
- Add tests for happy path and capture/encode failures.
- Verify: `go test ./internal/tools/...`

8. `feat(go): implement stdio MCP server`
- Implement stdio transport entrypoint in `cmd/server/main.go`.
- Register `take_screenshot` tool with same public name/description.
- Add integration test to start server, initialize MCP session, call tool.
- Verify: `go test ./... -run Stdio`

9. `feat(go): implement SSE MCP server`
- Implement SSE entrypoint in `cmd/server-sse/main.go` with `--port` default `3001`.
- Add integration test for SSE startup and tool call.
- Verify: `go test ./... -run SSE`

10. `feat(go): implement Go client CLI`
- Implement `cmd/client/main.go` equivalent to Python client.
- Accept output path argument, call tool over stdio, decode/write image.
- Add integration test that validates file creation and JPEG decoding.
- Verify: `go test ./... -run Client`

11. `test: run contract suite against Go binaries`
- Extend `tests/contract` to run against Go server/client as an alternate backend.
- Ensure parity with Python contract expectations.
- Verify: `go test ./...` and `uv run pytest tests/contract -q`

12. `build: preserve existing command names with wrappers`
- Add wrapper scripts or install targets so external callers keep using:
  - `screenshot_mcp_server-server`
  - `screenshot_mcp_server-server-sse`
  - `screenshot_mcp_server-client`
- Point wrappers to Go binaries.
- Verify: each command works end-to-end.

13. `ci: add Go build and parity checks`
- Add CI jobs for `go test`, `go build`, and contract tests.
- Keep Python contract tests during transition.
- Verify: CI green on main branch.

14. `docs: switch README and usage examples to Go`
- Update install/run instructions to Go-first workflow.
- Keep a short compatibility note for Python deprecation window.
- Verify: README commands run locally.

15. `chore: remove Python runtime implementation`
- Delete Python package runtime files:
  - `screenshot_mcp_server/server/*.py`
  - `screenshot_mcp_server/tools/*.py`
  - `screenshot_mcp_server/client/*.py`
- Remove Python runtime dependencies from `pyproject.toml` (or remove file if no longer used).
- Verify: Go commands pass all tests; no Python runtime path required.

16. `chore: finalize Go-only release prep`
- Add version injection (`-ldflags`) and release packaging notes.
- Remove migration-only compatibility shims if not needed.
- Verify: clean build from fresh clone with Go-only instructions.

## Execution Rules For Atomicity

- One logical change per commit.
- Each commit must compile and pass its relevant tests.
- Avoid mixed commits (for example, behavior changes + docs rewrites together).
- Keep Python behavior as reference until commit 15.
