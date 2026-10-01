// Contract test for the fork tree listing cursor: it is an opaque string.
// The client must send the cursor back byte for byte and decode nextCursor as
// a string, with JSON null (final page) decoding to nil.
package sdk_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	sdk "github.com/mountos-io/mountos-admin-sdk/go"
)

func TestForkTreeListStringCursor(t *testing.T) {
	const opaque = "n1.dGVzdC1fbmFtZQ"

	var queries []url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		var next any
		if r.URL.Query().Get("cursor") == "" {
			next = opaque
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success", "message": "ok",
			"data": map[string]any{
				"items":      []map[string]any{{"name": "a", "kind": "file", "inode": 5, "size": 1, "mtime": 2, "ctime": 3}},
				"nextCursor": next,
			},
		})
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	first, err := c.VolumeForkTrees.List(context.Background(), 7, "main", &sdk.VolumeForkTreeListOptions{Path: "/"})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if first.NextCursor == nil || *first.NextCursor != opaque {
		t.Fatalf("page 1 nextCursor = %v, want %q", first.NextCursor, opaque)
	}
	if queries[0].Has("cursor") {
		t.Errorf("page 1 sent cursor %q, want none", queries[0].Get("cursor"))
	}

	last, err := c.VolumeForkTrees.List(context.Background(), 7, "main", &sdk.VolumeForkTreeListOptions{Path: "/", Cursor: *first.NextCursor})
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if got := queries[1].Get("cursor"); got != opaque {
		t.Errorf("page 2 cursor = %q, want %q", got, opaque)
	}
	if last.NextCursor != nil {
		t.Errorf("final page nextCursor = %q, want nil", *last.NextCursor)
	}
}
