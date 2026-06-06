//! ED25519 JWT signing with local token caching.
//!
//! Mirrors the Go and TypeScript SDKs: tokens carry `sub=mountos:provider`,
//! `aud=mountos/appserv`, `scope=service`, and `kfp = hex(sha256(pubkey)[:16])`
//! (also the JWT `kid`). Tokens are cached for the full TTL minus a refresh
//! margin and re-signed on demand.

use std::sync::Mutex;
use std::time::{SystemTime, UNIX_EPOCH};

use base64::Engine;
use base64::engine::general_purpose::{STANDARD, URL_SAFE_NO_PAD};
use ed25519_dalek::{Signer, SigningKey};
use sha2::{Digest, Sha256};

use crate::errors::Error;

const TOKEN_TTL: i64 = 3600;
const REFRESH_MARGIN: i64 = 300;
const CLOCK_SKEW_LEEWAY: i64 = 5;

pub(crate) struct TokenCache {
    signing_key: SigningKey,
    kfp: String,
    cached: Mutex<Cached>,
}

#[derive(Default)]
struct Cached {
    token: String,
    expiry: i64,
}

impl TokenCache {
    pub(crate) fn new(private_key_base64: &str) -> Result<Self, Error> {
        let seed = ed25519_seed(private_key_base64)?;
        let signing_key = SigningKey::from_bytes(&seed);

        let mut hasher = Sha256::new();
        hasher.update(signing_key.verifying_key().as_bytes());
        let digest = hasher.finalize();
        let kfp = hex::encode(&digest[..16]);

        Ok(Self {
            signing_key,
            kfp,
            cached: Mutex::new(Cached::default()),
        })
    }

    /// Returns a valid bearer token, signing a fresh one when the cache is
    /// empty or within the refresh margin of expiry.
    ///
    /// The critical section is fully synchronous (no `.await`), so the lock is
    /// never held across a suspension point.
    pub(crate) fn token(&self) -> Result<String, Error> {
        let now = unix_now();
        // Recover from a poisoned lock rather than panicking: the guarded state
        // is a plain cache that is safe to rebuild.
        let mut cached = self.cached.lock().unwrap_or_else(|e| e.into_inner());

        if !cached.token.is_empty() && now < cached.expiry - REFRESH_MARGIN {
            return Ok(cached.token.clone());
        }

        let exp = now + TOKEN_TTL;
        let token = self.sign(now, exp);
        cached.token = token.clone();
        cached.expiry = exp;
        Ok(token)
    }

    fn sign(&self, now: i64, exp: i64) -> String {
        let header = format!(r#"{{"alg":"EdDSA","typ":"JWT","kid":"{}"}}"#, self.kfp);
        let payload = format!(
            r#"{{"sub":"mountos:provider","aud":"mountos/appserv","iat":{now},"nbf":{},"exp":{exp},"jti":"{}","scope":"service","kfp":"{}"}}"#,
            now - CLOCK_SKEW_LEEWAY,
            uuid::Uuid::new_v4(),
            self.kfp,
        );
        let signing_input = format!(
            "{}.{}",
            URL_SAFE_NO_PAD.encode(header.as_bytes()),
            URL_SAFE_NO_PAD.encode(payload.as_bytes()),
        );
        let sig = self.signing_key.sign(signing_input.as_bytes());
        format!("{signing_input}.{}", URL_SAFE_NO_PAD.encode(sig.to_bytes()))
    }
}

/// Decodes a base64 ED25519 private key (32-byte seed or 64-byte seed+pubkey)
/// into its 32-byte seed.
fn ed25519_seed(private_key_base64: &str) -> Result<[u8; 32], Error> {
    let raw = STANDARD
        .decode(private_key_base64.trim())
        .map_err(|e| Error::Key(e.to_string()))?;
    let seed = match raw.len() {
        32 | 64 => &raw[..32],
        n => return Err(Error::Key(format!("expected 32 or 64 bytes, got {n}"))),
    };
    seed.try_into()
        .map_err(|_| Error::Key("invalid ED25519 seed length".into()))
}

fn unix_now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0)
}
