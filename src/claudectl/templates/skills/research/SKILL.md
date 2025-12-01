---
name: research
description: Comprehensive research workflow for feature implementation planning. Use this skill when gathering information about how to implement features, understanding technology choices, researching APIs and libraries, or exploring implementation approaches. Combines web search for general research with Context7 for API documentation, and includes structured questioning to clarify requirements.
---

# Research

## When to Use

1. **Planning features** - "How do I implement..." or "help me plan..."
2. **Technology research** - Compare libraries, frameworks, approaches
3. **API/library docs** - Specific API references, versions, usage
4. **Implementation patterns** - Best practices, design patterns
5. **Requirement clarification** - Ambiguous or incomplete requirements
6. **Architecture decisions** - Evaluate different approaches

## Research Decision Tree

```
User requests feature implementation
│
├─ Requirements clear? NO → Clarifying Requirements Workflow
├─ Specific library/API? YES → API Documentation Workflow
├─ General "how to"? YES → General Implementation Workflow
└─ Compare options? YES → Comparative Research Workflow
```

## Four Core Workflows

### 1. Clarifying Requirements

**When**: Requirements ambiguous, incomplete, or multiple interpretations

**Process**:
1. Identify ambiguities (scope, tech choices, performance, security)
2. Structure 1-4 most important questions using AskUserQuestion
3. Present options with trade-offs in descriptions
4. Proceed with research after clarification

**Question patterns**:
- **Scope**: Scale? Performance requirements? Timeline?
- **Technical**: Integrate with existing systems? Tech preferences? Complexity acceptable?
- **Feature**: Core vs optional features? Error handling? Edge cases?

**Example**:
```
AskUserQuestion:
  Question: "Which authentication approach for this API?"
  Header: "Auth method"
  Options:
    - "JWT tokens" - Stateless, scalable, requires refresh logic
    - "Session cookies" - Traditional, stateful, simpler
    - "OAuth2" - Industry standard, third-party auth, complex
```

### 2. API Documentation Research

**When**: Need specific API docs, library usage, version info

**Process**:
```bash
# 1. Resolve library ID
mcp__context7__resolve-library-id: "library-name"

# 2. Get documentation
mcp__context7__get-library-docs: "/org/project"
  topic: "specific-feature"
  mode: "code"  # or "info" for concepts
  page: 1       # Try 2-4 if insufficient

# 3. Review multiple pages if needed
# 4. Document findings with code examples
```

**Mode selection**:
- `mode: "code"` - API references, code examples, function signatures
- `mode: "info"` - Architectural concepts, guides, explanations

### 3. General Implementation Research

**When**: Understand how to implement, find best practices, compare approaches

**Process**:
1. **Formulate 2-4 search queries**:
   ```
   "how to implement <feature> in <technology> <current-year>"
   "<feature> best practices <current-year>"
   "<technology> <feature> patterns examples"
   ```

2. **Execute WebSearch** (in parallel when possible)

3. **Analyze results** - Extract: approaches, code examples, trade-offs, patterns

4. **Synthesize findings**:
   - Available approaches (2-4 main options)
   - Recommended approach (based on requirements)
   - Key considerations (gotchas, requirements)
   - Code examples (concrete examples)
   - Next steps (decisions needed)

5. **Present with sources** (CRITICAL):
   ```markdown
   Sources:
   - [Source Title 1](url1)
   - [Source Title 2](url2)
   ```

### 4. Comparative Research

**When**: Compare libraries, frameworks, or implementation approaches

**Process**:
1. **Identify comparison criteria**: Functionality, performance, DX, ecosystem, maturity, maintenance, compatibility, learning curve

2. **Research each option**:
   - Context7 for library-specific info
   - WebSearch for comparisons: `"<tech-a> vs <tech-b> comparison 2025"`

3. **Create comparison matrix**:
   ```markdown
   | Criteria | Option A | Option B | Option C |
   |----------|----------|----------|----------|
   | Functionality | ... | ... | ... |
   | Performance | ... | ... | ... |
   | Ecosystem | ... | ... | ... |
   ```

4. **Make recommendation**: Which fits best? Why? Trade-offs? Migration path?

5. **Present with sources**

## Tool Selection Quick Reference

| Need | Primary Tool | Secondary | Ask Questions? |
|------|--------------|-----------|----------------|
| Unclear requirements | AskUserQuestion | - | Yes |
| API docs | Context7 | WebSearch | If needed |
| How to implement | WebSearch | Context7 | If unclear |
| Compare options | WebSearch | Context7 | If criteria unclear |
| Library versions | Context7 | WebSearch | No |

## Common Query Patterns

### By Technology

**Kubernetes**:
```
"kubernetes <feature> implementation 2025"
"kubernetes operator <use-case> example"
"kubernetes client-go <feature> example"
```

**Python/FastAPI**:
```
"fastapi <feature> implementation 2025"
"fastapi <pattern> best practices"
"python <library> <feature> example"
```

**Go**:
```
"golang <feature> implementation"
"go <pattern> best practices 2025"
"golang <library> example"
```

**React**:
```
"react <feature> implementation 2025"
"react <pattern> best practices"
"react hooks <use-case>"
```

**Database**:
```
"postgresql <feature> implementation"
"database <pattern> best practices"
"<database> <use-case> optimization"
```

### By Use Case

**Authentication**:
```
"<tech> authentication implementation 2025"
"JWT vs session authentication comparison"
"<auth-method> security best practices"
```

**API Design**:
```
"REST API design best practices 2025"
"API versioning strategies comparison"
"API rate limiting implementation <tech>"
```

**Background Jobs**:
```
"task queue comparison <language> 2025"
"<queue> <language> production setup"
"<queue> retry strategy patterns"
```

**Caching**:
```
"caching strategies <use-case> 2025"
"redis vs memcached comparison"
"cache invalidation patterns"
```

**Database**:
```
"<use-case> database comparison 2025"
"<database> <language> best practices"
"<database> schema design <use-case>"
```

## Context7 Topics by Library Type

**Web Frameworks**: routing, middleware, authentication, validation, error handling
**Database Clients**: connection, queries, transactions, migrations, relationships
**Kubernetes Clients**: resources, watch, informers, create update delete, custom resources
**React/Frontend**: hooks, state management, routing, forms, data fetching
**Testing**: fixtures, mocking, assertions, async testing

## Best Practices

### Research Quality
1. **Use current year** - Always include year from `<env>` in searches
2. **Multiple sources** - Verify across 2-3 sources
3. **Prefer official docs** - Use Context7 for authoritative API docs
4. **Recent information** - Prioritize within 1-2 years
5. **Cite sources** - Always include "Sources" section with links

### Asking Questions
1. **Be specific** - Concrete details, not vague preferences
2. **Provide context** - Explain trade-offs in option descriptions
3. **Limit questions** - Ask 1-4 most important, can follow up
4. **Actionable options** - Clear, distinct choices

### Presenting Research
1. **Start with summary** - 2-3 sentence overview
2. **Organize clearly** - Headers, lists, tables
3. **Show code examples** - Concrete implementation when available
4. **Highlight trade-offs** - Not just features, explain implications
5. **Recommend path** - Clear recommendation when possible
6. **Always cite sources** - Users need to verify and dive deeper

## Common Patterns

**Unknown library version**:
1. Use Context7 (gets latest by default)
2. Present version found
3. Ask if specific version needed

**Multiple related questions**:
1. Execute WebSearch in parallel for different aspects
2. Use Context7 concurrently for library items
3. Synthesize into cohesive presentation

**Insufficient Context7 results**:
1. Try `page: 2`, `page: 3`, etc.
2. Try different `topic` keywords
3. Switch between `mode: "code"` and `mode: "info"`
4. Fall back to WebSearch

**Feature implementation plan**:
1. Clarify requirements (use questions if needed)
2. Research approaches (General Implementation)
3. Research specific APIs (API Documentation)
4. Synthesize into structured plan with sources

## Integration with Planning

When used during feature planning:

**Before Planning**:
- Clarify requirements
- Research approaches
- Gather API docs
- Make tech recommendations

**Plan Structure**:
```markdown
# Feature Implementation Plan

## Research Summary
[2-3 sentence overview]

## Recommended Approach
[What to build and why]

## Technology Choices
[Libraries, frameworks, patterns]

## Implementation Steps
1. [Step with details]
2. [Step with details]

## Key Considerations
[Important details, gotchas]

## Code Examples
[Relevant examples from research]

## Resources
- [Source 1](url1)
- [Source 2](url2)
```

## Common Pitfalls

**Too broad research**:
```
❌ "authentication implementation"
✅ "FastAPI JWT authentication with refresh tokens 2025"
```

**Ignoring current year**:
```
❌ "kubernetes best practices"
✅ "kubernetes best practices 2025"
```

**Not using Context7 for libraries**:
```
❌ WebSearch: "fastapi authentication example"
✅ Context7: "fastapi" topic: "security" mode: "code"
```

**Asking too many questions**:
```
❌ Ask 10 questions about every detail
✅ Ask 2-3 critical questions, note assumptions
```

**No sources**:
```
❌ "Based on research, use approach X"
✅ "Based on research, use approach X

Sources:
- [Article 1](url)"
```

## Research Checklist

Before presenting research:
- [ ] Requirements clarified (or assumptions stated)
- [ ] Multiple sources consulted (2-3 minimum)
- [ ] Recent information prioritized (check year)
- [ ] Code examples included (when available)
- [ ] Trade-offs explained (not just features)
- [ ] Recommendation provided (with rationale)
- [ ] Sources cited (markdown links)
- [ ] Next steps identified

## Quick Tool Selection

```
Need authoritative API docs? → Context7
Need implementation examples? → WebSearch + Context7
Need to compare technologies? → WebSearch (multiple queries)
Need to understand concepts? → Context7 (mode: "info")
Need to clarify requirements? → AskUserQuestion
Need recent information? → WebSearch with current year
Need version-specific docs? → Context7 (latest or specified)
```

## Troubleshooting

**Context7 library not found**:
- Try alternative names
- Try broader terms
- Fall back to WebSearch
- Ask user for exact library name

**WebSearch irrelevant results**:
- Add more specific terms (tech, version, year)
- Try different phrasing
- Add filters: `site:stackoverflow.com` or `site:github.com`
- Combine with Context7

**Insufficient user answers**:
- Ask 1-2 follow-up questions
- Make reasonable assumptions (state them explicitly)
- Provide multiple plan variations
- Proceed with most common approach, note alternatives

**Research takes too long**:
- Focus on specific question, not comprehensive coverage
- Start with Context7 for authoritative source
- Limit WebSearch to 2-3 targeted queries
- Present summary first, offer to research deeper
