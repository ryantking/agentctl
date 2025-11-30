#!/usr/bin/env -S uv --script
"""Generate Homebrew formula from requirements.txt with hashes."""

import re
import sys
from pathlib import Path


def parse_requirements_with_hashes(req_file: Path) -> dict[str, dict]:
    """Parse requirements.txt and extract package info with versions and first hash."""
    packages = {}
    lines = req_file.read_text().splitlines()
    i = 0

    while i < len(lines):
        line = lines[i].rstrip()
        i += 1

        # Skip comments, empty lines, and editable installs
        if not line or line.startswith("#") or line.startswith("-e"):
            continue

        # Parse package==version line
        match = re.match(r"^([a-zA-Z0-9_\-\.]+)==([a-zA-Z0-9\._\-]+)", line)
        if not match:
            continue

        pkg_name = match.group(1)
        version = match.group(2)
        hashes = []

        # Collect all hashes for this package (which may span multiple lines with \)
        while line.endswith("\\") or (i < len(lines) and lines[i].strip().startswith("--hash=")):
            if line.endswith("\\"):
                line = lines[i].rstrip()
                i += 1
            else:
                line = lines[i].rstrip()
                i += 1

            # Extract hash from line
            hash_match = re.search(r"--hash=sha256:([a-f0-9]+)", line)
            if hash_match:
                hashes.append(hash_match.group(1))

        packages[pkg_name] = {"version": version, "hashes": hashes}

    return packages


def generate_formula(
    version: str,
    packages: dict[str, dict],
    url: str = None,
    sha256: str = None,
) -> str:
    """Generate the Homebrew formula Ruby code."""

    formula = f"""class Claudectl < Formula
  include Language::Python::Virtualenv

  desc "CLI tool for managing Claude Code configurations and workspaces"
  homepage "https://github.com/ryantking/claudectl"
  url "{url or 'https://github.com/carelesslisper/claudectl/releases/download/v' + version + '/claudectl-' + version + '.tar.gz'}"
  sha256 "{sha256 or 'PLACEHOLDER_UPDATE_AFTER_FIRST_RELEASE'}"
  license "MIT"

  depends_on "python@3.13"
"""

    # Add resources for each dependency (skip the main package itself and pywin32)
    for pkg_name in sorted(packages.keys()):
        # Skip the main package and Windows-only packages
        if pkg_name.lower() in ["claudectl", "pywin32"]:
            continue

        pkg_info = packages[pkg_name]
        version_str = pkg_info["version"]
        hashes = pkg_info["hashes"]

        if not hashes:
            continue

        sha = hashes[0]  # Use first hash

        # Standard PyPI URL pattern (works for most packages)
        # Package names in URLs use underscores instead of hyphens
        url_name = pkg_name.replace("-", "_")
        formula += f'''
  resource "{pkg_name}" do
    url "https://files.pythonhosted.org/packages/{url_name}-{version_str}.tar.gz"
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
        print(f"Error: {req_file} not found", file=sys.stderr)
        sys.exit(1)

    version = sys.argv[2] if len(sys.argv) > 2 else "0.1.0"
    url = sys.argv[3] if len(sys.argv) > 3 else None
    sha256 = sys.argv[4] if len(sys.argv) > 4 else None

    packages = parse_requirements_with_hashes(req_file)
    formula = generate_formula(version, packages, url, sha256)
    print(formula)
