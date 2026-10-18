//! Input structs are non_exhaustive, so this crate-external test builds them
//! through the public API only: new, Default and the with_* setters.

use mountos_admin_sdk::{
    CreateAccountRequest, DashboardUser, EditAccountRequest, EditVolumeRequest, RetentionPolicy,
    VolumeRetentionPolicy, VolumeVersioningPolicy,
};
use serde_json::json;

#[test]
fn edit_volume_request_carries_nested_policies() {
    let req = EditVolumeRequest::default()
        .with_retention(
            VolumeRetentionPolicy::default()
                .with_data_days(30)
                .with_grace_days(7)
                .with_fork_grace_days(3)
                .with_event_log_days(14),
        )
        .with_versioning(VolumeVersioningPolicy::default().with_content_window_seconds(60));

    let body = serde_json::to_value(&req).expect("serialize request");
    assert_eq!(
        body["retention"],
        json!({"dataDays": 30, "graceDays": 7, "forkGraceDays": 3, "eventLogDays": 14})
    );
    assert_eq!(body["versioning"], json!({"contentWindowSeconds": 60}));
}

#[test]
fn unset_optional_fields_are_omitted() {
    let req = EditVolumeRequest::default().with_retention(VolumeRetentionPolicy::default());
    let body = serde_json::to_value(&req).expect("serialize request");
    assert_eq!(body["retention"], json!({}));
    assert!(body.get("versioning").is_none());
}

#[test]
fn account_retention_and_dashboard_user_are_buildable() {
    let edit = EditAccountRequest::default()
        .with_retention(RetentionPolicy::default().with_client_session_days(5));
    let body = serde_json::to_value(&edit).expect("serialize edit");
    assert_eq!(body["retention"], json!({"clientSessionDays": 5}));

    let create = CreateAccountRequest::new("acme");
    assert_eq!(
        serde_json::to_value(&create).expect("serialize create")["name"],
        "acme"
    );

    let user = DashboardUser::new("u1", "Alice", "user")
        .with_account_id(7)
        .with_user_id(9);
    let body = serde_json::to_value(&user).expect("serialize user");
    assert_eq!(body["role"], "user");
    assert_eq!(body["accountId"], 7);
}
