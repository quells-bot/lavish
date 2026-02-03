package lavish_test

import (
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
		name      string
		src       string
		expectErr string
	}{
		{
			name: "empty_string.jsx",
			src:  "",
		},
		{
			name:      "syntax_error.jsx",
			src:       "let x = new Array(",
			expectErr: "syntax error in syntax_error.jsx: line 1: Unexpected end of file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lavish.CompileJSX(tt.name, tt.src, jsxOptions)
			if tt.expectErr != "" {
				require.EqualError(t, err, tt.expectErr, "expected error message to match")
			} else {
				require.NoError(t, err, "must compile jsx")
			}
		})
	}
}
