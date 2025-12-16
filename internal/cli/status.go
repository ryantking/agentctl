package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	agentclient "github.com/ryantking/agentctl/internal/agent"
	"github.com/spf13/cobra"
)

// getAgentCLIPathFromEnv returns the agent CLI path from environment variable.
// Used when flag is not available (e.g., in status command).
func getAgentCLIPathFromEnv() string {
	if envPath := os.Getenv("AGENTCTL_CLI_PATH"); envPath != "" {
		return envPath
	}
	return "claude"
}

// StatusInfo represents system status information.
type StatusInfo struct {
	Authenticated bool   `json:"authenticated"`
	AuthMethod    string `json:"auth_method,omitempty"`
	APIConnected  bool   `json:"api_connected,omitempty"`
}

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of agentctl and Anthropic API authentication",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := getStatusInfo()
			printStatus(info)
			return nil
		},
	}
	return cmd
}

func getStatusInfo() StatusInfo {
	var info StatusInfo

	// Check if claude CLI is available (handles auth automatically)
	if agentclient.IsConfigured() {
		info.Authenticated = true
		// Determine auth method
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			info.AuthMethod = "API key"
		} else {
			info.AuthMethod = "Claude Code session"
		}
		// Test connectivity
		info.APIConnected = testAPIConnectivity()
	} else {
		info.Authenticated = false
		info.APIConnected = false
	}

	return info
}

func testAPIConnectivity() bool {
	// Check if claude CLI is available and can execute
	cliPath := getAgentCLIPathFromEnv()
	agent := agentclient.NewAgent(agentclient.WithBinary(cliPath))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple test prompt
	_, err := agent.ExecuteWithLogger(ctx, "test", nil)
	if err != nil {
		// If it's just a content error (empty response), CLI is working
		if strings.Contains(err.Error(), "empty output") {
			return true
		}
		return false
	}

	return true
}

func printStatus(info StatusInfo) {
	fmt.Println("\n  agentctl Status")
	fmt.Println("  " + "----------------------------------------")
	
	if info.Authenticated {
		fmt.Print("  Authenticated: ")
		fmt.Println("Yes")
		if info.AuthMethod != "" {
			fmt.Printf("  Method:        %s\n", info.AuthMethod)
		}
		if info.APIConnected {
			fmt.Println("  API:           Connected")
		} else {
			fmt.Println("  API:           Connection test skipped")
		}
	} else {
		fmt.Print("  Authenticated: ")
		fmt.Println("No")
		fmt.Println("\n  To authenticate:")
		fmt.Println("    - Set ANTHROPIC_API_KEY environment variable")
		fmt.Println("    - Or use Claude Code (automatic authentication)")
		fmt.Println("    - Get your API key at https://console.anthropic.com/")
	}
	fmt.Println()
}
