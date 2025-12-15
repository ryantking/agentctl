package anthropic

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		wantInRepo bool
	}{
		{
			name:       "valid relative path",
			path:       "subdir",
			wantErr:    false,
			wantInRepo: true,
		},
		{
			name:       "current directory",
			path:       ".",
			wantErr:    false,
			wantInRepo: true,
		},
		{
			name:       "empty path",
			path:       "",
			wantErr:    false,
			wantInRepo: true,
		},
		{
			name:       "path traversal attempt",
			path:       "../outside",
			wantErr:    true,
			wantInRepo: false,
		},
		{
			name:       "nested path traversal",
			path:       "subdir/../../outside",
			wantErr:    true,
			wantInRepo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := validatePath(repoRoot, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				relPath, relErr := filepath.Rel(repoRoot, absPath)
				if relErr != nil {
					t.Errorf("Failed to get relative path: %v", relErr)
					return
				}
				if strings.HasPrefix(relPath, "..") != !tt.wantInRepo {
					t.Errorf("Path in repo = %v, want %v", !strings.HasPrefix(relPath, ".."), tt.wantInRepo)
				}
			}
		})
	}
}

func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	// Create test files and directories
	testDir := filepath.Join(repoRoot, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(repoRoot, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := listDirectory(repoRoot, ".")
	if err != nil {
		t.Fatalf("listDirectory() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	items, ok := resultMap["items"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected items array, got %T", resultMap["items"])
	}

	if len(items) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(items))
	}

	// Check that we have both file and directory
	hasFile := false
	hasDir := false
	for _, item := range items {
		if item["name"] == "test.txt" && item["type"] == "file" {
			hasFile = true
		}
		if item["name"] == "testdir" && item["type"] == "directory" {
			hasDir = true
		}
	}

	if !hasFile {
		t.Error("Expected to find test.txt file")
	}
	if !hasDir {
		t.Error("Expected to find testdir directory")
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	testFile := filepath.Join(repoRoot, "test.txt")
	testContent := "Hello, world!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := readFile(repoRoot, "test.txt")
	if err != nil {
		t.Fatalf("readFile() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	if resultMap["content"] != testContent {
		t.Errorf("Expected content %q, got %q", testContent, resultMap["content"])
	}
}

func TestReadFile_Binary(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	// Create a binary file (with null bytes)
	binaryFile := filepath.Join(repoRoot, "binary.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
	if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	_, err := readFile(repoRoot, "binary.bin")
	if err == nil {
		t.Error("Expected error for binary file, got nil")
	}
	if !strings.Contains(err.Error(), "binary") {
		t.Errorf("Expected binary file error, got: %v", err)
	}
}

func TestRegisterRepoTools(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	registry := NewToolRegistry()
	err := RegisterRepoTools(registry, repoRoot)
	if err != nil {
		t.Fatalf("RegisterRepoTools() error = %v", err)
	}

	// Check that tools are registered
	tools := registry.GetTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Check tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		if tool.OfTool != nil {
			toolNames[tool.OfTool.Name] = true
		}
	}

	if !toolNames["list_directory"] {
		t.Error("Expected list_directory tool to be registered")
	}
	if !toolNames["read_file"] {
		t.Error("Expected read_file tool to be registered")
	}
}

func TestToolExecution(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatalf("Failed to create repo root: %v", err)
	}

	testFile := filepath.Join(repoRoot, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	registry := NewToolRegistry()
	if err := RegisterRepoTools(registry, repoRoot); err != nil {
		t.Fatalf("RegisterRepoTools() error = %v", err)
	}

	// Test list_directory execution
	result, err := registry.ExecuteTool(context.Background(), "list_directory", map[string]interface{}{
		"path": ".",
	})
	if err != nil {
		t.Fatalf("ExecuteTool(list_directory) error = %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result from list_directory")
	}

	// Test read_file execution
	result, err = registry.ExecuteTool(context.Background(), "read_file", map[string]interface{}{
		"path": "test.txt",
	})
	if err != nil {
		t.Fatalf("ExecuteTool(read_file) error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	if resultMap["content"] != "test" {
		t.Errorf("Expected content 'test', got %v", resultMap["content"])
	}
}
