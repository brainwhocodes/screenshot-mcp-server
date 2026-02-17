package tools

import sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolResultFromText returns a simple MCP tool result containing only text.
func ToolResultFromText(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: text},
		},
	}
}
