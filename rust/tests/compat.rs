//! Regression guard for VE-017 (Pair -> Copyset renamed in 1.14.0 with no
//! aliases). Exercises compat.rs: the old method/type names still resolve
//! and hit the renamed Copyset routes, and the renamed request field is
//! translated on the wire.

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use mountos_admin_sdk::{Client, Config, PairState, UpdateVolumePairConfigRequest};
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use std::thread;

struct RecordedRequest {
    method: String,
    path: String,
    body: Option<serde_json::Value>,
}

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
    let seed = [7u8; 32];
    Client::new(Config { base_url, private_key: STANDARD.encode(seed), ..Default::default() }).expect("new client")
}

#[tokio::test]
async fn list_pairs_is_an_alias_for_list_copysets() {
    let (base_url, handle) = fixture_server(serde_json::json!([
        {"id": "p1", "storageId": "s1", "name": "mos-block-a", "state": "active", "volumeCount": 0, "tags": []},
    ]));
    let client = test_client(base_url);

    let pairs = client.storages.list_pairs(7, None, None).await.expect("list_pairs");
    assert_eq!(pairs.len(), 1);
    assert_eq!(pairs[0].state, PairState::Active);

    let recorded = handle.join().expect("server thread");
    assert_eq!(recorded.path, "/api/v1/storages/7/copysets");
}

#[tokio::test]
async fn get_pair_config_and_update_pair_config_alias_copyset_config_and_translate_fields() {
    let (base_url, handle) = fixture_server(serde_json::json!({
        "id": 5, "targetCopysetCount": 4, "copysetCountBefore": 3, "copysetsAdded": 1,
        "copysetsRemoved": 0, "copysetCountAfter": 4, "epoch": 2, "partial": false,
    }));
    let client = test_client(base_url);

    let resize = client
        .volumes
        .update_pair_config(5, &UpdateVolumePairConfigRequest { target_pair_count: 4 })
        .await
        .expect("update_pair_config");
    assert_eq!(resize.target_copyset_count, 4);
    assert_eq!(resize.copysets_added, 1);

    let recorded = handle.join().expect("server thread");
    assert_eq!(recorded.method, "PUT");
    assert_eq!(recorded.path, "/api/v1/volumes/5/copyset-config");
    let body = recorded.body.expect("request body");
    assert_eq!(body["targetCopysetCount"], serde_json::json!(4));
    assert!(body.get("targetPairCount").is_none(), "wire body must use the new field name only");
}
