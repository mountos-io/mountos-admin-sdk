package sdk

import "context"

// Backward-compatible aliases for the pre-1.14.0 "Pair" naming. 1.14.0
// renamed Pair -> Copyset (types and methods) without shipping migration
// aliases, which is a semver-breaking change in a minor release. This file
// restores the old names as aliases/wrappers so code written against
// <1.14.0 keeps compiling. New code should use the Copyset-named exports
// directly; the old names are deprecated.
//
// Go type aliases (type X = Y) make Pair and Copyset the identical type, so
// this is a lossless, zero-cost compatibility layer for the type-only
// renames below (Pair, PairState, DrainPairStorageResponse). It does NOT
// cover field-level renames on struct types that were never themselves
// renamed (BlockVolume.PairID -> CopysetID, VolumeBlockPlacementConfig.PairIds,
// VolumeBlockPlacementResizeResult's PairCount*/Pairs* fields): Go has no
// field-alias mechanism, so restoring those would require either editing
// the generated structs (adding the old field back) or a hand-written
// parallel type with custom JSON decoding. Neither is done here; no known
// Go consumer of this SDK reads those fields (see mountos-launcher, its
// only in-repo Go consumer). UpdateVolumePairConfigRequest is the one
// request type below that had its own field renamed, so it is a distinct
// struct translated at the call site rather than a plain alias.
//
// Hand-written, not touched by `make gen`.

// Deprecated: use CopysetState instead.
type PairState = CopysetState

const (
	// Deprecated: use CopysetStateActive instead.
	PairStateActive = CopysetStateActive
	// Deprecated: use CopysetStateDraining instead.
	PairStateDraining = CopysetStateDraining
	// Deprecated: use CopysetStateSyncedDrained instead.
	PairStateSyncedDrained = CopysetStateSyncedDrained
	// Deprecated: use CopysetStateRetired instead.
	PairStateRetired = CopysetStateRetired
)

// Deprecated: use ParseCopysetState instead.
func ParsePairState(s string) (PairState, error) { return ParseCopysetState(s) }

// Deprecated: use Copyset instead.
type Pair = Copyset

// Deprecated: use DrainCopysetStorageResponse instead.
type DrainPairStorageResponse = DrainCopysetStorageResponse

// UpdateVolumePairConfigRequest is the pre-1.14.0 request shape for
// UpdatePairConfig. Its field was itself renamed (TargetPairCount ->
// TargetCopysetCount), so unlike the type-only renames above this cannot be
// a plain alias: it is a distinct type with the old field name, translated
// to UpdateVolumeCopysetConfigRequest by UpdatePairConfig below.
//
// Deprecated: use UpdateVolumeCopysetConfigRequest instead.
type UpdateVolumePairConfigRequest struct {
	TargetPairCount int32 `json:"targetPairCount"`
}

func (r *UpdateVolumePairConfigRequest) toCopysetConfigRequest() *UpdateVolumeCopysetConfigRequest {
	return &UpdateVolumeCopysetConfigRequest{TargetCopysetCount: r.TargetPairCount}
}

// Deprecated: use StoragesService.ListCopysets instead.
func (s *StoragesService) ListPairs(ctx context.Context, storageID int64, state string, includeRetired bool) ([]Pair, error) {
	return s.ListCopysets(ctx, storageID, state, includeRetired)
}

// Deprecated: use StoragesService.GetCopysetStatus instead.
func (s *StoragesService) GetPairStatus(ctx context.Context, storageID int64, pairID string) (*Pair, error) {
	return s.GetCopysetStatus(ctx, storageID, pairID)
}

// Deprecated: use StoragesService.DrainCopyset instead.
func (s *StoragesService) DrainPair(ctx context.Context, storageID int64, pairID string) (*DrainPairStorageResponse, error) {
	return s.DrainCopyset(ctx, storageID, pairID, nil)
}

// Deprecated: use VolumesService.GetCopysetConfig instead.
func (s *VolumesService) GetPairConfig(ctx context.Context, volumeID int64) (*VolumeBlockPlacementConfig, error) {
	return s.GetCopysetConfig(ctx, volumeID)
}

// Deprecated: use VolumesService.UpdateCopysetConfig instead.
func (s *VolumesService) UpdatePairConfig(ctx context.Context, volumeID int64, req *UpdateVolumePairConfigRequest) (*VolumeBlockPlacementResizeResult, error) {
	return s.UpdateCopysetConfig(ctx, volumeID, req.toCopysetConfigRequest())
}
