package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDispatchMatrix protects writeGoMethod/writeTSMethod/writeRustMethod's
// endpoint-shape switch (gogen.go, tsgen.go, rustgen.go) across every
// combination of HTTP method, request body presence, and response shape it
// dispatches on. A no-request endpoint with a named responseType or an
// inline response on a mutating method (PUT/POST/DELETE) must emit that
// endpoint's own HTTP method, not silently fall back to GET or a wrong verb.
func TestDispatchMatrix(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	shapes := []string{"none", "inline", "named"}

	var endpoints []Endpoint
	for _, m := range methods {
		for _, shape := range shapes {
			endpoints = append(endpoints, matrixEndpoint(m, shape, false))
		}
	}
	// Request-body variants: the higher-priority branch in every generator's
	// switch, unaffected by the no-body dispatch fix, kept here so the
	// matrix documents (and guards) the full decision table in one place.
	endpoints = append(endpoints,
		matrixEndpoint("POST", "inline", true),
		matrixEndpoint("PUT", "named", true),
	)

	spec := &Spec{
		BasePath: "/api/v1",
		Types: map[string][]string{
			"Widget": {"id: string"},
		},
		Resources: []Resource{{
			Name:           "Widgets",
			Path:           "/widgets",
			PathParamTypes: map[string]string{"widgetId": "string"},
			Endpoints:      endpoints,
		}},
	}

	dir := t.TempDir()
	goOut := filepath.Join(dir, "go")
	tsOut := filepath.Join(dir, "ts")
	rustOut := filepath.Join(dir, "rust")

	generateGo(spec, goOut)
	generateTS(spec, tsOut)
	generateRust(spec, rustOut)

	goSrc := mustReadFile(t, filepath.Join(goOut, "resources_gen.go"))
	tsSrc := mustReadFile(t, filepath.Join(tsOut, "client_gen.ts"))
	rustSrc := mustReadFile(t, filepath.Join(rustOut, "client_gen.rs"))

	for _, ep := range endpoints {
		ep := ep
		t.Run(ep.Action, func(t *testing.T) {
			checkGoDispatch(t, goSrc, ep)
			checkTSDispatch(t, tsSrc, ep)
			checkRustDispatch(t, rustSrc, ep)
		})
	}
}

// matrixEndpoint builds one synthetic endpoint for (method, response shape),
// optionally carrying a request body. Action names encode the combination
// so a failing subtest names its own case.
func matrixEndpoint(method, shape string, withBody bool) Endpoint {
	action := strings.ToLower(method) + pascalCase(shape)
	if withBody {
		action += "Body"
	}
	ep := Endpoint{
		Action: action,
		Method: method,
		Path:   "/:widgetId/" + strings.ToLower(action),
	}
	if withBody {
		ep.Request = []string{"name: string!"}
	}
	switch shape {
	case "inline":
		ep.Response = []string{"ok: bool"}
	case "named":
		ep.ResponseType = "Widget"
	}
	return ep
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// sliceBlock returns the text from the first occurrence of start through
// the following occurrence of end (inclusive), or fails the test.
func sliceBlock(t *testing.T, src, start, end string) string {
	t.Helper()
	i := strings.Index(src, start)
	if i < 0 {
		t.Fatalf("marker %q not found in generated source", start)
	}
	rest := src[i:]
	j := strings.Index(rest, end)
	if j < 0 {
		t.Fatalf("terminator %q not found after %q", end, start)
	}
	return rest[:j+len(end)]
}

func goVerbs() map[string]string {
	return map[string]string{"GET": "s.c.get(", "POST": "s.c.post(", "PUT": "s.c.put(", "DELETE": "s.c.delete("}
}

func checkGoDispatch(t *testing.T, src string, ep Endpoint) {
	t.Helper()
	name := pascalCase(ep.Action)
	block := sliceBlock(t, src, "func (s *WidgetsService) "+name+"(", "\n}\n")
	verbs := goVerbs()
	want, ok := verbs[ep.Method]
	if !ok {
		t.Fatalf("no expected Go verb for method %s", ep.Method)
	}
	if !strings.Contains(block, want) {
		t.Errorf("Go %s: expected call %q, got:\n%s", name, want, block)
	}
	for m, v := range verbs {
		if m != ep.Method && strings.Contains(block, v) {
			t.Errorf("Go %s: unexpectedly called %s verb %q, got:\n%s", name, m, v, block)
		}
	}
}

func checkTSDispatch(t *testing.T, src string, ep Endpoint) {
	t.Helper()
	name := camelCase(ep.Action)
	block := sliceBlock(t, src, "  "+name+"(", "\n  }\n")
	want := "this.client.request('" + ep.Method + "'"
	if !strings.Contains(block, want) {
		t.Errorf("TS %s: expected call %q, got:\n%s", name, want, block)
	}
	for _, m := range []string{"GET", "POST", "PUT", "DELETE"} {
		if m != ep.Method && strings.Contains(block, "this.client.request('"+m+"'") {
			t.Errorf("TS %s: unexpectedly called %s, got:\n%s", name, m, block)
		}
	}
}

func checkRustDispatch(t *testing.T, src string, ep Endpoint) {
	t.Helper()
	name := rustFieldName(ep.Action)
	block := sliceBlock(t, src, "pub async fn "+name+"(", "\n    }\n")

	hasBody := len(ep.Request) > 0
	isVoid := !hasBody && len(ep.Response) == 0 && ep.ResponseType == ""

	verb := map[string]string{"GET": "get", "DELETE": "delete", "PUT": "put", "POST": "post"}[ep.Method]
	if !hasBody && (ep.Method == "POST" || ep.Method == "PUT") {
		verb += "_empty"
	}
	want := "self.inner." + verb
	if isVoid {
		want += "::<serde_json::Value>"
	} else {
		want += "("
	}
	if !strings.Contains(block, want) {
		t.Errorf("Rust %s: expected call %q, got:\n%s", name, want, block)
	}
}
