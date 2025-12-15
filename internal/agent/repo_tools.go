// Package anthropic provides Anthropic SDK client initialization and configuration.
package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
	return RegisterRepoToolsWithOptions(registry, repoRoot, false)
}

// RegisterRepoToolsWithOptions registers repository exploration tools with optional advanced tools.
// repoRoot: The root directory of the repository (for path validation)
// enableAdvanced: If true, registers additional analysis tools (search_files, get_file_info, list_git_files)
func RegisterRepoToolsWithOptions(registry *ToolRegistry, repoRoot string, enableAdvanced bool) error {
	// Define basic tools
	basicTools := []struct {
		name        string
		description string
		schema      map[string]interface{}
		handler     ToolHandler
	}{
		{
			name:        "list_directory",
			description: "List files and directories in a given path with type indicators (file/directory)",
			schema:      ListDirectorySchema,
			handler:     newListDirectoryHandler(repoRoot),
		},
		{
			name:        "read_file",
			description: "Read file contents from the repository. Returns error for binary files or files exceeding size limit",
			schema:      ReadFileSchema,
			handler:     newReadFileHandler(repoRoot),
		},
	}

	// Register basic tools
	for _, tool := range basicTools {
		if err := registry.RegisterTool(tool.name, tool.description, tool.schema, tool.handler); err != nil {
			return fmt.Errorf("failed to register %s tool: %w", tool.name, err)
		}
	}

	// Register advanced tools if enabled
	if enableAdvanced {
		if err := registerAdvancedTools(registry, repoRoot); err != nil {
			return fmt.Errorf("failed to register advanced tools: %w", err)
		}
	}

	return nil
}

// registerAdvancedTools registers optional advanced repository analysis tools.
func registerAdvancedTools(registry *ToolRegistry, repoRoot string) error {
	// Define advanced tools
	advancedTools := []struct {
		name        string
		description string
		schema      map[string]interface{}
		handler     ToolHandler
	}{
		{
			name:        "search_files",
			description: "Search for text patterns in files (grep-like functionality). Returns file paths and matching lines",
			schema:      SearchFilesSchema,
			handler:     newSearchFilesHandler(repoRoot),
		},
		{
			name:        "get_file_info",
			description: "Get file metadata: size, permissions, last modified time",
			schema:      GetFileInfoSchema,
			handler:     newGetFileInfoHandler(repoRoot),
		},
		{
			name:        "list_git_files",
			description: "List only files tracked by git (ignores untracked and ignored files)",
			schema:      ListGitFilesSchema,
			handler:     newListGitFilesHandler(repoRoot),
		},
	}

	// Register advanced tools
	for _, tool := range advancedTools {
		if err := registry.RegisterTool(tool.name, tool.description, tool.schema, tool.handler); err != nil {
			return fmt.Errorf("failed to register %s tool: %w", tool.name, err)
		}
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

// searchFiles searches for a pattern in files within the repository.
func searchFiles(repoRoot, pattern, searchPath string, caseSensitive bool) (interface{}, error) {
	absSearchPath, err := validatePath(repoRoot, searchPath)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	info, err := os.Stat(absSearchPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	var searchPaths []string
	if info.IsDir() {
		// Walk directory
		err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Skip ignored paths
			if isIgnored(repoRoot, path) {
				return nil
			}

			// Skip binary files (check first 512 bytes)
			if info.Size() > 0 && info.Size() <= MaxFileSize {
				//nolint:gosec // Path is validated to be within repository root
				data, err := os.ReadFile(path)
				if err == nil && !isBinaryFile(data) {
					searchPaths = append(searchPaths, path)
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else if !isIgnored(repoRoot, absSearchPath) {
		// Single file
		searchPaths = []string{absSearchPath}
	}

	// Compile pattern (treat as regex)
	var re *regexp.Regexp
	if caseSensitive {
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile("(?i)" + pattern)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var matches []map[string]interface{}
	for _, filePath := range searchPaths {
		// Check file size
		info, err := os.Stat(filePath)
		if err != nil || info.Size() > MaxFileSize {
			continue
		}

		// Read file
		//nolint:gosec // Path is validated to be within repository root
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Skip binary files
		if isBinaryFile(content) {
			continue
		}

		// Search for pattern
		lines := strings.Split(string(content), "\n")
		var matchingLines []map[string]interface{}
		for i, line := range lines {
			if re.MatchString(line) {
				matchingLines = append(matchingLines, map[string]interface{}{
					"line_number": i + 1,
					"content":     line,
				})
			}
		}

		if len(matchingLines) > 0 {
			relPath, _ := filepath.Rel(repoRoot, filePath)
			matches = append(matches, map[string]interface{}{
				"path":          relPath,
				"match_count":   len(matchingLines),
				"matching_lines": matchingLines,
			})
		}
	}

	return map[string]interface{}{
		"pattern": pattern,
		"path":    searchPath,
		"matches": matches,
		"total":   len(matches),
	}, nil
}

// getFileInfo returns file metadata.
func getFileInfo(repoRoot, path string) (interface{}, error) {
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

	// Get file mode (permissions)
	mode := info.Mode()
	permissions := fmt.Sprintf("%04o", mode.Perm())

	return map[string]interface{}{
		"path":        path,
		"size":        info.Size(),
		"permissions": permissions,
		"modified":    info.ModTime().Format(time.RFC3339),
		"is_readonly": mode&0200 == 0,
	}, nil
}

// listGitFiles lists files tracked by git.
func listGitFiles(repoRoot, path string) (interface{}, error) {
	absPath, err := validatePath(repoRoot, path)
	if err != nil {
		return nil, err
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Use git ls-files to get tracked files
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Run git ls-files
	gitPath := relPath
	if gitPath == "." {
		gitPath = ""
	}

	// Use git.RunGit if available, otherwise fallback to exec
	var trackedFiles []string
	switch {
	case info.IsDir():
		// List files in directory
		output, err := git.RunGit(repoRoot, "ls-files", gitPath)
		if err != nil {
			return nil, fmt.Errorf("failed to list git files: %w", err)
		}

		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				trackedFiles = append(trackedFiles, line)
			}
		}
	default:
		// Single file - check if tracked
		output, err := git.RunGit(repoRoot, "ls-files", "--error-unmatch", relPath)
		if err != nil {
			// File not tracked
			return map[string]interface{}{
				"path":  path,
				"files": []string{},
				"total": 0,
			}, nil
		}
		trackedFiles = []string{strings.TrimSpace(output)}
	}

	// Get file info for each tracked file
	var files []map[string]interface{}
	for _, file := range trackedFiles {
		filePath := filepath.Join(repoRoot, file)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue // Skip files that don't exist
		}

		files = append(files, map[string]interface{}{
			"path":     file,
			"size":     fileInfo.Size(),
			"modified": fileInfo.ModTime().Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"path":  path,
		"files": files,
		"total": len(files),
	}, nil
}

// NewRepoToolRegistry creates a new tool registry with repository exploration tools registered.
func NewRepoToolRegistry() (*ToolRegistry, string, error) {
	return NewRepoToolRegistryWithOptions(false)
}

// NewRepoToolRegistryWithOptions creates a new tool registry with optional advanced tools.
func NewRepoToolRegistryWithOptions(enableAdvanced bool) (*ToolRegistry, string, error) {
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get repository root: %w", err)
	}

	registry := NewToolRegistry()
	if err := RegisterRepoToolsWithOptions(registry, repoRoot, enableAdvanced); err != nil {
		return nil, "", fmt.Errorf("failed to register repository tools: %w", err)
	}

	return registry, repoRoot, nil
}
