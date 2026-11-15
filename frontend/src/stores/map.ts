import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  ApiClientError,
  fetchHeatmap,
  fetchPlaybackTimestamps,
  fetchPointSeries,
  fetchPublicNodes,
} from '../api/client'
import { getLegendSpec, type LegendSpec } from '../data/legends'
import {
  LAYER_AIR_QUALITY,
  type AirQualityStandard,
  type HeatmapResponse,
  type LayerKey,
  type PointResponse,
  type PublicNode,
} from '../types/telemetry'

export type PlaybackSpeed = 1 | 2 | 4

export interface SelectedPointMeta {
  lat: number
  lon: number
  nodeCode?: string
  nodeName?: string
}

const MAX_HEATMAP_CACHE = 320

const SPEED_INTERVAL_MS: Record<PlaybackSpeed, number> = {
  1: 950,
  2: 480,
  4: 240,
}

export const useMapStore = defineStore('map', () => {
  const activeLayer = ref<LayerKey>(LAYER_AIR_QUALITY)
  const activeStandard = ref<AirQualityStandard>('inca')
  const nodes = ref<PublicNode[]>([])
  const heatmap = ref<HeatmapResponse | null>(null)
  const timestamps = ref<string[]>([])
  const playbackIndex = ref<number>(0)
  const isPlaying = ref<boolean>(false)
  const playbackSpeed = ref<PlaybackSpeed>(1)

  const selectedPoint = ref<SelectedPointMeta | null>(null)
  const pointData = ref<PointResponse | null>(null)

  const isLoadingMap = ref<boolean>(false)
  const isLoadingPoint = ref<boolean>(false)
  const mapError = ref<string | null>(null)
  const pointError = ref<string | null>(null)
  const isOfflineDemo = ref<boolean>(false)
  const hasInitialized = ref<boolean>(false)

  const heatmapCache = new Map<string, { data: HeatmapResponse; isOfflineDemo: boolean }>()
  let heatmapRequestSeq = 0
  let pointRequestSeq = 0
  let playbackTimer: ReturnType<typeof setInterval> | null = null

  const isLive = computed<boolean>(() => {
    if (timestamps.value.length === 0) return true
    return playbackIndex.value >= timestamps.value.length - 1
  })

  const activeTimestamp = computed<string | null>(() => {
    if (isLive.value || timestamps.value.length === 0) return null
    return timestamps.value[playbackIndex.value] ?? null
  })

  const displayTimestamp = computed<string | null>(() => {
    return (
      heatmap.value?.recorded_at ??
      activeTimestamp.value ??
      timestamps.value[timestamps.value.length - 1] ??
      null
    )
  })

  const activeLegend = computed<LegendSpec>(() => {
    return getLegendSpec(activeLayer.value, activeStandard.value)
  })

  function cacheKey(layer: LayerKey, standard: AirQualityStandard, at: string | null): string {
    return `${layer}:${standard}:${at ?? 'live'}`
  }

  function writeCache(
    key: string,
    value: { data: HeatmapResponse; isOfflineDemo: boolean },
  ): void {
    if (heatmapCache.size >= MAX_HEATMAP_CACHE) {
      const oldestKey = heatmapCache.keys().next().value
      if (oldestKey !== undefined) {
        heatmapCache.delete(oldestKey)
      }
    }
    heatmapCache.set(key, value)
  }

  function prefetchStep(nextIdx: number): void {
    if (nextIdx < 0 || nextIdx >= timestamps.value.length) return
    const layer = activeLayer.value
    const standard = activeStandard.value
    const at = nextIdx >= timestamps.value.length - 1 ? null : (timestamps.value[nextIdx] ?? null)
    const key = cacheKey(layer, standard, at)
    if (heatmapCache.has(key)) return

    void fetchHeatmap(layer, standard, at)
      .then((res) => {
        writeCache(key, res)
      })
      .catch(() => {
        // Ignore silent prefetch failures; primary navigation surfaces errors directly.
      })
  }

  async function loadHeatmap(forceRefresh = false): Promise<void> {
    const seq = ++heatmapRequestSeq
    const layer = activeLayer.value
    const standard = activeStandard.value
    const at = activeTimestamp.value
    const key = cacheKey(layer, standard, at)

    if (!forceRefresh) {
      const cached = heatmapCache.get(key)
      if (cached) {
        heatmap.value = cached.data
        isOfflineDemo.value = cached.isOfflineDemo
        mapError.value = null
        return
      }
    }

    isLoadingMap.value = true
    mapError.value = null
    try {
      const res = await fetchHeatmap(layer, standard, at)
      if (seq !== heatmapRequestSeq) return
      writeCache(key, res)
      heatmap.value = res.data
      isOfflineDemo.value = res.isOfflineDemo
    } catch (err) {
      if (seq !== heatmapRequestSeq) return
      mapError.value =
        err instanceof ApiClientError ? err.message : 'No se pudo cargar el mapa de calor.'
    } finally {
      if (seq === heatmapRequestSeq) {
        isLoadingMap.value = false
      }
    }
  }

  async function refreshSelectedPoint(): Promise<void> {
    if (!selectedPoint.value) return
    const seq = ++pointRequestSeq
    const { lat, lon } = selectedPoint.value
    isLoadingPoint.value = true
    pointError.value = null

    try {
      const res = await fetchPointSeries(
        lat,
        lon,
        activeLayer.value,
        activeStandard.value,
        activeTimestamp.value,
      )
      if (seq !== pointRequestSeq) return
      pointData.value = res.data
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
    } catch (err) {
      if (seq !== pointRequestSeq) return
      pointError.value =
        err instanceof ApiClientError ? err.message : 'No se pudo consultar la serie temporal del punto.'
    } finally {
      if (seq === pointRequestSeq) {
        isLoadingPoint.value = false
      }
    }
  }

  async function initialize(force = false): Promise<void> {
    if (hasInitialized.value && !force) return
    isLoadingMap.value = true
    mapError.value = null
    if (force) {
      heatmapCache.clear()
    }

    try {
      const [nodesRes, tsRes] = await Promise.all([
        fetchPublicNodes(),
        fetchPlaybackTimestamps(),
      ])
      nodes.value = nodesRes.data.nodes
      timestamps.value = tsRes.data.timestamps
      playbackIndex.value = Math.max(0, tsRes.data.timestamps.length - 1)
      isOfflineDemo.value = nodesRes.isOfflineDemo || tsRes.isOfflineDemo
      hasInitialized.value = true
      await loadHeatmap(force)
      if (selectedPoint.value) {
        await refreshSelectedPoint()
      }
    } catch (err) {
      mapError.value =
        err instanceof ApiClientError ? err.message : 'No se pudo inicializar el mapa geoespacial.'
      isLoadingMap.value = false
    }
  }

  async function retryLiveApi(): Promise<void> {
    stopPlayback()
    await initialize(true)
  }

  async function setStandard(standard: AirQualityStandard): Promise<void> {
    if (activeStandard.value === standard && activeLayer.value === LAYER_AIR_QUALITY) {
      return
    }
    activeStandard.value = standard
    activeLayer.value = LAYER_AIR_QUALITY
    await Promise.all([loadHeatmap(), refreshSelectedPoint()])
  }

  async function setLayer(layer: LayerKey): Promise<void> {
    if (activeLayer.value === layer) return
    activeLayer.value = layer
    await Promise.all([loadHeatmap(), refreshSelectedPoint()])
  }

  async function selectTimestampIndex(nextIndex: number): Promise<void> {
    if (timestamps.value.length === 0) return
    const clamped = Math.max(0, Math.min(timestamps.value.length - 1, nextIndex))
    if (clamped === playbackIndex.value && heatmap.value) return
    playbackIndex.value = clamped
    await loadHeatmap()
    if (selectedPoint.value) {
      void refreshSelectedPoint()
    }
  }

  function stopPlayback(): void {
    isPlaying.value = false
    if (playbackTimer !== null) {
      clearInterval(playbackTimer)
      playbackTimer = null
    }
  }

  function restartInterval(): void {
    if (playbackTimer !== null) {
      clearInterval(playbackTimer)
      playbackTimer = null
    }
    if (!isPlaying.value) return

    playbackTimer = setInterval(() => {
      if (timestamps.value.length <= 1) {
        stopPlayback()
        return
      }
      if (playbackIndex.value >= timestamps.value.length - 1) {
        stopPlayback()
        return
      }
      const next = playbackIndex.value + 1
      void selectTimestampIndex(next)
      prefetchStep(next + 1)
      prefetchStep(next + 2)
    }, SPEED_INTERVAL_MS[playbackSpeed.value])
  }

  function togglePlayback(): void {
    if (timestamps.value.length <= 1) return
    if (isPlaying.value) {
      stopPlayback()
      return
    }
    if (playbackIndex.value >= timestamps.value.length - 1) {
      playbackIndex.value = 0
      void loadHeatmap()
    }
    isPlaying.value = true
    prefetchStep(playbackIndex.value + 1)
    prefetchStep(playbackIndex.value + 2)
    restartInterval()
  }

  function setPlaybackSpeed(speed: PlaybackSpeed): void {
    playbackSpeed.value = speed
    if (isPlaying.value) {
      restartInterval()
    }
  }

  async function goLive(): Promise<void> {
    stopPlayback()
    const lastIdx = Math.max(0, timestamps.value.length - 1)
    playbackIndex.value = lastIdx
    await loadHeatmap(true)
    if (selectedPoint.value) {
      await refreshSelectedPoint()
    }
  }

  async function inspectPoint(
    lat: number,
    lon: number,
    meta?: { nodeCode?: string; nodeName?: string },
  ): Promise<void> {
    selectedPoint.value = {
      lat,
      lon,
      nodeCode: meta?.nodeCode,
      nodeName: meta?.nodeName,
    }
    await refreshSelectedPoint()
  }

  function closePointDrawer(): void {
    selectedPoint.value = null
    pointData.value = null
    pointError.value = null
  }

  return {
    activeLayer,
    activeStandard,
    nodes,
    heatmap,
    timestamps,
    playbackIndex,
    isPlaying,
    playbackSpeed,
    selectedPoint,
    pointData,
    isLoadingMap,
    isLoadingPoint,
    mapError,
    pointError,
    isOfflineDemo,
    isLive,
    activeTimestamp,
    displayTimestamp,
    activeLegend,
    initialize,
    retryLiveApi,
    setStandard,
    setLayer,
    selectTimestampIndex,
    togglePlayback,
    setPlaybackSpeed,
    goLive,
    stopPlayback,
    inspectPoint,
    closePointDrawer,
  }
})
