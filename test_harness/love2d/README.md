# LÖVE Test Harness

A deterministic, safe test target for the screenshot-mcp-server automation system.

## Features

- **Fixed window size**: 800x600 pixels (non-resizable)
- **Stable window title**: "UI Automation Test Target"
- **Safe mode**: Starts paused by default - clicks won't trigger destructive actions
- **Deterministic**: Fixed RNG seed (12345) for reproducible behavior
- **Debug endpoint**: UDP socket on port 9876 for external control
- **Simple UI**: Three buttons with click counters for testing automation

## Running the Test Harness

### Prerequisites

Install LÖVE (Love2D) for macOS:
```bash
brew install love
```

Or download from: https://love2d.org/

### Launch

```bash
# From this directory
love .

# Or with absolute path
love /Users/wwilson/screenshot-mcp-server/test_harness/love2d
```

## Safety Features

1. **Auto-pause on focus loss**: Window automatically pauses when it loses focus
2. **Safe mode start**: Window starts in paused state - click "Start Test" or anywhere to activate
3. **Contained clicks**: All UI elements are within the window bounds
4. **No file operations**: Test harness doesn't write files or access the filesystem

## Debug Endpoint

The test harness exposes a UDP debug endpoint on port 9876.

### Commands

```bash
# Query current state
echo "GET_STATE" | nc -u 127.0.0.1 9876

# Reset to initial state
echo "RESET" | nc -u 127.0.0.1 9876

# Pause/Resume
echo "PAUSE" | nc -u 127.0.0.1 9876
echo "RESUME" | nc -u 127.0.0.1 9876

# Change scene
echo "SET_SCENE test" | nc -u 127.0.0.1 9876

# Quit the game
echo "QUIT" | nc -u 127.0.0.1 9876

# Health check
echo "PING" | nc -u 127.0.0.1 9876
```

### Response Format

All responses are JSON:
```json
{
  "status": "ok",
  "state": {
    "isPaused": true,
    "currentScene": "menu",
    "buttonClicks": {
      "start": 0,
      "reset": 0,
      "pause": 0
    },
    "windowInfo": {
      "title": "UI Automation Test Target",
      "width": 800,
      "height": 600
    }
  }
}
```

## UI Elements

The test harness displays three buttons for automation testing:

1. **Start Test** (green): Changes scene to "test", resumes from pause
2. **Reset State** (orange): Resets all state to initial values, re-seeds RNG
3. **Pause/Resume** (blue): Toggles pause state

Each button shows a click counter to the right for verifying automation actions.

## Coordinate System

- Window size: 800x600 pixels
- Origin (0,0): Top-left corner
- Button positions (approximate):
  - Start Test: center X, Y ≈ 200
  - Reset State: center X, Y ≈ 270
  - Pause/Resume: center X, Y ≈ 340
- Button size: 200x50 pixels

## Using with MCP Server

```bash
# 1. Start the test harness
love .

# 2. List windows to find it
./bin/screenshot_mcp_server list_windows

# 3. Run automation agent
./bin/agent \
  -goal "Click the Start Test button" \
  -window "UI Automation Test Target" \
  -run-dir ./run-artifacts
```

## Integration Testing

The test harness is designed for automated integration tests:

1. **Deterministic rendering**: Same visual output every run
2. **Predictable state**: Fixed seed means random operations are reproducible
3. **State introspection**: Debug endpoint allows verification of internal state
4. **Safe operations**: No risk of system modification or data loss

## DPI Handling

The test harness uses `highdpi = false` for simpler coordinate mapping:
- Screen points = Pixel coordinates
- No Retina scaling complications
- MCP server scale will be 1.0

This makes automation more reliable across different display types.
