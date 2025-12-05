"""Core initialization operations for agentctl init command."""

from __future__ import annotations

import json
import shutil
import subprocess
from pathlib import Path

from rich.console import Console

from agentctl.operations.settings_merge import merge_settings_smart


class ImportDirNotFoundError(Exception):
    """Raised when import directory cannot be found."""

    def __init__(self) -> None:
        super().__init__("Import directory not found. This may indicate a corrupted installation.")


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
        console: Console,
        force: bool = False,
        skip_index: bool = False,
    ) -> None:
        """Execute full initialization."""
        # 1. Install CLAUDE.md
        console.print("Installing CLAUDE.md...")
        self._install_file(
            self.template_dir / "CLAUDE.md",
            self.target / "CLAUDE.md",
            force,
            console,
        )

        # 2. Install agents
        console.print("Installing agents...")
        count = self._install_directory(
            self.template_dir / "agents",
            self.target / ".claude" / "agents",
            force,
            console,
            pattern="*.md",
        )
        console.print(f"  → Installed {count} agent(s)")

        # 3. Install skills
        console.print("Installing skills...")
        count = self._install_directory(
            self.template_dir / "skills",
            self.target / ".claude" / "skills",
            force,
            console,
            recursive=True,
        )
        console.print(f"  → Installed {count} skill(s)")

        # 4. Merge settings
        console.print("Merging settings.json...")
        self._merge_settings(
            self.template_dir / "settings.json",
            self.target / ".claude" / "settings.json",
            force,
            console,
        )

        # 5. Configure MCP servers
        console.print("Configuring MCP servers...")
        self._configure_mcp(
            self.target / ".mcp.json",
            force,
            console,
        )

        # 6. Index repository with claude CLI
        if not skip_index:
            self._index_repository(console)

        console.print("\n✓ Initialization complete")

    def _install_file(
        self,
        source: Path,
        dest: Path,
        force: bool,
        console: Console,
    ) -> None:
        """Install a single file."""
        if dest.exists() and not force:
            console.print(f"  • {dest.relative_to(self.target)} (skipped)")
            return

        dest.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, dest)
        status = "overwritten" if dest.exists() else "created"
        console.print(f"  • {dest.relative_to(self.target)} ({status})")

    def _install_directory(
        self,
        source: Path,
        dest: Path,
        force: bool,
        console: Console,
        pattern: str = "*",
        recursive: bool = False,
    ) -> int:
        """Install files from a directory."""
        count = 0
        dest.mkdir(parents=True, exist_ok=True)

        if recursive:
            # Copy entire directory trees (for skills)
            for item in source.iterdir():
                if item.is_dir():
                    dest_item = dest / item.name
                    existed = dest_item.exists()

                    if existed and not force:
                        console.print(f"  • {dest_item.relative_to(self.target)} (skipped)")
                        continue

                    shutil.copytree(item, dest_item, dirs_exist_ok=force)
                    status = "overwritten" if existed else "created"
                    console.print(f"  • {dest_item.relative_to(self.target)} ({status})")
                    count += 1
        else:
            # Copy matching files (for agents)
            for item in source.glob(pattern):
                if item.is_file():
                    self._install_file(item, dest / item.name, force, console)
                    count += 1

        return count

    def _merge_settings(
        self,
        source: Path,
        dest: Path,
        force: bool,
        console: Console,
    ) -> None:
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
            console.print(f"  • {dest.relative_to(self.target)} (created)")
            return

        # Existing settings - merge
        with open(dest) as f:
            existing_settings = json.load(f)

        if force:
            # Force: overwrite
            with open(dest, "w") as f:
                json.dump(new_settings, f, indent=2)
                f.write("\n")  # Add trailing newline
            console.print(f"  • {dest.relative_to(self.target)} (overwritten)")
            return

        # Smart merge
        merged = merge_settings_smart(existing_settings, new_settings)
        with open(dest, "w") as f:
            json.dump(merged, f, indent=2)
            f.write("\n")  # Add trailing newline
        console.print(f"  • {dest.relative_to(self.target)} (merged)")

    def _configure_mcp(
        self,
        dest: Path,
        force: bool,
        console: Console,
    ) -> None:
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
                console.print(f"  • {dest.relative_to(self.target)} (skipped)")
                return

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

        console.print(f"  • {dest.relative_to(self.target)} ({status})")

    def _index_repository(self, console: Console) -> None:
        """Generate repository index using Claude CLI with prompt."""
        if not shutil.which("claude"):
            return

        prompt = """Analyze this repository and provide a concise overview:
- Main purpose and key technologies
- Directory structure (2-3 levels max)
- Entry points and main files
- Build/run commands (check for package.json scripts, Makefile targets, Justfile recipes, etc.)
- Available scripts and automation tools

Format as clean markdown starting at heading level 3 (###), keep it brief (under 500 words)."""

        try:
            with console.status("Indexing repository with Claude CLI...", spinner="dots"):
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

            output = result.stdout.strip()
            if result.returncode == 0 and output and self._insert_repository_index(output):
                console.print("  → Repository indexed successfully")
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

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
