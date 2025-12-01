# Semtools CLI Reference

Complete reference documentation for all semtools commands.

## parse - Document Parser

Parse unsupported file formats (PDF, DOCX, PPTX, etc.) into searchable markdown.

### Usage

```bash
parse [OPTIONS] <FILES>...
```

### Arguments

- `<FILES>...` - Files to parse (required, multiple files accepted)

### Options

- `-c, --parse-config <PARSE_CONFIG>` - Path to config file (default: `~/.parse_config.json`)
- `-b, --backend <BACKEND>` - Backend type for parsing (default: `llama-parse`)
- `-h, --help` - Print help
- `-V, --version` - Print version

### Output

- Parsed files are cached to `~/.parse/`
- Output format is markdown with `.md` extension appended
- Example: `document.pdf` → `~/.parse/document.pdf.md`

### Examples

```bash
# Parse single PDF
parse document.pdf

# Parse multiple files
parse report.pdf data.xlsx presentation.pptx

# Parse directory
parse ./papers/*.pdf

# Parse with custom config
parse --parse-config ~/my_config.json *.pdf

# Pipe to search
parse *.pdf | xargs search "query"
```

### Environment Variables

- `LLAMA_CLOUD_API_KEY` - **Required** for parsing operations

### Cache Behavior

- First parse: Sends file to LlamaParse API, caches result
- Subsequent parses: Returns cached result instantly (no API call)
- Cache location: `~/.parse/`
- Cache is persistent across sessions

---

## search - Semantic Search

Perform local semantic keyword search using AI embeddings.

### Usage

```bash
search [OPTIONS] <QUERY> [FILES]...
```

### Arguments

- `<QUERY>` - Search query (required, positional argument)
- `[FILES]...` - Files or directories to search (optional, searches stdin if omitted)

### Options

- `-n, --n-lines <N_LINES>` - Lines of context before/after match (default: 3)
- `--top-k <TOP_K>` - Return top K results (default: 3, ignored if `--max-distance` set)
- `-m, --max-distance <MAX_DISTANCE>` - Distance threshold (0.0+, returns all results below threshold)
- `-i, --ignore-case` - Case-insensitive search (default: false)
- `-h, --help` - Print help
- `-V, --version` - Print version

### Distance Threshold Guidance

The `--max-distance` parameter controls semantic similarity (lower = more similar):

- **0.2** - Very strict, only highly similar matches
- **0.3** - Balanced, recommended default
- **0.4** - Broader, more exploratory
- **0.5+** - Very broad, may include loosely related content

### Examples

```bash
# Basic search in current directory
search "kubernetes authentication" .

# Search with recommended parameters
search "vault configuration" docs/ --ignore-case --n-lines 30 --max-distance 0.3

# Short form
search "crossplane XRD" docs/ -i -n 30 -m 0.3

# Search from stdin
cat config.yaml | search "secret management"
echo "some text" | search "content"

# Search specific files
search "API endpoints" *.yaml --top-k 5

# Search parsed PDFs
parse document.pdf | xargs cat | search "security" -i -n 30

# Chain with grep for pre-filtering
cat *.yaml | grep -i "vault" | search "authentication" -i -n 30 -m 0.3

# Multi-directory search
search "error handling" src/ docs/ -i -n 30 -m 0.3
```

### Workspace Integration

When `SEMTOOLS_WORKSPACE` environment variable is set, search will:
- Cache embeddings for faster repeated searches
- Automatically detect file changes and re-embed
- Store cache in `~/.semtools/workspaces/<workspace-name>/`

```bash
# Use with workspace
export SEMTOOLS_WORKSPACE=my-project
search "query" files/ -i -n 30 -m 0.3
```

### Performance Notes

- First search: Generates embeddings (slower)
- Cached searches: Uses stored embeddings (much faster)
- Workspace recommended for large file collections

---

## workspace - Workspace Management

Manage semtools workspaces for caching embeddings.

### Usage

```bash
workspace <COMMAND>
```

### Commands

#### use

Create or switch to a workspace.

```bash
workspace use <NAME>
```

**Output:**
```
Workspace '<NAME>' configured.
To activate it, run:
  export SEMTOOLS_WORKSPACE=<NAME>

Or add this to your shell profile (.bashrc, .zshrc, etc.)
```

**Example:**
```bash
workspace use research-project
export SEMTOOLS_WORKSPACE=research-project
```

#### status

Show active workspace and statistics.

```bash
workspace status
```

**Output:**
- Current workspace name (if active)
- Number of cached files
- Cache size
- Last modified date

#### prune

Remove stale or missing files from workspace cache.

```bash
workspace prune
```

**Use when:**
- Files have been deleted but cache remains
- Cache contains outdated embeddings
- Cleaning up disk space

### Workspace Directory Structure

```
~/.semtools/workspaces/
├── project-1/
│   ├── embeddings/
│   └── metadata.json
└── project-2/
    ├── embeddings/
    └── metadata.json
```

### Benefits of Workspaces

1. **Speed**: Embeddings cached, subsequent searches much faster
2. **Cost**: Avoid re-embedding same files repeatedly
3. **Isolation**: Different projects have separate caches
4. **Automatic**: Detects file changes and re-embeds as needed

### Workflow Example

```bash
# Setup
workspace use documentation-search
export SEMTOOLS_WORKSPACE=documentation-search

# First search (generates embeddings)
search "authentication" docs/ -i -n 30 -m 0.3

# Subsequent searches (uses cache, very fast)
search "authorization" docs/ -i -n 30 -m 0.3
search "secrets management" docs/ -i -n 30 -m 0.3

# Check workspace
workspace status

# Clean up if needed
workspace prune
```

---

## Command Chaining & Pipelines

Semtools is designed for Unix-style piping and chaining.

### parse → search

```bash
# Parse and search in one pipeline
parse *.pdf | xargs search "query" -i -n 30 -m 0.3

# Parse, then search cached content
parse documents/*.pdf
search "query" ~/.parse/*.pdf.md -i -n 30 -m 0.3
```

### grep → search

```bash
# Pre-filter with exact match, then semantic search
cat config/*.yaml | grep -i "vault" | search "authentication" -i -n 30 -m 0.3

# Find files, then search within them
grep -l "kubernetes" docs/*.md | xargs search "authentication" -i -n 30 -m 0.3
```

### find → parse → search

```bash
# Find files, parse, then search
find . -name "*.pdf" | xargs parse | xargs search "security" -i -n 30 -m 0.3
```

---

## Best Practices

### Always Use These Flags

For consistent, high-quality results:

```bash
search "query" files/ -i -n 30 -m 0.3
```

- `-i` / `--ignore-case`: Unknown capitalization in files
- `-n 30` / `--n-lines 30`: Default 3 is too small
- `-m 0.3` / `--max-distance 0.3`: Good precision/recall balance

### Workspace for Large Operations

Always create workspace before searching many files:

```bash
workspace use project-name
export SEMTOOLS_WORKSPACE=project-name
```

### Pre-parse PDFs

Parse once, search multiple times:

```bash
# Parse (one-time operation)
parse documents/*.pdf

# Search cached content (fast, repeatable)
search "topic 1" ~/.parse/ -i -n 30 -m 0.3
search "topic 2" ~/.parse/ -i -n 30 -m 0.3
```

### Adjust Distance Threshold

Start at 0.3, adjust based on results:

```bash
# Too many results? Tighten threshold
search "query" files/ -i -n 30 -m 0.2

# Too few results? Loosen threshold
search "query" files/ -i -n 30 -m 0.4
```

---

## Environment Variables

- `LLAMA_CLOUD_API_KEY` - Required for parse operations
- `SEMTOOLS_WORKSPACE` - Active workspace name for caching

---

## Exit Codes

All commands return standard Unix exit codes:
- `0` - Success
- `1` - General error
- `2` - Usage error (invalid arguments)
