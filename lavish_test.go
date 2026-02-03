package lavish_test

import (
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quells/lavish"
	"github.com/stretchr/testify/require"
)

func TestCompileJSX(t *testing.T) {
	jsxOptions := api.TransformOptions{
		Loader:      api.LoaderJSX,
		JSXFactory:  "h",
		JSXFragment: "Fragment",
	}

	tests := []struct {
		name             string
		src              string
		expectErrContain []string // error must contain all these strings
	}{
		{
			name: "empty_string.jsx",
			src:  "",
		},
		{
			name: "syntax_error.jsx",
			src:  "let x = new Array(",
			expectErrContain: []string{
				"syntax error in syntax_error.jsx",
				"line 1:",
				"Unexpected end of file",
				"let x = new Array(", // shows source line
				"^",                  // shows caret
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lavish.CompileJSX(tt.name, tt.src, jsxOptions)
			if len(tt.expectErrContain) > 0 {
				require.Error(t, err, "expected an error")
				errStr := err.Error()
				for _, s := range tt.expectErrContain {
					require.True(t, strings.Contains(errStr, s),
						"error should contain %q, got:\n%s", s, errStr)
				}
			} else {
				require.NoError(t, err, "must compile jsx")
			}
		})
	}
}
