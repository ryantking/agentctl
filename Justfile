# Default recipe - show available commands
default:
    @just --list

# Install system dependencies from Brewfile
deps:
    brew bundle --no-upgrade --file=hack/Brewfile
    gh extension install dlvhdr/gh-dash

# Install agentctl globally in editable mode
install:
    uv tool install --force --editable .

# Run linter checks
lint:
    uv run ruff check src
    uv run basedpyright src

# Format code
format:
    uv run ruff format src
    uv run ruff check --fix src

# Check formatting without making changes
format-check:
    uv run ruff format --check src

# Run tests
test:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -d tests ]; then
        uv run pytest tests
    else
        echo "No tests directory found, skipping tests"
    fi

# Run all checks (lint + test)
ci: lint test

# Build package (disables tool.uv.sources for distribution)
build:
    uv build --no-sources

# Clean build artifacts
clean:
    rm -rf dist/ build/ *.egg-info .pytest_cache .ruff_cache .coverage htmlcov

# Show current version
version:
    @uv version

# Create a release (bump version, commit, tag, push)
release bump:
    #!/usr/bin/env bash
    set -euo pipefail
    # Bump version
    uv version --bump {{bump}}
    NEW_VERSION=$(uv version)
    # Commit changes
    git add pyproject.toml uv.lock
    git commit -m "chore: bump version to ${NEW_VERSION}"
    # Create and push tag
    git tag "v${NEW_VERSION}"
    echo "✓ Created tag v${NEW_VERSION}"
    echo "Run 'git push && git push --tags' to trigger release workflow"
