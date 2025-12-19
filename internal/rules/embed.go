// Package rules provides access to embedded default rule files.
package rules

import (
	"embed"
	"path/filepath"
)

//go:embed all:rules
// FS is the embedded filesystem containing default rule files.
// Note: .agent/rules/ is the source of truth for rules. The rules/ directory
// here is kept synchronized with .agent/rules/ for embedding purposes, as Go's
// embed directive cannot reference paths outside the package directory.
var FS embed.FS

// GetRule reads a rule file from the embedded filesystem.
func GetRule(name string) ([]byte, error) {
	return FS.ReadFile(filepath.Join("rules", name))
}

// ReadRulesDir reads the rules directory from the embedded filesystem.
func ReadRulesDir() ([]string, error) {
	entries, err := FS.ReadDir("rules")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
