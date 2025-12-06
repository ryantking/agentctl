// Package templates provides access to embedded template files.
package templates

import (
	"embed"
	"path/filepath"
)

//go:embed all:templates
// FS is the embedded filesystem containing template files.
var FS embed.FS

// GetTemplate reads a template file from the embedded filesystem.
func GetTemplate(name string) ([]byte, error) {
	return FS.ReadFile(filepath.Join("templates", name))
}

// ReadDir reads a directory from the embedded filesystem.
func ReadDir(name string) ([]string, error) {
	entries, err := FS.ReadDir(filepath.Join("templates", name))
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}
