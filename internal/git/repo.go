package git

import (
	"fmt"
)

// ErrNotInGitRepo is returned when a git repository cannot be found.
var ErrNotInGitRepo = fmt.Errorf("not in a git repository")
