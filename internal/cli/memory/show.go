package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewMemoryShowCmd creates the memory show command.
func NewMemoryShowCmd() *cobra.Command {
	var resolve, jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show [file]",
		Short: "Display memory file contents",
		Long: `Display memory file contents. If no file is specified, shows both AGENTS.md and CLAUDE.md.
Use --resolve to expand @imports inline. Use --json for structured output.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var target string
			var err error

			target, err = git.GetRepoRoot()
			if err != nil {
				// Try current directory if not in git repo
				target, err = os.Getwd()
				if err != nil {
					output.Errorf("failed to determine target directory: %v", err)
					return err
				}
			}

			var files []string
			if len(args) > 0 {
				files = []string{args[0]}
			} else {
				files = []string{"AGENTS.md", "CLAUDE.md"}
			}

			if jsonOutput {
				return showJSON(target, files, resolve)
			}

			return showText(target, files, resolve)
		},
	}

	cmd.Flags().BoolVarP(&resolve, "resolve", "r", false, "Expand @imports inline")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}

func showText(target string, files []string, resolve bool) error {
	for _, fileName := range files {
		filePath := filepath.Join(target, fileName)
		data, err := os.ReadFile(filePath) //nolint:gosec // Path is controlled, reading user-specified files
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("File %s does not exist\n", fileName)
				continue
			}
			return fmt.Errorf("failed to read %s: %w", fileName, err)
		}

		content := string(data)
		if resolve {
			content = resolveImports(content, target)
		}

		if len(files) > 1 {
			fmt.Printf("=== %s ===\n\n", fileName)
		}
		fmt.Print(content)
		if len(files) > 1 {
			fmt.Println()
		}
	}

	return nil
}

func showJSON(target string, files []string, resolve bool) error {
	result := make(map[string]interface{})
	filesData := make(map[string]interface{})

	for _, fileName := range files {
		filePath := filepath.Join(target, fileName)
		data, err := os.ReadFile(filePath) //nolint:gosec // Path is controlled, reading user-specified files
		if err != nil {
			if os.IsNotExist(err) {
				filesData[fileName] = map[string]interface{}{
					"error": "file does not exist",
				}
				continue
			}
			return fmt.Errorf("failed to read %s: %w", fileName, err)
		}

		content := string(data)
		imports := extractImports(content)
		lineCount := len(strings.Split(content, "\n"))

		if resolve {
			content = resolveImports(content, target)
		}

		filesData[fileName] = map[string]interface{}{
			"content":   content,
			"lineCount": lineCount,
			"imports":   imports,
		}
	}

	result["files"] = filesData
	return output.WriteJSON(result)
}

func resolveImports(content, baseDir string) string {
	lines := strings.Split(content, "\n")
	var resolved []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "@") {
			importPath := strings.TrimSpace(strings.TrimPrefix(line, "@"))
			if strings.HasSuffix(importPath, ".md") {
				importFile := filepath.Join(baseDir, importPath)
				importData, err := os.ReadFile(importFile) //nolint:gosec // Path is controlled, reading template imports
				if err == nil {
					importContent := resolveImports(string(importData), baseDir)
					resolved = append(resolved, importContent)
					continue
				}
			}
		}
		resolved = append(resolved, line)
	}

	return strings.Join(resolved, "\n")
}

func extractImports(content string) []string {
	var imports []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@") {
			importPath := strings.TrimSpace(strings.TrimPrefix(trimmed, "@"))
			if strings.HasSuffix(importPath, ".md") {
				imports = append(imports, importPath)
			}
		}
	}

	return imports
}
