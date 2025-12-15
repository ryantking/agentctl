// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"errors"
)

// IsNotFound checks if an error is ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsTimeout checks if an error is ErrTimeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsCanceled checks if an error is ErrCanceled.
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled)
}

// IsEmptyOutput checks if an error is ErrEmptyOutput.
func IsEmptyOutput(err error) bool {
	return errors.Is(err, ErrEmptyOutput)
}

// IsValidationFailed checks if an error is ErrValidationFailed.
func IsValidationFailed(err error) bool {
	return errors.Is(err, ErrValidationFailed)
}

// AsAgentError checks if an error is an AgentError and extracts it.
func AsAgentError(err error, target **AgentError) bool {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		*target = agentErr
		return true
	}
	return false
}
