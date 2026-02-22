# Refactor Plan (Repo / Code Health)

Track UI automation feature work in `docs/automation-llm-vision-macos-plan.md`. This file is for code health, maintainability, and correctness refactors.

## Remaining TODO
- [ ] Split oversized files (`internal/mcpserver/server.go`, `internal/window/window.go`, `internal/mcpserver/tools_utility.go`).
- [ ] Clean up style/consistency backlog from secondary linters (`revive`, additional naming/comment polish).
- [x] Reduce cyclomatic complexity in core handlers (still pending on `internal/mcpserver/server.go`::`NewServer` and MCP registration flow).
- [ ] Split long functions (`internal/mcpserver/server.go`::`NewServer`, several `internal/window/window.go` screenshot/input helpers).
- [x] Reduce cognitive complexity (`internal/mcpserver/server.go`::`NewServer`).
- [x] Standardize error wrapping at package boundaries (especially MCP handlers and OS integration helpers).
- [x] Define pass-through error policy for `errors.Is`/`errors.As` versus wrapped context errors.
- [x] Address integer-cast warnings in hashing and pixel conversion code (`G115`) with explicit bounds checks.
- [x] Document accepted `gosec` suppressions in code comments where test-only behavior is intentional.
- [x] Remove avoidable built-in shadowing (e.g., `max` parameter name in `internal/window/window.go`).
- [ ] Split `internal/mcpserver/server.go` (~1338 LOC) into per-tool registration files.
- [ ] Split `internal/window/window.go` (~1081 LOC) into focused input/list/focus/screenshot/permission files.
- [x] Move long helper implementations out of MCP wiring code (keep `NewServer` mostly “register tools + call services”).
- [x] Centralize permission checks + messaging (screen recording / accessibility) so all tools fail consistently.
- [x] Standardize error shapes (error codes + actionable messages) for reliable agent behavior.
- [x] Remove or encapsulate package-level mutable state (e.g., recording state maps); guard with mutexes when state is required.
- [x] Add explicit timeouts for OS calls (`osascript`, process management) to avoid hangs during long test runs.
- [x] Avoid encode→decode loops in polling tools (pixel/region waiters) via raw image sampling path.
- [x] Make screenshot hashing explicit and stable (cursor inclusion, window vs full-screen target).
- [x] Add deterministic “artifacts directory” behavior (server-side allowlist + consistent file naming).
- [x] Add deterministic fixture strategy for image-based tests (stable codec options and metadata stripping).

## Priority Buckets

### P0 (Do First: correctness, safety, CI signal)

- [x] Wire lint into workflow: `make lint` uses `golangci-lint run`, add `.golangci.yml`, add CI lint job.
- [x] Fix current baseline lint failures in default run (`errcheck`, `govet`, `ineffassign`, `staticcheck`).
- [x] Address high-impact security/config items:
  - [x] add `ReadHeaderTimeout` and server safety limits,
  - [x] tighten default artifact/output permissions where appropriate,
  - [x] audit command execution and argument validation.
- [x] Hide or gate placeholder tools (`experimental` flag) until behavior is implemented or stabilized.
- [x] Reduce risk in largest hotspots:
  - [x] split `NewServer` registration into feature-group registrars,
  - [x] simplify `(*Agent).Run` control flow.

### P1 (Do Next: structure and maintainability)

- [ ] Split oversized files (`internal/mcpserver/server.go`, `internal/window/window.go`, `internal/mcpserver/tools_utility.go`).
- [x] Extract MCP-independent service layer and shared validators (window IDs, coords, timeouts, paths).
- [x] Standardize response and error patterns:
  - [x] add `tools.ToolResultFromJSON(any)`,
  - [x] enforce consistent `TextContent`/`ImageContent` conventions,
  - [x] standardize wrapped errors at package boundaries.
- [x] Improve platform boundaries:
  - [x] ensure darwin-specific build tagging/stubs are consistent,
  - [x] avoid advertising tools that always fail on non-darwin.
- [x] Deduplicate repeated window input mapping logic (`MouseDown`/`MouseUp` helper extraction).

### P2 (Then: quality expansion and long-term hardening)

- [x] Expand test hardening:
  - [x] add `-race` test pass,
- [x] add coverage gates for critical flows,
- [x] improve deterministic image fixture strategy.
- [x] Improve docs and consistency:
  - [x] tool support matrix by OS/permissions,
  - [x] keep README tool list synced (prefer generated table).
- [ ] Clean up style/consistency backlog from secondary linters (`revive`, additional naming/comment polish).

## Implementation Order

1. [x] Baseline tooling: `make lint` + `.golangci.yml` + CI lint job.
2. [x] Clear default lint blockers (`errcheck`, `govet`, `ineffassign`, `staticcheck`).
3. [x] Apply P0 security/server hardening (`gosec` high-value items, HTTP server timeouts).
4. [x] Gate/remove placeholder tools that are currently stubbed.
5. [x] Refactor `NewServer` into registration modules.
6. [x] Refactor `(*Agent).Run` and `GetAction` into smaller units.
7. [ ] Split `window.go` and `tools_utility.go`; extract shared validators/services.
8. [x] Unify response/error helpers and wrap policy across packages.
9. [x] Add race/coverage/determinism quality gates.
10. [x] Finish docs/support-matrix/style consistency backlog.

## Linting + CI hygiene

- [x] Run `golangci-lint run` and capture findings.
- [x] Adopt `golangci-lint` as the standard linter (local + CI).
- [x] Update `make lint` to run `golangci-lint run` (keep `make test` as `go test ./...`).
- [x] Add `.golangci.yml` to pin/curate enabled linters and timeouts.
- [x] Add CI job(s): `go test ./...` + `golangci-lint run`.

## Fix current `golangci-lint run` findings

- [x] Fix `errcheck` findings (handle or intentionally ignore with `_ = ...`):
  - [x] `cmd/agent/main.go`: ignore/check `fmt.Fprintln` / `fmt.Fprintf` writes to stderr.
  - [x] `internal/agent/openai.go`: ignore/check `resp.Body.Close()`.
  - [x] `internal/client/client.go`: ignore/check `session.Close()`.
  - [x] `internal/mcpserver/server_integration_test.go`: ignore/check `session.Close()`.
  - [x] `internal/mcpserver/sse_integration_test.go`: ignore/check `session.Close()`.
  - [x] `internal/mcpserver/tools_utility.go`: ignore/check `Close()` calls and `quitCmd.Run()` (intentional ignore should be explicit).
  - [x] `internal/testutil/fixture.go`: ignore/check fixture `file.Close()`.
- [x] Fix `govet` unreachable code in `internal/mcpserver/tools_utility.go` (template matching stub returns before using the screenshot variable).
- [x] Fix `ineffassign` in `internal/mcpserver/tools_utility.go` (template image + timeout defaults are unused due to placeholder returns).
- [x] Fix `staticcheck` S1016 in `internal/mcpserver/server.go` (convert `maskRegion` → `MaskRegion` directly).

## Reduce placeholder-tool debt

- [x] Move “placeholder” tools behind an explicit `experimental` flag (or remove from tool list until implemented).
- [x] Make stubs fail fast with consistent, typed errors (and ensure they still pass lint).
- [x] Split `internal/mcpserver/tools_utility.go` into focused files (paths, image diff, template matching, OCR, recording) to keep diffs small.

## Additional lint passes (refactor signals)

### Complexity hotspots

- [x] Run `golangci-lint run --enable-only=gocyclo`.
- [x] Reduce cyclomatic complexity:
  - [x] `internal/mcpserver/server.go`: `NewServer` complexity is currently flagged as very high.
  - [x] `internal/agent/agent.go`: `(*Agent).Run` complexity is currently flagged as high.

### Function length hotspots

- [x] Run `golangci-lint run --enable-only=funlen`.
- [x] Split long functions:
  - [x] `cmd/agent/main.go`: `run`
  - [x] `internal/agent/agent.go`: `(*Agent).Run`
  - [x] `internal/agent/openai.go`: `(*OpenAIVisionClient).GetAction`
  - [x] `internal/mcpserver/server.go`: `NewServer`
  - [x] `internal/mcpserver/tools_utility.go`: `compareImages`
  - [x] `internal/mcpserver/image_compare.go`: `assertScreenshotMatchesFixture`
  - [ ] `internal/window/window.go`: `TakeWindowScreenshot`, `TakeRegionScreenshot`, `TakeWindowScreenshotPNG`, `TakeRegionScreenshotPNG`

### Duplicate logic hotspots

- [x] Run `golangci-lint run --enable-only=dupl`.
- [x] Deduplicate common window + coordinate mapping logic:
  - [x] `internal/window/window.go`: `MouseDown` / `MouseUp` share large duplicated blocks (extract helper).

### Placeholder APIs

- [x] Run `golangci-lint run --enable-only=unparam`.
- [x] Either implement these placeholders or remove/reshape parameters so stubs don’t rot:
  - [x] `internal/mcpserver/tools_utility.go`: `findImageMatches` (unused `ctx`)
  - [x] `internal/mcpserver/tools_utility.go`: `assertScreenshotMatchesFixture` (always-nil result + placeholder error)
  - [x] `internal/mcpserver/tools_utility.go`: `waitForText`, `startRecording`, `stopRecording` (unused `ctx`)

### Cognitive complexity hotspots

- [x] Run `golangci-lint run --enable-only=gocognit`.
- [x] Reduce cognitive complexity:
  - [x] `internal/mcpserver/server.go`: `NewServer` is currently flagged as very high.
  - [x] `internal/agent/agent.go`: `(*Agent).Run` is currently flagged as high.

### Nested branching hotspots

- [x] Run `golangci-lint run --enable-only=nestif`.
- [x] Flatten nested flow in `internal/mcpserver/tools_utility.go` path setup (`SetAllowedRunDirectory`) by using early returns.

### Error-wrapping consistency

- [x] Run `golangci-lint run --enable-only=wrapcheck`.
- [x] Standardize error wrapping at package boundaries (especially MCP handlers and OS integration helpers).
- [x] Define where pass-through errors are allowed (`errors.Is`/`errors.As` cases) vs where context must be added.

### Context propagation

- [x] Run `golangci-lint run --enable-only=contextcheck`.
- [x] Fix non-inherited context usage in server shutdown path (`internal/mcpserver/server.go`).
- [x] Audit `context.Background()` use in runtime code and prefer inherited/request contexts.

### Security hardening signals

- [x] Run `golangci-lint run --enable-only=gosec`.
- [x] Tighten file permissions for run artifacts and outputs (prefer `0o750` dirs and `0o600` files unless intentionally public).
- [x] Add `ReadHeaderTimeout` (and other sane limits) on the SSE `http.Server`.
- [x] Review all `exec.CommandContext` call sites:
  - [x] validate/sanitize user-controlled arguments,
  - [x] centralize script generation/escaping for AppleScript.
- [x] Address integer-cast warnings in hashing and pixel conversion code (`G115`) with explicit bounds checks.
- [x] Document accepted `gosec` suppressions in code comments where test-only behavior is intentional.

### API docs + naming consistency

- [x] Run `golangci-lint run --enable-only=revive`.
- [x] Add package comments where missing (`cmd/agent`, `internal/testutil`, `internal/version`).
- [x] Add/normalize comments for exported API types and methods in `internal/agent`.
- [x] Remove avoidable built-in shadowing (e.g., `max` parameter name in `internal/window/window.go`).
- [ ] Normalize style issues called out by revive that improve readability without churn.

## Second-pass refactor ideas

### Split large files

- [ ] Split `internal/mcpserver/server.go` (~1338 LOC) into per-tool registration files (one tool/feature group per file).
- [ ] Split `internal/window/window.go` (~1081 LOC) into focused files (list/focus/screenshot/input/permissions).
- [x] Move long helper implementations out of MCP wiring code (keep `NewServer` mostly “register tools + call services”).

### Platform boundaries (macOS-only code)

- [x] Add `//go:build darwin` to macOS-only implementation files and provide non-darwin stubs where needed (keep `go test ./...` working on other OSes).
- [x] Ensure MCP tool registration is OS-aware (don’t advertise tools that will always error on the current platform).
- [x] Centralize permission checks + messaging (screen recording / accessibility) so all tools fail consistently.

### Dependency injection + testability

- [x] Pass `InputService` into `mcpserver.NewServer` (it’s created internally today), so tests can stub input without OS permissions.
- [x] Introduce interfaces for “window ops” and “screenshot capture” used by MCP tools (so unit tests can be hermetic).
- [x] Add targeted unit tests for coordinate mapping (points↔pixels) and clamping that don’t require real windows.

### Tool responses + errors

- [x] Add `tools.ToolResultFromJSON(any)` helper and use it across tools to reduce repeated `json.Marshal` boilerplate.
- [x] Standardize outputs: JSON always in `TextContent`; images always in `ImageContent` with explicit MIME types.
- [x] Standardize error shapes (error codes + actionable messages) for reliable agent behavior.

### Concurrency + state

- [x] Remove or encapsulate package-level mutable state (e.g., recording state maps); guard with mutexes when state is required.
- [x] Add explicit timeouts for OS calls (`osascript`, process management) to avoid hangs during long test runs.

### Performance + determinism

- [x] Avoid encode→decode loops in polling tools (e.g., pixel/region waiters): expose a raw-image path for sampling and hashing.
- [x] Make screenshot hashing explicit and stable (what it hashes, whether cursor included, window vs full screen).
- [x] Add deterministic “artifacts directory” behavior (server-side allowlist + consistent file naming).

### Docs + consistency

- [x] Keep README tool list in sync with registered tools (consider generating a tool table from code).
- [x] Add a “tool support matrix” (macOS vs other OSes; requires permissions vs not).

### Tool registration architecture

- [x] Replace monolithic `NewServer` registration with feature-group registrars:
  - [x] `registerScreenshotTools(...)`
  - [x] `registerWindowTools(...)`
  - [x] `registerInputTools(...)`
  - [x] `registerImageUtilities(...)`
- [x] Move request arg structs next to the tool they serve (or into dedicated `types_*.go` files) to reduce cross-file coupling.
- [x] Add a lightweight tool registry test that validates all registered tool names are unique.

### Service boundaries

- [x] Extract a thin service layer from MCP handlers to separate transport concerns from automation logic.
- [x] Introduce shared validators for common args (`window_id`, coordinates, timeout/poll defaults, file paths).
- [x] Centralize JSON serialization for tool responses to avoid repeated `json.Marshal` patterns.

### Tests and determinism

- [x] Add race test pass (`go test -race`) for packages with mutable process/global state.
- [x] Add coverage gates for critical flows (window targeting, coordinate mapping, click safety).
- [x] Add deterministic fixture strategy for image-based tests (stable codec options and metadata stripping where needed).
