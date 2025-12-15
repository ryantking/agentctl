// Package anthropic provides tool handler constructors for repository exploration tools.
package agent

import (
	"context"
	"fmt"
)

// newListDirectoryHandler creates a handler for the list_directory tool.
func newListDirectoryHandler(repoRoot string) ToolHandler {
	return func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		return listDirectory(repoRoot, path)
	}
}

// newReadFileHandler creates a handler for the read_file tool.
func newReadFileHandler(repoRoot string) ToolHandler {
	return func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		return readFile(repoRoot, path)
	}
}

// newSearchFilesHandler creates a handler for the search_files tool.
func newSearchFilesHandler(repoRoot string) ToolHandler {
	return func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		pattern, ok := input["pattern"].(string)
		if !ok {
			return nil, fmt.Errorf("pattern must be a string")
		}

		path := "."
		if p, ok := input["path"].(string); ok && p != "" {
			path = p
		}

		caseSensitive := false
		if cs, ok := input["case_sensitive"].(bool); ok {
			caseSensitive = cs
		}

		return searchFiles(repoRoot, pattern, path, caseSensitive)
	}
}

// newGetFileInfoHandler creates a handler for the get_file_info tool.
func newGetFileInfoHandler(repoRoot string) ToolHandler {
	return func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path, ok := input["path"].(string)
		if !ok {
			return nil, fmt.Errorf("path must be a string")
		}

		return getFileInfo(repoRoot, path)
	}
}

// newListGitFilesHandler creates a handler for the list_git_files tool.
func newListGitFilesHandler(repoRoot string) ToolHandler {
	return func(_ context.Context, input map[string]interface{}) (interface{}, error) {
		path := "."
		if p, ok := input["path"].(string); ok && p != "" {
			path = p
		}

		return listGitFiles(repoRoot, path)
	}
}
