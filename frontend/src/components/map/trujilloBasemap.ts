import type { Feature, FeatureCollection, Polygon } from 'geojson'
import type { StyleSpecification } from 'maplibre-gl'
import type { HeatmapCell, HeatmapResponse } from '../../types/telemetry'
import type { ThemeMode } from '../../stores/theme'

export const TRUJILLO_DEFAULT_CENTER: [number, number] = [-79.064, -8.101]
export const TRUJILLO_DEFAULT_ZOOM = 12.15
export const TRUJILLO_CORRIDOR_BOUNDS: [[number, number], [number, number]] = [
  [-79.132, -8.132],
  [-79.018, -8.068],
]

export interface CartoLabelSpec {
  id: string
  text: string
  lon: number
  lat: number
  kind: 'ocean' | 'sector'
}

export const OFFLINE_CARTO_LABELS: CartoLabelSpec[] = [
  { id: 'lbl-pacific', text: 'OCÉANO PACÍFICO', lon: -79.118, lat: -8.118, kind: 'ocean' },
  { id: 'lbl-huanchaco', text: 'HUANCHACO', lon: -79.113, lat: -8.069, kind: 'sector' },
  { id: 'lbl-centro', text: 'TRUJILLO CENTRO', lon: -79.042, lat: -8.092, kind: 'sector' },
  { id: 'lbl-larco', text: 'VÍCTOR LARCO', lon: -79.051, lat: -8.131, kind: 'sector' },
]

const METERS_PER_DEG_LAT = 111320

function buildGraticuleFeatures(): FeatureCollection {
  const features: Feature[] = []
  const minLon = -79.22
  const maxLon = -78.94
  const minLat = -8.2
  const maxLat = -8.02
  const step = 0.02

  for (let lon = minLon; lon <= maxLon + 1e-6; lon += step) {
    features.push({
      type: 'Feature',
      properties: { kind: 'graticule' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [lon, minLat],
          [lon, maxLat],
        ],
      },
    })
  }
  for (let lat = minLat; lat <= maxLat + 1e-6; lat += step) {
    features.push({
      type: 'Feature',
      properties: { kind: 'graticule' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [minLon, lat],
          [maxLon, lat],
        ],
      },
    })
  }
  return { type: 'FeatureCollection', features }
}

const LOCAL_COAST_WATER: FeatureCollection = {
  type: 'FeatureCollection',
  features: [
    {
      type: 'Feature',
      properties: { name: 'Océano Pacífico' },
      geometry: {
        type: 'Polygon',
        coordinates: [
          [
            [-79.35, -8.28],
            [-79.048, -8.28],
            [-79.042, -8.19],
            [-79.06, -8.152],
            [-79.092, -8.118],
            [-79.118, -8.084],
            [-79.124, -8.076],
            [-79.15, -8.042],
            [-79.19, -7.98],
            [-79.35, -7.98],
            [-79.35, -8.28],
          ],
        ],
      },
    },
  ],
}

const LOCAL_ARTERIALS: FeatureCollection = {
  type: 'FeatureCollection',
  features: [
    // Panamericana Norte + Av. Mansiche corridor
    {
      type: 'Feature',
      properties: { road: 'Av. Mansiche / Panamericana Norte' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [-79.118, -8.055],
          [-79.075, -8.082],
          [-79.045, -8.101],
          [-79.029, -8.111],
          [-79.01, -8.132],
          [-78.992, -8.158],
        ],
      },
    },
    // Carretera a Huanchaco
    {
      type: 'Feature',
      properties: { road: 'Carretera a Huanchaco' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [-79.045, -8.101],
          [-79.074, -8.094],
          [-79.104, -8.085],
          [-79.1206, -8.0794],
        ],
      },
    },
    // Av. América loop
    {
      type: 'Feature',
      properties: { road: 'Av. América' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [-79.042, -8.124],
          [-79.044, -8.106],
          [-79.0358, -8.096],
          [-79.018, -8.098],
          [-79.013, -8.115],
          [-79.026, -8.126],
          [-79.042, -8.124],
        ],
      },
    },
    // Av. Víctor Larco
    {
      type: 'Feature',
      properties: { road: 'Av. Larco' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [-79.0287, -8.1116],
          [-79.0395, -8.1204],
          [-79.054, -8.134],
        ],
      },
    },
    // Centro Histórico Av. España loop
    {
      type: 'Feature',
      properties: { road: 'Centro Histórico' },
      geometry: {
        type: 'LineString',
        coordinates: [
          [-79.033, -8.111],
          [-79.029, -8.107],
          [-79.024, -8.1115],
          [-79.028, -8.1165],
          [-79.033, -8.111],
        ],
      },
    },
  ],
}

const LOCAL_GRATICULE = buildGraticuleFeatures()

export const EMPTY_GEOJSON: FeatureCollection = {
  type: 'FeatureCollection',
  features: [],
}

export interface HeatmapCellProperties {
  value: number | null
  category: string
  color: string
  confidence: number
  has_data: boolean
  lat: number
  lon: number
}

export function cellsToGeoJson(
  response: HeatmapResponse | null,
): FeatureCollection<Polygon, HeatmapCellProperties> {
  if (!response || response.cells.length === 0) {
    return { type: 'FeatureCollection', features: [] }
  }

  const sizeM = response.cell_size_m > 0 ? response.cell_size_m : 50
  // Slight 5% overlap avoids hairline seams between adjacent 50m tiles under fractional WebGL scaling.
  const halfLat = (sizeM / METERS_PER_DEG_LAT) * 0.525

  const features: Array<Feature<Polygon, HeatmapCellProperties>> = response.cells.map(
    (cell: HeatmapCell) => {
      const cosLat = Math.max(0.01, Math.cos((cell.lat * Math.PI) / 180))
      const halfLon = (sizeM / (METERS_PER_DEG_LAT * cosLat)) * 0.525

      return {
        type: 'Feature',
        properties: {
          value: cell.value,
          category: cell.category,
          color: cell.color,
          confidence: cell.confidence,
          has_data: cell.has_data,
          lat: cell.lat,
          lon: cell.lon,
        },
        geometry: {
          type: 'Polygon',
          coordinates: [
            [
              [cell.lon - halfLon, cell.lat - halfLat],
              [cell.lon + halfLon, cell.lat - halfLat],
              [cell.lon + halfLon, cell.lat + halfLat],
              [cell.lon - halfLon, cell.lat + halfLat],
              [cell.lon - halfLon, cell.lat - halfLat],
            ],
          ],
        },
      }
    },
  )

  return {
    type: 'FeatureCollection',
    features,
  }
}

function buildCartoTileUrl(isDark: boolean): string {
  const customTemplate = (
    isDark ? import.meta.env.VITE_CARTO_TILE_URL_DARK : import.meta.env.VITE_CARTO_TILE_URL_LIGHT
  )?.trim()

  const baseUrl =
    customTemplate && customTemplate.length > 0
      ? customTemplate
      : isDark
        ? 'https://a.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}@2x.png'
        : 'https://a.basemaps.cartocdn.com/light_all/{z}/{x}/{y}@2x.png'

  const apiKey = import.meta.env.VITE_CARTO_API_KEY?.trim() ?? ''
  if (apiKey.length === 0) {
    return baseUrl
  }

  const encodedKey = encodeURIComponent(apiKey)
  if (baseUrl.includes('{api_key}')) {
    return baseUrl.replaceAll('{api_key}', encodedKey)
  }

  const separator = baseUrl.includes('?') ? '&' : '?'
  return `${baseUrl}${separator}api_key=${encodedKey}`
}

export function buildTrujilloMapStyle(
  theme: ThemeMode,
  enableExternalRaster: boolean,
): StyleSpecification {
  const isDark = theme === 'dark'

  const landBg = isDark ? '#1A1611' : '#EFE6D2'
  const oceanFill = isDark ? '#111A1C' : '#D5DFDC'
  const coastStroke = isDark ? '#4B3E2D' : '#8C7D65'
  const graticuleStroke = isDark ? '#2B241C' : '#D6C9AF'
  const arterialStroke = isDark ? '#544533' : '#B7A78B'
  const cartoTileUrl = buildCartoTileUrl(isDark)

  const sources: StyleSpecification['sources'] = {
    'local-ocean': {
      type: 'geojson',
      data: LOCAL_COAST_WATER,
    },
    'local-graticule': {
      type: 'geojson',
      data: LOCAL_GRATICULE,
    },
    'local-arterials': {
      type: 'geojson',
      data: LOCAL_ARTERIALS,
    },
    'heatmap-cells': {
      type: 'geojson',
      data: EMPTY_GEOJSON,
    },
    'active-nodes': {
      type: 'geojson',
      data: EMPTY_GEOJSON,
    },
    'selected-point': {
      type: 'geojson',
      data: EMPTY_GEOJSON,
    },
  }

  if (enableExternalRaster) {
    sources['carto-raster'] = {
      type: 'raster',
      tiles: [cartoTileUrl],
      tileSize: 256,
      attribution: '© OpenStreetMap · © CARTO',
    }
  }

  const layers: StyleSpecification['layers'] = [
    {
      id: 'warm-paper-bg',
      type: 'background',
      paint: {
        'background-color': landBg,
      },
    },
    {
      id: 'local-ocean-fill',
      type: 'fill',
      source: 'local-ocean',
      paint: {
        'fill-color': oceanFill,
        'fill-opacity': 0.9,
      },
    },
    {
      id: 'local-coastline',
      type: 'line',
      source: 'local-ocean',
      paint: {
        'line-color': coastStroke,
        'line-width': 1.4,
      },
    },
    {
      id: 'local-graticule-lines',
      type: 'line',
      source: 'local-graticule',
      paint: {
        'line-color': graticuleStroke,
        'line-width': 0.8,
        'line-dasharray': [3, 3],
      },
    },
    {
      id: 'local-arterial-lines',
      type: 'line',
      source: 'local-arterials',
      paint: {
        'line-color': arterialStroke,
        'line-width': 2.2,
      },
    },
  ]

  if (enableExternalRaster) {
    layers.push({
      id: 'carto-raster-layer',
      type: 'raster',
      source: 'carto-raster',
      paint: {
        'raster-opacity': 0.72,
        'raster-saturation': isDark ? -0.42 : -0.45,
        'raster-brightness-min': isDark ? 0.0 : 0.02,
        'raster-brightness-max': isDark ? 0.88 : 1.0,
        'raster-contrast': isDark ? 0.1 : 0.08,
      },
    })
  }

  layers.push(
    {
      id: 'heatmap-cells-fill',
      type: 'fill',
      source: 'heatmap-cells',
      paint: {
        'fill-color': ['get', 'color'],
        'fill-opacity': ['case', ['==', ['get', 'has_data'], false], 0.34, 0.72],
        'fill-antialias': false,
      },
    },
    {
      id: 'nodes-outer-halo',
      type: 'circle',
      source: 'active-nodes',
      paint: {
        'circle-radius': 11,
        'circle-color': isDark ? '#100D0A' : '#191D20',
        'circle-opacity': 0.32,
      },
    },
    {
      id: 'nodes-core',
      type: 'circle',
      source: 'active-nodes',
      paint: {
        'circle-radius': 6.5,
        'circle-color': [
          'case',
          ['==', ['get', 'status'], 'active'],
          isDark ? '#E6AA3C' : '#F7F1E3',
          isDark ? '#2B241C' : '#EFE6D2',
        ],
        'circle-stroke-width': 2.2,
        'circle-stroke-color': [
          'case',
          ['==', ['get', 'status'], 'active'],
          isDark ? '#BF4128' : '#191D20',
          isDark ? '#6B563A' : '#433A2E',
        ],
      },
    },
    {
      id: 'selected-point-ring',
      type: 'circle',
      source: 'selected-point',
      paint: {
        'circle-radius': 13,
        'circle-color': 'transparent',
        'circle-stroke-width': 2.4,
        'circle-stroke-color': isDark ? '#BF4128' : '#B83A28',
      },
    },
    {
      id: 'selected-point-dot',
      type: 'circle',
      source: 'selected-point',
      paint: {
        'circle-radius': 5,
        'circle-color': isDark ? '#BF4128' : '#B83A28',
        'circle-stroke-width': 2,
        'circle-stroke-color': isDark ? '#100D0A' : '#F7F1E3',
      },
    },
  )

  return {
    version: 8,
    sources,
    layers,
  }
}
