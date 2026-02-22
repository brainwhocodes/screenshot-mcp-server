package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIVisionClient is an OpenAI Chat Completions client for screenshot actions.
type OpenAIVisionClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAIVisionClient creates a new OpenAI client for screenshot-to-action decisions.
func NewOpenAIVisionClient(apiKey, model string) *OpenAIVisionClient {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIVisionClient{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type openAIMessage struct {
	Role    string          `json:"role"`
	Content []openAIContent `json:"content"`
}

type openAIContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openAIImageURL `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

const openAIVisionPrompt = `You are a UI automation agent. Your task is to analyze screenshots and decide the next action to achieve the given goal.

You MUST respond with ONLY a valid JSON object (no markdown, no code blocks) in this exact format:
{
  "action": "click" | "press_key" | "noop" | "done",
  "x": <pixel_x_coordinate>,
  "y": <pixel_y_coordinate>,
  "button": "left" | "right" | "middle",
  "clicks": 1 | 2,
  "key": "<key_name>",
  "modifiers": ["shift", "control", "option", "command"],
  "done": true | false,
  "why": "<brief explanation of why you chose this action>",
  "confidence": <0.0 to 1.0>
}

Rules:
1. Coordinates (x, y) are pixels from top-left of the screenshot image
2. For "click" action: provide x, y, optionally button (default: "left") and clicks (default: 1)
3. For "press_key" action: provide key name (e.g., "enter", "tab", "escape") and optional modifiers
4. For "noop" action: use when you are unsure; set done=false
5. For "done" action: set done=true when the goal is achieved
6. Set confidence (0-1) to indicate how certain you are about this action
7. If confidence < 0.5, stop with noop and done=false

SAFETY:
- Never click outside the visible window area
- If you cannot determine the action, use noop with low confidence
- Prefer being conservative over taking wrong actions`

// GetAction captures the next high-level automation action from the LLM.
func (c *OpenAIVisionClient) GetAction(ctx context.Context, screenshot []byte, goal string) (*Action, error) {
	response, err := c.requestAction(ctx, screenshot, goal)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = c.closeResponseBody(response)
	}()

	action, err := c.readAndParseAction(response)
	if err != nil {
		return nil, err
	}
	return &action, nil
}

func (c *OpenAIVisionClient) requestAction(ctx context.Context, screenshot []byte, goal string) (*http.Response, error) {
	payload := c.buildRequestPayload(screenshot, goal)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	request, err := c.buildHTTPRequest(ctx, requestBody)
	if err != nil {
		return nil, err
	}

	response, err := c.doRequest(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *OpenAIVisionClient) readAndParseAction(resp *http.Response) (Action, error) {
	openAIResp, err := c.readActionResponse(resp)
	if err != nil {
		return Action{}, err
	}

	var action Action
	if err := c.parseAction(openAIResp, &action); err != nil {
		return Action{}, err
	}
	return action, nil
}

// buildRequestPayload builds the OpenAI request payload for the vision model.
func (c *OpenAIVisionClient) buildRequestPayload(screenshot []byte, goal string) openAIRequest {
	imageBase64 := EncodeImageBase64(screenshot)
	return openAIRequest{
		Model: c.Model,
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: []openAIContent{
					{Type: "text", Text: openAIVisionPrompt},
				},
			},
			{
				Role: "user",
				Content: []openAIContent{
					{Type: "text", Text: fmt.Sprintf("Goal: %s\n\nAnalyze this screenshot and provide the next action as JSON:", goal)},
					{
						Type: "image_url",
						ImageURL: &openAIImageURL{
							URL:    fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
							Detail: "auto",
						},
					},
				},
			},
		},
		MaxTokens:   500,
		Temperature: 0.1,
	}
}

func (c *OpenAIVisionClient) buildHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	return request, nil
}

func (c *OpenAIVisionClient) doRequest(request *http.Request) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

func (c *OpenAIVisionClient) readActionResponse(resp *http.Response) (*openAIResponse, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var response openAIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}
	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &response, nil
}

func (c *OpenAIVisionClient) parseAction(response *openAIResponse, action *Action) error {
	content := response.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(content), action); err != nil {
		return fmt.Errorf("parse action JSON: %w (content: %s)", err, content)
	}
	return nil
}

func (c *OpenAIVisionClient) closeResponseBody(resp *http.Response) error {
	if resp == nil {
		return nil
	}
	return fmt.Errorf("close response body: %w", resp.Body.Close())
}
