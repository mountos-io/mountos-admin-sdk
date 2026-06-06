//! HTTP transport: request signing, the response envelope, and typed verbs the
//! generated client calls into.

use percent_encoding::{AsciiSet, NON_ALPHANUMERIC, utf8_percent_encode};
use reqwest::RequestBuilder;
use serde::Serialize;
use serde::de::DeserializeOwned;

use crate::auth::TokenCache;
use crate::dashboard_user::sign_dashboard_user;
use crate::errors::Error;
use crate::types_gen::{Config, DashboardUser};

/// Unreserved path characters (RFC 3986 `-._~` plus alphanumerics) are left
/// intact; everything else in a path segment is percent-encoded.
const PATH_SEGMENT: &AsciiSet = &NON_ALPHANUMERIC
    .remove(b'-')
    .remove(b'_')
    .remove(b'.')
    .remove(b'~');

/// Percent-encodes a single path segment (used for free-form string ids).
pub(crate) fn encode_segment(segment: &str) -> String {
    utf8_percent_encode(segment, PATH_SEGMENT).to_string()
}

/// Shared client state behind every resource service.
pub(crate) struct ClientInner {
    base_url: String,
    http: reqwest::Client,
    auth: TokenCache,
    dashboard_user: Option<DashboardUser>,
    private_key: String,
}

impl ClientInner {
    pub(crate) fn new(config: Config) -> Result<Self, Error> {
        let auth = TokenCache::new(&config.private_key)?;
        let http = reqwest::Client::builder().build().map_err(Error::Http)?;
        Ok(Self {
            base_url: config.base_url.trim_end_matches('/').to_string(),
            http,
            auth,
            dashboard_user: config.dashboard_user,
            private_key: config.private_key,
        })
    }

    fn url(&self, path: &str) -> String {
        format!("{}{}", self.base_url, path)
    }

    pub(crate) async fn get<T: DeserializeOwned>(
        &self,
        path: &str,
        query: &[(&str, String)],
    ) -> Result<T, Error> {
        self.send(self.http.get(self.url(path)).query(query)).await
    }

    pub(crate) async fn post<T: DeserializeOwned, B: Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, Error> {
        self.send(self.http.post(self.url(path)).json(body)).await
    }

    pub(crate) async fn post_empty<T: DeserializeOwned>(&self, path: &str) -> Result<T, Error> {
        self.send(self.http.post(self.url(path))).await
    }

    pub(crate) async fn put<T: DeserializeOwned, B: Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, Error> {
        self.send(self.http.put(self.url(path)).json(body)).await
    }

    // Emitted by the generator only for DELETE endpoints; the current spec has
    // none, so it is unused until one is added (kept for transport symmetry).
    #[allow(dead_code)]
    pub(crate) async fn delete<T: DeserializeOwned>(&self, path: &str) -> Result<T, Error> {
        self.send(self.http.delete(self.url(path))).await
    }

    async fn send<T: DeserializeOwned>(&self, request: RequestBuilder) -> Result<T, Error> {
        // Token signing is synchronous; the auth lock is released before the await.
        let token = self.auth.token()?;
        let mut request = request.bearer_auth(token);
        if let Some(user) = &self.dashboard_user {
            let header = sign_dashboard_user(user, &self.private_key)?;
            request = request.header("X-MountOS-Dashboard-User", header);
        }

        let response = request.send().await?;
        let status = response.status();

        let envelope: Envelope = match response.json().await {
            Ok(env) => env,
            Err(_) => {
                return Err(Error::Api {
                    message: format!(
                        "{} {}",
                        status.as_u16(),
                        status.canonical_reason().unwrap_or("request failed"),
                    ),
                    status: status.as_u16(),
                    error_code: 0,
                });
            }
        };

        if envelope.status != "success" {
            return Err(Error::Api {
                message: envelope.message,
                status: status.as_u16(),
                error_code: envelope.error_code,
            });
        }

        let data = envelope.data.unwrap_or(serde_json::Value::Null);
        Ok(serde_json::from_value(data)?)
    }
}

/// The standard `{status, message, data, errorCode}` response envelope.
#[derive(serde::Deserialize)]
struct Envelope {
    #[serde(default)]
    status: String,
    #[serde(default)]
    message: String,
    #[serde(default)]
    data: Option<serde_json::Value>,
    #[serde(default, rename = "errorCode")]
    error_code: i64,
}
