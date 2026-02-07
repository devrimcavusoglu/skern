package cli

// ValidationError represents a user input validation failure (exit code 2).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
