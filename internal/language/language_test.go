package language

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"testing"
)

func parseKeysForFile(t *testing.T, path string) map[string]bool {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", path, err)
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
	return keys
}

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

// TestReferencedKeysExist ensures every key referenced via language.String("KEY")
// (string-literal form) anywhere in the codebase is actually defined in en.go.
// This catches keys that are missing from ALL language files, which the parity
// check above cannot detect on its own.
func TestReferencedKeysExist(t *testing.T) {
	enKeys := parseKeysForFile(t, "en.go")

	root := "../.."
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	missing := make(map[string]bool)

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", path, err)
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "String" {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "language" {
				return true
			}
			if len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			if !enKeys[lit.Value] {
				missing[lit.Value] = true
			}
			return true
		})
	}

	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for k := range missing {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t.Errorf("key %s referenced via language.String but not defined in en.go", k)
		}
	}
}
