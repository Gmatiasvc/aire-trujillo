<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Map as MapLibreMap,
  Marker,
  NavigationControl,
  setWorkerUrl,
  type ErrorEvent,
  type GeoJSONSource,
  type MapMouseEvent,
} from 'maplibre-gl'
import maplibreWorkerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?url'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { FeatureCollection, Point } from 'geojson'
import UiIcon from '../UiIcon.vue'
import { useThemeStore } from '../../stores/theme'
import type { Node } from '../../types/telemetry'
import {
  EMPTY_GEOJSON,
  OFFLINE_CARTO_LABELS,
  TRUJILLO_CORRIDOR_BOUNDS,
  TRUJILLO_DEFAULT_CENTER,
  TRUJILLO_DEFAULT_ZOOM,
  buildTrujilloMapStyle,
} from '../map/trujilloBasemap'

const props = defineProps<{
  nodes: Node[]
  selectedNodeId: string | null
  draftLat: number | null
  draftLon: number | null
}>()

const emit = defineEmits<{
  'pick-coordinate': [payload: { lat: number; lon: number }]
  'select-node': [nodeId: string]
}>()

const themeStore = useThemeStore()
const containerEl = ref<HTMLElement | null>(null)
const isRasterOffline = ref<boolean>(typeof navigator !== 'undefined' ? !navigator.onLine : false)

let map: MapLibreMap | null = null
let resizeObserver: ResizeObserver | null = null
const nodeMarkers: Marker[] = []
const cartoMarkers: Marker[] = []

const hasValidDraft = computed(
  () =>
    props.draftLat !== null &&
    props.draftLon !== null &&
    Number.isFinite(props.draftLat) &&
    Number.isFinite(props.draftLon) &&
    props.draftLat >= -90 &&
    props.draftLat <= 90 &&
    props.draftLon >= -180 &&
    props.draftLon <= 180,
)

function nodesToGeoJson(nodeList: Node[]): FeatureCollection<Point> {
  return {
    type: 'FeatureCollection',
    features: nodeList.map((n) => ({
      type: 'Feature',
      properties: {
        id: n.id,
        code: n.code,
        name: n.name,
        status: n.status,
      },
      geometry: {
        type: 'Point',
        coordinates: [n.lon, n.lat],
      },
    })),
  }
}

function draftToGeoJson(): FeatureCollection {
  if (!hasValidDraft.value || props.draftLat === null || props.draftLon === null) {
    return EMPTY_GEOJSON
  }
  return {
    type: 'FeatureCollection',
    features: [
      {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Point',
          coordinates: [props.draftLon, props.draftLat],
        },
      },
    ],
  }
}

function syncCartoLabels(instance: MapLibreMap): void {
  while (cartoMarkers.length > 0) {
    cartoMarkers.pop()?.remove()
  }
  if (!isRasterOffline.value) return

  for (const spec of OFFLINE_CARTO_LABELS) {
    const labelEl = document.createElement('span')
    labelEl.className = `picker-carto-label is-${spec.kind}`
    labelEl.textContent = spec.text
    labelEl.setAttribute('aria-hidden', 'true')

    const marker = new Marker({ element: labelEl, anchor: 'center' })
      .setLngLat([spec.lon, spec.lat])
      .addTo(instance)

    cartoMarkers.push(marker)
  }
}

function syncNodeBadges(instance: MapLibreMap): void {
  while (nodeMarkers.length > 0) {
    nodeMarkers.pop()?.remove()
  }

  // Paint unselected nodes first and the selected station last so its label never gets occluded.
  const orderedNodes = [...props.nodes].sort((a, b) => {
    const aSel = a.id === props.selectedNodeId ? 1 : 0
    const bSel = b.id === props.selectedNodeId ? 1 : 0
    return aSel - bSel
  })

  for (const node of orderedNodes) {
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = 'picker-node-tag'
    if (node.id === props.selectedNodeId) {
      btn.classList.add('is-selected')
    }
    if (node.status === 'inactive') {
      btn.classList.add('is-inactive')
    }
    btn.textContent = node.code
    btn.title = `${node.code} · ${node.name}`
    btn.setAttribute('aria-label', `Seleccionar nodo ${node.code}: ${node.name}`)
    btn.addEventListener('click', (evt) => {
      evt.stopPropagation()
      emit('select-node', node.id)
    })

    const marker = new Marker({ element: btn, anchor: 'bottom', offset: [0, -7] })
      .setLngLat([node.lon, node.lat])
      .addTo(instance)

    nodeMarkers.push(marker)
  }
}

function syncGeoSources(instance: MapLibreMap): void {
  const nodesSrc = instance.getSource('active-nodes') as GeoJSONSource | undefined
  if (nodesSrc) {
    nodesSrc.setData(nodesToGeoJson(props.nodes))
  }
  const selectedSrc = instance.getSource('selected-point') as GeoJSONSource | undefined
  if (selectedSrc) {
    selectedSrc.setData(draftToGeoJson())
  }
}

function fitCorridor(instance: MapLibreMap, duration = 360): void {
  const isNarrow = (containerEl.value?.clientWidth ?? 420) < 380
  instance.fitBounds(TRUJILLO_CORRIDOR_BOUNDS, {
    padding: isNarrow
      ? { top: 32, bottom: 38, left: 52, right: 20 }
      : { top: 28, bottom: 28, left: 32, right: 24 },
    maxZoom: isNarrow ? 12.6 : 13.0,
    duration,
    essential: true,
  })
}

function handleCenterCorridor(): void {
  if (!map) return
  fitCorridor(map, 360)
}

onMounted(() => {
  const el = containerEl.value
  if (!el) return

  setWorkerUrl(maplibreWorkerUrl)

  const instance = new MapLibreMap({
    container: el,
    style: buildTrujilloMapStyle(themeStore.mode, !isRasterOffline.value),
    center: TRUJILLO_DEFAULT_CENTER,
    zoom: TRUJILLO_DEFAULT_ZOOM,
    minZoom: 10.5,
    maxZoom: 17.5,
    attributionControl: false,
  })
  map = instance

  instance.addControl(new NavigationControl({ showCompass: false, showZoom: true }), 'top-left')

  instance.on('load', () => {
    syncGeoSources(instance)
    syncNodeBadges(instance)
    syncCartoLabels(instance)
    fitCorridor(instance, 0)
  })

  instance.on('style.load', () => {
    syncGeoSources(instance)
    syncCartoLabels(instance)
  })

  instance.on('error', (evt: ErrorEvent) => {
    const sourceId = (evt as ErrorEvent & { sourceId?: string }).sourceId
    if (sourceId === 'carto-raster' && !isRasterOffline.value) {
      isRasterOffline.value = true
      if (instance.getLayer('carto-raster-layer')) {
        instance.removeLayer('carto-raster-layer')
      }
      if (instance.getSource('carto-raster')) {
        instance.removeSource('carto-raster')
      }
      syncCartoLabels(instance)
    }
  })

  instance.on('click', (e: MapMouseEvent) => {
    const lat = Math.round(e.lngLat.lat * 1e6) / 1e6
    const lon = Math.round(e.lngLat.lng * 1e6) / 1e6
    emit('pick-coordinate', { lat, lon })
  })

  resizeObserver = new ResizeObserver(() => {
    instance.resize()
  })
  resizeObserver.observe(el)
})

watch(
  () => [props.nodes, props.selectedNodeId],
  () => {
    if (!map) return
    syncGeoSources(map)
    syncNodeBadges(map)
  },
  { deep: true },
)

watch(
  () => [props.draftLat, props.draftLon],
  () => {
    if (!map) return
    syncGeoSources(map)
  },
)

watch(
  () => themeStore.mode,
  (nextTheme) => {
    if (!map) return
    map.setStyle(buildTrujilloMapStyle(nextTheme, !isRasterOffline.value))
  },
)

onBeforeUnmount(() => {
  while (nodeMarkers.length > 0) {
    nodeMarkers.pop()?.remove()
  }
  while (cartoMarkers.length > 0) {
    cartoMarkers.pop()?.remove()
  }
  resizeObserver?.disconnect()
  resizeObserver = null
  map?.remove()
  map = null
})
</script>

<template>
  <div class="coord-picker-card" :data-theme="themeStore.mode">
    <div class="picker-map-wrap">
      <div
        ref="containerEl"
        class="picker-map-canvas"
        aria-label="Selector interactivo de coordenadas sobre Trujillo y Huanchaco"
      ></div>

      <div v-if="isRasterOffline" class="offline-tile-pill">
        Cartografía local Trujillo–Huanchaco
      </div>

      <button
        type="button"
        class="center-map-btn"
        aria-label="Centrar mapa de selección"
        @click="handleCenterCorridor"
      >
        <UiIcon name="locate" :size="13" />
        <span>Centrar</span>
      </button>
    </div>

    <div class="picker-foot">
      <span class="picker-hint">Haz clic en el mapa para fijar la coordenada</span>
      <span class="picker-coords tabular">
        {{
          hasValidDraft && draftLat !== null && draftLon !== null
            ? `${draftLat.toFixed(6)}, ${draftLon.toFixed(6)}`
            : 'Sin coordenada válida'
        }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.coord-picker-card {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
}

.picker-map-wrap {
  position: relative;
  height: 230px;
  width: 100%;
  overflow: hidden;
  background: var(--surface-ivory-2);
}

.picker-map-canvas {
  width: 100%;
  height: 100%;
  cursor: crosshair;
}

.offline-tile-pill {
  position: absolute;
  bottom: 8px;
  left: 8px;
  z-index: 4;
  padding: 3px 8px;
  border: 1px solid var(--accent-amber);
  background: var(--surface-carbon);
  color: var(--accent-amber-fill);
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: 0.04em;
  pointer-events: none;
}

.center-map-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 4;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 9px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.16s var(--ease-expo),
    border-color 0.16s var(--ease-expo),
    color 0.16s var(--ease-expo);
}

.center-map-btn:hover {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.center-map-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.picker-foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 6px 12px;
  padding: 7px 10px;
  border-top: 1px solid var(--rule-soft);
  background: var(--surface-ivory-2);
}

.picker-hint {
  font-size: 11.5px;
  color: var(--ink-secondary);
}

.picker-coords {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-primary);
}

:deep(.picker-node-tag) {
  z-index: 1;
  padding: 1px 5px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition:
    background-color 0.15s var(--ease-expo),
    border-color 0.15s var(--ease-expo);
}

:deep(.picker-node-tag.is-inactive) {
  background: var(--surface-ivory-2);
  color: var(--ink-secondary);
}

:deep(.picker-node-tag.is-selected) {
  z-index: 5;
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

:deep(.picker-node-tag:hover) {
  z-index: 6;
  border-color: var(--accent-amber);
}

:deep(.picker-carto-label) {
  pointer-events: none;
  user-select: none;
  font-family: var(--font-mono);
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  white-space: nowrap;
  color: rgba(67, 58, 46, 0.72);
  text-shadow: 0 1px 0 rgba(247, 241, 227, 0.85);
}

:deep(.picker-carto-label.is-ocean) {
  font-style: italic;
  color: rgba(31, 61, 79, 0.68);
}

.coord-picker-card[data-theme='dark'] :deep(.picker-carto-label) {
  color: rgba(203, 188, 163, 0.72);
  text-shadow: 0 1px 2px rgba(16, 13, 10, 0.9);
}

.coord-picker-card[data-theme='dark'] :deep(.picker-carto-label.is-ocean) {
  color: rgba(154, 184, 206, 0.72);
}

:deep(.maplibregl-ctrl-group) {
  border-radius: 0;
  border: 1.5px solid var(--rule-color);
  box-shadow: none;
  background: var(--surface-ivory);
}

:deep(.maplibregl-ctrl-group button) {
  border-radius: 0;
  background: var(--surface-ivory);
}

.coord-picker-card[data-theme='dark'] :deep(.maplibregl-ctrl-group button) {
  background: var(--surface-ivory-2);
}

.coord-picker-card[data-theme='dark'] :deep(.maplibregl-ctrl-group button .maplibregl-ctrl-icon) {
  filter: invert(0.94) sepia(0.2) brightness(1.15);
}

@media (max-width: 720px) {
  .picker-map-wrap {
    height: 196px;
  }
}
</style>
