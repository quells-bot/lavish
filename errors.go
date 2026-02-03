package lavish

import (
	"fmt"
	"strings"
)

// RenderError represents an error that occurred during JSX rendering.
type RenderError struct {
	Name    string // JSX file name
	Phase   string // "compile", "execute", or "render"
	Message string
	Cause   error
}

func (e *RenderError) Error() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "%s failed during %s", e.Name, e.Phase)
	if e.Message != "" {
		_, _ = fmt.Fprintf(&b, ": %s", e.Message)
	}
	if e.Cause != nil {
		_, _ = fmt.Fprintf(&b, ": %s", e.Cause.Error())
	}
	return b.String()
}

func (e *RenderError) Unwrap() error {
	return e.Cause
}

// MissingRenderCallError indicates the JSX program did not call render().
type MissingRenderCallError struct {
	Name           string
	RenderFunction string
}

func (e *MissingRenderCallError) Error() string {
	return fmt.Sprintf(
		"%s did not produce output: ensure your JSX calls %s(<Component />) at the end",
		e.Name,
		e.RenderFunction,
	)
}

// RuntimeError wraps JavaScript runtime errors with context.
type RuntimeError struct {
	Name    string
	Message string
	Cause   error
}

func (e *RuntimeError) Error() string {
	// Try to make common goja errors more readable
	msg := e.Cause.Error()

	// Simplify common error patterns
	if strings.Contains(msg, "Cannot read property") {
		return fmt.Sprintf("%s: attempted to access property on null or undefined value: %s", e.Name, msg)
	}
	if strings.Contains(msg, "is not defined") {
		return fmt.Sprintf("%s: referenced undefined variable: %s", e.Name, msg)
	}
	if strings.Contains(msg, "has no member") {
		return fmt.Sprintf("%s: method or property does not exist: %s", e.Name, msg)
	}
	if strings.Contains(msg, "not an object") || strings.Contains(msg, "Value is not an object") {
		return fmt.Sprintf("%s: expected an object but got a primitive value: %s", e.Name, msg)
	}

	return fmt.Sprintf("%s: runtime error: %s", e.Name, msg)
}

func (e *RuntimeError) Unwrap() error {
	return e.Cause
}
