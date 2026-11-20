<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  AttributionControl,
  Map as MapLibreMap,
  Marker,
  NavigationControl,
  setWorkerUrl,
  type ErrorEvent,
  type GeoJSONSource,
  type ImageSource,
  type MapMouseEvent,
} from 'maplibre-gl'
import maplibreWorkerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?url'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { FeatureCollection, Point } from 'geojson'
import { useThemeStore } from '../../stores/theme'
import type { SelectedPointMeta } from '../../stores/map'
import {
  NO_DATA_COLOR,
  NO_DATA_LABEL,
  type HeatmapResponse,
  type PublicNode,
} from '../../types/telemetry'
import {
  EMPTY_GEOJSON,
  OFFLINE_CARTO_LABELS,
  TRUJILLO_CORRIDOR_BOUNDS,
  TRUJILLO_DEFAULT_CENTER,
  TRUJILLO_DEFAULT_ZOOM,
  buildTrujilloMapStyle,
} from './trujilloBasemap'
import {
  buildHeatmapSurface,
  createEmptyHeatmapImage,
  lookupHeatmapCell,
  type HeatmapGridIndex,
} from './heatmapSurface'

interface HoverReadout {
  x: number
  y: number
  title: string
  category: string
  valueText: string
  color: string
  confidencePct: number
  hasData: boolean
  isNode: boolean
}

const props = defineProps<{
  heatmap: HeatmapResponse | null
  nodes: PublicNode[]
  selectedPoint: SelectedPointMeta | null
  unit: string
}>()

const emit = defineEmits<{
  'select-point': [payload: { lat: number; lon: number; nodeCode?: string; nodeName?: string }]
}>()

const themeStore = useThemeStore()
const containerEl = ref<HTMLElement | null>(null)
const hoverReadout = ref<HoverReadout | null>(null)
const isRasterOffline = ref<boolean>(typeof navigator !== 'undefined' ? !navigator.onLine : false)
const tooltipLeftPx = ref<number>(12)
const tooltipTopPx = ref<number>(12)

let map: MapLibreMap | null = null
let resizeObserver: ResizeObserver | null = null
let hasUserMoved = false
let lastFitWidth = 0
let heatmapGridIndex: HeatmapGridIndex | null = null
let emptyHeatmapImage: ImageData | null = null
const nodeMarkers: Marker[] = []
const cartoMarkers: Marker[] = []

function getEmptyHeatmapImage(): ImageData {
  if (!emptyHeatmapImage) {
    emptyHeatmapImage = createEmptyHeatmapImage()
  }
  return emptyHeatmapImage
}

function nodesToGeoJson(nodeList: PublicNode[]): FeatureCollection<Point> {
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

function selectedPointToGeoJson(sel: SelectedPointMeta | null): FeatureCollection {
  if (!sel) return EMPTY_GEOJSON
  return {
    type: 'FeatureCollection',
    features: [
      {
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Point',
          coordinates: [sel.lon, sel.lat],
        },
      },
    ],
  }
}

function syncNodeBadges(instance: MapLibreMap): void {
  while (nodeMarkers.length > 0) {
    nodeMarkers.pop()?.remove()
  }

  for (const node of props.nodes) {
    const tag = document.createElement('button')
    tag.type = 'button'
    tag.className = `map-node-label ${node.status === 'active' ? 'is-active' : 'is-inactive'}`
    tag.textContent = node.code
    tag.title = `${node.code} · ${node.name}`
    tag.addEventListener('click', (ev) => {
      ev.stopPropagation()
      emit('select-point', {
        lat: node.lat,
        lon: node.lon,
        nodeCode: node.code,
        nodeName: node.name,
      })
    })

    const marker = new Marker({
      element: tag,
      anchor: 'bottom',
      offset: [0, -10],
    })
      .setLngLat([node.lon, node.lat])
      .addTo(instance)

    nodeMarkers.push(marker)
  }
}

function syncCartoLabels(instance: MapLibreMap): void {
  while (cartoMarkers.length > 0) {
    cartoMarkers.pop()?.remove()
  }

  if (!isRasterOffline.value) return

  for (const spec of OFFLINE_CARTO_LABELS) {
    const labelEl = document.createElement('span')
    labelEl.className = `map-carto-label is-${spec.kind}`
    labelEl.textContent = spec.text
    labelEl.setAttribute('aria-hidden', 'true')

    const marker = new Marker({
      element: labelEl,
      anchor: 'center',
    })
      .setLngLat([spec.lon, spec.lat])
      .addTo(instance)

    cartoMarkers.push(marker)
  }
}

function syncGeoSources(instance: MapLibreMap): void {
  const surface = buildHeatmapSurface(props.heatmap)
  heatmapGridIndex = surface?.gridIndex ?? null

  const heatSrc = instance.getSource('heatmap-surface') as ImageSource | undefined
  if (heatSrc) {
    if (surface?.image && surface.coordinates) {
      heatSrc.updateImage({
        image: surface.image,
        coordinates: surface.coordinates,
      })
    } else {
      heatSrc.updateImage({
        image: getEmptyHeatmapImage(),
      })
    }
  }
  const nodesSrc = instance.getSource('active-nodes') as GeoJSONSource | undefined
  if (nodesSrc) {
    nodesSrc.setData(nodesToGeoJson(props.nodes))
  }
  const selectedSrc = instance.getSource('selected-point') as GeoJSONSource | undefined
  if (selectedSrc) {
    selectedSrc.setData(selectedPointToGeoJson(props.selectedPoint))
  }
}

function resolveHoverAtPoint(instance: MapLibreMap, event: MapMouseEvent): HoverReadout {
  const pt = event.point
  const hitBox: [[number, number], [number, number]] = [
    [pt.x - 6, pt.y - 6],
    [pt.x + 6, pt.y + 6],
  ]

  const nodeHits = instance.queryRenderedFeatures(hitBox, { layers: ['nodes-core'] })
  const topCell = lookupHeatmapCell(heatmapGridIndex, event.lngLat.lat, event.lngLat.lng)

  const topNode = nodeHits[0]?.properties as { code?: string; name?: string } | undefined

  const hasData = Boolean(topCell?.has_data)
  const rawVal = topCell?.value
  const valueText =
    hasData && rawVal !== null && rawVal !== undefined
      ? `${Number(rawVal).toFixed(1)} ${props.unit}`
      : NO_DATA_LABEL

  return {
    x: pt.x,
    y: pt.y,
    title: topNode ? `${topNode.code ?? ''} · ${topNode.name ?? ''}` : 'Celda IDW 50 m',
    category: hasData ? String(topCell?.category ?? '') : NO_DATA_LABEL,
    valueText,
    color: hasData && topCell?.color ? String(topCell.color) : NO_DATA_COLOR,
    confidencePct:
      hasData && typeof topCell?.confidence === 'number'
        ? Math.round(topCell.confidence * 100)
        : 0,
    hasData,
    isNode: Boolean(topNode),
  }
}

function fitCorridor(instance: MapLibreMap, duration: number): void {
  const hostWidth = containerEl.value?.clientWidth ?? 1200
  const isMobile = hostWidth <= 900
  const hasDesktopDrawer = !isMobile && Boolean(props.selectedPoint)
  lastFitWidth = hostWidth

  instance.fitBounds(TRUJILLO_CORRIDOR_BOUNDS, {
    padding: isMobile
      ? { top: 88, bottom: 138, left: 18, right: 18 }
      : {
          top: 48,
          bottom: 112,
          left: 268,
          right: hasDesktopDrawer ? 440 : 120,
        },
    maxZoom: isMobile ? 12.8 : 13.4,
    duration,
    essential: true,
  })
}

function ensureSelectedPointClear(instance: MapLibreMap, sel: SelectedPointMeta): void {
  const hostWidth = containerEl.value?.clientWidth ?? 0
  if (hostWidth <= 900) return

  const projected = instance.project([sel.lon, sel.lat])
  // Right drawer occupies 392px + 12px margin (404px); leave 80px extra margin so node badge + 600m rim stay unobstructed.
  const safeRightEdge = hostWidth - 484
  if (projected.x > safeRightEdge) {
    const deltaX = Math.round(projected.x - safeRightEdge)
    instance.panBy([deltaX, 0], { duration: 280, essential: true })
  }
}

function resetView(): void {
  if (!map) return
  hasUserMoved = false
  fitCorridor(map, 420)
}

defineExpose({ resetView })

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

  instance.addControl(
    new NavigationControl({ showCompass: true, showZoom: true }),
    'top-left',
  )
  instance.addControl(new AttributionControl({ compact: true }), 'top-left')

  instance.on('movestart', (evt) => {
    if ('originalEvent' in evt && evt.originalEvent) {
      hasUserMoved = true
    }
  })

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
    // Gracefully strip unreachable external raster tiles when offline while keeping local coastal vectors + heatmap active.
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

  instance.on('mousemove', (e: MapMouseEvent) => {
    if (typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches) {
      return
    }
    const hostWidth = el.clientWidth || 1200
    const hostHeight = el.clientHeight || 700
    const reservedRight = props.selectedPoint && hostWidth > 900 ? 416 : 16
    const tooltipWidth = 266
    const maxTooltipLeft = Math.max(12, hostWidth - reservedRight - tooltipWidth)
    const preferredLeft = e.point.x > maxTooltipLeft ? e.point.x - tooltipWidth - 8 : e.point.x + 14
    tooltipLeftPx.value = Math.max(10, Math.min(preferredLeft, maxTooltipLeft))
    tooltipTopPx.value = Math.max(10, Math.min(e.point.y - 18, hostHeight - 104))
    hoverReadout.value = resolveHoverAtPoint(instance, e)
  })

  instance.on('mouseout', () => {
    hoverReadout.value = null
  })

  instance.on('click', (e: MapMouseEvent) => {
    hoverReadout.value = null
    const hitBox: [[number, number], [number, number]] = [
      [e.point.x - 8, e.point.y - 8],
      [e.point.x + 8, e.point.y + 8],
    ]
    const nodeHits = instance.queryRenderedFeatures(hitBox, { layers: ['nodes-core'] })
    const topNode = nodeHits[0]?.properties as { code?: string; name?: string } | undefined
    if (topNode && topNode.code) {
      const matched = props.nodes.find((n) => n.code === topNode.code)
      if (matched) {
        emit('select-point', {
          lat: matched.lat,
          lon: matched.lon,
          nodeCode: matched.code,
          nodeName: matched.name,
        })
        return
      }
    }

    emit('select-point', {
      lat: Math.round(e.lngLat.lat * 1e6) / 1e6,
      lon: Math.round(e.lngLat.lng * 1e6) / 1e6,
    })
  })

  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      hoverReadout.value = null
      instance.resize()
      const nextWidth = el.clientWidth || 0
      if (!hasUserMoved && nextWidth > 0 && Math.abs(nextWidth - lastFitWidth) > 8) {
        fitCorridor(instance, 0)
      }
    })
    resizeObserver.observe(el)
  }
})

watch(
  [() => props.heatmap, () => props.nodes, () => props.selectedPoint],
  () => {
    hoverReadout.value = null
    if (!map) return
    syncGeoSources(map)
  },
)

watch(
  () => props.selectedPoint,
  (sel, prevSel) => {
    if (!map || !sel) return
    const isSame = prevSel && prevSel.lat === sel.lat && prevSel.lon === sel.lon
    if (!isSame) {
      ensureSelectedPointClear(map, sel)
    }
  },
)

watch(
  () => props.nodes,
  () => {
    if (!map) return
    syncNodeBadges(map)
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
  <div class="map-canvas-shell" :data-theme="themeStore.mode">
    <div ref="containerEl" class="maplibre-container" aria-label="Mapa interactivo de Trujillo y Huanchaco"></div>

    <div v-if="isRasterOffline" class="offline-tile-pill">
      Base geográfica local (Trujillo–Huanchaco)
    </div>

    <div
      v-if="hoverReadout"
      class="hover-tooltip"
      :style="{
        transform: `translate(${tooltipLeftPx}px, ${tooltipTopPx}px)`,
      }"
      role="status"
      aria-live="polite"
    >
      <div class="tooltip-head">
        <span class="tooltip-title">{{ hoverReadout.title }}</span>
        <span v-if="hoverReadout.hasData" class="tooltip-conf tabular">
          {{ hoverReadout.confidencePct }}% conf.
        </span>
      </div>
      <div class="tooltip-body">
        <span
          class="tooltip-swatch"
          :style="{ backgroundColor: hoverReadout.color }"
          aria-hidden="true"
        ></span>
        <strong class="tooltip-val tabular">{{ hoverReadout.valueText }}</strong>
        <span v-if="hoverReadout.hasData" class="tooltip-cat">· {{ hoverReadout.category }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-canvas-shell {
  position: relative;
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: var(--surface-ivory-2);
}

.maplibre-container {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  cursor: crosshair;
}

.offline-tile-pill {
  position: absolute;
  top: 12px;
  left: 52px;
  z-index: 6;
  padding: 4px 10px;
  background: var(--surface-ivory);
  color: var(--ink-secondary);
  border: 1px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  pointer-events: none;
}

.hover-tooltip {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 12;
  min-width: 185px;
  max-width: 260px;
  padding: 7px 10px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border: 1.5px solid var(--accent-amber-fill);
  pointer-events: none;
}

.tooltip-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--ink-muted-on-carbon);
}

.tooltip-body {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12.5px;
}

.tooltip-swatch {
  width: 11px;
  height: 11px;
  border: 1px solid #f7f1e3;
  flex-shrink: 0;
}

.tooltip-val {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
}

.tooltip-cat {
  font-size: 11.5px;
  color: var(--ink-muted-on-carbon);
}

:deep(.map-node-label) {
  padding: 2px 6px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition: background-color 0.16s var(--ease-expo);
}

:deep(.map-node-label:hover) {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
}

:deep(.map-carto-label) {
  pointer-events: none;
  user-select: none;
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  white-space: nowrap;
  color: rgba(67, 58, 46, 0.72);
  text-shadow: 0 1px 0 rgba(247, 241, 227, 0.85);
}

:deep(.map-carto-label.is-ocean) {
  font-style: italic;
  letter-spacing: 0.18em;
  color: rgba(31, 61, 79, 0.68);
}

.map-canvas-shell[data-theme='dark'] :deep(.map-carto-label) {
  color: rgba(203, 188, 163, 0.72);
  text-shadow: 0 1px 2px rgba(16, 13, 10, 0.9);
}

.map-canvas-shell[data-theme='dark'] :deep(.map-carto-label.is-ocean) {
  color: rgba(154, 184, 206, 0.72);
}

@media (max-width: 600px) {
  .offline-tile-pill {
    top: 10px;
    left: 48px;
    padding: 3px 8px;
    font-size: 9.5px;
    max-width: calc(100% - 58px);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
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

.map-canvas-shell[data-theme='dark'] :deep(.maplibregl-ctrl-group button) {
  background: var(--surface-ivory-2);
}

.map-canvas-shell[data-theme='dark'] :deep(.maplibregl-ctrl-group button .maplibregl-ctrl-icon) {
  filter: invert(0.94) sepia(0.2) brightness(1.15);
}

:deep(.maplibregl-ctrl-attrib) {
  border-radius: 0;
  font-family: var(--font-mono);
  font-size: 10px;
  background: var(--surface-ivory);
  color: var(--ink-secondary);
}
</style>
