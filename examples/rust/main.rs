//! Example exercising the mountOS Admin Rust SDK against a running appserv.
//!
//! Set credentials via environment:
//!   MOUNTOS_BASE_URL     (default https://appserv.example.com)
//!   MOUNTOS_PRIVATE_KEY  base64 ED25519 provider key

use std::env;

use mountos_admin_sdk::{
    AccountListOptions, AddUserRequest, AuditLogListOptions, Client, Config, CreateAccountRequest,
    CreateRegionRequest, CreateStorageRequest, Error, UpdateVolumeQuotaRequest, UserListOptions,
    serde_json,
};

#[tokio::main]
async fn main() -> Result<(), Error> {
    let client = Client::new(Config {
        base_url: env::var("MOUNTOS_BASE_URL")
            .unwrap_or_else(|_| "https://appserv.example.com".into()),
        private_key: env::var("MOUNTOS_PRIVATE_KEY").unwrap_or_default(),
        ..Default::default()
    })?;

    // --- License ---
    let license = client.license.get().await?;
    println!("License: {} ({})", license.licensee, license.status);

    // --- Accounts ---
    let created = client
        .accounts
        .create(&CreateAccountRequest {
            name: "Acme Corp".into(),
            description: Some("Demo account".into()),
            icon_url: None,
            provider_info: Some(serde_json::json!({ "tier": "enterprise" })),
        })
        .await?;
    println!("Created account ID: {}", created.id);

    let account = client.accounts.get(created.id).await?;
    println!(
        "Account: {} (active={}, locked={})",
        account.name, account.is_active, account.locked
    );

    let page = client
        .accounts
        .list(Some(&AccountListOptions {
            page: Some(1),
            limit: Some(10),
            ..Default::default()
        }))
        .await?;
    println!(
        "Accounts: {} total, page {} of {}",
        page.pagination.total, page.pagination.page, page.pagination.total_pages
    );

    client.accounts.lock(created.id).await?;
    client.accounts.unlock(created.id).await?;

    // --- Users ---
    let user = client
        .users
        .add(&AddUserRequest {
            account_id: created.id,
            username: format!("alice-{}", created.id),
            email: format!("alice-{}@example.com", created.id),
            name: Some("Alice".into()),
            provider_info: None,
        })
        .await?;
    println!("Created user ID: {}", user.id);

    let users = client
        .users
        .list(Some(&UserListOptions {
            account_id: created.id,
            limit: Some(10),
            ..Default::default()
        }))
        .await?;
    println!("Users in account: {}", users.pagination.total);

    // --- Regions ---
    let region = client
        .regions
        .create(&CreateRegionRequest {
            account_id: created.id,
            name: format!("us-east-{}", created.id),
            dns: format!("us-east-{}.example.com", created.id),
        })
        .await?;
    println!("Created region ID: {}", region.id);

    // --- Storages (dummy S3 payload; appserv may reject on validation) ---
    match client
        .storages
        .create(&CreateStorageRequest {
            account_id: created.id,
            region_id: region.id,
            name: format!("prod-s3-{}", created.id),
            description: None,
            storage_type: "object".into(),
            provider_type: "s3".into(),
            endpoint: "https://s3.example.com".into(),
            region: Some("us-east-1".into()),
            bucket: Some("demo-bucket".into()),
            base: None,
            block_region: None,
            block_type: None,
            block_size: None,
            access_key: Some("AKIAEXAMPLE".into()),
            secret_key: Some("secret".into()),
        })
        .await
    {
        Ok(resp) => println!("Created storage ID: {}", resp.id),
        Err(Error::Api { message, status, .. }) => {
            println!("Create storage rejected: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Volumes (expected to fail if volume 1 does not exist) ---
    match client
        .volumes
        .update_quota(1, &UpdateVolumeQuotaRequest { quota_limit: 10 << 30 })
        .await
    {
        Ok(resp) => println!("Updated volume quota for {}", resp.id),
        Err(Error::Api { message, status, .. }) => {
            println!("UpdateQuota error: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Volume stats (GET with inline response; 404 if volume 1 is absent) ---
    match client.volumes.stats(1).await {
        Ok(stats) => println!("Volume 1 live bytes: {}", stats.live_volume),
        Err(Error::Api { message, status, .. }) => {
            println!("Volume stats: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Audit logs (cursor pagination) ---
    let logs = client
        .audit_logs
        .list(Some(&AuditLogListOptions {
            account_id: Some(created.id),
            limit: Some(5),
            ..Default::default()
        }))
        .await?;
    println!("Audit logs: {} entries", logs.items.len());

    // --- Service nodes ---
    match client.service_nodes.list(region.id, None, None, None, None).await {
        Ok(nodes) => println!("Service nodes: {}", nodes.len()),
        Err(e) => println!("List nodes: {e}"),
    }

    // --- Vault ---
    match client.vault.resync().await {
        Ok(()) => println!("Vault resynced"),
        Err(e) => println!("Vault resync: {e}"),
    }

    println!("Done.");
    Ok(())
}
