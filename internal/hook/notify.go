package hook

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ryantking/agentctl/internal/notify"
)

// detectAgent detects the agent type and returns (appName, sender).
// Returns ("Claude Code", sender) or ("Cursor", sender) or ("Cursor Agent", sender) based on environment.
func detectAgent() (string, string) {
	// Check for Cursor Agent TUI (terminal-based)
	// CURSOR_AGENT is set to "1" when running in Cursor Agent TUI
	if cursorAgent := os.Getenv("CURSOR_AGENT"); cursorAgent != "" {
		// If CURSOR_CLI is also set, it's Cursor IDE, not Agent TUI
		if os.Getenv("CURSOR_CLI") == "" {
			return "Cursor Agent", notify.SenderCursor
		}
		// CURSOR_CLI is set, so it's Cursor IDE
		return "Cursor", notify.SenderCursor
	}
	
	// Check for Cursor IDE (desktop app) - CURSOR_CLI or CURSOR_CLI_MODE indicates IDE
	if os.Getenv("CURSOR_CLI") != "" || os.Getenv("CURSOR_CLI_MODE") != "" {
		return "Cursor", notify.SenderCursor
	}
	
	// Check for Claude Code environment variables
	// Claude Code typically sets these when running
	if os.Getenv("CLAUDE_CODE") != "" || 
	   os.Getenv("ANTHROPIC_CLAUDE") != "" ||
	   os.Getenv("CLAUDE_DESKTOP") != "" {
		return "Claude Code", notify.SenderClaudeCode
	}
	
	// Check for explicit sender override
	if sender := os.Getenv("AGENTCTL_NOTIFICATION_SENDER"); sender != "" {
		// Try to infer app name from sender
		if sender == notify.SenderCursor {
			return "Cursor", sender
		}
		if sender == notify.SenderClaudeCode {
			return "Claude Code", sender
		}
		return "Agent", sender
	}
	
	// No known agent detected - return empty sender (no custom icon)
	return "Claude Code", ""
}

// NotifyInput sends notification when Claude needs input.
func NotifyInput(message string) error {
	appName, sender := detectAgent()
	return NotifyInputWithSender(message, appName, sender)
}

// NotifyInputWithSender sends notification with a custom sender.
func NotifyInputWithSender(message string, appName, sender string) error {
	projectName := getProjectName()
	if message == "" {
		message = "Claude needs your input to continue"
	}
	return notify.Send(notify.Options{
		Title:    appName,
		Subtitle: projectName,
		Message:  message,
		Sound:    "",
		Group:    fmt.Sprintf("claude-code-%s", projectName),
		Sender:   sender,
	})
}

// NotifyStop sends notification when Claude completes a task.
func NotifyStop(transcriptPath string) error {
	appName, sender := detectAgent()
	return NotifyStopWithSender(transcriptPath, appName, sender)
}

// NotifyStopWithSender sends stop notification with a custom sender.
func NotifyStopWithSender(transcriptPath string, appName, sender string) error {
	projectName := getProjectName()
	timeStr := getTime()

	message := fmt.Sprintf("Completed at %s", timeStr)
	if transcriptPath != "" {
		if finalResponse := extractFinalResponse(transcriptPath, 200); finalResponse != "" {
			message = finalResponse
		}
	}

	return notify.Send(notify.Options{
		Title:    fmt.Sprintf("✅ %s", appName),
		Subtitle: projectName,
		Message:  message,
		Sound:    "",
		Group:    fmt.Sprintf("claude-code-%s", projectName),
		Sender:   sender,
	})
}

// NotifyError sends error notification.
func NotifyError(message string) error {
	appName, sender := detectAgent()
	return NotifyErrorWithSender(message, appName, sender)
}

// NotifyErrorWithSender sends error notification with a custom sender.
func NotifyErrorWithSender(message string, appName, sender string) error {
	projectName := getProjectName()
	if message == "" {
		message = "An error occurred during task execution"
	}
	return notify.Send(notify.Options{
		Title:    appName,
		Subtitle: projectName,
		Message:  message,
		Sound:    "Basso",
		Group:    fmt.Sprintf("claude-code-%s", projectName),
		Sender:   sender,
	})
}

func getProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(cwd)
}

func getTime() string {
	return time.Now().Format("3:04 PM")
}

func extractFinalResponse(transcriptPath string, maxLength int) string {
	path := filepath.Clean(transcriptPath)
	if !filepath.IsAbs(path) {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(home, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	var lastResponse string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry["type"] == "assistant" {
			if message, ok := entry["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].([]interface{}); ok {
					for _, block := range content {
						if blockMap, ok := block.(map[string]interface{}); ok {
							if blockMap["type"] == "text" {
								if text, ok := blockMap["text"].(string); ok {
									lastResponse = text
								}
							}
						} else if text, ok := block.(string); ok {
							lastResponse = text
						}
					}
				}
			}
		}
	}

	if lastResponse == "" {
		return ""
	}

	// Truncate and clean up for notification
	text := strings.TrimSpace(lastResponse)
	firstLine := strings.Split(text, "\n")[0]

	// Strip markdown formatting
	re := regexp.MustCompile(`\*\*(.+?)\*\*`)
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile(`\*(.+?)\*`)
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile("`(.+?)`")
	firstLine = re.ReplaceAllString(firstLine, "$1")
	re = regexp.MustCompile(`^#+\s*`)
	firstLine = re.ReplaceAllString(firstLine, "")

	if len(firstLine) > maxLength {
		return firstLine[:maxLength-3] + "..."
	}
	return firstLine
}
