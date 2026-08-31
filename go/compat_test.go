// Regression guard for VE-017 (Pair -> Copyset renamed in 1.14.0 with no
// aliases). Exercises compat.go: the old method/type names still resolve
// and hit the renamed Copyset routes.
package sdk_test

import (
	"context"
	"testing"

	sdk "github.com/mountos-io/mountos-admin-sdk/go"
)

func TestListPairsIsAnAliasForListCopysets(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/copysets": []map[string]any{
			{"id": "p1", "storageId": "s1", "state": "active", "memberA": "bv1", "memberB": "bv2"},
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	pairs, err := c.Storages.ListPairs(context.Background(), 7, "", false)
	if err != nil {
		t.Fatalf("ListPairs: %v", err)
	}
	if len(pairs) != 1 || pairs[0].ID != "p1" || pairs[0].State != sdk.PairStateActive {
		t.Fatalf("unexpected pairs: %+v", pairs)
	}
}

func TestGetPairStatusAndDrainPairAreAliases(t *testing.T) {
	_, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/storages/7/copysets/c1":        map[string]any{"id": "c1", "storageId": "s1", "state": "active"},
		"POST /api/v1/storages/7/copysets/c1/drain": map[string]any{"id": "c1", "state": "draining"},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	status, err := c.Storages.GetPairStatus(context.Background(), 7, "c1")
	if err != nil {
		t.Fatalf("GetPairStatus: %v", err)
	}
	if status.State != sdk.PairStateActive {
		t.Fatalf("unexpected status: %+v", status)
	}

	var drained *sdk.DrainPairStorageResponse
	drained, err = c.Storages.DrainPair(context.Background(), 7, "c1")
	if err != nil {
		t.Fatalf("DrainPair: %v", err)
	}
	if drained.State != "draining" {
		t.Fatalf("unexpected drain response: %+v", drained)
	}
}

func TestGetPairConfigAndUpdatePairConfigAreAliases(t *testing.T) {
	fs, srv := newFixtureServer(t, map[string]any{
		"GET /api/v1/volumes/5/copyset-config": map[string]any{"id": 5, "targetCopysetCount": 3, "currentEpoch": 1, "copysetIds": []string{"c1", "c2"}},
		"PUT /api/v1/volumes/5/copyset-config": map[string]any{
			"id": 5, "targetCopysetCount": 4, "copysetCountBefore": 3, "copysetsAdded": 1, "copysetsRemoved": 0,
			"copysetCountAfter": 4, "epoch": 2, "partial": false,
		},
	})
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	cfg, err := c.Volumes.GetPairConfig(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetPairConfig: %v", err)
	}
	if cfg.TargetCopysetCount != 3 || len(cfg.CopysetIDs) != 2 {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	resize, err := c.Volumes.UpdatePairConfig(context.Background(), 5, &sdk.UpdateVolumePairConfigRequest{TargetPairCount: 4})
	if err != nil {
		t.Fatalf("UpdatePairConfig: %v", err)
	}
	if resize.TargetCopysetCount != 4 || resize.CopysetsAdded != 1 {
		t.Fatalf("unexpected resize result: %+v", resize)
	}
	putCall := fs.calls[len(fs.calls)-1]
	if putCall.body["targetCopysetCount"] != float64(4) {
		t.Fatalf("expected targetPairCount to be translated to targetCopysetCount on the wire, got body %+v", putCall.body)
	}
}
