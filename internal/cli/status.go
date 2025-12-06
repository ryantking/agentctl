package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// StatusInfo represents system status information.
type StatusInfo struct {
	Claude struct {
		Installed bool   `json:"installed"`
		Version   string `json:"version,omitempty"`
		Path      string `json:"path,omitempty"`
	} `json:"claude"`
}

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of Claude Code",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := getClaudeInfo()
			printStatus(info)
			return nil
		},
	}
	return cmd
}

func getClaudeInfo() StatusInfo {
	var info StatusInfo

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return info
	}

	info.Claude.Installed = true
	info.Claude.Path = claudePath

	// Try to get version
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err == nil {
		info.Claude.Version = string(output)
	}

	return info
}

func printStatus(info StatusInfo) {
	fmt.Println("\n  Claude Code")
	fmt.Println("  " + "----------------------------------------")
	if info.Claude.Installed {
		fmt.Print("  Status:   ")
		fmt.Println("installed")
		version := info.Claude.Version
		if version == "" {
			version = "unknown"
		}
		fmt.Printf("  Version:  %s\n", version)
		fmt.Printf("  Path:     %s\n", info.Claude.Path)
	} else {
		fmt.Print("  Status:   ")
		fmt.Println("not installed")
	}
	fmt.Println()
}
