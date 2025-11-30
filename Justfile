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

# Bump patch version (0.1.0 -> 0.1.1)
bump-patch:
    uv version --bump patch

# Bump minor version (0.1.0 -> 0.2.0)
bump-minor:
    uv version --bump minor

# Bump major version (0.1.0 -> 1.0.0)
bump-major:
    uv version --bump major

# Generate Homebrew formula from requirements.txt
formula: build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Exporting dependencies to requirements.txt..."
    uv export --format requirements.txt --no-dev > /tmp/requirements.txt
    echo "Generating Homebrew formula..."
    VERSION=$(uv version | awk '{print $NF}')
    python3 scripts/generate_formula.py /tmp/requirements.txt "$VERSION" > Formula/claudectl.rb
    echo "✓ Formula generated at Formula/claudectl.rb"

# Verify Homebrew formula matches current dependencies
verify-formula: build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Exporting dependencies to requirements.txt..."
    uv export --format requirements.txt --no-dev > /tmp/requirements.txt
    echo "Generating formula from requirements..."
    VERSION=$(uv version | awk '{print $NF}')
    python3 scripts/generate_formula.py /tmp/requirements.txt "$VERSION" > /tmp/generated-formula.rb
    echo "Comparing formulas..."
    if diff -u Formula/claudectl.rb /tmp/generated-formula.rb > /tmp/formula-diff.txt 2>&1; then
        echo "✓ Formula matches current dependencies"
    else
        echo "✗ Formula does not match current dependencies:"
        cat /tmp/formula-diff.txt
        exit 1
    fi

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
