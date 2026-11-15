<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import UiIcon from '../components/UiIcon.vue'
import MapCanvas from '../components/map/MapCanvas.vue'
import MapLegend from '../components/map/MapLegend.vue'
import MapTopToolbar from '../components/map/MapTopToolbar.vue'
import PlaybackBar from '../components/map/PlaybackBar.vue'
import PointInspectionDrawer from '../components/map/PointInspectionDrawer.vue'
import { useMapStore, type PlaybackSpeed } from '../stores/map'
import type { AirQualityStandard, LayerKey } from '../types/telemetry'

const mapStore = useMapStore()
const mapCanvasRef = ref<InstanceType<typeof MapCanvas> | null>(null)

function handleSelectStandard(standard: AirQualityStandard): void {
  void mapStore.setStandard(standard)
}

function handleSelectLayer(layer: LayerKey): void {
  void mapStore.setLayer(layer)
}

function handleResetView(): void {
  mapCanvasRef.value?.resetView()
}

function handleSelectPoint(payload: {
  lat: number
  lon: number
  nodeCode?: string
  nodeName?: string
}): void {
  void mapStore.inspectPoint(payload.lat, payload.lon, {
    nodeCode: payload.nodeCode,
    nodeName: payload.nodeName,
  })
}

function handleSetSpeed(speed: PlaybackSpeed): void {
  mapStore.setPlaybackSpeed(speed)
}

function handleScrub(index: number): void {
  void mapStore.selectTimestampIndex(index)
}

onMounted(() => {
  void mapStore.initialize()
})

onBeforeUnmount(() => {
  mapStore.stopPlayback()
})
</script>

<template>
  <main class="map-viewport">
    <MapTopToolbar
      :active-layer="mapStore.activeLayer"
      :active-standard="mapStore.activeStandard"
      :active-node-count="mapStore.nodes.length"
      :display-timestamp="mapStore.displayTimestamp"
      :is-live="mapStore.isLive"
      @select-standard="handleSelectStandard"
      @select-layer="handleSelectLayer"
      @reset-view="handleResetView"
    />

    <div v-if="mapStore.mapError" class="status-banner is-error" role="alert">
      <div class="banner-copy">
        <UiIcon name="alert" :size="15" />
        <span>{{ mapStore.mapError }}</span>
      </div>
      <button type="button" class="banner-btn" @click="mapStore.retryLiveApi">
        <UiIcon name="refresh" :size="13" />
        Reintentar
      </button>
    </div>

    <div v-else-if="mapStore.isOfflineDemo" class="status-banner is-demo" role="status">
      <div class="banner-copy">
        <UiIcon name="alert" :size="15" />
        <span>Sin conexión al backend :8080 — vista de muestra</span>
      </div>
      <button type="button" class="banner-btn" @click="mapStore.retryLiveApi">
        <UiIcon name="refresh" :size="13" />
        Reintentar API
      </button>
    </div>

    <div class="map-stage">
      <MapCanvas
        ref="mapCanvasRef"
        :heatmap="mapStore.heatmap"
        :nodes="mapStore.nodes"
        :selected-point="mapStore.selectedPoint"
        :unit="mapStore.activeLegend.unit"
        @select-point="handleSelectPoint"
      />

      <div class="legend-dock" :class="{ 'is-drawer-open': Boolean(mapStore.selectedPoint) }">
        <MapLegend :legend="mapStore.activeLegend" />
      </div>

      <PointInspectionDrawer
        :selected-point="mapStore.selectedPoint"
        :point-data="mapStore.pointData"
        :is-loading="mapStore.isLoadingPoint"
        :error="mapStore.pointError"
        :layer-title="mapStore.activeLegend.title"
        :active-timestamp="mapStore.displayTimestamp"
        @close="mapStore.closePointDrawer"
        @retry="
          mapStore.selectedPoint &&
            mapStore.inspectPoint(mapStore.selectedPoint.lat, mapStore.selectedPoint.lon, {
              nodeCode: mapStore.selectedPoint.nodeCode,
              nodeName: mapStore.selectedPoint.nodeName,
            })
        "
      />

      <div class="playback-dock">
        <PlaybackBar
          :timestamps="mapStore.timestamps"
          :playback-index="mapStore.playbackIndex"
          :is-playing="mapStore.isPlaying"
          :playback-speed="mapStore.playbackSpeed"
          :is-live="mapStore.isLive"
          :display-timestamp="mapStore.displayTimestamp"
          :is-loading-map="mapStore.isLoadingMap"
          @toggle-play="mapStore.togglePlayback"
          @set-speed="handleSetSpeed"
          @scrub-index="handleScrub"
          @go-live="mapStore.goLive"
        />
      </div>
    </div>
  </main>
</template>

<style scoped>
.map-viewport {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100svh;
  min-height: 540px;
  overflow: hidden;
  background: var(--bed-bg);
}

.status-banner {
  position: relative;
  z-index: 18;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 16px;
  border-bottom: 1.5px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
}

.status-banner.is-demo {
  background: var(--accent-amber-fill);
  color: #181a1c;
}

.status-banner.is-error {
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.banner-copy {
  display: flex;
  align-items: center;
  gap: 8px;
}

.banner-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border: 1px solid currentColor;
  background: transparent;
  color: inherit;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
}

.banner-btn:hover {
  background: rgba(25, 29, 32, 0.14);
}

.map-stage {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
}

.legend-dock {
  position: absolute;
  left: 12px;
  bottom: 94px;
  z-index: 10;
}

.playback-dock {
  position: absolute;
  left: 12px;
  right: 12px;
  bottom: 12px;
  z-index: 14;
}

@media (max-width: 900px) {
  .legend-dock.is-drawer-open {
    display: none;
  }

  .legend-dock {
    left: 10px;
    bottom: 128px;
  }

  .playback-dock {
    left: 10px;
    right: 10px;
    bottom: 10px;
  }
}

@media (max-width: 600px) {
  .status-banner {
    padding: 5px 10px;
    font-size: 10.5px;
    gap: 8px;
  }

  .banner-btn {
    padding: 3px 8px;
    font-size: 10px;
    white-space: nowrap;
    flex-shrink: 0;
  }
}
</style>
