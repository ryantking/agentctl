#!/usr/bin/env python3
"""Generate Homebrew formula from requirements.txt"""

import re
import sys
from pathlib import Path


def parse_requirements(req_file: Path) -> dict[str, tuple[str, str]]:
    """Parse requirements.txt and extract package names, versions, and hashes."""
    packages = {}
    current_package = None
    current_hashes = []

    with open(req_file) as f:
        for line in f:
            line = line.rstrip()

            # Skip comments and empty lines
            if not line or line.startswith("#"):
                continue

            # Handle line continuations
            if line.endswith("\\"):
                line = line[:-2].strip()

            # Skip the -e . line (editable install of current package)
            if line == "-e .":
                continue

            # Check if this is a package line (starts with a letter, digit, or underscore)
            if re.match(r"^[a-zA-Z0-9_-]", line):
                # Save previous package if any
                if current_package:
                    packages[current_package] = tuple(current_hashes)
                    current_hashes = []

                # Extract package name and version
                match = re.match(r"^([a-zA-Z0-9_-]+)==([^ \\]+)", line)
                if match:
                    current_package = match.group(1)
                    version = match.group(2)
                    # Download URL for PyPI
                    url = f"https://files.pythonhosted.org/packages/{current_package.replace('-', '_')}-{version}.tar.gz"
            # Check if this is a hash line
            elif "--hash=" in line:
                match = re.search(r"--hash=sha256:([a-f0-9]+)", line)
                if match:
                    current_hashes.append(match.group(1))

        # Don't forget the last package
        if current_package:
            packages[current_package] = tuple(current_hashes)

    return packages


def generate_formula(
    version: str, packages: dict[str, tuple[str, str]], url: str = None, sha256: str = None
) -> str:
    """Generate the Homebrew formula Ruby code."""

    formula = f"""class Claudectl < Formula
  include Language::Python::Virtualenv

  desc "CLI tool for managing Claude Code configurations and workspaces"
  homepage "https://github.com/carelesslisper/claudectl"
  url "{url or 'https://github.com/carelesslisper/claudectl/releases/download/v' + version + '/claudectl-' + version + '.tar.gz'}"
  sha256 "{sha256 or 'PLACEHOLDER_UPDATE_AFTER_FIRST_RELEASE'}"
  license "MIT"

  depends_on "python@3.13"
"""

    # Add resources for each dependency
    for pkg_name, hashes in sorted(packages.items()):
        # Convert package name to snake_case for resource names
        resource_name = pkg_name.lower().replace("-", "_")

        # Skip the main package itself
        if resource_name == "claudectl":
            continue

        if hashes:
            sha = hashes[0]  # Use first hash
        else:
            sha = "PLACEHOLDER"

        formula += f'''
  resource "{pkg_name}" do
    url "https://files.pythonhosted.org/packages/{pkg_name.replace('-', '_')}-{version}.tar.gz"
    sha256 "{sha}"
  end
'''

    formula += """
  def install
    virtualenv_install_with_resources
  end

  test do
    system "claudectl", "--version"
  end
end
"""

    return formula


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: generate_formula.py <requirements.txt> [version] [url] [sha256]")
        sys.exit(1)

    req_file = Path(sys.argv[1])
    if not req_file.exists():
        print(f"Error: {req_file} not found")
        sys.exit(1)

    version = sys.argv[2] if len(sys.argv) > 2 else "0.1.0"
    url = sys.argv[3] if len(sys.argv) > 3 else None
    sha256 = sys.argv[4] if len(sys.argv) > 4 else None

    packages = parse_requirements(req_file)
    formula = generate_formula(version, packages, url, sha256)
    print(formula)
