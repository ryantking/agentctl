// Package notify provides macOS notification functionality with graceful fallback.
package notify

import (
	"fmt"
	"os/exec"
)

// Common sender bundle IDs for macOS notifications.
const (
	// SenderClaudeCode is the bundle ID for Claude Desktop.
	SenderClaudeCode = "com.anthropic.claudefordesktop"
	
	// SenderCursor is the bundle ID for Cursor.
	SenderCursor = "com.todesktop.230313mzl4w4u92"
)

// Options contains notification options.
type Options struct {
	Title    string
	Subtitle string
	Message  string
	Sound    string
	Group    string
	Sender   string // macOS bundle ID for custom icon (e.g., "com.anthropic.claudefordesktop", "com.todesktop.230313mzl4w4u92")
}

// Send sends a macOS notification with the given options.
// Uses terminal-notifier if available (supports custom sender/icons), otherwise falls back to osascript.
func Send(opts Options) error {
	if hasTerminalNotifier() {
		return sendWithTerminalNotifier(opts)
	}
	return sendWithOSAScript(opts)
}

// hasTerminalNotifier checks if terminal-notifier is available.
func hasTerminalNotifier() bool {
	_, err := exec.LookPath("terminal-notifier")
	return err == nil
}

// sendWithTerminalNotifier sends notification using terminal-notifier (supports custom sender).
func sendWithTerminalNotifier(opts Options) error {
	args := []string{
		"-title", opts.Title,
		"-subtitle", opts.Subtitle,
		"-message", opts.Message,
	}
	
	// Add sender if provided (for custom icons)
	if opts.Sender != "" {
		args = append(args, "-sender", opts.Sender)
	}
	
	if opts.Sound != "" {
		args = append(args, "-sound", opts.Sound)
	}
	if opts.Group != "" {
		args = append(args, "-group", opts.Group)
	}

	cmd := exec.Command("terminal-notifier", args...) //nolint:gosec // terminal-notifier is a trusted local binary
	return cmd.Run()
}

// sendWithOSAScript sends notification using osascript (fallback, no custom sender support).
func sendWithOSAScript(opts Options) error {
	soundClause := ""
	if opts.Sound != "" {
		soundClause = fmt.Sprintf(` sound name "%s"`, opts.Sound)
	}
	script := fmt.Sprintf(`display notification "%s" with title "%s" subtitle "%s"%s`,
		opts.Message, opts.Title, opts.Subtitle, soundClause)
	cmd := exec.Command("osascript", "-e", script) //nolint:gosec // osascript is a trusted system binary
	return cmd.Run()
}

// HasTerminalNotifier returns whether terminal-notifier is available.
// Useful for showing installation hints to users.
func HasTerminalNotifier() bool {
	return hasTerminalNotifier()
}
