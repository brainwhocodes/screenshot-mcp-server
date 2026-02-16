# Full UI Automation Plan (macOS-first, LLM Vision)

This document describes how to extend `screenshot-mcp-server` into a full “see → decide → act” UI automation stack:

- **MCP server**: window discovery, window screenshots, and input injection (click/type).
- **Agent client**: uses **LLM vision** to interpret screenshots and decide actions.

## Goals

- Target a specific application **window** (not just full-screen screenshots).
- Take repeatable **window-scoped screenshots** with metadata needed for coordinate mapping.
- Perform **safe, bounded clicks** based on screenshot pixel coordinates.
- Enable an LLM-vision agent loop: `focus → screenshot → decide → act → verify`.

## Non-goals (v1)

- Cross-platform support (macOS only first).
- Complex gestures (drag/scroll), multi-touch.
- DOM/accessibility element trees (pixel-only to start).

## Architecture Overview

1. **MCP Server (this repo)**
   - Exposes tools: window enumeration, focus, window screenshots, click/type primitives.
   - Enforces safety policies (only click inside focused target window).
2. **Automation Agent (new client)**
   - Repeatedly requests a screenshot, sends it to a vision model with a goal, parses a strict JSON action, and executes it via MCP tools.

## MCP Tool Contract (Proposed)

### 1) `list_windows`

Returns a list of candidate windows:

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

Notes:
- `bounds` are **screen coordinates in points** (Quartz coordinate system as returned by `CGWindowListCopyWindowInfo`).
- Filter out:
  - windows with tiny bounds (e.g. < 50×50),
  - windows without an owning app,
  - obviously system/overlay windows (heuristics).

### 2) `focus_window`

Input:

```json
{ "window_id": 123 }
```

Behavior:
- Brings the owning app to the foreground and raises the window.
- If raising via APIs is unreliable, **fallback** to clicking a safe point (e.g. window center) to ensure focus.

### 3) `take_window_screenshot`

Input:

```json
{ "window_id": 123 }
```

Output:
- MCP image content (JPEG bytes) plus metadata:

```json
{
  "window_id": 123,
  "bounds": { "x": 10, "y": 50, "width": 1200, "height": 800 },
  "image_width": 2400,
  "image_height": 1600,
  "scale": 2.0
}
```

Coordinate contract:
- The screenshot image is the **window content cropped to `bounds`**.
- Pixel coordinates `(x_px, y_px)` are measured from **top-left of the returned image**.
- Convert back to global screen points:
  - `x_pt = bounds.x + (x_px / scale)`
  - `y_pt = bounds.y + (y_px / scale)`
- The server is responsible for mapping and sending the correct system mouse event.

### 4) `click`

Input:

```json
{ "window_id": 123, "x": 400, "y": 220, "button": "left", "clicks": 1 }
```

Rules:
- `(x,y)` are **pixels in the most recent screenshot** for that window (or the server recomputes fresh bounds before clicking).
- The server clamps the click to `[0..image_width) × [0..image_height)`.
- The server verifies:
  - target window still exists,
  - window bounds are sane,
  - window is focused (or re-focuses),
  - the mapped global point lies within the window bounds.

Optional inputs:
- `button`: `"left" | "right" | "middle"` (start with left only).
- `clicks`: `1 | 2` (single or double).

### 5) Optional (later): `type_text`, `press_key`

- `type_text(window_id, text)`
- `press_key(window_id, key, modifiers)`

For v1, clicks alone are sufficient for many flows, but typing is the next highest-value primitive.

## macOS Implementation Notes

### Window enumeration and bounds

- Use `CGWindowListCopyWindowInfo` to list windows and their bounds.
- Capture `kCGWindowNumber`, `kCGWindowOwnerName`, `kCGWindowOwnerPID`, `kCGWindowName`, `kCGWindowBounds`.

### Window screenshots

Preferred:
- `CGWindowListCreateImage` with window bounds cropping.

Fallback:
- Take full screen screenshot then crop to the window bounds (using existing capture path), but ensure correct multi-monitor + Retina scaling.

### Input injection (click)

- Use `CGEventCreateMouseEvent` + `CGEventPost`.
- Post `kCGEventLeftMouseDown` then `kCGEventLeftMouseUp` at the computed global location.

### Permissions (must handle gracefully)

macOS will require:
- **Screen Recording** permission to capture the screen/window.
- **Accessibility** permission to post mouse/keyboard events.

The server should:
- Detect permission failures and return explicit errors telling the user what to enable in:
  - System Settings → Privacy & Security → Screen Recording
  - System Settings → Privacy & Security → Accessibility

## LLM Vision Agent (Client) Plan

### Agent loop

1. Select a target window (`list_windows` → choose by title/app/pid).
2. `focus_window`.
3. `take_window_screenshot`.
4. Send to LLM vision with:
   - the screenshot image,
   - the goal/instructions,
   - strict constraints: click-only, window-bounded, return JSON only.
5. Parse action JSON and execute `click`.
6. Repeat until `"done": true` or safety limit triggers.

### Action JSON schema (strict)

```json
{
  "action": "click",
  "x": 123,
  "y": 456,
  "done": false,
  "why": "Click the 'Continue' button"
}
```

Allowed actions (v1):
- `"click"`
- `"noop"` (agent is unsure; request another screenshot or stop)
- `"done"`

### Safety constraints for the agent

- Max steps (e.g. 25), max wall time (e.g. 2 minutes).
- Stop if the LLM returns out-of-bounds coordinates or malformed JSON.
- Optional: require the LLM to also return a `confidence` (0–1) and stop below a threshold.

## Testing and Debugging

### Server-side tests

- Unit tests with fakes for:
  - window lookup,
  - coordinate mapping and clamping,
  - “refuse to click when not focused / bounds invalid”.
- Integration tests (manual or CI-skipped on non-macOS):
  - list windows,
  - take a window screenshot,
  - click a harmless location within a test app.

### Agent-side debugging artifacts

Persist per-step artifacts to a run folder:
- screenshot JPEG
- chosen window metadata
- LLM prompt/response
- executed action (click coords + mapped global coords)

## Milestones

1. `list_windows` (stable enumeration + filtering).
2. `take_window_screenshot` (correct cropping + scale metadata).
3. `focus_window` (reliable enough for repeated loops).
4. `click` (safe clamped clicks with focus enforcement).
5. Minimal agent CLI (`cmd/agent`) that runs the loop with an LLM vision model.
6. Add `type_text`/`press_key` for broader workflows.

