package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// guardAllowlist names exported structs that are exempt from the keyed-literal
// guard, each with the reason.
var guardAllowlist = map[string]string{
	"AzureCredentialFields": "documentation anchor with no fields",
}

// TestGoStructsRejectPositionalLiterals fails when an exported struct in the
// Go SDK can be built by a positional literal outside the package but has no
// leading blank `_ struct{}` field. Without the guard such a literal compiles
// today and breaks when a field is added. A struct with any unexported field
// (Client, the services) cannot be built positionally outside the package and
// needs no guard.
func TestGoStructsRejectPositionalLiterals(t *testing.T) {
	files, err := filepath.Glob("../go/*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob ../go/*.go: %v (%d files)", err, len(files))
	}
	fset := token.NewFileSet()
	checked := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				st, ok := ts.Type.(*ast.StructType)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				if _, skip := guardAllowlist[ts.Name.Name]; skip {
					continue
				}
				if hasUnexportedField(st) {
					continue
				}
				checked++
				if !hasLeadingGuard(st) {
					t.Errorf("%s: exported struct %s lacks a leading `_ struct{}` guard", filepath.Base(path), ts.Name.Name)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no exported structs checked; the scan is broken")
	}
}

func hasUnexportedField(st *ast.StructType) bool {
	for _, fld := range st.Fields.List {
		for _, n := range fld.Names {
			if n.Name != "_" && !n.IsExported() {
				return true
			}
		}
	}
	return false
}

func hasLeadingGuard(st *ast.StructType) bool {
	if len(st.Fields.List) == 0 {
		return false
	}
	first := st.Fields.List[0]
	if len(first.Names) != 1 || first.Names[0].Name != "_" {
		return false
	}
	lit, ok := first.Type.(*ast.StructType)
	return ok && len(lit.Fields.List) == 0
}
