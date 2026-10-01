//! Contract test for the fork tree listing cursor: it is an opaque string.
//! The client sends it back unchanged and reads `next_cursor` as a string,
//! with JSON null (final page) read as `None`.

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use mountos_admin_sdk::{Client, Config, VolumeForkTreeListOptions};
use std::io::{Read, Write};
use std::net::TcpListener;
use std::thread;

const OPAQUE: &str = "n1.dGVzdC1fbmFtZQ";

/// Serves two canned pages on one listener and returns each request line.
fn fixture_server() -> (String, thread::JoinHandle<Vec<String>>) {
    let listener = TcpListener::bind("127.0.0.1:0").expect("bind ephemeral port");
    let addr = listener.local_addr().expect("local addr");

    let handle = thread::spawn(move || {
        let mut request_lines = Vec::new();
        for page in 0..2 {
            let (mut stream, _) = listener.accept().expect("accept");
            let mut buf = [0u8; 4096];
            let n = stream.read(&mut buf).expect("read request");
            let text = String::from_utf8_lossy(&buf[..n]).to_string();
            request_lines.push(text.lines().next().expect("request line").to_string());

            let next = if page == 0 { serde_json::json!(OPAQUE) } else { serde_json::Value::Null };
            let envelope = serde_json::json!({
                "status": "success", "message": "ok",
                "data": {
                    "items": [{"name": "a", "kind": "file", "inode": 5, "size": 1, "mtime": 2, "ctime": 3}],
                    "nextCursor": next,
                },
            });
            let body = serde_json::to_vec(&envelope).expect("serialize envelope");
            let head = format!(
                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
                body.len()
            );
            stream.write_all(head.as_bytes()).expect("write head");
            stream.write_all(&body).expect("write body");
            stream.flush().ok();
        }
        request_lines
    });

    (format!("http://{}", addr), handle)
}

#[tokio::test]
async fn tree_list_cursor_is_an_opaque_string_round_trip() {
    let (base_url, handle) = fixture_server();
    let seed = [7u8; 32];
    let client = Client::new(Config { base_url, private_key: STANDARD.encode(seed), ..Default::default() })
        .expect("new client");

    let first = client.volume_fork_trees.list(7, "main", None).await.expect("page 1");
    assert_eq!(first.next_cursor.as_deref(), Some(OPAQUE));

    let opts = VolumeForkTreeListOptions { cursor: first.next_cursor.clone(), ..Default::default() };
    let last = client.volume_fork_trees.list(7, "main", Some(&opts)).await.expect("page 2");
    assert!(last.next_cursor.is_none());

    let lines = handle.join().expect("server thread");
    assert!(!lines[0].contains("cursor="), "page 1 must not send a cursor: {}", lines[0]);
    assert!(lines[1].contains(&format!("cursor={OPAQUE}")), "page 2 must send the cursor unchanged: {}", lines[1]);
}
