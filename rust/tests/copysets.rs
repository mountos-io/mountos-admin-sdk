//! Fixture/mock-server contract test for the block copyset placement admin
//! surface: exercises the generated client against a hand-rolled
//! single-request TCP fixture, no live appserv and no new dev-dependency
//! (no mock-server crate available under `cargo --locked`). Covers the
//! "accepted, not completed" response shapes (drainCopyset/cancelDrain) and
//! regression-guards the reactivateMember GET-vs-POST generator bug (a
//! no-request endpoint with a named responseType on a mutating method was
//! silently generated as GET in all three SDK languages until fixed
//! alongside this test).

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use mountos_admin_sdk::{
    Client, Config, CopysetState, RegisterStorageCopysetRequest, RegisterStorageCopysetsBulkRequest,
};
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use std::thread;

struct RecordedRequest {
    method: String,
    path: String,
    body: Option<serde_json::Value>,
}

/// Binds an ephemeral port, serves exactly one request with the given JSON
/// `data` payload wrapped in this SDK's `{status, message, data}` envelope,
/// then closes. Returns the base URL to hand to `Config` and a join handle
/// that yields what the server actually received, for the caller to assert
/// method/path/body against.
fn fixture_server(response_data: serde_json::Value) -> (String, thread::JoinHandle<RecordedRequest>) {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind ephemeral port");
    let addr = listener.local_addr().expect("local addr");

    let handle = thread::spawn(move || {
        let (mut stream, _) = listener.accept().expect("accept one connection");
        let recorded = read_request(&mut stream);

        let envelope = serde_json::json!({"status": "success", "message": "ok", "data": response_data});
        let body = serde_json::to_vec(&envelope).expect("serialize envelope");
        let response = format!(
            "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
            body.len()
        );
        stream.write_all(response.as_bytes()).expect("write status line/headers");
        stream.write_all(&body).expect("write body");
        stream.flush().ok();

        recorded
    });

    (format!("http://{}", addr), handle)
}

/// Reads a single HTTP/1.1 request off `stream`: request line, headers up to
/// the blank line, then exactly Content-Length body bytes if present. Good
/// enough for reqwest's client behavior against a `Connection: close`
/// single-shot server - not a general-purpose HTTP parser.
fn read_request(stream: &mut TcpStream) -> RecordedRequest {
    let mut buf = Vec::new();
    let mut chunk = [0u8; 512];
    let header_end = loop {
        let n = stream.read(&mut chunk).expect("read from socket");
        assert!(n > 0, "connection closed before headers completed");
        buf.extend_from_slice(&chunk[..n]);
        if let Some(pos) = find_subslice(&buf, b"\r\n\r\n") {
            break pos;
        }
    };

    let header_text = String::from_utf8_lossy(&buf[..header_end]).to_string();
    let mut lines = header_text.split("\r\n");
    let request_line = lines.next().expect("request line");
    let mut parts = request_line.split_whitespace();
    let method = parts.next().expect("method").to_string();
    let path = parts.next().expect("path").to_string();

    let content_length: usize = lines
        .find_map(|l| l.to_ascii_lowercase().strip_prefix("content-length:").map(|v| v.trim().to_string()))
        .and_then(|v| v.parse().ok())
        .unwrap_or(0);

    let mut body_bytes = buf[header_end + 4..].to_vec();
    while body_bytes.len() < content_length {
        let n = stream.read(&mut chunk).expect("read body");
        assert!(n > 0, "connection closed before body completed");
        body_bytes.extend_from_slice(&chunk[..n]);
    }

    let body = if body_bytes.is_empty() { None } else { serde_json::from_slice(&body_bytes).ok() };
    RecordedRequest { method, path, body }
}

fn find_subslice(haystack: &[u8], needle: &[u8]) -> Option<usize> {
    haystack.windows(needle.len()).position(|w| w == needle)
}

fn test_client(base_url: String) -> Client {
    // Fixed 32-byte seed: structurally valid ED25519 key material, no
    // randomness needed for a fixture test.
    let seed = [7u8; 32];
    Client::new(Config { base_url, private_key: STANDARD.encode(seed), ..Default::default() }).expect("new client")
}

#[tokio::test]
async fn list_copysets_full_state_per_copyset() {
    let (base_url, handle) = fixture_server(serde_json::json!([
        {"id": "p1", "storageId": "s1", "name": "mos-block-a", "state": "active", "memberA": "bv1", "memberB": "bv2", "tags": []},
        {"id": "p2", "storageId": "s1", "name": "mos-block-b", "state": "draining", "memberA": "bv3", "memberB": "bv4", "pendingSyncJobsA": 3, "pendingSyncJobsB": 0, "tags": ["east"]},
    ]));
    let client = test_client(base_url);

    let copysets = client.storages.list_copysets(7, None, None).await.expect("list_copysets");
    assert_eq!(copysets.len(), 2);
    assert_eq!(copysets[1].state, CopysetState::Draining);
    assert_eq!(copysets[1].pending_sync_jobs_a, Some(3));

    let recorded = handle.join().expect("server thread");
    assert_eq!(recorded.method, "GET");
    assert_eq!(recorded.path, "/api/v1/storages/7/copysets");
}

#[tokio::test]
async fn drain_copyset_idempotent_ack() {
    // response reads "draining", not "drained" - an accepted-transition
    // ack, never a completion promise.
    let (base_url, handle) = fixture_server(serde_json::json!({"id": "p1", "state": "draining"}));
    let client = test_client(base_url);

    let res = client.storages.drain_copyset(7, "p1").await.expect("drain_copyset");
    assert_eq!(res.state, "draining");
    assert_eq!(handle.join().unwrap().method, "POST");
}

#[tokio::test]
async fn cancel_drain_active_again() {
    let (base_url, handle) = fixture_server(serde_json::json!({"id": "p1", "state": "active"}));
    let client = test_client(base_url);

    let res = client.storages.cancel_drain(7, "p1").await.expect("cancel_drain");
    assert_eq!(res.state, "active");
    assert_eq!(handle.join().unwrap().path, "/api/v1/storages/7/copysets/p1/cancel-drain");
}

/// Regression guard: reactivateMember is a mutating (POST) action with a
/// named responseType and no request body - a generator dispatch bug made
/// this silently emit a GET in all three SDK languages (fixed alongside
/// this test). Asserting the recorded method here, not just that the call
/// succeeds, is the point.
#[tokio::test]
async fn reactivate_member_sends_post() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": "bv1", "name": "originator", "regionId": 1, "memberState": "active",
    }));
    let client = test_client(base_url);

    let res = client.storages.reactivate_member(7, "bv1").await.expect("reactivate_member");
    assert_eq!(res.member_state, "active");
    assert_eq!(handle.join().unwrap().method, "POST");
}

#[tokio::test]
async fn register_copyset_explicit_name() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": "p5", "storageId": "s1", "name": "mos-block-a", "state": "active", "memberA": "bv5", "memberB": "bv6", "tags": [],
    }));
    let client = test_client(base_url);

    let res = client
        .storages
        .register_copyset(7, &RegisterStorageCopysetRequest { name: Some("mos-block-a".into()) })
        .await
        .expect("register_copyset");
    assert_eq!(res.name, "mos-block-a");
    assert_eq!(res.member_a.as_deref(), Some("bv5"));
    assert_eq!(res.member_b.as_deref(), Some("bv6"));
    assert_eq!(handle.join().unwrap().body.unwrap()["name"], "mos-block-a");
}

// name is optional on the wire: None must be dropped from the request body
// entirely (never sent as `"name": null`), letting the server auto-fill it -
// both members then derive from whatever it picks.
#[tokio::test]
async fn register_copyset_omitted_name() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": "p6", "storageId": "s1", "name": "riveted-truss-4f2a", "state": "active", "memberA": "bv7", "memberB": "bv8", "tags": [],
    }));
    let client = test_client(base_url);

    let res = client
        .storages
        .register_copyset(7, &RegisterStorageCopysetRequest { name: None })
        .await
        .expect("register_copyset");
    assert_eq!(res.state, CopysetState::Active);
    assert_eq!(res.name, "riveted-truss-4f2a");
    let body = handle.join().unwrap().body.unwrap();
    assert!(body.get("name").is_none(), "expected no name key in request body when omitted, got {body:?}");
}

#[tokio::test]
async fn register_copysets_bulk_count_only() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "copysets": [
            {"id": "p10", "storageId": "s1", "name": "riveted-truss-1a2b", "state": "active", "memberA": "bv10", "memberB": "bv11", "tags": []},
            {"id": "p11", "storageId": "s1", "name": "coupled-beam-3c4d", "state": "active", "memberA": "bv12", "memberB": "bv13", "tags": []},
        ],
    }));
    let client = test_client(base_url);

    let res = client
        .storages
        .register_copysets_bulk(7, &RegisterStorageCopysetsBulkRequest { count: 2 })
        .await
        .expect("register_copysets_bulk");
    assert_eq!(res.copysets.len(), 2);
    assert_eq!(res.copysets[0].name, "riveted-truss-1a2b");
    assert_eq!(res.copysets[1].name, "coupled-beam-3c4d");
    assert_eq!(handle.join().unwrap().body.unwrap()["count"], 2);
}

// The new member's name is always derived server-side from the copyset's
// own name, never operator-supplied: no request body to send.
#[tokio::test]
async fn add_copyset_member_replaces_vacant_slot() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": "bv9", "name": "mos-block-a-b", "regionId": 1, "memberState": "active",
    }));
    let client = test_client(base_url);

    let res = client.storages.add_copyset_member(7, "p1").await.expect("add_copyset_member");
    assert_eq!(res.member_state, "active");
    assert_eq!(res.name, "mos-block-a-b");
    assert!(handle.join().unwrap().body.is_none(), "expected no request body");
}

// Regression coverage for the generator emitting crate::http::encode_segment
// on every &str :path param (not just the one endpoint that originally
// needed it): a raw slash, space, or non-ASCII character in a string id
// must reach the wire as one escaped segment, not split into extra
// segments. `recorded.path` here is the literal request-line bytes read
// off the socket, not a re-parsed/decoded view, so it can actually tell an
// escaped "%2F" apart from a real segment boundary.
#[tokio::test]
async fn string_path_params_are_url_encoded() {
    let raw_copyset_id = "copyset/id with spaces/café";
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": raw_copyset_id, "storageId": "s1", "name": "mos-block-a", "state": "active", "tags": [],
    }));
    let client = test_client(base_url);

    let copyset = client.storages.get_copyset_status(7, raw_copyset_id).await.expect("get_copyset_status");
    assert_eq!(copyset.id, raw_copyset_id);

    let path = handle.join().unwrap().path;
    assert!(
        path.contains("copyset%2Fid%20with%20spaces%2Fcaf%C3%A9"),
        "request path {path:?} did not contain the escaped segment"
    );
    assert!(!path.contains("/copysets/copyset/"), "request path {path:?} sent the '/' unescaped, splitting the path");
}
