//! Rust SDK for the mountOS provider API.
//!
//! Wraps the appserv provider API (`/api/v1/*`) with ED25519 JWT authentication.
//! Construct a [`Client`] from a [`Config`] and call typed methods grouped by
//! resource, e.g. `client.accounts.list(..)`.
//!
//! ```no_run
//! use mountos_admin_sdk::{Client, Config};
//!
//! # async fn run() -> Result<(), mountos_admin_sdk::Error> {
//! let client = Client::new(Config {
//!     base_url: "https://appserv.example.com".into(),
//!     private_key: "<base64 ED25519 key>".into(),
//!     ..Default::default()
//! })?;
//!
//! let account = client.accounts.get(1).await?;
//! println!("{}", account.name);
//! # Ok(())
//! # }
//! ```
//!
//! The `private_key` is a base64-encoded ED25519 key — either a 32-byte seed or
//! a 64-byte seed+public-key. JWT tokens are signed locally and cached for ~55
//! minutes (1h TTL with a 5-minute refresh margin).

#![forbid(unsafe_code)]

mod auth;
mod client_gen;
mod dashboard_user;
mod errors;
mod http;
mod providers;
mod types_gen;

pub use client_gen::*;
pub use dashboard_user::sign_dashboard_user;
pub use errors::Error;
pub use providers::*;
pub use types_gen::*;

// Re-exported so callers can build `serde_json::Value` fields (e.g.
// `provider_info`) without taking their own version-matched dependency.
pub use serde_json;
