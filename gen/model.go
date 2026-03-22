package main

import "strings"

type Spec struct {
	Version    string              `yaml:"version"`
	BasePath   string              `yaml:"basePath"`
	JWT        JWTSpec             `yaml:"jwt"`
	ErrorCodes []ErrorCodeDef      `yaml:"errorCodes"`
	Enums      map[string][]string `yaml:"enums"`
	Types      map[string][]string `yaml:"types"`
	Resources  []Resource          `yaml:"resources"`
}

type JWTSpec struct {
	Algorithm string            `yaml:"algorithm"`
	Subject   string            `yaml:"subject"`
	Audience  string            `yaml:"audience"`
	Scope     string            `yaml:"scope"`
	KeyFormat string            `yaml:"keyFormat"`
	Claims    map[string]string `yaml:"claims"`
}

type ErrorCodeDef struct {
	Code int    `yaml:"code"`
	Name string `yaml:"name"`
}

type Resource struct {
	Name           string            `yaml:"name"`
	Path           string            `yaml:"path"`
	PathParamTypes map[string]string `yaml:"pathParamTypes,omitempty"`
	Endpoints      []Endpoint        `yaml:"endpoints"`
}

type Endpoint struct {
	Action        string   `yaml:"action"`
	Method        string   `yaml:"method"`
	Path          string   `yaml:"path"`
	Stub          bool     `yaml:"stub,omitempty"`
	Pagination    string   `yaml:"pagination,omitempty"`
	Query         []string `yaml:"query,omitempty"`
	Request       []string `yaml:"request,omitempty"`
	Response      []string `yaml:"response,omitempty"`
	ResponseType  string   `yaml:"responseType,omitempty"`
	ResponseArray bool     `yaml:"responseArray,omitempty"`
}

type Field struct {
	Name     string
	Type     string
	Required bool
	Optional bool
	Default  string
}

func parseField(s string) Field {
	// "name: type!", "name: type?", "name: type=X", "name: type"
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		fatalf("invalid field notation %q: expected 'name: type'", s)
	}
	name := strings.TrimSpace(parts[0])
	rest := strings.TrimSpace(parts[1])

	f := Field{Name: name}
	if i := strings.Index(rest, "="); i >= 0 {
		f.Default = rest[i+1:]
		f.Type = rest[:i]
		return f
	}
	if strings.HasSuffix(rest, "!") {
		f.Required = true
		f.Type = rest[:len(rest)-1]
		return f
	}
	if strings.HasSuffix(rest, "?") {
		f.Optional = true
		f.Type = rest[:len(rest)-1]
		return f
	}
	f.Type = rest
	return f
}
