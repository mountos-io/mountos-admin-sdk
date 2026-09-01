// Fixture/mock-server contract test for the block copyset placement admin
// surface: exercises the generated client against an httptest.Server, no
// live appserv. Covers the "accepted, not completed" response shapes
// (drainCopyset/cancelDrain) and regression-guards the GET-vs-POST
// generator bug (a no-request endpoint with a named responseType on a
// mutating method was silently generated as GET in all three SDK languages
// until fixed) via TestAddCopysetMember, the surviving action with that
// same shape.
package sdk_test

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/mountos-io/mountos-admin-sdk/go"
)

// fixtureServer answers one canned envelope per "METHOD path" and records
// every call for assertion.
type fixtureServer struct {
	t         *testing.T
	responses map[string]any
	calls     []recordedCall
}

type recordedCall struct {
	method string
	path   string
	body   map[string]any
}

func newFixtureServer(t *testing.T, responses map[string]any) (*fixtureServer, *httptest.Server) {
	fs := &fixtureServer{t: t, responses: responses}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		var body map[string]any
		if r.ContentLength != 0 {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		fs.calls = append(fs.calls, recordedCall{method: r.Method, path: r.URL.Path, body: body})
		data, ok := responses[key]
		if !ok {
			t.Errorf("fixture has no response for %s", key)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "message": "ok", "data": data})
	}))
	return fs, srv
}

func newTestClient(t *testing.T, baseURL string) *sdk.Client {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	c, err := sdk.NewClient(sdk.Config{BaseURL: baseURL, PrivateKey: base64.StdEncoding.EncodeToString(priv.Seed())})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

// strPtr builds a *string for constructing optional request fields, which
// the generator pointer-wraps so callers can distinguish "not set" from an
// explicitly empty string.
func strPtr(s string) *string { return &s }

func TestListCopysets(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/copysets": []map[string]any{
			{"id": "p1", "storageId": "s1", "state": "active", "memberA": "bv1", "memberB": "bv2"},
			{"id": "p2", "storageId": "s1", "state": "draining", "memberA": "bv3", "memberB": "bv4", "pendingSyncJobsA": 3, "pendingSyncJobsB": 0},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	copysets, err := c.Storages.ListCopysets(context.Background(), 7, "", false)
	if err != nil {
		t.Fatalf("ListCopysets: %v", err)
	}
	if len(copysets) != 2 || copysets[1].State != "draining" || copysets[1].PendingSyncJobsA == nil || *copysets[1].PendingSyncJobsA != 3 {
		t.Fatalf("unexpected copysets: %+v", copysets)
	}
}

// TestCopysetPendingSyncJobsNullVsZero proves R3-010's fix: PendingSyncJobsA/B
// is a *int32, so a wire "null" (not yet observed) and a wire "0" (confirmed
// zero backlog) decode to distinct Go values, a nil pointer versus a pointer
// to zero, instead of collapsing onto the same int32 zero value.
func TestCopysetPendingSyncJobsNullVsZero(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/copysets": []map[string]any{
			{"id": "p1", "storageId": "s1", "state": "active", "memberA": "bv1", "memberB": "bv2", "pendingSyncJobsA": nil, "pendingSyncJobsB": 0},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	copysets, err := c.Storages.ListCopysets(context.Background(), 7, "", false)
	if err != nil {
		t.Fatalf("ListCopysets: %v", err)
	}
	if len(copysets) != 1 {
		t.Fatalf("expected 1 copyset, got %+v", copysets)
	}
	if copysets[0].PendingSyncJobsA != nil {
		t.Fatalf("expected PendingSyncJobsA nil (not observed), got %v", *copysets[0].PendingSyncJobsA)
	}
	if copysets[0].PendingSyncJobsB == nil || *copysets[0].PendingSyncJobsB != 0 {
		t.Fatalf("expected PendingSyncJobsB non-nil zero (confirmed empty), got %+v", copysets[0].PendingSyncJobsB)
	}
}

func TestDrainCopysetIdempotentAck(t *testing.T) {
	// D9: response reads "draining", not "drained" - an accepted-transition
	// ack, never a completion promise.
	_, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets/p1/drain": map[string]any{"id": "p1", "state": "draining"},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.DrainCopyset(context.Background(), 7, "p1")
	if err != nil {
		t.Fatalf("DrainCopyset: %v", err)
	}
	if res.State != "draining" {
		t.Fatalf("expected state=draining, got %q", res.State)
	}
}

func TestCancelDrain(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets/p1/cancel-drain": map[string]any{"id": "p1", "state": "active"},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.CancelDrain(context.Background(), 7, "p1")
	if err != nil {
		t.Fatalf("CancelDrain: %v", err)
	}
	if res.State != "active" {
		t.Fatalf("expected state=active, got %q", res.State)
	}
}

func TestRegisterCopysetExplicitName(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets": map[string]any{
			"id": "p5", "storageId": "s1", "name": "mos-block-a", "state": "active", "memberA": "bv5", "memberB": "bv6",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.RegisterCopyset(context.Background(), 7, &sdk.RegisterStorageCopysetRequest{Name: strPtr("mos-block-a")})
	if err != nil {
		t.Fatalf("RegisterCopyset: %v", err)
	}
	if res.Name != "mos-block-a" || res.MemberA == nil || *res.MemberA != "bv5" || res.MemberB == nil || *res.MemberB != "bv6" {
		t.Fatalf("unexpected copyset: %+v", res)
	}
	if got := fs.calls[0].body["name"]; got != "mos-block-a" {
		t.Fatalf("expected name in request body, got %+v", fs.calls[0].body)
	}
}

// TestRegisterCopysetOmittedName confirms name is optional on the wire: an
// omitted Name never sends its key at all (omitempty), letting the server
// auto-fill it - and both members derive from whatever it picks.
func TestRegisterCopysetOmittedName(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets": map[string]any{
			"id": "p6", "storageId": "s1", "name": "riveted-truss-4f2a", "state": "active", "memberA": "bv7", "memberB": "bv8",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	_, err := c.Storages.RegisterCopyset(context.Background(), 7, &sdk.RegisterStorageCopysetRequest{})
	if err != nil {
		t.Fatalf("RegisterCopyset: %v", err)
	}
	if _, ok := fs.calls[0].body["name"]; ok {
		t.Fatalf("expected no name key in request body when omitted, got %+v", fs.calls[0].body)
	}
}

// TestRegisterCopysetsBulk covers the count-only bulk path: no explicit
// names, every copyset in the batch auto-generated server-side.
func TestRegisterCopysetsBulk(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets/bulk": map[string]any{
			"copysets": []map[string]any{
				{"id": "p10", "storageId": "s1", "name": "riveted-truss-1a2b", "state": "active", "memberA": "bv10", "memberB": "bv11"},
				{"id": "p11", "storageId": "s1", "name": "coupled-beam-3c4d", "state": "active", "memberA": "bv12", "memberB": "bv13"},
			},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.RegisterCopysetsBulk(context.Background(), 7, &sdk.RegisterStorageCopysetsBulkRequest{Count: 2})
	if err != nil {
		t.Fatalf("RegisterCopysetsBulk: %v", err)
	}
	if len(res.Copysets) != 2 || res.Copysets[0].Name != "riveted-truss-1a2b" || res.Copysets[1].Name != "coupled-beam-3c4d" {
		t.Fatalf("unexpected bulk result: %+v", res)
	}
	if got := fs.calls[0].body["count"]; got != float64(2) {
		t.Fatalf("expected count=2 in request body, got %+v", fs.calls[0].body)
	}
}

// TestAddCopysetMember covers the narrow vacant-slot replacement path,
// distinct from registerCopyset's atomic two-member creation. Takes no
// request body - the new member's name is always derived server-side from
// the copyset's own name. Also the regression guard for the GET-vs-POST
// generator bug this file's own doc comment describes: asserting the
// recorded method, not just that the call succeeds, is the point.
func TestAddCopysetMember(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/copysets/p1/members": map[string]any{
			"id": "bv9", "name": "mos-block-a-b", "regionId": 1, "memberState": "active",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.AddCopysetMember(context.Background(), 7, "p1")
	if err != nil {
		t.Fatalf("AddCopysetMember: %v", err)
	}
	if res.Name != "mos-block-a-b" {
		t.Fatalf("unexpected member: %+v", res)
	}
	if len(fs.calls) != 1 || fs.calls[0].method != http.MethodPost {
		t.Fatalf("expected exactly one POST, got calls: %+v", fs.calls)
	}
	if fs.calls[0].body != nil {
		t.Fatalf("expected no request body, got %+v", fs.calls[0].body)
	}
}
