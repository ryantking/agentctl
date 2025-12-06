// Package context provides utilities for managing context files.
package context //nolint:revive // Package name matches purpose, conflict with stdlib is acceptable

import (
	"io"
	"os"
	"path/filepath"
)

// CopyClaudeContext copies Claude local settings and CLAUDE.md to a workspace.
// Copies files that aren't tracked by git but are needed for Claude
// to have proper context and permissions in the new workspace.
func CopyClaudeContext(workspacePath, sourceRoot string) ([]string, error) {
	var copied []string

	// Files to copy (relative to repo root)
	filesToCopy := []string{
		".claude/settings.local.json",
		"CLAUDE.md",
		".mcp.json",
	}

	for _, relPath := range filesToCopy {
		sourceFile := filepath.Join(sourceRoot, relPath)
		destFile := filepath.Join(workspacePath, relPath)

		// Check if source exists
		if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
			continue
		}

		// Check if dest already exists
		if _, err := os.Stat(destFile); err == nil {
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil { //nolint:gosec // Context directories need to be readable
			return nil, err
		}

		// Copy file
		if err := copyFile(sourceFile, destFile); err != nil {
			return nil, err
		}

		copied = append(copied, relPath)
	}

	return copied, nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src) //nolint:gosec // Path is controlled, opening context files
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	dest, err := os.Create(dst) //nolint:gosec // Path is controlled, creating context files
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	// Copy file mode
	sourceInfo, err := source.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
