// Backward-compatible aliases for the pre-1.14.0 "Pair" naming. 1.14.0
// renamed Pair -> Copyset (types, methods, and some response fields)
// without shipping migration aliases, which is a semver-breaking change in
// a minor release. This file restores the old names/fields alongside the
// new ones so a caller written against <1.14.0 keeps compiling and working
// against the current server contract. New code should use the
// Copyset-named exports directly; the old names are @deprecated.
//
// Hand-written, not touched by `make gen` - kept separate from
// types_gen.ts/client_gen.ts so a regeneration cannot silently drop it.

import type {
  BlockVolume,
  Copyset,
  CopysetState,
  UpdateVolumeCopysetConfigRequest,
  VolumeBlockPlacementConfig,
  VolumeBlockPlacementResizeResult,
} from './types_gen.js'
import { StoragesResource, VolumesResource } from './client_gen.js'

// ── Type aliases ──────────────────────────────────────────────────────

/** @deprecated Use CopysetState instead. */
export type PairState = CopysetState

/** @deprecated Use Copyset instead. */
export type Pair = Copyset

/** @deprecated Use UpdateVolumeCopysetConfigRequest instead. */
export interface UpdateVolumePairConfigRequest {
  targetPairCount: number
}

function toCopysetConfigRequest(req: UpdateVolumePairConfigRequest): UpdateVolumeCopysetConfigRequest {
  return { targetCopysetCount: req.targetPairCount }
}

// ── Response field shims ─────────────────────────────────────────────
// The server (mountos-servers) only sends the new field names now; a bare
// type alias would compile but silently drop data for a caller still
// reading the old key. These interface merges add the old key back onto
// the existing exported types, populated by the wrappers below.

declare module './types_gen.js' {
  interface BlockVolume {
    /** @deprecated Use copysetId instead. Populated for compatibility only. */
    pairId?: string
  }
  interface VolumeBlockPlacementConfig {
    /** @deprecated Use targetCopysetCount instead. Populated for compatibility only. */
    targetPairCount?: number
    /** @deprecated Use copysetIds instead. Populated for compatibility only. */
    pairIds?: string[]
  }
  interface VolumeBlockPlacementResizeResult {
    /** @deprecated Use targetCopysetCount instead. Populated for compatibility only. */
    targetPairCount?: number
    /** @deprecated Use copysetCountBefore instead. Populated for compatibility only. */
    pairCountBefore?: number
    /** @deprecated Use copysetsAdded instead. Populated for compatibility only. */
    pairsAdded?: number
    /** @deprecated Use copysetsRemoved instead. Populated for compatibility only. */
    pairsRemoved?: number
    /** @deprecated Use copysetCountAfter instead. Populated for compatibility only. */
    pairCountAfter?: number
  }
}

function shimBlockVolume(bv: BlockVolume): BlockVolume {
  bv.pairId = bv.copysetId
  return bv
}

function shimVolumeBlockPlacementConfig(c: VolumeBlockPlacementConfig): VolumeBlockPlacementConfig {
  c.targetPairCount = c.targetCopysetCount
  c.pairIds = c.copysetIds
  return c
}

function shimVolumeBlockPlacementResizeResult(r: VolumeBlockPlacementResizeResult): VolumeBlockPlacementResizeResult {
  r.targetPairCount = r.targetCopysetCount
  r.pairCountBefore = r.copysetCountBefore
  r.pairsAdded = r.copysetsAdded
  r.pairsRemoved = r.copysetsRemoved
  r.pairCountAfter = r.copysetCountAfter
  return r
}

// ── Method aliases ────────────────────────────────────────────────────

declare module './client_gen.js' {
  interface StoragesResource {
    /** @deprecated Use listCopysets instead. */
    listPairs(storageId: number, state?: string, includeRetired?: boolean, signal?: AbortSignal): Promise<Pair[]>
    /** @deprecated Use getCopysetStatus instead. */
    getPairStatus(storageId: number, pairId: string, signal?: AbortSignal): Promise<Pair>
    /** @deprecated Use drainCopyset instead. */
    drainPair(storageId: number, pairId: string, signal?: AbortSignal): Promise<{ id: string; state: string }>
  }
  interface VolumesResource {
    /** @deprecated Use getCopysetConfig instead. */
    getPairConfig(volumeId: number, signal?: AbortSignal): Promise<VolumeBlockPlacementConfig>
    /** @deprecated Use updateCopysetConfig instead. */
    updatePairConfig(volumeId: number, req: UpdateVolumePairConfigRequest, signal?: AbortSignal): Promise<VolumeBlockPlacementResizeResult>
  }
}

StoragesResource.prototype.listPairs = function (
  this: StoragesResource, storageId: number, state?: string, includeRetired?: boolean, signal?: AbortSignal,
): Promise<Copyset[]> {
  return this.listCopysets(storageId, state, includeRetired, signal)
}

StoragesResource.prototype.getPairStatus = function (
  this: StoragesResource, storageId: number, pairId: string, signal?: AbortSignal,
): Promise<Copyset> {
  return this.getCopysetStatus(storageId, pairId, signal)
}

StoragesResource.prototype.drainPair = function (
  this: StoragesResource, storageId: number, pairId: string, signal?: AbortSignal,
): Promise<{ id: string; state: string }> {
  return this.drainCopyset(storageId, pairId, {}, signal)
}

VolumesResource.prototype.getPairConfig = function (
  this: VolumesResource, volumeId: number, signal?: AbortSignal,
): Promise<VolumeBlockPlacementConfig> {
  return this.getCopysetConfig(volumeId, signal)
}

VolumesResource.prototype.updatePairConfig = function (
  this: VolumesResource, volumeId: number, req: UpdateVolumePairConfigRequest, signal?: AbortSignal,
): Promise<VolumeBlockPlacementResizeResult> {
  return this.updateCopysetConfig(volumeId, toCopysetConfigRequest(req), signal)
}

// listBlockVolumes was never renamed - only some of its response fields
// were. Wrap it in place so every caller, old or new, gets both field name
// sets on the response.

const genListBlockVolumes = StoragesResource.prototype.listBlockVolumes
StoragesResource.prototype.listBlockVolumes = function (
  this: StoragesResource, storageId: number, signal?: AbortSignal,
): Promise<BlockVolume[]> {
  return genListBlockVolumes.call(this, storageId, signal).then((list) => list.map(shimBlockVolume))
}

// getCopysetConfig/updateCopysetConfig responses also carry the deprecated
// fields, so a caller who migrated the method name but not every read site
// still works.

const genGetCopysetConfig = VolumesResource.prototype.getCopysetConfig
VolumesResource.prototype.getCopysetConfig = function (
  this: VolumesResource, volumeId: number, signal?: AbortSignal,
): Promise<VolumeBlockPlacementConfig> {
  return genGetCopysetConfig.call(this, volumeId, signal).then(shimVolumeBlockPlacementConfig)
}

const genUpdateCopysetConfig = VolumesResource.prototype.updateCopysetConfig
VolumesResource.prototype.updateCopysetConfig = function (
  this: VolumesResource, volumeId: number, req: UpdateVolumeCopysetConfigRequest, signal?: AbortSignal,
): Promise<VolumeBlockPlacementResizeResult> {
  return genUpdateCopysetConfig.call(this, volumeId, req, signal).then(shimVolumeBlockPlacementResizeResult)
}
