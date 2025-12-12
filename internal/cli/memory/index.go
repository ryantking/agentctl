// Package memory provides commands for managing agent memory files (AGENTS.md and CLAUDE.md).
package memory

import (
	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewMemoryIndexCmd creates the memory index command.
func NewMemoryIndexCmd() *cobra.Command {
	var timeout int

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Generate repository overview and inject into AGENTS.md",
		Long: `Generate repository overview using Claude CLI and inject it into AGENTS.md
between the REPOSITORY_INDEX_START and REPOSITORY_INDEX_END markers.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var target string
			var err error

			target, err = git.GetRepoRoot()
			if err != nil {
				output.Errorf("%v\n\nRun from inside a git repository", err)
				return err
			}

			if err := IndexRepository(target); err != nil {
				output.Error(err)
				return err
			}

			if err := output.SuccessJSON(map[string]string{"status": "indexed", "file": "AGENTS.md"}); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&timeout, "timeout", 90, "Timeout in seconds for Claude CLI (default: 90)")

	return cmd
}
