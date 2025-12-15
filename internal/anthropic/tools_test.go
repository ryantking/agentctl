package anthropic

import (
	"context"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestToolRegistry_RegisterTool(t *testing.T) {
	registry := NewToolRegistry()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path to read",
			},
		},
		"required": []interface{}{"path"},
	}

	err := registry.RegisterTool("read_file", "Read a file from the filesystem", schema, func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
		return "file content", nil
	})
	if err != nil {
		t.Fatalf("RegisterTool() error = %v", err)
	}

	if len(registry.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(registry.tools))
	}

	if registry.tools[0].OfTool == nil {
		t.Error("Expected tool to be set")
	}

	if registry.tools[0].OfTool.Name != "read_file" {
		t.Errorf("Expected tool name 'read_file', got %q", registry.tools[0].OfTool.Name)
	}
}

func TestToolRegistry_ExecuteTool(t *testing.T) {
	registry := NewToolRegistry()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"value": map[string]interface{}{
				"type": "string",
			},
		},
	}

	registry.RegisterTool("test_tool", "Test tool", schema, func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
		value, _ := input["value"].(string)
		return map[string]interface{}{"result": value}, nil
	})

	result, err := registry.ExecuteTool(context.Background(), "test_tool", map[string]interface{}{"value": "test"})
	if err != nil {
		t.Fatalf("ExecuteTool() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	if resultMap["result"] != "test" {
		t.Errorf("Expected result 'test', got %v", resultMap["result"])
	}
}

func TestToolRegistry_ExecuteToolNotFound(t *testing.T) {
	registry := NewToolRegistry()

	_, err := registry.ExecuteTool(context.Background(), "nonexistent", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestNewConversation(t *testing.T) {
	client, err := NewClientOrNil()
	if err != nil {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	registry := NewToolRegistry()
	conv := NewConversation(client, registry)

	if conv.client.Options == nil {
		t.Error("Expected client to be set")
	}

	if conv.registry != registry {
		t.Error("Expected registry to be set")
	}

	if len(conv.messages) != 0 {
		t.Errorf("Expected empty messages, got %d", len(conv.messages))
	}
}

func TestConversation_SetSystem(t *testing.T) {
	client, err := NewClientOrNil()
	if err != nil {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	registry := NewToolRegistry()
	conv := NewConversation(client, registry)

	conv.SetSystem("You are a helpful assistant")
	if len(conv.system) != 1 {
		t.Errorf("Expected 1 system message, got %d", len(conv.system))
	}

	if conv.system[0].Text != "You are a helpful assistant" {
		t.Errorf("Expected system text 'You are a helpful assistant', got %q", conv.system[0].Text)
	}
}

func TestConversation_AddUserMessage(t *testing.T) {
	client, err := NewClientOrNil()
	if err != nil {
		t.Skip("ANTHROPIC_API_KEY not set, skipping test")
	}

	registry := NewToolRegistry()
	conv := NewConversation(client, registry)

	conv.AddUserMessage("Hello")
	if len(conv.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(conv.messages))
	}
}
