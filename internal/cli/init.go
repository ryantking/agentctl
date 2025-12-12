package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryantking/agentctl/internal/cli/memory"
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
		Long: `Initialize Claude Code configuration. Installs agents, skills, settings, and memory files (AGENTS.md and CLAUDE.md) from the bundled templates directory.
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

			if err := manager.Install(force, true); err != nil {
				output.Error(err)
				return err
			}

			// Install memory files (AGENTS.md and CLAUDE.md)
			if err := memory.InstallTemplate("AGENTS.md", target, force); err != nil {
				output.Error(err)
				return err
			}
			if err := memory.InstallTemplate("CLAUDE.md", target, force); err != nil {
				output.Error(err)
				return err
			}

			// Optionally run indexing (unless --no-index or --global)
			if !noIndex && !globalInstall {
				if err := memory.IndexRepository(target); err != nil {
					// Non-fatal: warn but continue
					fmt.Printf("  → Repository indexing skipped: %v\n", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&globalInstall, "global", "g", false, "Install to $HOME/.claude instead of current repository")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "Skip Claude CLI repository indexing")

	return cmd
}
