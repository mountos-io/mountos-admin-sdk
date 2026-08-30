// Fixture/mock-server contract test for the block HA-pair placement admin
// surface (admin-sdk.md §5 step 2): exercises the generated client against
// an httptest.Server, no live appserv. Covers the "accepted, not completed"
// response shapes (drainPair/cancelDrain/updateConfig) and regression-guards
// the reactivateMember GET-vs-POST generator bug (a no-request endpoint with
// a named responseType on a mutating method was silently generated as GET
// in all three SDK languages until fixed alongside this test).
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

func TestGetConfigReadsBackK(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/config": map[string]any{"id": "s1", "k": 3, "algorithmVersion": 1, "epochPolicyVersion": 1},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	cfg, err := c.Storages.GetConfig(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.K != 3 {
		t.Fatalf("expected k=3, got %+v", cfg)
	}
}

func TestListPairs(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/pairs": []map[string]any{
			{"id": "p1", "storageId": "s1", "state": "active", "memberA": "bv1", "memberB": "bv2"},
			{"id": "p2", "storageId": "s1", "state": "draining", "memberA": "bv3", "memberB": "bv4", "pendingSyncJobsA": 3, "pendingSyncJobsB": 0},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	pairs, err := c.Storages.ListPairs(context.Background(), 7, "", false)
	if err != nil {
		t.Fatalf("ListPairs: %v", err)
	}
	if len(pairs) != 2 || pairs[1].State != "draining" || pairs[1].PendingSyncJobsA == nil || *pairs[1].PendingSyncJobsA != 3 {
		t.Fatalf("unexpected pairs: %+v", pairs)
	}
}

// TestPairPendingSyncJobsNullVsZero proves R3-010's fix: PendingSyncJobsA/B
// is a *int32, so a wire "null" (not yet observed) and a wire "0" (confirmed
// zero backlog) decode to distinct Go values, a nil pointer versus a pointer
// to zero, instead of collapsing onto the same int32 zero value.
func TestPairPendingSyncJobsNullVsZero(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/pairs": []map[string]any{
			{"id": "p1", "storageId": "s1", "state": "active", "memberA": "bv1", "memberB": "bv2", "pendingSyncJobsA": nil, "pendingSyncJobsB": 0},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	pairs, err := c.Storages.ListPairs(context.Background(), 7, "", false)
	if err != nil {
		t.Fatalf("ListPairs: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %+v", pairs)
	}
	if pairs[0].PendingSyncJobsA != nil {
		t.Fatalf("expected PendingSyncJobsA nil (not observed), got %v", *pairs[0].PendingSyncJobsA)
	}
	if pairs[0].PendingSyncJobsB == nil || *pairs[0].PendingSyncJobsB != 0 {
		t.Fatalf("expected PendingSyncJobsB non-nil zero (confirmed empty), got %+v", pairs[0].PendingSyncJobsB)
	}
}

func TestDrainPairIdempotentAck(t *testing.T) {
	// D9: response reads "draining", not "drained" - an accepted-transition
	// ack, never a completion promise.
	_, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/pairs/p1/drain": map[string]any{"id": "p1", "state": "draining"},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.DrainPair(context.Background(), 7, "p1")
	if err != nil {
		t.Fatalf("DrainPair: %v", err)
	}
	if res.State != "draining" {
		t.Fatalf("expected state=draining, got %q", res.State)
	}
}

func TestCancelDrain(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/pairs/p1/cancel-drain": map[string]any{"id": "p1", "state": "active"},
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

func TestUpdateConfigPartialSurfacesReason(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"PUT /api/v1/storages/7/config": map[string]any{
			"id": "s1", "targetK": 3, "activePairCountBefore": 1, "pairsNeeded": 2,
			"pairsFormed": 1, "activePairCountAfter": 2, "partial": true,
			"reason": "placement cluster B has no unused members",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.UpdateConfig(context.Background(), 7, &sdk.UpdateStorageConfigRequest{K: 3})
	if err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
	if !res.Partial || res.Reason == "" {
		t.Fatalf("expected partial=true with a reason, got %+v", res)
	}
	if res.TargetK != 3 || res.ActivePairCountBefore != 1 || res.PairsNeeded != 2 || res.PairsFormed != 1 || res.ActivePairCountAfter != 2 {
		t.Fatalf("unexpected explicit before/needed/formed/after fields: %+v", res)
	}
}

// Regression guard: reactivateMember is a mutating (POST) action with a
// named responseType and no request body - a generator dispatch bug made
// this silently emit a GET in all three SDK languages (fixed alongside
// this test). Asserting the recorded method here, not just that the call
// succeeds, is the point.
func TestReactivateMemberSendsPost(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/members/bv1/reactivate": map[string]any{
			"id": "bv1", "name": "originator", "regionId": 1, "regionClusterId": 2, "memberState": "active",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.ReactivateMember(context.Background(), 7, "bv1")
	if err != nil {
		t.Fatalf("ReactivateMember: %v", err)
	}
	if res.MemberState != "active" {
		t.Fatalf("unexpected member: %+v", res)
	}
	if len(fs.calls) != 1 || fs.calls[0].method != http.MethodPost {
		t.Fatalf("expected exactly one POST, got calls: %+v", fs.calls)
	}
}

func TestRegisterMemberRequiredName(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"POST /api/v1/storages/7/members": map[string]any{
			"id": "bv5", "name": "new-member", "regionId": 1, "regionClusterId": 3, "memberState": "active",
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Storages.RegisterMember(context.Background(), 7, &sdk.RegisterStorageMemberRequest{RegionClusterID: 3, Name: "new-member"})
	if err != nil {
		t.Fatalf("RegisterMember: %v", err)
	}
	if res.MemberState != "active" {
		t.Fatalf("unexpected member: %+v", res)
	}
	if got := fs.calls[0].body["name"]; got != "new-member" {
		t.Fatalf("expected name in request body, got %+v", fs.calls[0].body)
	}
}
