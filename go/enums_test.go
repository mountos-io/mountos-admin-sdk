// Regression coverage for the generated closed string-enum types (gen/gogen.go
// writeGoEnum): each must accept every one of its own defined wire values via
// IsValid/Parse<Name>, and reject an unrecognized string rather than silently
// treating it as a legitimate value.
package sdk_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/mountos-io/mountos-admin-sdk/go"
)

func TestPairState_IsValid(t *testing.T) {
	for _, v := range []sdk.PairState{
		sdk.PairStateActive, sdk.PairStateDraining, sdk.PairStateSyncedDrained, sdk.PairStateRetired,
	} {
		if !v.IsValid() {
			t.Errorf("PairState(%q).IsValid() = false, want true", v)
		}
	}
	if sdk.PairState("bogus").IsValid() {
		t.Error(`PairState("bogus").IsValid() = true, want false`)
	}
}

func TestParsePairState(t *testing.T) {
	for _, want := range []sdk.PairState{
		sdk.PairStateActive, sdk.PairStateDraining, sdk.PairStateSyncedDrained, sdk.PairStateRetired,
	} {
		got, err := sdk.ParsePairState(string(want))
		if err != nil {
			t.Errorf("ParsePairState(%q): unexpected error: %v", want, err)
		}
		if got != want {
			t.Errorf("ParsePairState(%q) = %q, want %q", want, got, want)
		}
	}
	if _, err := sdk.ParsePairState("bogus"); err == nil {
		t.Error(`ParsePairState("bogus") returned no error, want one`)
	}
}

func TestClientSessionStatus_IsValid(t *testing.T) {
	for _, v := range []sdk.ClientSessionStatus{
		sdk.ClientSessionStatusConnected, sdk.ClientSessionStatusActive, sdk.ClientSessionStatusDegraded,
		sdk.ClientSessionStatusDisconnected, sdk.ClientSessionStatusExpired, sdk.ClientSessionStatusUnknown,
	} {
		if !v.IsValid() {
			t.Errorf("ClientSessionStatus(%q).IsValid() = false, want true", v)
		}
	}
	if sdk.ClientSessionStatus("bogus").IsValid() {
		t.Error(`ClientSessionStatus("bogus").IsValid() = true, want false`)
	}
}

func TestParseClientSessionStatus(t *testing.T) {
	got, err := sdk.ParseClientSessionStatus("degraded")
	if err != nil || got != sdk.ClientSessionStatusDegraded {
		t.Errorf("ParseClientSessionStatus(%q) = (%q, %v), want (%q, nil)", "degraded", got, err, sdk.ClientSessionStatusDegraded)
	}
	if _, err := sdk.ParseClientSessionStatus("bogus"); err == nil {
		t.Error(`ParseClientSessionStatus("bogus") returned no error, want one`)
	}
}

func TestLicenseStatus_IsValid(t *testing.T) {
	for _, v := range []sdk.LicenseStatus{
		sdk.LicenseStatusValid, sdk.LicenseStatusExpiring, sdk.LicenseStatusGrace,
		sdk.LicenseStatusExpiredAccess, sdk.LicenseStatusExpired,
	} {
		if !v.IsValid() {
			t.Errorf("LicenseStatus(%q).IsValid() = false, want true", v)
		}
	}
	if sdk.LicenseStatus("bogus").IsValid() {
		t.Error(`LicenseStatus("bogus").IsValid() = true, want false`)
	}
	if _, err := sdk.ParseLicenseStatus("bogus"); err == nil {
		t.Error(`ParseLicenseStatus("bogus") returned no error, want one`)
	}
}

func TestLicenseQuotaState_IsValid(t *testing.T) {
	for _, v := range []sdk.LicenseQuotaState{sdk.LicenseQuotaStateOk, sdk.LicenseQuotaStateExceeded} {
		if !v.IsValid() {
			t.Errorf("LicenseQuotaState(%q).IsValid() = false, want true", v)
		}
	}
	if sdk.LicenseQuotaState("bogus").IsValid() {
		t.Error(`LicenseQuotaState("bogus").IsValid() = true, want false`)
	}
	if _, err := sdk.ParseLicenseQuotaState("bogus"); err == nil {
		t.Error(`ParseLicenseQuotaState("bogus") returned no error, want one`)
	}
}

// TestPairState_String proves the enum's String method returns the wire
// value, matching the pre-existing print/format expectations a plain string
// alias used to satisfy implicitly.
func TestPairState_String(t *testing.T) {
	if got := sdk.PairStateDraining.String(); got != "draining" {
		t.Errorf("PairStateDraining.String() = %q, want %q", got, "draining")
	}
}

// TestPairState_JSONRoundTrip proves the defined type still marshals and
// unmarshals as a plain JSON string, unaffected by the switch away from a
// bare string alias.
func TestPairState_JSONRoundTrip(t *testing.T) {
	type wrapper struct {
		State sdk.PairState `json:"state"`
	}
	data := []byte(`{"state":"synced_drained"}`)

	var got wrapper
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.State != sdk.PairStateSyncedDrained {
		t.Fatalf("unmarshalled state = %q, want %q", got.State, sdk.PairStateSyncedDrained)
	}

	out, err := json.Marshal(wrapper{State: sdk.PairStateSyncedDrained})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"state":"synced_drained"}` {
		t.Fatalf("marshalled = %s, want %s", out, `{"state":"synced_drained"}`)
	}
}
