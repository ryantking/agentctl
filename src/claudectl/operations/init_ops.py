"""Core initialization operations for claudectl init command."""

from __future__ import annotations

import json
import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path

from claudectl.cli.output import Result
from claudectl.domain.exceptions import ImportDirNotFoundError
from claudectl.operations.settings_merge import merge_settings_smart


@dataclass
class FileResult:
    """Result of a single file operation."""

    path: str
    status: str  # "created", "skipped", "overwritten", "merged"


class InitManager:
    """Manages Claude Code initialization."""

    def __init__(self, target_dir: Path) -> None:
        self.target = target_dir
        self.template_dir = self._find_template_dir()

    def _find_template_dir(self) -> Path:
        """Locate bundled templates directory."""
        # Check relative to package
        package_templates = Path(__file__).parent.parent / "templates"
        if package_templates.exists():
            return package_templates
        raise ImportDirNotFoundError()

    def install(
        self,
        force: bool = False,
        skip_index: bool = False,
        verbose: bool = False,
    ) -> Result:
        """Execute full initialization."""
        results: dict[str, list[str]] = {
            "installed": [],
            "skipped": [],
            "overwritten": [],
            "merged": [],
        }

        # 1. Install CLAUDE.md
        claude_md = self._install_file(
            self.template_dir / "CLAUDE.md",
            self.target / "CLAUDE.md",
            force,
        )
        self._track_result(results, claude_md)

        # 2. Install agents
        agents_results = self._install_directory(
            self.template_dir / "agents",
            self.target / ".claude" / "agents",
            force,
            pattern="*.md",
        )
        for r in agents_results:
            self._track_result(results, r)

        # 3. Install skills
        skills_results = self._install_directory(
            self.template_dir / "skills",
            self.target / ".claude" / "skills",
            force,
            recursive=True,
        )
        for r in skills_results:
            self._track_result(results, r)

        # 4. Merge settings
        settings_result = self._merge_settings(
            self.template_dir / "settings.json",
            self.target / ".claude" / "settings.json",
            force,
        )
        self._track_result(results, settings_result)

        # 5. Configure MCP servers
        mcp_result = self._configure_mcp(
            self.target / ".mcp.json",
            force,
        )
        self._track_result(results, mcp_result)

        # 6. Index repository with claude CLI
        if not skip_index:
            index_result = self._index_repository()
            if verbose and index_result:
                results["indexed"] = ["CLAUDE.md"]

        # Build result message
        message = self._build_message(results, verbose)

        return Result(
            success=True,
            message=message,
            data=results,
        )

    def _install_file(
        self,
        source: Path,
        dest: Path,
        force: bool,
    ) -> FileResult:
        """Install a single file."""
        status = "created"
        if dest.exists():
            if not force:
                return FileResult(str(dest.relative_to(self.target)), "skipped")
            status = "overwritten"

        dest.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, dest)

        return FileResult(str(dest.relative_to(self.target)), status)

    def _install_directory(
        self,
        source: Path,
        dest: Path,
        force: bool,
        pattern: str = "*",
        recursive: bool = False,
    ) -> list[FileResult]:
        """Install files from a directory."""
        results = []
        dest.mkdir(parents=True, exist_ok=True)

        if recursive:
            # Copy entire directory trees (for skills)
            for item in source.iterdir():
                if item.is_dir():
                    dest_item = dest / item.name
                    existed = dest_item.exists()

                    if existed and not force:
                        results.append(
                            FileResult(
                                str(dest_item.relative_to(self.target)),
                                "skipped",
                            )
                        )
                        continue

                    shutil.copytree(item, dest_item, dirs_exist_ok=force)
                    status = "overwritten" if existed else "created"
                    results.append(
                        FileResult(
                            str(dest_item.relative_to(self.target)),
                            status,
                        )
                    )
        else:
            # Copy matching files (for agents)
            for item in source.glob(pattern):
                if item.is_file():
                    result = self._install_file(item, dest / item.name, force)
                    results.append(result)

        return results

    def _merge_settings(
        self,
        source: Path,
        dest: Path,
        force: bool,
    ) -> FileResult:
        """Merge settings.json with smart deep merge."""
        dest.parent.mkdir(parents=True, exist_ok=True)

        # Load source settings
        with open(source) as f:
            new_settings = json.load(f)

        if not dest.exists():
            # No existing settings - just copy
            with open(dest, "w") as f:
                json.dump(new_settings, f, indent=2)
                f.write("\n")  # Add trailing newline
            return FileResult(
                str(dest.relative_to(self.target)),
                "created",
            )

        # Existing settings - merge
        with open(dest) as f:
            existing_settings = json.load(f)

        if force:
            # Force: overwrite
            with open(dest, "w") as f:
                json.dump(new_settings, f, indent=2)
                f.write("\n")  # Add trailing newline
            return FileResult(
                str(dest.relative_to(self.target)),
                "overwritten",
            )

        # Smart merge
        merged = merge_settings_smart(existing_settings, new_settings)
        with open(dest, "w") as f:
            json.dump(merged, f, indent=2)
            f.write("\n")  # Add trailing newline

        return FileResult(
            str(dest.relative_to(self.target)),
            "merged",
        )

    def _configure_mcp(
        self,
        dest: Path,
        force: bool,
    ) -> FileResult:
        """Configure MCP servers (.mcp.json) with Context7 and Linear."""
        # New MCP servers to add
        new_servers = {
            # Context7 automatically loads CONTEXT7_API_KEY from environment
            "context7": {"type": "http", "url": "https://mcp.context7.com/mcp"},
            # Linear with SSE transport
            "linear": {"type": "sse", "url": "https://mcp.linear.app/sse"},
        }

        dest.parent.mkdir(parents=True, exist_ok=True)

        # Load existing config and merge, or create new one
        if dest.exists() and not force:
            with open(dest) as f:
                existing_config = json.load(f)

            # Ensure mcpServers key exists
            if "mcpServers" not in existing_config:
                existing_config["mcpServers"] = {}

            # Merge new servers (don't overwrite existing ones)
            added_any = False
            for server_name, server_config in new_servers.items():
                if server_name not in existing_config["mcpServers"]:
                    existing_config["mcpServers"][server_name] = server_config
                    added_any = True

            # If no new servers were added, skip
            if not added_any:
                return FileResult(str(dest.relative_to(self.target)), "skipped")

            mcp_config = existing_config
            status = "merged"
        else:
            # Create new config or force overwrite
            mcp_config = {"mcpServers": new_servers}
            status = "overwritten" if dest.exists() else "created"

        # Write MCP configuration
        with open(dest, "w") as f:
            json.dump(mcp_config, f, indent=2)
            f.write("\n")  # Add trailing newline

        return FileResult(str(dest.relative_to(self.target)), status)

    def _index_repository(self) -> bool:
        """Generate repository index using Claude CLI with prompt."""
        if not shutil.which("claude"):
            return False

        prompt = """Analyze this repository and provide a concise overview:
- Main purpose and key technologies
- Directory structure (2-3 levels max)
- Entry points and main files
- Build/run commands (check for package.json scripts, Makefile targets, Justfile recipes, etc.)
- Available scripts and automation tools

Format as clean markdown starting at heading level 3 (###), keep it brief (under 500 words)."""

        try:
            result = subprocess.run(
                [
                    "claude",
                    "--print",
                    "--output-format",
                    "text",
                    prompt,
                ],
                cwd=self.target,
                capture_output=True,
                text=True,
                timeout=90,
                check=False,
            )

            if result.returncode == 0 and result.stdout.strip():
                return self._insert_repository_index(result.stdout.strip())
            return False
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False

    def _insert_repository_index(self, index_content: str) -> bool:
        """Insert generated repository index into CLAUDE.md."""
        claude_md_path = self.target / "CLAUDE.md"

        if not claude_md_path.exists():
            return False

        with open(claude_md_path) as f:
            content = f.read()

        # Find placeholder markers
        start_marker = "<!-- REPOSITORY_INDEX_START -->"
        end_marker = "<!-- REPOSITORY_INDEX_END -->"

        if start_marker not in content or end_marker not in content:
            return False

        # Replace content between markers
        start_idx = content.find(start_marker) + len(start_marker)
        end_idx = content.find(end_marker)

        updated_content = content[:start_idx] + "\n" + index_content + "\n" + content[end_idx:]

        with open(claude_md_path, "w") as f:
            f.write(updated_content)

        return True

    def _track_result(self, results: dict[str, list[str]], file_result: FileResult) -> None:
        """Track file operation result."""
        # Map status to result key
        status_map = {
            "created": "installed",
            "overwritten": "overwritten",
            "skipped": "skipped",
            "merged": "merged",
        }
        key = status_map.get(file_result.status, file_result.status)
        results[key].append(file_result.path)

    def _build_message(self, results: dict[str, list[str]], verbose: bool) -> str:
        """Build user-facing message."""
        installed = len(results.get("installed", []))
        skipped = len(results.get("skipped", []))
        overwritten = len(results.get("overwritten", []))
        merged = len(results.get("merged", []))

        if verbose:
            # Detailed message
            lines = ["Initialized Claude Code configuration\n"]
            if installed:
                lines.append(f"  Installed {installed} file(s):")
                for f in results["installed"]:
                    lines.append(f"    • {f}")
            if overwritten:
                lines.append(f"  Overwritten {overwritten} file(s):")
                for f in results["overwritten"]:
                    lines.append(f"    • {f}")
            if merged:
                lines.append(f"  Merged {merged} file(s):")
                for f in results["merged"]:
                    lines.append(f"    • {f}")
            if skipped:
                lines.append(f"  Skipped {skipped} existing file(s):")
                for f in results["skipped"]:
                    lines.append(f"    • {f}")
            if "indexed" in results:
                lines.append("  Repository indexed with Claude CLI")
            return "\n".join(lines)
        else:
            # Summary message
            parts = []
            if installed:
                parts.append(f"{installed} installed")
            if overwritten:
                parts.append(f"{overwritten} overwritten")
            if merged:
                parts.append(f"{merged} merged")
            if skipped:
                parts.append(f"{skipped} skipped")

            summary = ", ".join(parts) if parts else "nothing to do"
            return f"Initialized Claude Code configuration ({summary})"
