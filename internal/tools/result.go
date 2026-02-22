package tools

import (
	"encoding/json"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolResultFromText returns a simple MCP tool result containing only text.
func ToolResultFromText(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: text},
		},
	}
}

// ToolResultFromJSON returns an MCP tool result containing JSON text.
func ToolResultFromJSON(value any) (*sdkmcp.CallToolResult, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: string(data)},
		},
	}, nil
}

// ToolResultFromJSONWithImage returns image/text JSON content in a single response.
func ToolResultFromJSONWithImage(value any, imageData []byte, mimeType string) (*sdkmcp.CallToolResult, error) {
	result, err := ToolResultFromJSON(value)
	if err != nil {
		return nil, err
	}
	result.Content = append(result.Content, &sdkmcp.ImageContent{
		Data:     imageData,
		MIMEType: mimeType,
	})
	return result, nil
}
