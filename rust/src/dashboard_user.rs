//! Signs the `X-MountOS-Dashboard-User` header for dashboard operator context.
//!
//! Mirrors the Go and TypeScript SDKs. The HMAC key is derived from the
//! ED25519 verification key (itself derived from the signing private key) with
//! domain separation, and the value is
//! `base64url(json).base64url(hmac-sha256(base64url(json), derived_key))`.

use std::time::{SystemTime, UNIX_EPOCH};

use base64::Engine;
use base64::engine::general_purpose::{STANDARD, URL_SAFE_NO_PAD};
use ed25519_dalek::SigningKey;
use hmac::{Hmac, Mac};
use sha2::Sha256;

use crate::errors::Error;
use crate::types_gen::DashboardUser;

type HmacSha256 = Hmac<Sha256>;

const DEFAULT_TTL_SECS: i64 = 10 * 60;

/// Produces the signed `X-MountOS-Dashboard-User` header value, stamping `exp`
/// at now + 10 minutes.
pub fn sign_dashboard_user(user: &DashboardUser, private_key: &str) -> Result<String, Error> {
    let verification_key = derive_verification_key(private_key)?;

    let mut user = user.clone();
    user.exp = Some(unix_now() + DEFAULT_TTL_SECS);
    let payload = serde_json::to_vec(&user)?;
    let encoded = URL_SAFE_NO_PAD.encode(&payload);

    let key = derive_hmac_key(&verification_key)?;
    let mut mac =
        HmacSha256::new_from_slice(&key).map_err(|e| Error::Key(e.to_string()))?;
    mac.update(encoded.as_bytes());
    let sig = URL_SAFE_NO_PAD.encode(mac.finalize().into_bytes());

    Ok(format!("{encoded}.{sig}"))
}

/// Derives the standard-base64 ED25519 public key from a base64 signing key.
fn derive_verification_key(signing_key_base64: &str) -> Result<String, Error> {
    let raw = STANDARD
        .decode(signing_key_base64.trim())
        .map_err(|e| Error::Key(e.to_string()))?;
    let seed: [u8; 32] = match raw.len() {
        32 | 64 => raw[..32]
            .try_into()
            .map_err(|_| Error::Key("invalid ED25519 seed length".into()))?,
        n => return Err(Error::Key(format!("signing key must be 32 or 64 bytes, got {n}"))),
    };
    let public = SigningKey::from_bytes(&seed).verifying_key();
    Ok(STANDARD.encode(public.as_bytes()))
}

/// Derives a dedicated HMAC key from the verification key with domain separation.
fn derive_hmac_key(verification_key: &str) -> Result<Vec<u8>, Error> {
    let mut mac = HmacSha256::new_from_slice(verification_key.as_bytes())
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
