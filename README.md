# Screenshot MCP Server

A Go implementation of an MCP server and client for full-screen screenshots.

## Features

- MCP tool: `take_screenshot`
- JPEG output (`image/jpeg`) with default quality `60`
- stdio transport server
- SSE transport server (default port `3001`)
- CLI client that saves screenshot output to a file

## Requirements

- Go `1.25+`
- macOS, Linux, or Windows

## Build

```bash
make build
```

Built binary:

- `./bin/screenshot_mcp_server`

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

## Compatibility Command Wrappers

The repo includes wrapper scripts with the legacy command names:

- `./screenshot_mcp_server-server`
- `./screenshot_mcp_server-server-sse`
- `./screenshot_mcp_server-client`
- `./screenshot_mcp_server`

These wrappers run `go run` against the Go commands.

## Testing

Run all unit and integration tests:

```bash
go test ./...
```

For deterministic tests in headless environments, the server supports:

- `SCREENSHOT_MCP_TEST_IMAGE_PATH=/path/to/fixture.jpg`

## License

MIT
