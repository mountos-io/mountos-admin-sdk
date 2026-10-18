# mountOS Admin SDK for Rust

Async Rust SDK for [mountOS](https://mountos.io), a cross-platform, forkable,
time-traveling distributed filesystem over object storage. Automate accounts, users, volumes, forks, quotas, and audit
logs. Built on `reqwest`/`tokio` with ED25519 JWT authentication.

## Install

```bash
cargo add mountos-admin-sdk
```

Requires Rust 1.85+ (edition 2024). TLS uses `rustls` with the OS trust store,
so self-hosted appserv behind a private/corporate CA works out of the box.

## Usage

```rust,no_run
use mountos_admin_sdk::{Client, Config, CreateAccountRequest};

#[tokio::main]
async fn main() -> Result<(), mountos_admin_sdk::Error> {
    // 32-byte seed or 64-byte seed+pubkey, base64 encoded
    let client = Client::new(Config::new("https://appserv.example.com", "<base64 ED25519 key>"))?;

    // Accounts
    let created = client.accounts.create(&CreateAccountRequest::new("Acme")).await?;
    let account = client.accounts.get(created.id).await?;
    println!("{} (active={})", account.name, account.is_active);

    // Page-based list
    let page = client.accounts.list(None).await?;
    println!("{} accounts", page.pagination.total);

    // Volumes
    client.volumes.update_quota(
        created.id,
        &mountos_admin_sdk::UpdateVolumeQuotaRequest::new(1 << 30),
    ).await?;

    Ok(())
}
```

Resource accessors map to `client.<resource>` in snake_case (`client.volume_fork_trees`,
`client.region_audit_logs`, ...). Methods are `async` and take `&self`.

## Error Handling

```rust,no_run
use mountos_admin_sdk::{Client, Config, Error};

#[tokio::main]
async fn main() -> Result<(), Error> {
    let client = Client::new(Config::new("https://appserv.example.com", "<base64 ED25519 key>"))?;

    match client.accounts.get(999).await {
        Ok(account) => println!("{}", account.name),
        // Error is non_exhaustive: keep a wildcard arm.
        Err(Error::Api { message, status, error_code }) => {
            eprintln!("{message} (status={status}, code={error_code})");
        }
        Err(e) => eprintln!("transport error: {e}"),
    }
    Ok(())
}
```

## Auth

`private_key` is a base64-encoded ED25519 key - a 32-byte seed or a 64-byte
seed+public-key. JWT tokens are signed locally and cached for ~55 minutes (1h
TTL with a 5-minute refresh margin); the cache is `Mutex`-guarded and never held
across an `.await`.

Pass `dashboard_user` in `Config` to sign an `X-MountOS-Dashboard-User` header
into every request for operator-scoped access.

## Reference

Full API reference: [docs/rust.md](../docs/rust.md). Generated from
[api.yaml](../api.yaml); see also the language-neutral [api.md](../api.md).

## License

Apache-2.0
