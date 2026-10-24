package pkg_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-openapi/testify/require"
)

var forbiddenImports = []string{
	"github.com/adm87/finch/content",
	"github.com/adm87/finch/plugins",
}

// TestPkgForbiddenImports keeps the engine core free of game content and third-party tool integrations.
func TestPkgForbiddenImports(t *testing.T) {
	fset := token.NewFileSet()
	var violations []string

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			for _, b := range forbiddenImports {
				if importPath == b || strings.HasPrefix(importPath, b+"/") {
					violations = append(violations, "pkg/"+filepath.ToSlash(path)+" imports "+importPath)
				}
			}
		}
		return nil
	})

	require.NoError(t, err)
	require.Empty(t, violations)
}
