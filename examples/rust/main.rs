//! Example exercising the mountOS Admin Rust SDK against a running appserv.
//!
//! Set credentials via environment:
//!   MOUNTOS_BASE_URL     (default https://appserv.example.com)
//!   MOUNTOS_PRIVATE_KEY  base64 ED25519 provider key

use std::env;

use mountos_admin_sdk::{
    AccountListOptions, AddUserRequest, AuditLogListOptions, Client, Config, CreateAccountRequest,
    CreateRegionRequest, CreateStorageRequest, Error, LoadLicenseRequest, UpdateVolumeQuotaRequest,
    UserListOptions, serde_json,
};

#[tokio::main]
async fn main() -> Result<(), Error> {
    let client = Client::new(Config::new(
        env::var("MOUNTOS_BASE_URL").unwrap_or_else(|_| "https://appserv.example.com".into()),
        env::var("MOUNTOS_PRIVATE_KEY").unwrap_or_default(),
    ))?;

    // --- License ---
    let license = client.license.get().await?;
    println!("License: {} ({})", license.licensee, license.status);

    // Upload signed payloads; the HUB verifies each and rejects the batch if any is invalid.
    match client
        .license
        .load(&LoadLicenseRequest::new(vec![
            env::var("MOUNTOS_LICENSE_PAYLOAD").unwrap_or_default(),
        ]))
        .await
    {
        Ok(result) => println!(
            "License loaded: {} new, {} ignored",
            result.loaded, result.ignored
        ),
        Err(Error::Api {
            message, status, ..
        }) => {
            println!("License load rejected: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    let records = client.license.list().await?;
    println!("Stored license payloads: {}", records.items.len());
    for r in &records.items {
        println!(
            "  {}: {} ({:?}, expires {})",
            r.key, r.licensee, r.status, r.expires_at
        );
    }

    // --- Accounts ---
    let created = client
        .accounts
        .create(
            &CreateAccountRequest::new("Acme Corp")
                .with_description("Demo account")
                .with_provider_info(serde_json::json!({ "tier": "enterprise" })),
        )
        .await?;
    println!("Created account ID: {}", created.id);

    let account = client.accounts.get(created.id).await?;
    println!(
        "Account: {} (active={}, locked={})",
        account.name, account.is_active, account.locked
    );

    let page = client
        .accounts
        .list(Some(
            &AccountListOptions::default().with_page(1).with_limit(10),
        ))
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
        .add(
            &AddUserRequest::new(
                created.id,
                format!("alice-{}", created.id),
                format!("alice-{}@example.com", created.id),
            )
            .with_name("Alice"),
        )
        .await?;
    println!("Created user ID: {}", user.id);

    let users = client
        .users
        .list(&UserListOptions::new(created.id).with_limit(10))
        .await?;
    println!("Users in account: {}", users.pagination.total);

    // --- Regions ---
    let region = client
        .regions
        .create(&CreateRegionRequest::new(
            created.id,
            format!("us-east-{}", created.id),
        ))
        .await?;
    println!("Created region ID: {}", region.id);

    // --- Storages (dummy S3 payload; appserv may reject on validation) ---
    match client
        .storages
        .create(
            &CreateStorageRequest::new(
                created.id,
                region.id,
                format!("prod-s3-{}", created.id),
                "object",
                "s3",
                "https://s3.example.com",
            )
            .with_region("us-east-1")
            .with_bucket("demo-bucket")
            .with_access_key("AKIAEXAMPLE")
            .with_secret_key("secret"),
        )
        .await
    {
        Ok(resp) => println!("Created storage ID: {}", resp.id),
        Err(Error::Api {
            message, status, ..
        }) => {
            println!("Create storage rejected: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Volumes (expected to fail if volume 1 does not exist) ---
    match client
        .volumes
        .update_quota(1, &UpdateVolumeQuotaRequest::new(10 << 30))
        .await
    {
        Ok(resp) => println!("Updated volume quota for {}", resp.id),
        Err(Error::Api {
            message, status, ..
        }) => {
            println!("UpdateQuota error: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Volume stats (GET with inline response; 404 if volume 1 is absent) ---
    match client.volumes.stats(1).await {
        Ok(stats) => println!("Volume 1 live bytes: {}", stats.live_volume),
        Err(Error::Api {
            message, status, ..
        }) => {
            println!("Volume stats: {message} (status={status})")
        }
        Err(e) => return Err(e),
    }

    // --- Audit logs (cursor pagination) ---
    let logs = client
        .audit_logs
        .list(&AuditLogListOptions::new(created.id).with_limit(5))
        .await?;
    println!("Audit logs: {} entries", logs.items.len());

    // --- Service nodes ---
    match client
        .service_nodes
        .list(region.id, None, None, None, None)
        .await
    {
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
