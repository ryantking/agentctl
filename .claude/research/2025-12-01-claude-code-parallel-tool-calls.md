# Research: Claude Code Parallel Tool Calls and Bash Command Patterns
Date: 2025-12-01
Focus: Understanding parallel tool calls, bash command chaining, and orchestration patterns
Agent: researcher

## Summary

Claude Code supports two distinct forms of parallelism: (1) multiple tool calls within a single response that execute simultaneously, and (2) subagents that run in parallel with separate context windows. For bash commands specifically, independent commands should be made as separate parallel tool calls, while dependent commands should be chained with `&&` or `;` within a single call.

## Key Findings

### 1. Parallel Tool Calls Within Single Response
- Claude can include multiple `tool_use` blocks in one assistant message [Anthropic Tool Use Docs](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview)
- All corresponding `tool_result` blocks must be returned together in the subsequent user message
- This reduces conversation turns from 5+ to 1 for independent operations [Continue Blog](https://blog.continue.dev/parallel-tool-calling/)
- Claude 4.x models (especially Sonnet 4.5) are particularly aggressive in parallel tool execution

### 2. Bash Tool Specific Guidance
From the [Claude Code system prompt](https://gist.github.com/wong2/e0f34aac66caf890a332f7b6f9e2ba8f):
- Use `;` or `&&` to chain commands **within a single bash call**
- **DO NOT use newlines** to separate commands (newlines OK in quoted strings)
- For **independent operations**, make multiple Bash tool calls in a single message (parallel)
- For **dependent operations**, chain with `&&` in a single call (sequential)

### 3. When to Chain vs Split

**Chain commands (single Bash call with `&&`):**
- When output of one command feeds into another
- When commands must execute in order (mkdir before cp, git add before git commit)
- When working directory changes matter
- Example: `git add . && git commit -m "message" && git push`

**Split commands (multiple parallel Bash tool calls):**
- When commands are independent and can run simultaneously
- When gathering information from multiple sources
- When each command has its own timeout/error handling needs
- Example: Run `git status`, `git diff`, and `git log` as three separate parallel calls

### 4. Performance Benefits

| Approach | Latency | When to Use |
|----------|---------|-------------|
| Single chained call | 1 network round-trip | Dependent commands |
| Multiple parallel calls | 1 network round-trip (all execute simultaneously) | Independent commands |
| Sequential separate calls | N network round-trips | When results inform next call |

Performance data from [Continue Blog](https://blog.continue.dev/parallel-tool-calling/):
- Parallel execution reduces 30+ second codebase exploration to near-instant
- Eliminates multiple rounds of network latency and model inference
- Provides approximately n-times speedup for n parallel operations

### 5. Subagent Parallelism (Higher Level)

For larger-scale parallelism, Claude Code supports [subagents](https://code.claude.com/docs/en/sub-agents):
- Maximum 10 concurrent agents
- Each subagent has its own 200k context window
- Subagents cannot spawn other subagents (prevents infinite nesting)
- Explore subagent restricted to read-only bash: `ls, git status, git log, git diff, find, cat, head, tail`

### 6. Model Differences

- **Claude 4.x/4.5**: Excellent parallel tool use by default
- **Claude 3.7 Sonnet**: May be less likely to make parallel calls; workaround is "batch tool" meta-pattern
- **Opus 4.5**: Accuracy improves from 79.5% to 88.1% with Tool Search enabled [Anthropic Engineering](https://www.anthropic.com/engineering/advanced-tool-use)

## Detailed Analysis

### The Two Levels of Parallelism

**Level 1: Tool-Level Parallelism**
Within a single Claude response, multiple tool calls can execute simultaneously. This is controlled by:
- `disable_parallel_tool_use=true` in API settings to force sequential (one tool at a time)
- Default behavior allows Claude to intelligently batch independent operations

**Level 2: Agent-Level Parallelism**
Multiple Claude instances (subagents) running in parallel:
- Each with isolated context windows
- Orchestrated by a parent agent
- Results synthesized back to parent

### Best Practice Patterns

**Pattern 1: Gather Multiple Independent Pieces of Information**
```
Bad:  Execute git status, wait, execute git diff, wait, execute git log
Good: Send single message with 3 Bash tool calls (git status, git diff, git log)
```

**Pattern 2: Sequential Operations with Dependencies**
```
Bad:  Separate calls for mkdir, then cp (race condition if parallel)
Good: Single call: mkdir -p /path && cp file /path/
```

**Pattern 3: Git Workflows**
```
# Independent gathering (parallel)
Tool Call 1: git status
Tool Call 2: git diff HEAD
Tool Call 3: git log --oneline -5

# Dependent operations (chained)
Tool Call 4: git add . && git commit -m "message" && git push
```

### Token and Performance Considerations

From [Anthropic Bash Tool Docs](https://platform.claude.com/docs/en/agents-and-tools/tool-use/bash-tool):
- Each Bash tool call adds ~245 input tokens regardless of complexity
- Output truncated at ~100 lines (configurable)
- Default timeout: 120,000ms (2 minutes), max 600,000ms (10 minutes)

**Optimization strategies:**
- Batch independent operations to reduce round-trips
- Use CLAUDE.md to reduce context-gathering overhead
- Chain related commands to reduce token overhead from multiple tool definitions

### How Other Agents Handle This

From [Zach Wills' Subagent Guide](https://zachwills.net/how-to-use-claude-code-subagents-to-parallelize-development/):
- Specialist agents dispatched in parallel (PM, UX, Engineer)
- Results synthesized by orchestrator
- Separate output files per agent for audit trails

From [Continue IDE](https://blog.continue.dev/parallel-tool-calling/):
- Message structures store `toolCallStates[]` instead of single state
- Streaming logic handles multiple tool call deltas independently
- Individual approval/rejection buttons per pending operation

## Applicable Patterns

For the claudectl codebase and similar agent orchestration systems:

1. **Discovery Phase**: Use parallel tool calls for file exploration
   - Glob, Grep, Read calls can be batched
   - Git history queries (`git log`, `git blame`) can run parallel

2. **Implementation Phase**: Chain dependent operations
   - File creation sequences
   - Git commit workflows

3. **Subagent Spawning**: Use Task tool for true parallelism
   - Explore agents (1-3) for codebase understanding
   - Researcher agents (3-5) for external information
   - Cap at 10 concurrent to stay within limits

## Sources

- [Claude Code System Prompt (GitHub Gist)](https://gist.github.com/wong2/e0f34aac66caf890a332f7b6f9e2ba8f)
- [Anthropic Tool Use Overview](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview)
- [Anthropic Bash Tool Documentation](https://platform.claude.com/docs/en/agents-and-tools/tool-use/bash-tool)
- [Claude Code Best Practices (Anthropic Engineering)](https://www.anthropic.com/engineering/claude-code-best-practices)
- [Advanced Tool Use (Anthropic Engineering)](https://www.anthropic.com/engineering/advanced-tool-use)
- [Parallel Tool Calling: Making AI Agents Work Faster (Continue Blog)](https://blog.continue.dev/parallel-tool-calling/)
- [Claude Code Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [How to Use Claude Code Subagents (Zach Wills)](https://zachwills.net/how-to-use-claude-code-subagents-to-parallelize-development/)
- [Multi-Agent Orchestration (DEV Community)](https://dev.to/bredmond1019/multi-agent-orchestration-running-10-claude-instances-in-parallel-part-3-29da)
