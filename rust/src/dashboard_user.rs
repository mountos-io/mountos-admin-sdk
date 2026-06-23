//! Signs the `X-MountOS-Dashboard-User` header for dashboard operator context.
//!
//! Mirrors the Go and TypeScript SDKs. The HMAC key is derived from a dedicated
//! secret (appserv `DASHBOARD_USER_HMAC_KEY`, distinct from the public
//! verification key) with domain separation, and the value is
//! `base64url(json).base64url(hmac-sha256(base64url(json), derived_key))`.

use std::time::{SystemTime, UNIX_EPOCH};

use base64::Engine;
use base64::engine::general_purpose::URL_SAFE_NO_PAD;
use hmac::{Hmac, Mac};
use sha2::Sha256;

use crate::errors::Error;
use crate::types_gen::DashboardUser;

type HmacSha256 = Hmac<Sha256>;

const DEFAULT_TTL_SECS: i64 = 10 * 60;

/// Produces the signed `X-MountOS-Dashboard-User` header value, stamping `exp`
/// at now + 10 minutes. `hmac_secret` is the dedicated dashboard-user HMAC
/// secret shared with appserv (`DASHBOARD_USER_HMAC_KEY`), never the public
/// verification key.
pub fn sign_dashboard_user(user: &DashboardUser, hmac_secret: &str) -> Result<String, Error> {
    if hmac_secret.is_empty() {
        return Err(Error::Key(
            "dashboard HMAC secret is required to sign a dashboard user header".into(),
        ));
    }

    let mut user = user.clone();
    user.exp = Some(unix_now() + DEFAULT_TTL_SECS);
    let payload = serde_json::to_vec(&user)?;
    let encoded = URL_SAFE_NO_PAD.encode(&payload);

    let key = derive_hmac_key(hmac_secret)?;
    let mut mac =
        HmacSha256::new_from_slice(&key).map_err(|e| Error::Key(e.to_string()))?;
    mac.update(encoded.as_bytes());
    let sig = URL_SAFE_NO_PAD.encode(mac.finalize().into_bytes());

    Ok(format!("{encoded}.{sig}"))
}

/// Derives a dedicated HMAC key from the shared secret with domain separation.
fn derive_hmac_key(secret: &str) -> Result<Vec<u8>, Error> {
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes())
        .map_err(|e| Error::Key(e.to_string()))?;
    mac.update(b"dashboard-user");
    Ok(mac.finalize().into_bytes().to_vec())
}

fn unix_now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0)
}
