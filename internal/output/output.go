// Package output provides utilities for formatting and displaying command output.
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// Error writes an error message to stderr.
func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// Errorf writes a formatted error message to stderr.
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(format, args...))
}

// SuccessJSON writes a successful result as JSON.
func SuccessJSON(data interface{}) error {
	result := Result{
		Success: true,
		Data:    data,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// ErrorJSON writes an error result as JSON.
func ErrorJSON(err error) error {
	result := Result{
		Success: false,
		Message: err.Error(),
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// WriteJSON writes raw data as JSON (without Result wrapper).
func WriteJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
