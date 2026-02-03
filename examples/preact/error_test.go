package main

import (
	"strings"
	"testing"

	"github.com/quells/lavish"
	"github.com/quells/lavish/engine/preact10"
)

func TestInvalidJSXSyntax(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tests := []struct {
		name         string
		jsx          string
		expectErrContains string
	}{
		{
			name: "unclosed tag",
			jsx: `
				const App = () => (
					<div>
						<span>Hello
					</div>
				);
				render(<App />)
			`,
			expectErrContains: "Expected closing",
		},
		{
			name: "unclosed bracket",
			jsx: `
				const App = () => (
					<div>{data.items.map(x => <span>{x}</span>}</div>
				);
				render(<App />)
			`,
			expectErrContains: "Expected",
		},
		{
			name: "missing closing parenthesis",
			jsx: `
				const App = () => (
					<div>Hello</div>
				;
				render(<App />)
			`,
			expectErrContains: "Expected",
		},
		{
			name: "invalid attribute syntax",
			jsx: `
				const App = () => (
					<div class==>Hello</div>
				);
				render(<App />)
			`,
			expectErrContains: "Expected",
		},
		{
			name: "unterminated string literal",
			jsx: `
				const App = () => (
					<div>{"Hello</div>
				);
				render(<App />)
			`,
			expectErrContains: "Unterminated string literal",
		},
		{
			name: "completely malformed jsx",
			jsx:  `<<<<>>>>`,
			expectErrContains: "Unexpected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bundle.RenderJSX("invalid.jsx", tt.jsx, nil)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expectErrContains) {
				t.Fatalf("expected error containing %q, got: %s", tt.expectErrContains, err.Error())
			}
		})
	}
}

func TestRuntimeErrors(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tests := []struct {
		name              string
		jsx               string
		data              any
		expectErrContains string
	}{
		{
			name: "undefined variable access",
			jsx: `
				const App = () => (
					<div>{undefinedVar.property}</div>
				);
				render(<App />)
			`,
			data:              nil,
			expectErrContains: "not defined",
		},
		{
			name: "null property access",
			jsx: `
				const App = () => (
					<div>{data.nested.property}</div>
				);
				render(<App />)
			`,
			data:              nil,
			expectErrContains: "Cannot read property",
		},
		{
			name: "call non-function",
			jsx: `
				const notAFunction = "hello";
				const App = () => (
					<div>{notAFunction()}</div>
				);
				render(<App />)
			`,
			data:              nil,
			expectErrContains: "not an object",
		},
		{
			name: "map on non-array",
			jsx: `
				const App = () => (
					<div>{data.items.map(x => <span>{x}</span>)}</div>
				);
				render(<App />)
			`,
			data:              map[string]any{"items": "not an array"},
			expectErrContains: "has no member 'map'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bundle.RenderJSX("runtime_error.jsx", tt.jsx, tt.data)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expectErrContains) {
				t.Fatalf("expected error containing %q, got: %s", tt.expectErrContains, err.Error())
			}
		})
	}
}

func TestMissingRenderCall(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	jsx := `
		const App = () => (
			<div>Hello</div>
		);
		// Forgot to call render!
	`

	_, err := bundle.RenderJSX("no_render.jsx", jsx, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "did not produce output") {
		t.Fatalf("expected error about missing render call, got: %s", err.Error())
	}
	// Should also suggest the fix
	if !strings.Contains(err.Error(), "ensure your JSX calls render(") {
		t.Fatalf("expected error to suggest fix, got: %s", err.Error())
	}
}

func TestInvalidComponentLoader(t *testing.T) {
	renderer := preact10.RenderEngine

	// Create a bundle with an invalid component that has syntax errors
	invalidComponentJSX := `
		const BrokenComponent = () => (
			<div>
				<span>Unclosed
			</div>
		);
	`

	bundle := lavish.NewBundle(
		renderer,
		lavish.ComponentJSX("broken.jsx", invalidComponentJSX, renderer.GetJSXOptions()),
	)

	validJSX := `
		render(<div>Hello</div>)
	`

	_, err := bundle.RenderJSX("valid.jsx", validJSX, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "failed to compile component") {
		t.Fatalf("expected error about component compilation, got: %s", err.Error())
	}
}

func TestTransformErrorDetails(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	// This JSX has an error at a specific location
	jsx := `const x = (`

	_, err := bundle.RenderJSX("location_test.jsx", jsx, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	errStr := err.Error()
	// Error should clearly indicate it's a syntax error
	if !strings.Contains(errStr, "syntax error") {
		t.Fatalf("expected error to indicate syntax error, got: %s", errStr)
	}
	// Error should contain the file name
	if !strings.Contains(errStr, "location_test.jsx") {
		t.Fatalf("expected error to contain filename, got: %s", errStr)
	}
	// Error should contain line number
	if !strings.Contains(errStr, "line 1") {
		t.Fatalf("expected error to contain line number, got: %s", errStr)
	}
	// Error should show the source line
	if !strings.Contains(errStr, "const x = (") {
		t.Fatalf("expected error to show source line, got: %s", errStr)
	}
	// Error should show a caret pointing to the error
	if !strings.Contains(errStr, "^") {
		t.Fatalf("expected error to show caret, got: %s", errStr)
	}
}

func TestErrorShowsSourceContext(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tests := []struct {
		name           string
		jsx            string
		expectContains []string
	}{
		{
			name: "unclosed tag shows both lines",
			jsx: `const App = () => (
    <div>
        <span>Hello
    </div>
);`,
			expectContains: []string{
				"</div>",      // shows the problematic line
				"^",           // caret pointing to error
				"span",        // mentions the unclosed tag
			},
		},
		{
			name: "typo in JSX attribute",
			jsx:  `const App = () => <div class=="test">Hi</div>;`,
			expectContains: []string{
				`class=="test"`, // shows the source with typo
				"^",             // caret
			},
		},
		{
			name: "missing closing brace",
			jsx: `const App = () => (
    <div>{items.map(x => <span>{x}</span>}</div>
);`,
			expectContains: []string{
				"items.map", // shows the source line
				"^",         // caret
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bundle.RenderJSX("test.jsx", tt.jsx, nil)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			errStr := err.Error()
			for _, s := range tt.expectContains {
				if !strings.Contains(errStr, s) {
					t.Errorf("error should contain %q, got:\n%s", s, errStr)
				}
			}
		})
	}
}

func TestImprovedErrorMessages(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	t.Run("syntax error is clearly labeled", func(t *testing.T) {
		jsx := `<div>unclosed`
		_, err := bundle.RenderJSX("test.jsx", jsx, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.HasPrefix(err.Error(), "syntax error in test.jsx") {
			t.Fatalf("error should start with 'syntax error in': %s", err.Error())
		}
	})

	t.Run("single syntax error says 'syntax error' (singular)", func(t *testing.T) {
		jsx := `const a = (`
		_, err := bundle.RenderJSX("single.jsx", jsx, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		// Single error should say "syntax error" (singular)
		if !strings.Contains(err.Error(), "syntax error in single.jsx") {
			t.Fatalf("expected singular 'syntax error in': %s", err.Error())
		}
	})

	t.Run("runtime error shows context", func(t *testing.T) {
		jsx := `
			const App = () => <div>{undefined.foo}</div>;
			render(<App />)
		`
		_, err := bundle.RenderJSX("runtime.jsx", jsx, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		errStr := err.Error()
		// Should include the file name
		if !strings.Contains(errStr, "runtime.jsx") {
			t.Fatalf("error should include filename: %s", errStr)
		}
	})

	t.Run("missing render suggests fix", func(t *testing.T) {
		jsx := `const App = () => <div>Hello</div>;`
		_, err := bundle.RenderJSX("norender.jsx", jsx, nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		errStr := err.Error()
		// Should mention the file
		if !strings.Contains(errStr, "norender.jsx") {
			t.Fatalf("error should include filename: %s", errStr)
		}
		// Should suggest the fix
		if !strings.Contains(errStr, "ensure your JSX calls render(") {
			t.Fatalf("error should suggest calling render(): %s", errStr)
		}
	})
}

func TestEmptyJSX(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	// Empty JSX should fail because render is never called
	_, err := bundle.RenderJSX("empty.jsx", "", nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "did not produce output") {
		t.Fatalf("expected error about missing render call, got: %s", err.Error())
	}
}

func TestWhitespaceOnlyJSX(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	// Whitespace-only JSX should fail because render is never called
	_, err := bundle.RenderJSX("whitespace.jsx", "   \n\t\n   ", nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "did not produce output") {
		t.Fatalf("expected error about missing render call, got: %s", err.Error())
	}
}

// TestData is a struct used to test property access from JSX
type TestData struct {
	Name    string
	Count   int
	Items   []string
	Nested  *NestedData
	private string
}

type NestedData struct {
	Value string
}

func TestNonexistentPropertyAccess(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tests := []struct {
		name        string
		jsx         string
		data        any
		expectErr   bool
		errContains string
		expectHTML  string // if no error expected, check output
	}{
		{
			name: "access nonexistent field on struct",
			jsx: `
				const App = () => <div>{data.NonExistent}</div>;
				render(<App />)
			`,
			data:       TestData{Name: "test"},
			expectErr:  false,
			expectHTML: "<div></div>", // undefined renders as empty
		},
		{
			name: "access nonexistent nested field",
			jsx: `
				const App = () => <div>{data.NonExistent.Deep}</div>;
				render(<App />)
			`,
			data:        TestData{Name: "test"},
			expectErr:   true,
			errContains: "Cannot read property",
		},
		{
			name: "access nonexistent key on map",
			jsx: `
				const App = () => <div>{data.missing}</div>;
				render(<App />)
			`,
			data:       map[string]string{"present": "value"},
			expectErr:  false,
			expectHTML: "<div></div>", // undefined renders as empty
		},
		{
			name: "access nested nonexistent key on map",
			jsx: `
				const App = () => <div>{data.missing.nested}</div>;
				render(<App />)
			`,
			data:        map[string]string{"present": "value"},
			expectErr:   true,
			errContains: "Cannot read property",
		},
		{
			name: "access property on nil nested struct",
			jsx: `
				const App = () => <div>{data.Nested.Value}</div>;
				render(<App />)
			`,
			data:        TestData{Name: "test", Nested: nil},
			expectErr:   true,
			errContains: "Cannot read property",
		},
		{
			name: "access valid nested struct property",
			jsx: `
				const App = () => <div>{data.Nested.Value}</div>;
				render(<App />)
			`,
			data:       TestData{Name: "test", Nested: &NestedData{Value: "hello"}},
			expectErr:  false,
			expectHTML: "<div>hello</div>",
		},
		{
			name: "access property with typo",
			jsx: `
				const App = () => <div>{data.Naem}</div>;
				render(<App />)
			`,
			data:       TestData{Name: "test"},
			expectErr:  false,
			expectHTML: "<div></div>", // undefined due to typo
		},
		{
			name: "access array index out of bounds",
			jsx: `
				const App = () => <div>{data.Items[99]}</div>;
				render(<App />)
			`,
			data:       TestData{Items: []string{"a", "b"}},
			expectErr:  false,
			expectHTML: "<div></div>", // undefined for out of bounds
		},
		{
			name: "iterate over nil slice",
			jsx: `
				const App = () => (
					<ul>{data.Items.map(x => <li>{x}</li>)}</ul>
				);
				render(<App />)
			`,
			data:       TestData{Items: nil},
			expectErr:  false,
			expectHTML: "<ul></ul>", // goja converts nil slice to empty array
		},
		{
			name: "iterate over empty slice",
			jsx: `
				const App = () => (
					<ul>{data.Items.map(x => <li>{x}</li>)}</ul>
				);
				render(<App />)
			`,
			data:       TestData{Items: []string{}},
			expectErr:  false,
			expectHTML: "<ul></ul>",
		},
		{
			name: "access unexported field",
			jsx: `
				const App = () => <div>{data.private}</div>;
				render(<App />)
			`,
			data:       TestData{Name: "test"},
			expectErr:  false,
			expectHTML: "<div></div>", // unexported fields not visible
		},
		{
			name: "wrong data type - expected object got string",
			jsx: `
				const App = () => <div>{data.Name}</div>;
				render(<App />)
			`,
			data:       "just a string",
			expectErr:  false,
			expectHTML: "<div></div>", // string has no .Name property
		},
		{
			name: "wrong data type - expected array got object",
			jsx: `
				const App = () => (
					<ul>{data.map(x => <li>{x}</li>)}</ul>
				);
				render(<App />)
			`,
			data:        TestData{Name: "test"},
			expectErr:   true,
			errContains: "has no member 'map'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bundle.RenderJSX("test.jsx", tt.jsx, tt.data)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got result: %s", result)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got: %s", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err.Error())
				}
				if result != tt.expectHTML {
					t.Fatalf("expected HTML %q, got %q", tt.expectHTML, result)
				}
			}
		})
	}
}
