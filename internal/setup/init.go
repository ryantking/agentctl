// Package setup provides initialization and template installation functionality.
package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/config"
	"github.com/ryantking/agentctl/internal/templates"
)

// Manager manages Claude Code initialization.
type Manager struct {
	target      string
	templateDir string
}

// NewManager creates a new initialization manager.
func NewManager(target string) (*Manager, error) {
	return &Manager{
		target:      target,
		templateDir: "templates", // Embedded templates path
	}, nil
}

// Install executes full initialization.
func (m *Manager) Install(force, _ bool) error {
	// 1. Install agents
	fmt.Println("Installing agents...")
	count, err := m.installDirectory("agents", filepath.Join(m.target, ".claude", "agents"), force, false, "*.md")
	if err != nil {
		return err
	}
	fmt.Printf("  → Installed %d agent(s)\n", count)

	// 2. Install skills
	fmt.Println("Installing skills...")
	count, err = m.installDirectory("skills", filepath.Join(m.target, ".claude", "skills"), force, true, "")
	if err != nil {
		return err
	}
	fmt.Printf("  → Installed %d skill(s)\n", count)

	// 3. Merge settings
	fmt.Println("Merging settings.json...")
	if err := m.mergeSettings(force); err != nil {
		return err
	}

	// 4. Configure MCP servers
	fmt.Println("Configuring MCP servers...")
	if err := m.configureMCP(force); err != nil {
		return err
	}

	fmt.Println("\n✓ Initialization complete")
	return nil
}

func (m *Manager) installFile(templatePath, destPath string, force bool) error {
	if _, err := os.Stat(destPath); err == nil && !force {
		relPath, _ := filepath.Rel(m.target, destPath)
		fmt.Printf("  • %s (skipped)\n", relPath)
		return nil
	}

	data, err := templates.GetTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil { //nolint:gosec // Template directories need to be readable
		return err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil { //nolint:gosec // Template files need to be readable
		return err
	}

	relPath, _ := filepath.Rel(m.target, destPath)
	status := "overwritten"
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		status = "created"
	}
	fmt.Printf("  • %s (%s)\n", relPath, status)
	return nil
}

func (m *Manager) installDirectory(templateDir, destDir string, force, recursive bool, pattern string) (int, error) {
	count := 0

	if recursive {
		// Copy entire directory trees (for skills)
		entries, err := templates.FS.ReadDir("templates/" + templateDir)
		if err != nil {
			return 0, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			destItem := filepath.Join(destDir, entry.Name())
			existed := false
			if _, err := os.Stat(destItem); err == nil {
				existed = true
				if !force {
					relPath, _ := filepath.Rel(m.target, destItem)
					fmt.Printf("  • %s (skipped)\n", relPath)
					continue
				}
			}

			if err := m.copyTree(
				filepath.Join("templates", templateDir, entry.Name()),
				destItem,
			); err != nil {
				return count, err
			}

			relPath, _ := filepath.Rel(m.target, destItem)
			status := "overwritten"
			if !existed {
				status = "created"
			}
			fmt.Printf("  • %s (%s)\n", relPath, status)
			count++
		}
	} else {
		// Copy matching files (for agents)
		entries, err := templates.FS.ReadDir("templates/" + templateDir)
		if err != nil {
			return 0, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if pattern != "" && !matchPattern(entry.Name(), pattern) {
				continue
			}

			destPath := filepath.Join(destDir, entry.Name())
			if err := m.installFile(
				filepath.Join(templateDir, entry.Name()),
				destPath,
				force,
			); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

func (m *Manager) copyTree(srcPath, destPath string) error {
	if err := os.MkdirAll(destPath, 0755); err != nil { //nolint:gosec // Template directories need to be readable
		return err
	}

	entries, err := templates.FS.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcItem := filepath.Join(srcPath, entry.Name())
		destItem := filepath.Join(destPath, entry.Name())

		if entry.IsDir() {
			if err := m.copyTree(srcItem, destItem); err != nil {
				return err
			}
		} else {
			data, err := templates.FS.ReadFile(srcItem)
			if err != nil {
				return err
			}
			if err := os.WriteFile(destItem, data, 0644); err != nil { //nolint:gosec // Template files need to be readable
				return err
			}
		}
	}

	return nil
}

func (m *Manager) mergeSettings(force bool) error {
	sourcePath := filepath.Join(m.target, ".claude", "settings.json")
	destPath := sourcePath

	// Load source settings
	sourceData, err := templates.GetTemplate("settings.json")
	if err != nil {
		return fmt.Errorf("failed to read template settings.json: %w", err)
	}

	newSettings, err := config.LoadJSON(sourceData)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil { //nolint:gosec // Template directories need to be readable
		return err
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		// No existing settings - just copy
		data, err := config.SaveJSON(newSettings)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, append(data, '\n'), 0644); err != nil { //nolint:gosec // Template files need to be readable //nolint:gosec // Template files need to be readable
			return err
		}
		relPath, _ := filepath.Rel(m.target, destPath)
		fmt.Printf("  • %s (created)\n", relPath)
		return nil
	}

	// Existing settings - merge
	existingData, err := os.ReadFile(destPath) //nolint:gosec // Path is controlled, reading template files
	if err != nil {
		return err
	}

	existingSettings, err := config.LoadJSON(existingData)
	if err != nil {
		return err
	}

	if force {
		// Force: overwrite
		data, err := config.SaveJSON(newSettings)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, append(data, '\n'), 0644); err != nil { //nolint:gosec // Template files need to be readable //nolint:gosec // Template files need to be readable
			return err
		}
		relPath, _ := filepath.Rel(m.target, destPath)
		fmt.Printf("  • %s (overwritten)\n", relPath)
		return nil
	}

	// Smart merge
	merged := config.Merge(existingSettings, newSettings)
	data, err := config.SaveJSON(merged)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destPath, append(data, '\n'), 0644); err != nil { //nolint:gosec // Template files need to be readable
		return err
	}
	relPath, _ := filepath.Rel(m.target, destPath)
	fmt.Printf("  • %s (merged)\n", relPath)
	return nil
}

func (m *Manager) configureMCP(force bool) error {
	destPath := filepath.Join(m.target, ".mcp.json")

	// New MCP servers to add
	newServers := map[string]interface{}{
		"context7": map[string]interface{}{
			"type": "http",
			"url":  "https://mcp.context7.com/mcp",
		},
		"linear": map[string]interface{}{
			"type": "sse",
			"url":  "https://mcp.linear.app/sse",
		},
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil { //nolint:gosec // Template directories need to be readable
		return err
	}

	var mcpConfig map[string]interface{}
	var status string

	if _, err := os.Stat(destPath); err == nil && !force {
		// Load existing config
		data, err := os.ReadFile(destPath) //nolint:gosec // Path is controlled, reading template files
		if err != nil {
			return err
		}

		existingConfig, err := config.LoadJSON(data)
		if err != nil {
			return err
		}

		// Ensure mcpServers key exists
		if _, ok := existingConfig["mcpServers"]; !ok {
			existingConfig["mcpServers"] = make(map[string]interface{})
		}

		servers := existingConfig["mcpServers"].(map[string]interface{})
		addedAny := false

		// Merge new servers (don't overwrite existing ones)
		for serverName, serverConfig := range newServers {
			if _, exists := servers[serverName]; !exists {
				servers[serverName] = serverConfig
				addedAny = true
			}
		}

		if !addedAny {
			relPath, _ := filepath.Rel(m.target, destPath)
			fmt.Printf("  • %s (skipped)\n", relPath)
			return nil
		}

		mcpConfig = existingConfig
		status = "merged"
	} else {
		// Create new config or force overwrite
		mcpConfig = map[string]interface{}{
			"mcpServers": newServers,
		}
		if _, err := os.Stat(destPath); err == nil {
			status = "overwritten"
		} else {
			status = "created"
		}
	}

	// Write MCP configuration
	data, err := config.SaveJSON(mcpConfig)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destPath, append(data, '\n'), 0644); err != nil { //nolint:gosec // Template files need to be readable
		return err
	}

	relPath, _ := filepath.Rel(m.target, destPath)
	fmt.Printf("  • %s (%s)\n", relPath, status)
	return nil
}

func matchPattern(name, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*.") {
		ext := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(name, ext)
	}
	return name == pattern
}
