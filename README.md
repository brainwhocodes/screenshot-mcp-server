# Screenshot MCP Server

A Go implementation of an MCP server and client for full-screen screenshots, with LLM vision automation capabilities.

## Features

- MCP tools:
  - `take_screenshot`
  - `take_screenshot_png`
  - `screenshot_hash`
  - `list_windows`
  - `focus_window`
  - `take_window_screenshot`
  - `take_window_screenshot_png`
  - `take_region_screenshot`
  - `take_region_screenshot_png`
  - `click`
  - `click_screen` (screen coordinates, no window_id)
  - `mouse_move`
  - `mouse_down`
  - `mouse_up`
  - `drag`
  - `scroll`
  - `press_key`
  - `type_text`
  - `key_down`
  - `key_up`
  - `wait_for_pixel`
  - `wait_for_region_stable`
  - `launch_app`
  - `quit_app`
  - `wait_for_process`
  - `kill_process`
  - `wait_for_image_match`
  - `find_image_matches`
  - `compare_images`
  - `assert_screenshot_matches_fixture`
  - `set_clipboard`
  - `get_clipboard`
  - `start_recording` *(experimental)*
  - `stop_recording` *(experimental)*
  - `wait_for_text` *(experimental)*
  - `restart_app` *(experimental)*
  - `take_screenshot_with_cursor` *(experimental)*
- JPEG/PNG outputs and image hash support
- Optional `--experimental` tools for non-production workflows
- stdio transport server
- SSE transport server (default port `3001`)
- CLI client that saves screenshot output to a file
- LLM Vision automation agent CLI (`cmd/agent`)

## Requirements

- Go `1.25+`
- macOS is required for window + input automation tools (window listing/focus/window screenshots/clicks/keys).
- Other OSes can use full-screen screenshot tools (`take_screenshot`, `take_screenshot_png`, and `screenshot_hash` with `target: "screen"`) where the screenshot backend is supported.

## Build

```bash
make build
```

Built binaries:

- `./bin/screenshot_mcp_server` - MCP server
- `./bin/agent` - Automation agent CLI

## Run

Start stdio server:

```bash
./bin/screenshot_mcp_server server
```

Start SSE server:

```bash
./bin/screenshot_mcp_server sse --port 3001
```

Take a screenshot with the CLI client:

```bash
./bin/screenshot_mcp_server client output.jpg
```

## Automation Agent

The agent CLI enables automated UI interaction using LLM vision:

```bash
./bin/agent -goal "Click the Continue button" -window "My App" -run-dir ./run-artifacts
```

### Agent Options

- `-goal` (required): The automation goal for the LLM
- `-window`: Target window title (partial match)
- `-app`: Target application name
- `-max-steps`: Maximum automation steps (default: 25)
- `-timeout`: Maximum duration (default: 2m)
- `-run-dir`: Directory to save artifacts
- `-api-key`: OpenAI API key (or set `OPENAI_API_KEY`)
- `-model`: OpenAI model to use (default: gpt-4o)
- `-dry-run`: Run without executing actions (for testing)

### Agent Loop

The agent follows this loop:

1. Select target window by title/app
2. Focus the window
3. Take a window screenshot
4. Send to LLM vision with goal and constraints
5. Parse JSON action response
6. Execute action (click/press_key)
7. Repeat until `done: true` or safety limits

### Safety Constraints

- Maximum steps (default: 25)
- Maximum duration (default: 2 minutes)
- Confidence threshold (stops if < 0.5)
- Clicks are clamped to window bounds
- All actions logged to run artifacts

## MCP Tools

### `list_windows`

Returns all visible application windows with metadata (ID, owner, title, bounds).

### `focus_window`

Brings a window to the foreground and activates its application.

### Tool Support Matrix

| Tool | macOS | Other OSes | Notes |
| --- | :---: | :---: | --- |
| `take_screenshot`, `take_screenshot_png` | ✅ | ✅ | Full-screen screenshot capture via `github.com/kbinani/screenshot` |
| `screenshot_hash` | ✅ | ✅ | Hashes the full screen; `target: "window"` requires macOS window tools |
| `list_windows`, `focus_window`, `take_window_screenshot*` | ✅ | ❌ | Window automation requires macOS APIs (not registered on other OSes) |
| input + wait tools (`click`, `click_screen`, `press_key`, `wait_for_pixel`, etc.) | ✅ | ❌ | Require macOS accessibility APIs (not registered on other OSes) |
| app/process helpers (`launch_app`, `quit_app`, etc.) | ✅ | ❌ | macOS-specific commands (not registered on other OSes) |
| experimental tools (`wait_for_text`, recording, cursor capture, etc.) | ✅ | ❌ | Behind `--experimental`; feature availability depends on host tools (`tesseract`, `screencapture`, `ffmpeg`) |

### `take_screenshot`

Captures the full screen and returns image bytes (JPEG output with metadata in `TextContent`).

### `take_window_screenshot`

Captures a specific window and returns image bytes plus metadata for coordinate mapping.

### `click`

Performs a mouse click at specified pixel coordinates within a window.

### `click_screen`

Performs a mouse click at absolute screen coordinates (default `coord_space: "points"`; set `coord_space: "pixels"` for raw pixel inputs).

### `press_key`

Sends a key press (with optional modifiers) to the focused window.

## Testing

Run all unit and integration tests:

```bash
go test ./...
```

For deterministic tests in headless environments, the server supports:

- `SCREENSHOT_MCP_TEST_IMAGE_PATH=/path/to/fixture.jpg`

## macOS Permissions

The server requires these macOS permissions:

- **Screen Recording**: System Settings → Privacy & Security → Screen Recording
- **Accessibility**: System Settings → Privacy & Security → Accessibility

## License

MIT
