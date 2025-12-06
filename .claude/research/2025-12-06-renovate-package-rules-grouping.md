# Research: Renovate packageRules and Grouping Strategies
Date: 2025-12-06
Focus: How to configure Renovate to group dependency updates by ecosystem, manager, and update type
Agent: researcher

## Summary

Renovate's `packageRules` configuration provides powerful grouping capabilities through `groupName`, `matchManagers`, `matchUpdateTypes`, and `matchDatasources`. Dependencies can be grouped by package manager using `{{manager}}` template variables, by update type (major/minor/patch), or by specific package patterns. For Go-based projects with GitHub Actions, the recommended approach is to use separate rules targeting `gomod` and `github-actions` managers with appropriate grouping strategies.

## Key Findings

- **Grouping by manager**: Use `matchManagers` with `{{manager}}` template variable to auto-group by package manager ([GitHub Discussion #16419](https://github.com/renovatebot/renovate/discussions/16419))
- **Update type separation**: `separateMajorMinor: true` and `separateMinorPatch: true` control PR splitting by severity ([Renovate Docs](https://docs.renovatebot.com/configuration-options/))
- **Built-in presets**: `group:allNonMajor` groups all minor/patch updates; `group:recommended` aggregates common groupings ([Group Presets](https://docs.renovatebot.com/presets-group/))
- **Monorepo support**: Use `extends: ["monorepo:<name>"]` to group packages from the same monorepo ([Group Presets](https://docs.renovatebot.com/presets-group/))
- **Manager-specific files**: `gomod` matches `go.mod`, `github-actions` matches `.github/workflows/*.yml` ([Manager Docs](https://docs.renovatebot.com/modules/manager/))

## Detailed Analysis

### Core Configuration Options

#### groupName
Assigns a common name to group multiple dependency updates into a single PR:
```json
{
  "packageRules": [
    {
      "matchPackageNames": ["/eslint/"],
      "groupName": "eslint"
    }
  ]
}
```

#### groupSlug
URL-safe identifier for branch naming. Auto-generated from `groupName` if not specified:
```json
{
  "packageRules": [
    {
      "groupName": "devDependencies (non-major)",
      "groupSlug": "dev-dependencies"
    }
  ]
}
```

#### matchManagers
Filters rules to specific package managers:
```json
{
  "packageRules": [
    {
      "matchManagers": ["gomod"],
      "groupName": "Go modules"
    },
    {
      "matchManagers": ["github-actions"],
      "groupName": "GitHub Actions"
    }
  ]
}
```

#### matchUpdateTypes
Filters by update severity:
- `major` - Breaking version changes
- `minor` - Feature additions
- `patch` - Bug fixes
- `digest` - Content hash changes (Docker, etc.)
- `pin` - Version pinning

```json
{
  "packageRules": [
    {
      "matchUpdateTypes": ["minor", "patch"],
      "automerge": true
    }
  ]
}
```

#### matchDatasources
Filters by package registry type:
```json
{
  "packageRules": [
    {
      "matchDatasources": ["go"],
      "labels": ["go-dependency"]
    }
  ]
}
```

### Grouping Strategies

#### 1. Group by Package Manager
Uses template variable `{{manager}}` to create separate PRs per ecosystem:
```json
{
  "packageRules": [
    {
      "matchPackagePatterns": ["*"],
      "groupName": "{{manager}} dependencies",
      "groupSlug": "{{manager}}"
    }
  ]
}
```

**Or explicitly list managers:**
```json
{
  "packageRules": [
    {
      "matchManagers": ["gomod", "github-actions", "npm", "dockerfile"],
      "groupName": "{{manager}}"
    }
  ]
}
```

#### 2. Group by Update Type
Separate major updates from non-major:
```json
{
  "separateMajorMinor": true,
  "packageRules": [
    {
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "all non-major dependencies",
      "groupSlug": "all-minor-patch"
    }
  ]
}
```

#### 3. Combine Manager + Update Type
Best practice: group by manager AND exclude major updates:
```json
{
  "packageRules": [
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "Go non-major dependencies",
      "groupSlug": "go-minor-patch"
    },
    {
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "GitHub Actions non-major",
      "groupSlug": "gha-minor-patch"
    }
  ]
}
```

#### 4. Monorepo Grouping
Group packages from the same monorepo:
```json
{
  "packageRules": [
    {
      "extends": ["monorepo:storybook"],
      "groupName": "storybook monorepo",
      "matchUpdateTypes": ["digest", "patch", "minor", "major"]
    }
  ]
}
```

### Available Managers (Relevant)

| Manager | File Pattern | Datasources |
|---------|-------------|-------------|
| `gomod` | `go.mod` | `go`, `golang-version` |
| `github-actions` | `.github/workflows/*.yml`, `action.yml` | `github-tags`, `github-runners` |
| `npm` | `package.json` | `npm` |
| `pip_requirements` | `requirements*.txt` | `pypi` |
| `dockerfile` | `Dockerfile*` | `docker` |

### Recommended Configuration for Go + GitHub Actions

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended"
  ],
  "separateMajorMinor": true,
  "packageRules": [
    {
      "description": "Group Go module non-major updates",
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "Go dependencies (non-major)",
      "groupSlug": "go-non-major"
    },
    {
      "description": "Group GitHub Actions non-major updates",
      "matchManagers": ["github-actions"],
      "matchUpdateTypes": ["minor", "patch", "digest"],
      "groupName": "GitHub Actions (non-major)",
      "groupSlug": "gha-non-major"
    },
    {
      "description": "Group GitHub Artifact Actions major updates",
      "matchManagers": ["github-actions"],
      "matchPackageNames": [
        "actions/download-artifact",
        "actions/upload-artifact"
      ],
      "matchUpdateTypes": ["major"],
      "groupName": "GitHub Artifact Actions"
    },
    {
      "description": "Label major updates for review",
      "matchUpdateTypes": ["major"],
      "labels": ["major-update", "needs-review"]
    }
  ],
  "postUpdateOptions": ["gomodTidy"]
}
```

### Best Practices

1. **Use `separateMajorMinor: true`** - Major updates require more review; keep them separate
2. **Group non-major by ecosystem** - Reduces PR noise while maintaining visibility
3. **Use descriptive `description` fields** - Documents intent for future maintainers
4. **Label major updates** - Makes them easy to filter in PR lists
5. **Consider automerge for patches** - Use `matchUpdateTypes: ["patch"]` with `automerge: true`
6. **Use `postUpdateOptions: ["gomodTidy"]`** - Keeps go.sum clean for Go projects
7. **Pin GitHub Actions digests** - Use `helpers:pinGitHubActionDigests` for security

## Applicable Patterns

For this codebase (Go project with GitHub Actions):

1. **Group Go module updates** by minor/patch to reduce PR count
2. **Group GitHub Actions updates** separately from Go modules
3. **Keep major updates separate** for careful review
4. **Use `gomodTidy`** post-update option
5. **Consider automerge** for low-risk patch updates

## Sources

- [Renovate Configuration Options](https://docs.renovatebot.com/configuration-options/)
- [Group Presets Documentation](https://docs.renovatebot.com/presets-group/)
- [GitHub Actions Manager](https://docs.renovatebot.com/modules/manager/github-actions/)
- [Go Modules Manager](https://docs.renovatebot.com/modules/manager/gomod/)
- [GitHub Discussion #16419 - Group by Manager](https://github.com/renovatebot/renovate/discussions/16419)
- [Managers List](https://docs.renovatebot.com/modules/manager/)
