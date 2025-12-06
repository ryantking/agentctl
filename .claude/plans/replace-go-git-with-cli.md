# Plan: Replace go-git with Git CLI

## Problem

go-git has known limitations with git worktrees. When opening a repository at a worktree path, it fails to correctly read the worktree-specific HEAD and index files, causing `Status()` to report incorrect changes (showing all files as "Added" when comparing against the wrong commit).

## Solution

Replace all go-git library usage with direct git CLI commands. The git CLI correctly handles worktrees and is already installed on all target systems.

## Files to Modify

### 1. `internal/git/exec.go` (NEW FILE)
Create a helper module for executing git commands safely:
- `RunGit(repoPath string, args ...string) (string, error)` - runs git with `-C` flag
- `RunGitLines(repoPath string, args ...string) ([]string, error)` - returns output as lines
- Handles stderr, exit codes, and error wrapping

### 2. `internal/git/repo.go`
**Remove**: `OpenRepo`, `OpenRepoWithDiscover`, `Repo` struct wrapping go-git
**Keep**: `discoverRepoRoot` (uses file system, not go-git)
**Add**: `GetRepoRootFromPath(path string) (string, error)` using `git rev-parse --show-toplevel`

### 3. `internal/git/status.go`
**Replace**: `IsWorktreeClean` implementation
- Use `git -C <path> status --porcelain` instead of go-git
- Parse porcelain output to count staged/modified/untracked

**Replace**: `GetStatusSummary` to use the new implementation

### 4. `internal/git/git.go`
**Replace**: `GetCurrentBranch`
- Use `git -C <path> rev-parse --abbrev-ref HEAD`
- Remove go-git fallback code

**Replace**: `BranchExists`
- Use `git -C <path> show-ref --verify refs/heads/<branch>` for local
- Use `git -C <path> ls-remote --heads origin <branch>` for remote

### 5. `internal/git/worktree.go`
**Replace**: `AddWorktree`
- Use `git -C <repoRoot> worktree add <path> <branch>` for existing branch
- Use `git -C <repoRoot> worktree add -b <branch> <path> <base>` for new branch

**Replace**: `RemoveWorktree`
- Use `git -C <repoRoot> worktree remove <path>` (with `--force` if needed)

**Keep**: `ListWorktrees`, `parseWorktreeDir`, `getMainWorktree` - these already parse files directly

**Remove**: `generateWorktreeID` (not needed with CLI)

### 6. `internal/hook/autocommit.go`
**Replace**: `gitAddAndCommit` and `gitAddAndCommitNewFile`
- Use `git -C <path> add <file>` to stage
- Use `git -C <path> diff --cached --quiet <file>` to check if staged
- Use `git -C <path> commit -m "<message>"` to commit

**Remove**: go-git import

### 7. `internal/hook/context.go`
**Replace**: `getAllGitBranches`
- Use `git -C <path> for-each-ref --format='%(refname:short)' refs/heads/`

**Remove**: `plumbing` import

### 8. `internal/workspace/manager.go`
**Replace**: `calculateAheadBehind` and `getCommitList`
- Use `git -C <path> rev-list --count <local>..<remote>` for behind
- Use `git -C <path> rev-list --count <remote>..<local>` for ahead

**Replace**: `GetWorkspaceDiff`
- Use `git -C <path> diff <target>..<head>`

**Remove**: go-git imports

### 9. `internal/github/client.go`
**Replace**: `getRepoInfo`
- Use `git -C <path> remote get-url origin` to get remote URL

**Remove**: go-git import

### 10. `go.mod`
**Remove**: `github.com/go-git/go-git/v5` dependency

## Git CLI Command Reference

| Operation | Command |
|-----------|---------|
| Get repo root | `git rev-parse --show-toplevel` |
| Get current branch | `git rev-parse --abbrev-ref HEAD` |
| Check branch exists | `git show-ref --verify refs/heads/<branch>` |
| Check remote branch | `git ls-remote --heads origin <branch>` |
| List branches | `git for-each-ref --format='%(refname:short)' refs/heads/` |
| Get status (porcelain) | `git status --porcelain` |
| Stage file | `git add <file>` |
| Check staged changes | `git diff --cached --quiet` |
| Commit | `git commit -m "<message>"` |
| Add worktree | `git worktree add [-b <branch>] <path> [<base>]` |
| Remove worktree | `git worktree remove [--force] <path>` |
| List worktrees | `git worktree list --porcelain` |
| Get remote URL | `git remote get-url origin` |
| Ahead/behind count | `git rev-list --count <a>..<b>` |
| Diff branches | `git diff <target>..<head>` |
| Get HEAD commit | `git rev-parse HEAD` |
| Get short commit | `git rev-parse --short HEAD` |

## Implementation Order

1. Create `internal/git/exec.go` with git command helpers
2. Update `internal/git/status.go` (fixes the original bug)
3. Update `internal/git/git.go`
4. Update `internal/git/worktree.go`
5. Update `internal/git/repo.go` (remove Repo type)
6. Update `internal/hook/autocommit.go`
7. Update `internal/hook/context.go`
8. Update `internal/workspace/manager.go`
9. Update `internal/github/client.go`
10. Remove go-git from `go.mod` and run `go mod tidy`
11. Run tests and verify workspace list shows correct status

## Testing

After implementation, verify:
```bash
# Create a clean worktree
agentctl workspace create test/cli-migration

# Check status shows "clean"
agentctl workspace list

# Make changes in worktree, verify status updates correctly
agentctl workspace status test/cli-migration

# Cleanup
agentctl workspace delete test/cli-migration
```
