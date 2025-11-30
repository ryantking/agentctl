# Default recipe - show available commands
default:
    @just --list

# Install system dependencies from Brewfile
deps:
    brew bundle --no-upgrade --file=import/Brewfile

# Install claudectl globally in editable mode
install: deps
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
    uv run pytest tests

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

# Bump patch version (0.1.0 -> 0.1.1)
bump-patch:
    uv version --bump patch

# Bump minor version (0.1.0 -> 0.2.0)
bump-minor:
    uv version --bump minor

# Bump major version (0.1.0 -> 1.0.0)
bump-major:
    uv version --bump major

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
    echo "After release is published, update Formula/claudectl.rb with new version and SHA256"
