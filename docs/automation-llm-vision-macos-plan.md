# Full UI Automation Checklist (macOS-first, LLM Vision)

Use this as an implementation checklist to extend `screenshot-mcp-server` into a full "see → decide → act" UI automation stack.

## Scope

- [x] MCP server provides: window discovery, window screenshots, and input injection primitives (click/type).
- [x] Automation agent client uses LLM vision to interpret screenshots and decide actions.
- [x] Target engine/game under test: LÖVE (Love2D).

## Goals (v1)

- [x] Target a specific application **window** (not just full-screen screenshots).
- [x] Take repeatable **window-scoped screenshots** with metadata needed for coordinate mapping.
- [x] Perform **safe, bounded clicks** based on screenshot pixel coordinates.
- [x] Enable an agent loop: `focus → screenshot → decide → act → verify`.

## Non-goals (v1) — confirm

- [x] Confirm macOS-only for v1 (no cross-platform support yet).
- [x] Confirm no complex gestures (drag/scroll) or multi-touch in v1.
- [x] Confirm no DOM/accessibility element trees in v1 (pixel-only).

## Architecture

- [x] **MCP Server (this repo)**
  - [x] Expose tools: window enumeration, focus, window screenshots, click/type primitives.
  - [x] Enforce safety policies (only click inside focused target window).
- [x] **Automation Agent (new client)**
  - [x] Repeatedly request a screenshot, send it to a vision model with a goal, parse a strict JSON action, and execute it via MCP tools.
  - [x] Enforce client-side safety limits (max steps, timeouts).

## LÖVE (Love2D) Test Harness (recommended)

- [x] Run the game in a deterministic "test mode" (seed RNG, fixed timestep, stable assets, predictable startup state).
- [x] Force windowed mode with a fixed size (avoid fullscreen/resizable for stable screenshot diffs).
- [x] Set a stable, unique window title via `love.window.setTitle(...)` so `list_windows` selection is reliable.
- [x] Decide DPI strategy and document it:
  - [x] prefer `highdpi=false` for simpler mapping, or
  - [x] allow `highdpi=true` but ensure screenshot `scale` metadata is correct.
- [x] Provide a "safe focus" behavior (e.g., pause menu) so clicks/keys never trigger destructive actions during tests.
- [x] (Optional, high value) Add an in-game debug endpoint (UDP/TCP/file) to:
  - [x] reset game state / load a scene,
  - [x] expose assertions ("current level", "player HP", etc.),
  - [x] trigger an in-engine framebuffer screenshot (lossless PNG).

## MCP Tool Contract (Proposed)

### `list_windows`

- [x] Implement tool `list_windows`.
- [x] Return a list of candidate windows in this JSON shape:

```json
{
  "windows": [
    {
      "window_id": 123,
      "owner_name": "Safari",
      "pid": 9999,
      "title": "Example - Safari",
      "bounds": { "x": 10, "y": 50, "width": 1200, "height": 800 }
    }
  ]
}
```

- [x] `bounds` are **screen coordinates in points** (Quartz coordinate system as returned by `CGWindowListCopyWindowInfo`).
- [x] Filter out:
  - [x] windows with tiny bounds (e.g. < 50×50),
  - [x] windows without an owning app,
  - [x] obviously system/overlay windows (heuristics).

### `focus_window`

- [x] Implement tool `focus_window`.
- [x] Accept this input:

```json
{ "window_id": 123 }
```

- [x] Bring the owning app to the foreground and raise the window.
- [x] If raising via APIs is unreliable, **fallback** to clicking a safe point (e.g. window center) to ensure focus.

### `take_window_screenshot`

- [x] Implement tool `take_window_screenshot`.
- [x] Accept this input:

```json
{ "window_id": 123 }
```

- [x] Return MCP image content (JPEG bytes) plus metadata:

```json
{
  "window_id": 123,
  "bounds": { "x": 10, "y": 50, "width": 1200, "height": 800 },
  "image_width": 2400,
  "image_height": 1600,
  "scale": 2.0
}
```

- [x] Coordinate contract:
  - [x] The screenshot image is the **window content cropped to `bounds`**.
  - [x] Pixel coordinates `(x_px, y_px)` are measured from **top-left of the returned image**.
  - [x] Convert back to global screen points:
    - [x] `x_pt = bounds.x + (x_px / scale)`
    - [x] `y_pt = bounds.y + (y_px / scale)`
  - [x] The server maps and sends the correct system mouse event.

### `click`

- [x] Implement tool `click`.
- [x] Accept this input:

```json
{ "window_id": 123, "x": 400, "y": 220, "button": "left", "clicks": 1 }
```

- [x] Rules:
  - [x] `(x,y)` are **pixels in the most recent screenshot** for that window (or the server recomputes fresh bounds before clicking).
  - [x] Clamp the click to `[0..image_width) × [0..image_height)`.
  - [x] Verify:
    - [x] target window still exists,
    - [x] window bounds are sane,
    - [x] window is focused (or re-focuses),
    - [x] the mapped global point lies within the window bounds.
- [x] Optional inputs:
  - [x] `button`: `"left" | "right" | "middle"` (start with left only).
  - [x] `clicks`: `1 | 2` (single or double).

### Optional (later): `type_text`, `press_key`

- [x] Add `type_text(window_id, text)`.
- [x] Add `press_key(window_id, key, modifiers)`.
- [x] Confirm: clicks-only is sufficient for v1 flows; typing is the next highest-value primitive.

## Additional Automation Tools (Backlog)

### Screenshot & Video

- [x] Add `take_region_screenshot` (rect in points/pixels, with explicit coordinate space).
- [x] Add lossless variants for baseline testing:
  - [x] `take_screenshot_png`
  - [x] `take_window_screenshot_png`
- [x] Add cursor controls:
  - [x] `take_screenshot_with_cursor` / `take_window_screenshot_with_cursor` (API exposed, returns informative error - requires CGWindowListCreateImage with cursor flag for full implementation)
- [x] Add `screenshot_hash` (stable hash for "did the UI change?" checks).
- [x] Add recording primitives (API exposed, stubs only - requires AVFoundation for full implementation):
  - [x] `start_recording(window_id, fps, format)`
  - [x] `stop_recording(run_id)` (return video path/bytes)

### Mouse

- [x] Add `mouse_move(window_id, x, y)`.
- [x] Add `mouse_down(window_id, x, y, button)` and `mouse_up(window_id, x, y, button)`.
- [x] Add `drag(window_id, from_x, from_y, to_x, to_y, button, duration_ms)`.
- [x] Add `scroll(window_id, delta_x, delta_y)`.

### Keyboard & Text

- [x] Add `key_down(window_id, key, modifiers)` and `key_up(window_id, key, modifiers)` (for hold actions).
- [x] Add `type_text(window_id, text)` with optional `delay_ms` for reliability.
- [x] Add clipboard helpers (optional but high leverage): `set_clipboard(text)`, `get_clipboard()`.

### Wait & Assert (test-centric)

- [x] Add `sleep(ms)` (or keep this client-side only).
- [x] Add pixel-level waits:
  - [x] `wait_for_pixel(window_id, x, y, rgba, tolerance, timeout_ms)`
  - [x] `wait_for_region_stable(window_id, rect, timeout_ms)` (no-change heuristic)
- [x] Add template matching:
  - [x] `wait_for_image_match(window_id, template_image, threshold, timeout_ms)`
  - [x] `find_image_matches(window_id, template_image, threshold)` (return coordinates)
- [x] Add baseline testing helpers:
  - [x] `compare_images` (metric + diff image)
  - [x] `assert_screenshot_matches_fixture` (golden + masks/tolerance)
- [x] (Optional) Add `wait_for_text` (OCR) only if needed for LÖVE UI (API exposed, stub only - requires Vision framework).

### App / Process Lifecycle (optional)

- [x] Add `launch_app(...)` and `quit_app(...)` helpers for a predictable test lifecycle.
- [x] Add `wait_for_process(...)` and `kill_process(...)` for cleanup.
- [x] Add "restart target app" flow for flaky-state recovery.

### Artifacts & Debugging (CI-friendly)

- [x] Add a run directory concept (e.g. `start_run(name)` → returns `run_id`).
- [x] Save per-step artifacts (screenshot, metadata, actions) into the run directory.
- [x] Ensure any file-path tools are allowlisted to the run directory only.

## macOS Implementation Notes

### Window enumeration and bounds

- [x] Use `CGWindowListCopyWindowInfo` to list windows and their bounds.
- [x] Capture `kCGWindowNumber`, `kCGWindowOwnerName`, `kCGWindowOwnerPID`, `kCGWindowName`, `kCGWindowBounds`.

### Window screenshots

- [x] ~~Preferred: `CGWindowListCreateImage` with window bounds cropping~~ - **DEPRECATED in macOS 15.0** - API is obsolete and replaced by ScreenCaptureKit. The fallback method is the supported approach.
- [x] Fallback: take full screen screenshot then crop to the window bounds (using existing capture path), but ensure correct multi-monitor + Retina scaling.
- [x] Return metadata sufficient to map pixels → global points reliably on:
  - [x] Retina/non-Retina displays
  - [x] multi-monitor (including negative origins)
  - [x] mixed DPI setups

### Input injection (click)

- [x] Use `CGEventCreateMouseEvent` + `CGEventPost`.
- [x] Post `kCGEventLeftMouseDown` then `kCGEventLeftMouseUp` at the computed global location.

### Permissions (must handle gracefully)

macOS will require:
- **Screen Recording** permission to capture the screen/window.
- **Accessibility** permission to post mouse/keyboard events.

The server should:
- [x] Detect permission failures and return explicit errors telling the user what to enable in:
  - [x] System Settings → Privacy & Security → Screen Recording
  - [x] System Settings → Privacy & Security → Accessibility

## LLM Vision Agent (Client) Plan

### Agent loop

- [x] Select a target window (`list_windows` → choose by title/app/pid).
- [x] Call `focus_window`.
- [x] Call `take_window_screenshot`.
- [x] Send to LLM vision with:
  - [x] the screenshot image,
  - [x] the goal/instructions,
  - [x] strict constraints: click-only, window-bounded, return JSON only.
- [x] Parse action JSON and execute `click`.
- [x] Repeat until `"done": true` or safety limit triggers.

### Action JSON schema (strict)

- [x] Enforce that the model output conforms to a strict JSON schema, for example:

```json
{
  "action": "click",
  "x": 123,
  "y": 456,
  "done": false,
  "why": "Click the 'Continue' button"
}
```

- [x] Allowed actions (v1):
  - [x] `"click"`
  - [x] `"noop"` (agent is unsure; request another screenshot or stop)
  - [x] `"done"`

### Safety constraints for the agent

- [x] Add max steps (e.g. 25) and max wall time (e.g. 2 minutes).
- [x] Stop if the LLM returns out-of-bounds coordinates or malformed JSON.
- [x] Optional: require the LLM to also return a `confidence` (0–1) and stop below a threshold.

## Testing and Debugging

### Server-side tests

- [x] Unit tests with fakes for:
  - [x] window lookup,
  - [x] coordinate mapping and clamping,
  - [x] "refuse to click when not focused / bounds invalid".
- [x] Integration tests (manual or CI-skipped on non-macOS):
  - [x] list windows,
  - [x] take a window screenshot,
  - [x] click a harmless location within a test app.

### Agent-side debugging artifacts

- [x] Persist per-step artifacts to a run folder:
  - [x] screenshot JPEG
  - [x] chosen window metadata
  - [x] LLM prompt/response
  - [x] executed action (click coords + mapped global coords)

## Repo Refactors

- [ ] Track code health refactors in `docs/refactor-plan.md`.

## Milestones

- [x] `list_windows` (stable enumeration + filtering).
- [x] `take_window_screenshot` (correct cropping + scale metadata).
- [x] `focus_window` (reliable enough for repeated loops).
- [x] `click` (safe clamped clicks with focus enforcement).
- [x] Minimal agent CLI (`cmd/agent`) that runs the loop with an LLM vision model.
- [x] Add `type_text`/`press_key` for broader workflows.
