package lavish

import (
	"fmt"
	"regexp"
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

// Patterns to extract useful info from goja errors
var (
	// Matches: Cannot read property 'PropName' of undefined
	cannotReadPropRe = regexp.MustCompile(`Cannot read property '([^']+)' of (undefined|null)`)
	// Matches: Object has no member 'memberName'
	noMemberRe = regexp.MustCompile(`Object has no member '([^']+)'`)
	// Matches: varName is not defined
	notDefinedRe = regexp.MustCompile(`(\w+) is not defined`)
	// Matches: at FuncName (file:line:col)
	locationRe = regexp.MustCompile(`at (\w+) \(([^)]+)\)`)
)

func (e *RuntimeError) Error() string {
	msg := e.Cause.Error()
	var b strings.Builder

	// Extract location info if present
	location := ""
	if matches := locationRe.FindStringSubmatch(msg); matches != nil {
		funcName := matches[1]
		loc := matches[2]
		// Parse "file:line:col(offset)" format
		if parts := strings.Split(loc, ":"); len(parts) >= 2 {
			location = fmt.Sprintf(" (in %s at line %s)", funcName, parts[1])
		}
	}

	// Handle "Cannot read property 'X' of undefined/null"
	if matches := cannotReadPropRe.FindStringSubmatch(msg); matches != nil {
		propName := matches[1]
		nullOrUndef := matches[2]
		_, _ = fmt.Fprintf(&b, "%s: cannot access '%s' because the parent value is %s%s\n",
			e.Name, propName, nullOrUndef, location)
		_, _ = fmt.Fprint(&b, "  hint: check that all intermediate properties exist before accessing nested values\n")
		_, _ = fmt.Fprintf(&b, "  hint: use optional chaining in your JSX: data?.Parent?.%s", propName)
		return b.String()
	}

	// Handle "Object has no member 'X'"
	if matches := noMemberRe.FindStringSubmatch(msg); matches != nil {
		memberName := matches[1]
		_, _ = fmt.Fprintf(&b, "%s: '%s' does not exist on this value%s\n",
			e.Name, memberName, location)
		if memberName == "map" || memberName == "forEach" || memberName == "filter" {
			_, _ = fmt.Fprint(&b, "  hint: you're calling an array method on a non-array value\n")
			_, _ = fmt.Fprint(&b, "  hint: ensure the data property is an array, not an object or primitive")
		} else {
			_, _ = fmt.Fprintf(&b, "  hint: check that the property name '%s' is spelled correctly", memberName)
		}
		return b.String()
	}

	// Handle "X is not defined"
	if matches := notDefinedRe.FindStringSubmatch(msg); matches != nil {
		varName := matches[1]
		_, _ = fmt.Fprintf(&b, "%s: '%s' is not defined%s\n",
			e.Name, varName, location)
		if varName == "data" {
			_, _ = fmt.Fprint(&b, "  hint: the 'data' variable should be passed from Go via bundle.RenderJSX()")
		} else {
			_, _ = fmt.Fprintf(&b, "  hint: '%s' may be misspelled, or you may need to import/define it", varName)
		}
		return b.String()
	}

	// Handle "Value is not an object"
	if strings.Contains(msg, "not an object") || strings.Contains(msg, "Value is not an object") {
		_, _ = fmt.Fprintf(&b, "%s: tried to call a function on a non-object value%s\n",
			e.Name, location)
		_, _ = fmt.Fprint(&b, "  hint: you may be calling a method on a string, number, or other primitive")
		return b.String()
	}

	// Default: include the original error with some context
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %s", e.Name, e.Message, msg)
	}
	return fmt.Sprintf("%s: runtime error: %s", e.Name, msg)
}

func (e *RuntimeError) Unwrap() error {
	return e.Cause
}
