package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionInfo = struct {
		version string
		commit  string
		date    string
	}{
		version: "dev",
		commit:  "none",
		date:    "unknown",
	}
)

// SetVersion sets the version information.
func SetVersion(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the current version",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("agentctl %s\n", versionInfo.version)
			return nil
		},
	}
	return cmd
}
