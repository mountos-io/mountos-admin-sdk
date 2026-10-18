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
	Claims    map[string]string `yaml:"claims"`
	Algorithm string            `yaml:"algorithm"`
	Subject   string            `yaml:"subject"`
	Audience  string            `yaml:"audience"`
	Scope     string            `yaml:"scope"`
	KeyFormat string            `yaml:"keyFormat"`
}

type ErrorCodeDef struct {
	Name string `yaml:"name"`
	Code int    `yaml:"code"`
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
	Pagination    string   `yaml:"pagination,omitempty"`
	ResponseType  string   `yaml:"responseType,omitempty"`
	Query         []string `yaml:"query,omitempty"`
	Request       []string `yaml:"request,omitempty"`
	Response      []string `yaml:"response,omitempty"`
	Stub          bool     `yaml:"stub,omitempty"`
	ResponseArray bool     `yaml:"responseArray,omitempty"`
}

// cursorResponseType names the wrapper type of a cursor-paginated endpoint.
// A "cursor" query field typed string selects the opaque string cursor
// wrapper. Any other cursor type selects the int64 wrapper.
func cursorResponseType(ep Endpoint) string {
	for _, qs := range ep.Query {
		if f := parseField(qs); f.Name == "cursor" && f.Type == "string" {
			return "StringCursorPaginatedResponse"
		}
	}
	return "CursorPaginatedResponse"
}

type Field struct {
	Name     string
	Type     string
	Default  string
	Required bool
	Optional bool
}

// splitLiteralQuery separates a literal "?a=b&c=d" suffix baked into a spec
// path from the bare path, returning the path and the literal key/value pairs.
// Generators fold these pairs into their query builders so a dynamically
// appended query string uses a single, correctly-separated "?".
func splitLiteralQuery(path string) (string, [][2]string) {
	i := strings.IndexByte(path, '?')
	if i < 0 {
		return path, nil
	}
	var pairs [][2]string
	for _, kv := range strings.Split(path[i+1:], "&") {
		if kv == "" {
			continue
		}
		k, v, _ := strings.Cut(kv, "=")
		pairs = append(pairs, [2]string{k, v})
	}
	return path[:i], pairs
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
