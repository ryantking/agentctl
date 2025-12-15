// Package anthropic provides Anthropic SDK client initialization and configuration.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
)

// ToolHandler is a function that executes a tool and returns a result.
// The input is the tool's input parameters as a JSON object (map[string]interface{}).
// Returns the result as a JSON-serializable value, or an error.
type ToolHandler func(ctx context.Context, input map[string]interface{}) (interface{}, error)

// ToolRegistry manages tool definitions and handlers.
type ToolRegistry struct {
	tools   []anthropic.ToolUnionParam
	handlers map[string]ToolHandler
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:   []anthropic.ToolUnionParam{},
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterTool registers a tool with the registry.
// name: Tool name (must be unique)
// description: Tool description
// inputSchema: JSON schema for tool input (as map[string]interface{})
// handler: Function to execute when tool is called
func (r *ToolRegistry) RegisterTool(name, description string, inputSchema map[string]interface{}, handler ToolHandler) error {
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("tool %q already registered", name)
	}

	// Convert inputSchema to ToolInputSchemaParam
	// Extract properties and required fields from schema
	var properties map[string]interface{}
	var required []string

	if props, ok := inputSchema["properties"].(map[string]interface{}); ok {
		properties = props
	}
	if req, ok := inputSchema["required"].([]interface{}); ok {
		required = make([]string, 0, len(req))
		for _, reqItem := range req {
			if s, ok := reqItem.(string); ok {
				required = append(required, s)
			}
		}
	}

	schemaParam := anthropic.ToolInputSchemaParam{
		Type:       constant.Object("object"),
		Properties: properties,
		Required:   required,
	}

	toolParam := anthropic.ToolParam{
		Name:        name,
		Description: param.NewOpt(description),
		InputSchema: schemaParam,
	}

	tool := anthropic.ToolUnionParam{
		OfTool: &toolParam,
	}

	r.tools = append(r.tools, tool)
	r.handlers[name] = handler

	return nil
}

// GetTools returns the list of registered tools for use in API calls.
func (r *ToolRegistry) GetTools() []anthropic.ToolUnionParam {
	return r.tools
}

// ExecuteTool executes a tool by name with the given input.
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	handler, exists := r.handlers[name]
	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}

	return handler(ctx, input)
}

// Conversation manages a multi-turn conversation with tool use support.
type Conversation struct {
	client     anthropic.Client
	registry   *ToolRegistry
	messages   []anthropic.MessageParam
	system     []anthropic.TextBlockParam
	verbose    bool
	toolCalls  int
	maxToolCalls int
	logger     func(string, ...interface{})
}

// NewConversation creates a new conversation with tool use support.
func NewConversation(client anthropic.Client, registry *ToolRegistry) *Conversation {
	return &Conversation{
		client:      client,
		registry:    registry,
		messages:    []anthropic.MessageParam{},
		system:      []anthropic.TextBlockParam{},
		verbose:     false,
		toolCalls:   0,
		maxToolCalls: 50, // Default max tool calls per session
		logger:      func(string, ...interface{}) {}, // No-op logger by default
	}
}

// SetVerbose enables verbose logging of tool executions.
func (c *Conversation) SetVerbose(verbose bool) {
	c.verbose = verbose
	if verbose {
		c.logger = func(format string, args ...interface{}) {
			fmt.Printf("  [tool] "+format+"\n", args...)
		}
	} else {
		c.logger = func(string, ...interface{}) {}
	}
}

// SetMaxToolCalls sets the maximum number of tool calls allowed per session.
func (c *Conversation) SetMaxToolCalls(max int) { //nolint:revive // Method name is clear
	c.maxToolCalls = max
}

// SetSystem sets the system prompt for the conversation.
func (c *Conversation) SetSystem(systemPrompt string) {
	c.system = []anthropic.TextBlockParam{
		{
			Text: systemPrompt,
			Type: constant.Text("text"),
		},
	}
}

// AddUserMessage adds a user message to the conversation.
func (c *Conversation) AddUserMessage(content string) {
	c.messages = append(c.messages, anthropic.NewUserMessage(anthropic.ContentBlockParamUnion{
		OfText: &anthropic.TextBlockParam{
			Text: content,
			Type: constant.Text("text"),
		},
	}))
}

// Send sends a message and handles tool use automatically.
// Returns the final assistant response text, or an error.
func (c *Conversation) Send(ctx context.Context, model anthropic.Model, maxTokens int64) (string, error) {
	maxIterations := 10 // Prevent infinite loops
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Check tool call limit
		if c.toolCalls >= c.maxToolCalls {
			return "", fmt.Errorf("max tool calls (%d) exceeded", c.maxToolCalls)
		}

		// Build message request
		params := anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: maxTokens,
			Messages: c.messages,
			Tools:     c.registry.GetTools(),
		}

		if len(c.system) > 0 {
			params.System = c.system
		}

		// Call Messages API
		msg, err := c.client.Messages.New(ctx, params)
		if err != nil {
			return "", EnhanceSDKError(fmt.Errorf("failed to send message: %w", err))
		}

		// Check for tool use requests
		hasToolUse := false
		for _, block := range msg.Content {
			if block.Type == "tool_use" {
				hasToolUse = true
				break
			}
		}

		if !hasToolUse {
			// No tool use, return the text response
			var response strings.Builder
			for _, block := range msg.Content {
				if block.Type == "text" && block.Text != "" {
					response.WriteString(block.Text)
				}
			}
			return strings.TrimSpace(response.String()), nil
		}

		// Handle tool use - add assistant message with tool use, then execute tools
		var toolResults []anthropic.ContentBlockParamUnion

		for _, block := range msg.Content {
			if block.Type == "tool_use" {
				toolUseID := block.ID
				toolName := block.Name

				// Extract input from tool use block
				var toolInput map[string]interface{}
				if len(block.Input) > 0 {
					if err := json.Unmarshal(block.Input, &toolInput); err != nil {
						return "", fmt.Errorf("failed to unmarshal tool input: %w", err)
					}
				} else {
					toolInput = make(map[string]interface{})
				}

				// Increment tool call counter
				c.toolCalls++

				// Log tool execution
				c.logger("Executing tool: %s (call #%d)", toolName, c.toolCalls)
				if path, ok := toolInput["path"].(string); ok {
					c.logger("  Path: %s", path)
				}

				// Execute tool
				result, err := c.registry.ExecuteTool(ctx, toolName, toolInput)
				if err != nil {
					c.logger("  Error: %v", err)
					// Return error result
					toolResults = append(toolResults, anthropic.ContentBlockParamUnion{
						OfToolResult: &anthropic.ToolResultBlockParam{
							ToolUseID: toolUseID,
							IsError:   param.NewOpt(true),
							Content: []anthropic.ToolResultBlockParamContentUnion{
								{
									OfText: &anthropic.TextBlockParam{
										Text: fmt.Sprintf("Error: %v", err),
										Type: constant.Text("text"),
									},
								},
							},
							Type: constant.ToolResult("tool_result"),
						},
					})
					continue
				}

				// Log successful result (summary)
				if resultMap, ok := result.(map[string]interface{}); ok {
					if path, ok := resultMap["path"].(string); ok {
						c.logger("  Result: accessed %s", path)
					}
					if items, ok := resultMap["items"].([]map[string]interface{}); ok {
						c.logger("  Result: listed %d items", len(items))
					}
					if size, ok := resultMap["size"].(int); ok {
						c.logger("  Result: read %d bytes", size)
					}
				}

				// Serialize result to JSON
				resultJSON, err := json.Marshal(result)
				if err != nil {
					return "", fmt.Errorf("failed to marshal tool result: %w", err)
				}

				// Add tool result
				toolResults = append(toolResults, anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: toolUseID,
						IsError:   param.NewOpt(false),
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{
								OfText: &anthropic.TextBlockParam{
									Text: string(resultJSON),
									Type: constant.Text("text"),
								},
							},
						},
						Type: constant.ToolResult("tool_result"),
					},
				})
			}
		}

		// Convert ContentBlockUnion to ContentBlockParamUnion for assistant message
		var assistantBlocks []anthropic.ContentBlockParamUnion
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				assistantBlocks = append(assistantBlocks, anthropic.ContentBlockParamUnion{
					OfText: &anthropic.TextBlockParam{
						Text: block.Text,
						Type: constant.Text("text"),
					},
				})
			case "tool_use":
				// Convert tool use block
				var toolInput any
				if len(block.Input) > 0 {
					if err := json.Unmarshal(block.Input, &toolInput); err != nil {
						return "", fmt.Errorf("failed to unmarshal tool use input: %w", err)
					}
				}
				assistantBlocks = append(assistantBlocks, anthropic.ContentBlockParamUnion{
					OfToolUse: &anthropic.ToolUseBlockParam{
						ID:    block.ID,
						Name:  block.Name,
						Input: toolInput,
						Type:  constant.ToolUse("tool_use"),
					},
				})
			}
		}

		// Add assistant message with tool use to conversation
		c.messages = append(c.messages, anthropic.NewAssistantMessage(assistantBlocks...))

		// Add tool results as user message
		if len(toolResults) > 0 {
			c.messages = append(c.messages, anthropic.NewUserMessage(toolResults...))
		}
	}

	return "", fmt.Errorf("max iterations (%d) reached in tool use loop", maxIterations)
}
