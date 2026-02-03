package lavish_test

import (
	"bytes"
	"testing"

	"github.com/quells/lavish"
	"github.com/quells/lavish/engine/preact10"
)

func TestHashBuilder(t *testing.T) {
	// Same inputs should produce same hash
	h1 := lavish.NewHashBuilder().String("hello").Int(42).Bool(true).Sum()
	h2 := lavish.NewHashBuilder().String("hello").Int(42).Bool(true).Sum()
	if h1 != h2 {
		t.Errorf("same inputs produced different hashes: %d != %d", h1, h2)
	}

	// Different inputs should produce different hashes
	h3 := lavish.NewHashBuilder().String("hello").Int(43).Bool(true).Sum()
	if h1 == h3 {
		t.Errorf("different inputs produced same hash: %d", h1)
	}

	// Strings helper
	h4 := lavish.NewHashBuilder().Strings([]string{"a", "b", "c"}).Sum()
	h5 := lavish.NewHashBuilder().String("a").String("b").String("c").Sum()
	if h4 != h5 {
		t.Errorf("Strings helper produced different hash than individual String calls")
	}
}

func TestMustCompile(t *testing.T) {
	renderer := preact10.RenderEngine

	// Valid JSX should compile
	tmpl := lavish.MustCompile("test.jsx", `render(<div>Hello</div>)`, renderer)
	if tmpl == nil {
		t.Fatal("MustCompile returned nil")
	}
	if tmpl.Name() != "test.jsx" {
		t.Errorf("expected name 'test.jsx', got %q", tmpl.Name())
	}
}

func TestMustCompilePanics(t *testing.T) {
	renderer := preact10.RenderEngine

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustCompile should panic on invalid JSX")
		}
	}()

	lavish.MustCompile("invalid.jsx", `<div>unclosed`, renderer)
}

func TestCompile(t *testing.T) {
	renderer := preact10.RenderEngine

	// Valid JSX
	tmpl, err := lavish.Compile("test.jsx", `render(<div>Hello</div>)`, renderer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("Compile returned nil template")
	}

	// Invalid JSX
	_, err = lavish.Compile("invalid.jsx", `<div>unclosed`, renderer)
	if err == nil {
		t.Fatal("expected error for invalid JSX")
	}
}

func TestTemplateRender(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tmpl := lavish.MustCompile("test.jsx", `
		const App = () => <div>Hello {data.Name}</div>;
		render(<App />)
	`, renderer)

	data := map[string]string{"Name": "World"}
	result, err := tmpl.Render(&bundle, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "<div>Hello World</div>" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestTemplateRenderTo(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	tmpl := lavish.MustCompile("test.jsx", `render(<div>Hello</div>)`, renderer)

	var buf bytes.Buffer
	err := tmpl.RenderTo(&bundle, &buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "<div>Hello</div>" {
		t.Errorf("unexpected result: %s", buf.String())
	}
}

func TestRenderJSXTo(t *testing.T) {
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(renderer)

	var buf bytes.Buffer
	err := bundle.RenderJSXTo(&buf, "test.jsx", `render(<div>Hello</div>)`, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "<div>Hello</div>" {
		t.Errorf("unexpected result: %s", buf.String())
	}
}

// Example of simplified Hashable implementation
type UserData struct {
	ID    int
	Name  string
	Roles []string
}

func (u UserData) Hash64() uint64 {
	return lavish.NewHashBuilder().
		Int(u.ID).
		String(u.Name).
		Strings(u.Roles).
		Sum()
}

func TestHashableWithBuilder(t *testing.T) {
	u1 := UserData{ID: 1, Name: "Alice", Roles: []string{"admin", "user"}}
	u2 := UserData{ID: 1, Name: "Alice", Roles: []string{"admin", "user"}}
	u3 := UserData{ID: 2, Name: "Bob", Roles: []string{"user"}}

	if u1.Hash64() != u2.Hash64() {
		t.Error("identical data should produce same hash")
	}
	if u1.Hash64() == u3.Hash64() {
		t.Error("different data should produce different hash")
	}
}
