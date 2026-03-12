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
	browserOut := flag.String("browser-client-out", "", "Browser client output directory (optional)")
	docOut := flag.String("doc-out", ".", "api.md output directory")
	flag.Parse()

	spec := loadSpec(*specPath)
	generateGo(spec, *goOut)
	generateTS(spec, *tsOut)
	generateBrowserClient(spec, *browserOut)
	generateDoc(spec, *docOut)
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
