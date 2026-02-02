# CLAUDE.md

## Project Overview

Lavish is an HTML templating library for Go that enables JSX syntax for server-side rendering. It compiles JSX templates into JavaScript, executes them in a pure Go runtime (goja), and renders HTML using Preact.

## Tech Stack

- **Go 1.18+** - Primary language
- **goja** - Pure Go JavaScript runtime
- **esbuild** - JSX transpiler
- **Preact v10** - Lightweight React alternative for SSR
- **hashicorp/golang-lru/v2** - Cache implementations

## Project Structure

```
lavish.go           # Core API: CompileJSX, Loader, RenderEngine interfaces
bundle.go           # Bundle struct: orchestrates rendering with caching
cache.go            # Cache interface definitions
cache/              # LRU and TwoQueue cache implementations
engine/preact10/    # Public Preact engine interface
internal/preact10/  # Preact engine implementation and bundled JS
examples/preact/    # Demo HTTP server and benchmark tests
```

## Common Commands

```bash
# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./examples/preact/

# Run the example server (serves on :3000)
go run ./examples/preact/

# Regenerate Preact bundle (if modifying internal/preact10/preact-glue.js)
cd internal/preact10 && go generate
```

## Key Interfaces

- **`Loader`** - Module loading into goja runtime
- **`RenderEngine`** - Framework for rendering (JSX options, render function)
- **`ProgramCache`** - Stores compiled goja.Program by hash
- **`RenderCache`** - Stores final HTML by (program hash, data hash)
- **`Hashable`** - Interface for data objects that can be hashed for caching

## Architecture Notes

- **Two-level caching**: Program cache (compiled JS) and render cache (final HTML)
- **Functional options pattern**: Use `bundle.With*()` functions for configuration
- **Interface-based design**: Engines, caches, and loaders are pluggable
- Uses FNV-64a hashing for cache keys

## JSX Conventions

- Data is passed via the `data` variable (configurable with `WithDataVariable()`)
- Render is called with `render(<Component />)` (configurable with `WithRenderFunction()`)
- Uses Preact's `h` factory and `Fragment` for JSX transpilation
