package output

// Result represents a command result for JSON output.
type Result struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success creates a successful result.
func Success(data interface{}) Result {
	return Result{
		Success: true,
		Data:    data,
	}
}

// ErrorResult creates an error result.
func ErrorResult(msg string) Result {
	return Result{
		Success: false,
		Message: msg,
	}
}
