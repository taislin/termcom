package language

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

func TestLanguageKeyConsistency(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	langKeys := make(map[string]map[string]bool)
	fset := token.NewFileSet()

	for _, file := range files {
		if file == "language_test.go" || file == "language.go" {
			continue
		}

		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", file, err)
			continue
		}

		keys := make(map[string]bool)
		ast.Inspect(f, func(n ast.Node) bool {
			if kv, ok := n.(*ast.KeyValueExpr); ok {
				if lit, ok := kv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					keys[lit.Value] = true
				}
			}
			return true
		})
		langKeys[file] = keys
	}

	// Compare en.go against all other files
	enKeys := langKeys["en.go"]
	for file, keys := range langKeys {
		if file == "en.go" {
			continue
		}
		for key := range enKeys {
			if !keys[key] {
				t.Errorf("%s is missing key: %s", file, key)
			}
		}
		for key := range keys {
			if !enKeys[key] {
				t.Errorf("%s has extra key: %s", file, key)
			}
		}
	}
}
