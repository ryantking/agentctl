package cli

import (
	"os"
	"path/filepath"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/setup"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	var globalInstall, force, noIndex bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Claude Code configuration",
		Long: `Initialize Claude Code configuration. Installs CLAUDE.md, agents, skills, and settings from the bundled templates directory.
By default, skips existing files.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var target string
			var err error

			if globalInstall {
				home, err := os.UserHomeDir()
				if err != nil {
					output.Errorf("failed to get home directory: %v", err)
					return err
				}
				target = filepath.Join(home, ".claude")
			} else {
				target, err = git.GetRepoRoot()
				if err != nil {
					output.Errorf("%v\n\nRun from inside a git repository or use --global", err)
					return err
				}
			}

			manager, err := setup.NewManager(target)
			if err != nil {
				output.Error(err)
				return err
			}

			if err := manager.Install(force, noIndex || globalInstall); err != nil {
				output.Error(err)
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&globalInstall, "global", "g", false, "Install to $HOME/.claude instead of current repository")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "Skip Claude CLI repository indexing")

	return cmd
}
