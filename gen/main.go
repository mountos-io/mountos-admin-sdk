package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	specPath := flag.String("spec", "api.yaml", "path to API spec")
	goOut := flag.String("go-out", "go", "Go output directory")
	tsOut := flag.String("ts-out", "ts/src", "TS output directory")
	rustOut := flag.String("rust-out", "rust/src", "Rust output directory")
	docOut := flag.String("doc-out", ".", "api.md output directory")
	docsOut := flag.String("docs-out", "docs", "language-specific docs output directory (ts.md, go.md, rust.md)")
	flag.Parse()

	spec := loadSpec(*specPath)
	validateSpec(spec)
	generateGo(spec, *goOut)
	generateTS(spec, *tsOut)
	generateRust(spec, *rustOut)
	generateDoc(spec, *docOut)
	generateDocTS(spec, *docsOut)
	generateDocGo(spec, *docsOut)
	generateDocRust(spec, *docsOut)
	fmt.Println("generation complete")
}

func loadSpec(path string) *Spec {
	data, err := os.ReadFile(path)
	if err != nil {
		fatalf("read spec: %v", err)
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		fatalf("parse spec: %v", err)
	}
	return &spec
}

// validateSpec catches endpoint shapes the generators don't fully cover.
// method: QUERY carries its parameters in a JSON request body instead of the
// URL (see docs/design/query-verb.md) -- that only makes sense when the
// endpoint actually declares fields to put there. A QUERY endpoint with
// neither request nor query fields is a plain GET; the Rust and Go toggle/
// void writers don't plumb a body for that shape (Rust would silently mis-map
// to POST, Go would fail to compile), so reject it here instead.
func validateSpec(spec *Spec) {
	for _, res := range spec.Resources {
		for _, ep := range res.Endpoints {
			if ep.Method != "QUERY" {
				continue
			}
			if len(ep.Request) == 0 && len(ep.Query) == 0 {
				fatalf("%s.%s: method QUERY requires request or query fields (use GET for a parameterless read)", res.Name, ep.Action)
			}
		}
	}
}

func writeFile(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
