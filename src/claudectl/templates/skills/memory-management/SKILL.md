---
name: memory-management
description: Guide for creating and maintaining repository-local CLAUDE.md files that capture project-specific context, patterns, and conventions. Use when initializing repositories for Claude Code, updating project memory, or understanding effective repository memory structure.
---

# Memory Management

Guide for creating and maintaining repository-local CLAUDE.md files that transform Claude Code from a general-purpose assistant into a project-specialized expert.

## Overview

Repository memory serves as persistent context that:

- Captures project-specific architecture, patterns, and conventions
- Eliminates repeated explanations across sessions
- Provides immediate context to new Claude instances
- Documents non-obvious decisions and rationales
- Guides consistent implementation across the codebase

The global `~/.claude/CLAUDE.md` configures Claude's behavior. Repository-local `.claude/CLAUDE.md` files specialize Claude for specific projects.

## When to Create/Update

**Create** when:
- Initializing a repository for Claude Code usage
- Onboarding to a new project (capture learnings)
- Project has non-standard patterns or conventions
- Repeated context needed across sessions
- Team wants consistent Claude interactions

**Update** when:
- Architecture changes (new services, refactored modules)
- Patterns emerge or evolve (new conventions adopted)
- Deployment process changes
- Testing strategy shifts
- New team conventions established
- Previous guidance becomes stale or incorrect

## CLAUDE.md Structure

Target **under 100 lines** for quick AI consumption. Use XML tags for clear section boundaries:

```xml
<architecture>
High-level system structure, key modules, data flow
</architecture>

<patterns>
Code patterns, idioms, conventions specific to this project
</patterns>

<conventions>
Naming, formatting, organizational rules
</conventions>

<testing>
Test structure, commands, coverage expectations
</testing>

<deployment>
Build process, deployment steps, environment setup
</deployment>
```

Tags enable efficient parsing and selective context loading.

## Recommended Sections

**architecture**: System structure, key modules, service boundaries, data flow patterns. Focus on relationships and responsibilities.

**patterns**: Project-specific code patterns, idioms, common abstractions. Examples: error handling approach, state management pattern, API response structure.

**conventions**: Naming rules, file organization, import ordering, comment style. Enforceable through tooling (linters, formatters) but document exceptions.

**testing**: Test file structure, naming conventions, commands to run tests, coverage expectations. Include mock/fixture patterns if non-standard.

**deployment**: Build commands, deployment process, environment variables, CI/CD specifics. Focus on developer-facing operations.

**Optional sections**:
- `<dependencies>`: Key libraries and usage guidelines
- `<troubleshooting>`: Common issues and solutions
- `<development>`: Setup steps, required tools, local dev workflow

## Keep Concise

**Use bullet points**: Quick to scan, easy to update, AI-friendly
**Avoid obvious information**: Skip standard practices (e.g., "use descriptive names")
**Focus on project-specific**: Only include what differs from defaults
**Reference external docs**: Link to ADRs, wikis, or detailed guides instead of duplicating
**Update, don't accumulate**: Remove outdated guidance rather than leaving contradictions

Example of concise vs verbose:

**Verbose** (avoid):
```
<patterns>
When creating API endpoints, you should follow RESTful conventions.
Use GET for retrieval, POST for creation, PUT for updates, and DELETE
for deletion. Always return appropriate status codes like 200 for success,
404 for not found, and 500 for server errors.
</patterns>
```

**Concise** (prefer):
```
<patterns>
- API responses: { data, error, meta } structure
- Error handling: throw custom AppError subclasses
- Auth: JWT in Authorization header, validated in middleware
</patterns>
```

## Progressive Disclosure

Information hierarchy:

1. **CLAUDE.md** (100 lines): Immediate context, quick reference
2. **docs/ directory**: Detailed design docs, ADRs, API specs
3. **repo-local skills**: Complex workflows, specialized tools

CLAUDE.md provides the map. Skills and docs provide the details.

Example reference pattern:

```xml
<architecture>
- Microservices: auth, api, worker, notifications
- Event-driven: RabbitMQ for inter-service communication
- See docs/architecture/services.md for detailed flow diagrams
</architecture>
```

Link to deeper documentation but keep core patterns inline.

## Repository-Local Skills

Create `.claude/skills/` when workflows involve:

**Create skills when**:
- 5+ operations in sequence (e.g., release process)
- Complex domain logic (e.g., data migration patterns)
- Tech-stack-specific patterns (e.g., Kubernetes deployment)
- Tool integration (e.g., custom CLI workflows)
- Multi-step validation (e.g., pre-deployment checks)

**Do NOT create skills for**:
- 1-2 simple commands (document in CLAUDE.md)
- Standard tech patterns (global skills handle these)
- Obvious workflows (e.g., "run tests")
- Over-specification (trust Claude's judgment)

Example repository skill structure:

```
.claude/
├── CLAUDE.md                    # Core context (<100 lines)
└── skills/
    ├── release/SKILL.md         # Multi-step release workflow
    └── migration/SKILL.md       # Database migration patterns
```

## Integration Pattern

Repository-local CLAUDE.md supplements (not replaces) global configuration:

**Global** (`~/.claude/CLAUDE.md`):
- Claude's operating model (agent orchestration, git workflow)
- Cross-project conventions (commit format, PR templates)
- Universal patterns (testing, documentation)

**Repository-local** (`.claude/CLAUDE.md`):
- Project architecture and structure
- Codebase-specific patterns
- Tech stack conventions
- Deployment specifics

Claude merges both contexts. Repository-local wins for conflicts.

## Maintenance

**Review cadence**:
- After major refactors or architecture changes
- Quarterly for active projects
- When onboarding surfaces outdated info

**Monitor for staleness**:
- Commands that fail (outdated CLI usage)
- Mismatched patterns (code evolved, docs didn't)
- Contradictions (old guidance conflicts with new)

**Keep lightweight**: Remove rather than archive. Git history preserves old decisions.

## Example

Realistic ~80-line CLAUDE.md for a Next.js SaaS application:

```markdown
# MyApp - SaaS Platform

<architecture>
- Next.js 14 (App Router) monorepo
- `/app`: Frontend routes and components
- `/lib`: Shared utilities, API clients, db access
- `/workers`: Background job processors (separate service)
- Postgres (Prisma ORM) + Redis (caching, sessions)
- Deployed on Vercel (frontend) + Railway (workers)
</architecture>

<patterns>
- Server actions: use "use server" for mutations, return { data, error }
- API routes: minimal, only for webhooks or third-party integrations
- Database: single Prisma client instance, exported from lib/db
- Auth: NextAuth v5, session stored in Redis, JWT for API tokens
- Error handling: throw AppError subclasses, caught in error.tsx boundaries
- State: React Query for server state, Zustand for client-only UI state
</patterns>

<conventions>
- Components: PascalCase files in feature directories (e.g., dashboard/UserTable.tsx)
- Utilities: camelCase in lib/ (e.g., lib/formatCurrency.ts)
- Server actions: suffix with "Action" (e.g., updateProfileAction)
- API routes: /app/api/webhooks/* only (prefer server actions)
- Database: snake_case for Postgres, camelCase in Prisma schema
- Env vars: NEXT_PUBLIC_* for client, plain for server
</conventions>

<testing>
- Unit tests: Vitest, colocated as *.test.ts
- Integration: Playwright for critical flows (auth, checkout)
- E2E: tests/e2e/, run via npm run test:e2e
- Coverage: aim for >80% on lib/, >60% on components
- Run: npm test (unit), npm run test:e2e (playwright)
</testing>

<deployment>
- Frontend: auto-deploy from main via Vercel (preview on PRs)
- Workers: Railway deploy via Dockerfile, manual trigger
- Migrations: npm run db:migrate before deploying workers
- Env sync: Vercel CLI for frontend, Railway dashboard for workers
- Staging: branch "staging" deploys to staging.myapp.com
</deployment>

<dependencies>
- shadcn/ui: component library, add via npm run add-component
- Prisma: ORM, schema in prisma/schema.prisma
- React Query: server state, configured in lib/query-client.ts
- NextAuth: authentication, config in lib/auth.ts
- Stripe: payments, webhook handling in app/api/webhooks/stripe/
</dependencies>

<troubleshooting>
- Hydration errors: check for window/localStorage in server components
- Prisma errors: ensure migrations run, check DATABASE_URL env var
- Auth loops: clear cookies, verify NEXTAUTH_SECRET is set
- Build fails: check for TypeScript errors in server actions
</troubleshooting>
```

This example demonstrates:
- Concise bullet points throughout
- Project-specific information only
- Clear XML section boundaries
- Focus on non-obvious patterns
- Practical command references
- Under 100 lines total

Adapt sections to your project's needs. Remove sections that don't apply.
