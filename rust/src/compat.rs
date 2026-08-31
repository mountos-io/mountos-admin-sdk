//! Backward-compatible aliases for the pre-1.14.0 "Pair" naming. 1.14.0
//! renamed Pair -> Copyset (types and methods) without shipping migration
//! aliases, a semver-breaking change in a minor release. This module
//! restores the old names as aliases/wrapper methods so code written
//! against <1.14.0 keeps compiling; new code should use the Copyset-named
//! exports directly. The old names are deprecated.
//!
//! `type X = Y` is a genuine Rust type alias (X and Y are the same type),
//! so this is a lossless, zero-cost compatibility layer for the type-only
//! renames below (`Pair`, `PairState`, `DrainPairStorageResponse`). It does
//! NOT cover field-level renames on structs that were never themselves
//! renamed (`BlockVolume::copyset_id`, `UpdateConfigResult`'s
//! `*copyset*`/`active_copyset_count_*` fields,
//! `VolumeBlockPlacementConfig::copyset_ids`,
//! `VolumeBlockPlacementResizeResult`'s `copyset_count_*`/`copysets_*`
//! fields): Rust has no field-alias mechanism, so restoring those would
//! need a hand-written parallel struct with its own Deserialize impl.
//! Neither is done here; no known Rust consumer of this SDK exists in the
//! mountOS suite. `UpdateVolumePairConfigRequest` is the one request type
//! below that had its own field renamed, so it is a distinct struct
//! translated at the call site rather than a plain alias.
//!
//! Hand-written, not touched by `make gen`.

use crate::{
    Copyset, CopysetState, DrainCopysetStorageResponse, Error, StoragesService,
    UpdateVolumeCopysetConfigRequest, VolumeBlockPlacementConfig, VolumeBlockPlacementResizeResult,
    VolumesService,
};

/// Deprecated: use [`CopysetState`] instead.
pub type PairState = CopysetState;

/// Deprecated: use [`Copyset`] instead.
pub type Pair = Copyset;

/// Deprecated: use [`DrainCopysetStorageResponse`] instead.
pub type DrainPairStorageResponse = DrainCopysetStorageResponse;

/// Deprecated: use [`UpdateVolumeCopysetConfigRequest`] instead.
#[derive(Debug, Clone)]
pub struct UpdateVolumePairConfigRequest {
    pub target_pair_count: i32,
}

impl From<&UpdateVolumePairConfigRequest> for UpdateVolumeCopysetConfigRequest {
    fn from(req: &UpdateVolumePairConfigRequest) -> Self {
        UpdateVolumeCopysetConfigRequest { target_copyset_count: req.target_pair_count }
    }
}

impl StoragesService {
    /// Deprecated: use [`StoragesService::list_copysets`] instead.
    pub async fn list_pairs(
        &self,
        storage_id: i64,
        state: Option<&str>,
        include_retired: Option<bool>,
    ) -> Result<Vec<Pair>, Error> {
        self.list_copysets(storage_id, state, include_retired).await
    }

    /// Deprecated: use [`StoragesService::get_copyset_status`] instead.
    pub async fn get_pair_status(&self, storage_id: i64, pair_id: &str) -> Result<Pair, Error> {
        self.get_copyset_status(storage_id, pair_id).await
    }

    /// Deprecated: use [`StoragesService::drain_copyset`] instead.
    pub async fn drain_pair(&self, storage_id: i64, pair_id: &str) -> Result<DrainPairStorageResponse, Error> {
        self.drain_copyset(storage_id, pair_id).await
    }
}

impl VolumesService {
    /// Deprecated: use [`VolumesService::get_copyset_config`] instead.
    pub async fn get_pair_config(&self, volume_id: i64) -> Result<VolumeBlockPlacementConfig, Error> {
        self.get_copyset_config(volume_id).await
    }

    /// Deprecated: use [`VolumesService::update_copyset_config`] instead.
    pub async fn update_pair_config(
        &self,
        volume_id: i64,
        req: &UpdateVolumePairConfigRequest,
    ) -> Result<VolumeBlockPlacementResizeResult, Error> {
        self.update_copyset_config(volume_id, &UpdateVolumeCopysetConfigRequest::from(req)).await
    }
}
