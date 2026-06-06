//! Error type returned by all SDK operations.

/// An error from the mountOS Admin API or the transport beneath it.
#[derive(Debug, thiserror::Error)]
pub enum Error {
    /// The API returned a non-success envelope (`status != "success"`).
    #[error("mountos: {message} (status={status}, code={error_code})")]
    Api {
        /// Human-readable message from the API.
        message: String,
        /// HTTP status code.
        status: u16,
        /// mountOS application error code (0 when absent).
        error_code: i64,
    },

    /// The HTTP request failed (connection, timeout, TLS, ...).
    #[error("mountos: http transport error: {0}")]
    Http(#[from] reqwest::Error),

    /// A request or response body could not be (de)serialized.
    #[error("mountos: serialization error: {0}")]
    Serde(#[from] serde_json::Error),

    /// The configured private key could not be parsed.
    #[error("mountos: invalid private key: {0}")]
    Key(String),
}
