// Package agent provides tool schema definitions for repository exploration tools.
package agent

// Tool schemas for repository exploration tools

// ListDirectorySchema defines the schema for the list_directory tool.
var ListDirectorySchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Directory path to list (relative to repository root)",
		},
	},
	"required": []interface{}{"path"},
}

// ReadFileSchema defines the schema for the read_file tool.
var ReadFileSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "File path to read (relative to repository root)",
		},
	},
	"required": []interface{}{"path"},
}

// SearchFilesSchema defines the schema for the search_files tool.
var SearchFilesSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"pattern": map[string]interface{}{
			"type":        "string",
			"description": "Search pattern (regex or plain text)",
		},
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Directory path to search in (relative to repository root, defaults to root)",
		},
		"case_sensitive": map[string]interface{}{
			"type":        "boolean",
			"description": "Whether search is case sensitive (default: false)",
		},
	},
	"required": []interface{}{"pattern"},
}

// GetFileInfoSchema defines the schema for the get_file_info tool.
var GetFileInfoSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "File path (relative to repository root)",
		},
	},
	"required": []interface{}{"path"},
}

// ListGitFilesSchema defines the schema for the list_git_files tool.
var ListGitFilesSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Directory path to list tracked files in (relative to repository root, defaults to root)",
		},
	},
	"required": []interface{}{},
}
