// Regression coverage for the generator emitting url.PathEscape on every
// string :path param (not just the one endpoint that originally needed it):
// a raw slash, space, or non-ASCII character in a string id must reach the
// wire as one escaped path segment, not split into extra segments or sent
// unescaped. r.URL.Path (net/http's own decoded view) cannot tell the two
// cases apart - a literal '/' used as a real separator and an escaped
// "%2F" both decode back to the same string - so this asserts on
// r.RequestURI, the literal bytes that were sent on the wire.
package sdk_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStringPathParamsAreURLEncoded(t *testing.T) {
	const rawCopysetID = "copyset/id with spaces/café"
	const wantEscaped = "copyset%2Fid%20with%20spaces%2Fcaf%C3%A9"

	var requestURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","message":"ok","data":{"id":"` + rawCopysetID + `","storageId":"s1","state":"active"}}`))
	}))
	defer srv.Close()
	client := newTestClient(t, srv.URL)

	copyset, err := client.Storages.GetCopysetStatus(context.Background(), 7, rawCopysetID)
	if err != nil {
		t.Fatalf("GetCopysetStatus: %v", err)
	}
	if copyset.ID != rawCopysetID {
		t.Fatalf("got copyset id %q, want %q", copyset.ID, rawCopysetID)
	}
	if !strings.Contains(requestURI, wantEscaped) {
		t.Fatalf("request URI %q does not contain escaped segment %q", requestURI, wantEscaped)
	}
	if strings.Contains(requestURI, "/copysets/copyset/") {
		t.Fatalf("request URI %q sent the '/' unescaped, splitting the path", requestURI)
	}
}
