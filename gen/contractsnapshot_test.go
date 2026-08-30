package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateContractSnapshot proves the snapshot faithfully reflects
// parseField's required/optional/type computation for a small hand-built
// Spec, and that enum values pass through unchanged.
func TestGenerateContractSnapshot(t *testing.T) {
	spec := &Spec{
		Version: "1",
		Types: map[string][]string{
			"UpdateConfigResult": {
				"id: string",
				"activePairCountAfter: int32!",
				"note: string?",
			},
		},
		Enums: map[string][]string{
			"PairState": {"active", "draining", "synced_drained", "retired"},
		},
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "contract-snapshot.json")
	if err := generateContractSnapshot(spec, outPath); err != nil {
		t.Fatalf("generateContractSnapshot: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	var got contractSnapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	if got.Version != "1" {
		t.Errorf("version = %q, want %q", got.Version, "1")
	}

	ty, ok := got.Types["UpdateConfigResult"]
	if !ok {
		t.Fatalf("missing type UpdateConfigResult in snapshot")
	}

	wantFields := map[string]snapshotField{
		"id":                   {Type: "string", Required: false, Optional: false},
		"activePairCountAfter": {Type: "int32", Required: true, Optional: false},
		"note":                 {Type: "string", Required: false, Optional: true},
	}
	if len(ty.Fields) != len(wantFields) {
		t.Fatalf("field count = %d, want %d (%v)", len(ty.Fields), len(wantFields), ty.Fields)
	}
	for name, want := range wantFields {
		got, ok := ty.Fields[name]
		if !ok {
			t.Errorf("missing field %q", name)
			continue
		}
		if got != want {
			t.Errorf("field %q = %+v, want %+v", name, got, want)
		}
	}

	wantEnum := []string{"active", "draining", "synced_drained", "retired"}
	gotEnum, ok := got.Enums["PairState"]
	if !ok {
		t.Fatalf("missing enum PairState in snapshot")
	}
	if len(gotEnum) != len(wantEnum) {
		t.Fatalf("enum values = %v, want %v", gotEnum, wantEnum)
	}
	for i, v := range wantEnum {
		if gotEnum[i] != v {
			t.Errorf("enum[%d] = %q, want %q", i, gotEnum[i], v)
		}
	}
}
