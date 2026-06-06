//! Stable provider-type identifiers accepted by the storage API for
//! `Storage.provider_type` / `CreateStorageRequest.provider_type`.
//!
//! For Azure (`PROVIDER_TYPE_AZURE`), the generic credential fields map as:
//! `endpoint` → `https://<account>.blob.core.windows.net` (or Azurite URL),
//! `bucket` → container name, `access_key` → storage account name,
//! `secret_key` → base64 account key, `region` → informational only.

pub const PROVIDER_TYPE_S3: &str = "s3";
pub const PROVIDER_TYPE_BACKBLAZE: &str = "backblaze";
pub const PROVIDER_TYPE_CLOUDFLARE: &str = "cloudflare";
pub const PROVIDER_TYPE_DIGITAL_OCEAN: &str = "digitalocean";
pub const PROVIDER_TYPE_IBM_CLOUD: &str = "ibmcloud";
pub const PROVIDER_TYPE_IMPOSSIBLE_CLOUD: &str = "impossiblecloud";
pub const PROVIDER_TYPE_LYVE: &str = "lyve";
pub const PROVIDER_TYPE_WASABI: &str = "wasabi";
pub const PROVIDER_TYPE_S3_COMPATIBLE: &str = "s3compatible";
pub const PROVIDER_TYPE_AZURE: &str = "azure";
pub const PROVIDER_TYPE_MOUNTOS: &str = "mountOS";
