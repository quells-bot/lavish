package lavish

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
)

type Bundle struct {
	engine       RenderEngine
	modules      []Loader
	programCache ProgramCache
	renderCache  RenderCache

	dataVariable   string
	renderFunction string
}

func NewBundle(engine RenderEngine, modules ...Loader) Bundle {
	return Bundle{
		engine:  engine,
		modules: modules,

		dataVariable:   "data",
		renderFunction: "render",
	}
}

func (b Bundle) WithDataVariable(varName string) Bundle {
	b.dataVariable = varName
	return b
}

func (b Bundle) WithRenderFunction(funcName string) Bundle {
	b.renderFunction = funcName
	return b
}

func (b Bundle) WithProgramCache(cache ProgramCache) Bundle {
	b.programCache = cache
	return b
}

func (b Bundle) WithRenderCache(cache RenderCache) Bundle {
	b.renderCache = cache
	return b
}

func (b Bundle) Render(program *goja.Program, data any) (rendered string, err error) {
	return b.renderWithName(program, data, "program")
}

func (b Bundle) renderWithName(program *goja.Program, data any, name string) (rendered string, err error) {
	vm := goja.New()
	for _, loader := range b.modules {
		if err = loader.Load(vm); err != nil {
			err = fmt.Errorf("failed to load module: %w", err)
			return
		}
	}
	if err = b.engine.Load(vm); err != nil {
		err = fmt.Errorf("failed to load render engine: %w", err)
		return
	}
	var engineRender goja.Callable
	engineRender, err = b.engine.GetRenderFunction(vm)
	if err != nil {
		err = fmt.Errorf("failed to get render function from engine: %w", err)
		return
	}
	renderCalled := false
	buf := new(bytes.Buffer)
	var renderErr error
	if err = vm.Set(b.renderFunction, func(call goja.ConstructorCall) *goja.Object {
		renderCalled = true
		var renderResult goja.Value
		renderResult, renderErr = engineRender(goja.Undefined(), call.Argument(0))
		if renderErr == nil {
			buf.WriteString(renderResult.String())
		}
		return nil
	}); err != nil {
		return
	}
	if err = vm.Set(b.dataVariable, data); err != nil {
		err = fmt.Errorf("failed to set data variable %q: %w", b.dataVariable, err)
		return
	}
	if _, err = vm.RunProgram(program); err != nil {
		err = &RuntimeError{Name: name, Cause: err}
		return
	}

	if !renderCalled {
		err = &MissingRenderCallError{Name: name, RenderFunction: b.renderFunction}
		return
	}
	if renderErr != nil {
		err = &RuntimeError{Name: name, Message: "error in component render", Cause: renderErr}
		return
	}

	rendered = buf.String()
	return
}

func (b *Bundle) RenderJSX(name, jsx string, data any) (rendered string, err error) {
	var program *goja.Program
	var programHash uint64
	var dataHash uint64
	if b.programCache != nil {
		programHash = hash(jsx)
		program = b.programCache.GetProgram(programHash)

		if b.renderCache != nil {
			if data == nil {
				dataHash = hash("")
			} else if hashableData, ok := data.(Hashable); ok {
				dataHash = hashableData.Hash64()
			}
			if dataHash != 0 {
				var ok bool
				if rendered, ok = b.renderCache.GetRender(programHash, dataHash); ok {
					return
				}
			}
		}
	}

	if program == nil {
		program, err = CompileJSX(name, jsx, b.engine.GetJSXOptions())
		if err != nil {
			// TransformError already has good context, don't wrap it
			return
		}
	}

	if b.programCache != nil {
		b.programCache.AddProgram(programHash, program)
	}

	rendered, err = b.renderWithName(program, data, name)
	if err != nil {
		return
	}

	if b.renderCache != nil && dataHash != 0 {
		b.renderCache.AddRender(programHash, dataHash, rendered)
	}
	return
}
