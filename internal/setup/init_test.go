package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstall(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test installation
	err = manager.Install(false, true) // force=false, skipIndex=true
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify agents directory exists
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("Agents directory should be created")
	}

	// Verify skills directory exists
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Error("Skills directory should be created")
	}

	// Verify settings.json exists
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.json should be created")
	}

	// Verify .mcp.json exists
	mcpPath := filepath.Join(tmpDir, ".mcp.json")
	if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
		t.Error(".mcp.json should be created")
	}

	// Verify CLAUDE.md is NOT installed by setup (memory init handles it)
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		t.Error("CLAUDE.md should NOT be installed by setup.Install() - memory init handles it")
	}
}

func TestInstallWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// First installation
	err = manager.Install(false, true)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Second installation with force
	err = manager.Install(true, true)
	if err != nil {
		t.Fatalf("Install() with force error = %v", err)
	}

	// Verify files still exist
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.json should still exist after force install")
	}
}
