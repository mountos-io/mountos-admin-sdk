//! Regression coverage for the generated closed wire-enum types (gen/rustgen.go
//! writeRustEnum): a known value must deserialize to its own variant, an
//! unrecognized value must deserialize to the fallback variant carrying the
//! original string rather than a hard parse failure, and the round trip back
//! to JSON must reproduce the original wire string.

use mountos_admin_sdk::{ClientSessionStatus, PairState};

#[test]
fn pair_state_deserializes_known_values() {
    for (wire, want) in [
        ("active", PairState::Active),
        ("draining", PairState::Draining),
        ("synced_drained", PairState::SyncedDrained),
        ("retired", PairState::Retired),
    ] {
        let json = format!("\"{wire}\"");
        let got: PairState = serde_json::from_str(&json).expect("deserialize known value");
        assert_eq!(got, want, "wire value {wire:?}");
        assert!(got.is_known());
        assert_eq!(got.as_str(), wire);
    }
}

#[test]
fn pair_state_deserializes_unknown_value_to_fallback() {
    let got: PairState = serde_json::from_str("\"future_state\"").expect("deserialize unknown value");
    assert_eq!(got, PairState::Unknown("future_state".to_string()));
    assert!(!got.is_known());
    assert_eq!(got.as_str(), "future_state");
}

#[test]
fn pair_state_serialize_round_trips_known_and_unknown() {
    let known = PairState::Draining;
    assert_eq!(serde_json::to_string(&known).unwrap(), "\"draining\"");

    let unknown = PairState::Unknown("future_state".to_string());
    assert_eq!(serde_json::to_string(&unknown).unwrap(), "\"future_state\"");
}

#[test]
fn pair_state_display_matches_as_str() {
    assert_eq!(PairState::Retired.to_string(), "retired");
    assert_eq!(PairState::Unknown("x".to_string()).to_string(), "x");
}

// ClientSessionStatus has a legitimate wire value literally named "unknown"
// (see api.yaml), which is exactly why the generator's fallback variant is
// named UnknownValue instead of Unknown for this enum - this test guards
// that collision-avoidance, not just the general enum shape covered above.
#[test]
fn client_session_status_unknown_value_is_the_declared_variant_not_the_fallback() {
    let got: ClientSessionStatus = serde_json::from_str("\"unknown\"").expect("deserialize");
    assert_eq!(got, ClientSessionStatus::Unknown);
    assert!(got.is_known());
}

#[test]
fn client_session_status_deserializes_unrecognized_value_to_fallback() {
    let got: ClientSessionStatus =
        serde_json::from_str("\"totally_new_status\"").expect("deserialize unrecognized value");
    assert_eq!(got, ClientSessionStatus::UnknownValue("totally_new_status".to_string()));
    assert!(!got.is_known());
}
