// Package anthropic provides Anthropic SDK client initialization and configuration.
package anthropic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
)

const (
	// MaxFileSize is the maximum file size to read (1MB)
	MaxFileSize = 1024 * 1024
	// BinaryCheckSize is the number of bytes to check for binary content
	BinaryCheckSize = 512
	// MaxDirectoryDepth is the maximum directory depth for list_directory
	MaxDirectoryDepth = 10
)

// RegisterRepoTools registers repository exploration tools (list_directory, read_file).
// repoRoot: The root directory of the repository (for path validation)
func RegisterRepoTools(registry *ToolRegistry, repoRoot string) error {
	// Register list_directory tool
	listDirSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list (relative to repository root)",
			},
		},
		"required": []interface{}{"path"},
	}

	err := registry.RegisterTool("list_directory", "List files and directories in a given path with type indicators (file/directory)", listDirSchema, func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		return listDirectory(repoRoot, path)
	})
	if err != nil {
		return fmt.Errorf("failed to register list_directory tool: %w", err)
	}

	// Register read_file tool
	readFileSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path to read (relative to repository root)",
			},
		},
		"required": []interface{}{"path"},
	}

	err = registry.RegisterTool("read_file", "Read file contents from the repository. Returns error for binary files or files exceeding size limit", readFileSchema, func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		return readFile(repoRoot, path)
	})
	if err != nil {
		return fmt.Errorf("failed to register read_file tool: %w", err)
	}

	return nil
}

// validatePath ensures the path is within the repository root and returns the absolute path.
func validatePath(repoRoot, path string) (string, error) {
	// Clean and resolve the path
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == "" {
		return repoRoot, nil
	}

	// Resolve relative to repo root
	absPath := filepath.Join(repoRoot, cleanPath)
	absPath = filepath.Clean(absPath)

	// Ensure the resolved path is within repo root
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check for path traversal attempts
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", path)
	}

	return absPath, nil
}

// isIgnored checks if a path should be ignored based on .gitignore rules.
// This is a simplified implementation - for production use, consider a proper gitignore parser.
func isIgnored(repoRoot, path string) bool {
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return false
	}

	// Check common ignored patterns
	ignoredPatterns := []string{
		".git/",
		"node_modules/",
		".claude/scratch/",
		"vendor/",
		"__pycache__/",
		".pytest_cache/",
		".mypy_cache/",
		"*.pyc",
		"*.pyo",
		".DS_Store",
	}

	pathParts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range pathParts {
		for _, pattern := range ignoredPatterns {
			switch {
			case strings.HasPrefix(pattern, "*"):
				// Simple suffix match
				if strings.HasSuffix(part, strings.TrimPrefix(pattern, "*")) {
					return true
				}
			case strings.HasSuffix(pattern, "/"):
				// Directory match
				if part == strings.TrimSuffix(pattern, "/") {
					return true
				}
			default:
				// Exact match
				if part == pattern {
					return true
				}
			}
		}
	}

	return false
}

// listDirectory lists files and directories in the given path.
func listDirectory(repoRoot, path string) (interface{}, error) {
	return listDirectoryWithDepth(repoRoot, path, 0)
}

// listDirectoryWithDepth lists files and directories with depth tracking.
func listDirectoryWithDepth(repoRoot, path string, depth int) (interface{}, error) {
	// Check depth limit
	if depth >= MaxDirectoryDepth {
		return nil, fmt.Errorf("directory depth limit (%d) exceeded", MaxDirectoryDepth)
	}

	absPath, err := validatePath(repoRoot, path)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// If it's a file, return file info
	if !info.IsDir() {
		return map[string]interface{}{
			"path":  path,
			"type":  "file",
			"error": "path is a file, not a directory",
		}, nil
	}

	// Read directory contents
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var items []map[string]interface{}
	for _, entry := range entries {
		absEntryPath := filepath.Join(absPath, entry.Name())

		// Skip ignored paths
		if isIgnored(repoRoot, absEntryPath) {
			continue
		}

		item := map[string]interface{}{
			"name": entry.Name(),
			"type": "directory",
		}

		if !entry.IsDir() {
			item["type"] = "file"
			// Get file size
			if info, err := entry.Info(); err == nil {
				item["size"] = info.Size()
			}
		}

		items = append(items, item)
	}

	return map[string]interface{}{
		"path":  path,
		"items": items,
	}, nil
}

// isBinaryFile checks if a file is binary by examining its content.
func isBinaryFile(content []byte) bool {
	// Check for null bytes (common in binary files)
	checkSize := BinaryCheckSize
	if len(content) < checkSize {
		checkSize = len(content)
	}

	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return true
		}
	}

	return false
}

// readFile reads file contents from the repository.
func readFile(repoRoot, path string) (interface{}, error) {
	absPath, err := validatePath(repoRoot, path)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Check file size
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", info.Size(), MaxFileSize)
	}

	// Check if ignored
	if isIgnored(repoRoot, absPath) {
		return nil, fmt.Errorf("file is ignored (matches .gitignore patterns)")
	}

	// Read file
	// Path is validated by validatePath to be within repo root
	//nolint:gosec // Path is validated to be within repository root
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if binary
	if isBinaryFile(content) {
		return nil, fmt.Errorf("file appears to be binary (contains null bytes)")
	}

	return map[string]interface{}{
		"path":    path,
		"content": string(content),
		"size":    len(content),
	}, nil
}

// NewRepoToolRegistry creates a new tool registry with repository exploration tools registered.
func NewRepoToolRegistry() (*ToolRegistry, string, error) {
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get repository root: %w", err)
	}

	registry := NewToolRegistry()
	if err := RegisterRepoTools(registry, repoRoot); err != nil {
		return nil, "", fmt.Errorf("failed to register repository tools: %w", err)
	}

	return registry, repoRoot, nil
}
