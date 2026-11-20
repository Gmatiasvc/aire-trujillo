import {
  LAYER_AIR_QUALITY,
  METRIC_KEYS,
  METRIC_SPECS,
  NO_DATA_COLOR,
  NO_DATA_LABEL,
  type AdminNodesResponse,
  type AdminReadingItem,
  type AdminReadingsResponse,
  type AirQualityStandard,
  type BatchReadingsRequest,
  type BatchReadingsResponse,
  type CreateNodeRequest,
  type HeatmapCell,
  type HeatmapResponse,
  type LayerKey,
  type MetricKey,
  type Node,
  type PlaybackTimestampsResponse,
  type PointResponse,
  type PointSeriesItem,
  type PublicNode,
  type PublicNodesResponse,
  type UpdateNodeRequest,
} from '../types/telemetry'

interface SampleStation {
  node: PublicNode
  baseInca: number
  baseEpa: number
  baseMetrics: Record<MetricKey, number>
}

const SAMPLE_STATIONS: SampleStation[] = [
  {
    node: {
      id: 'demo-trj-01',
      code: 'TRJ-01',
      name: 'Plaza de Armas',
      lat: -8.1116,
      lon: -79.0287,
      status: 'active',
    },
    baseInca: 42,
    baseEpa: 56,
    baseMetrics: {
      pm25: 14.2,
      pm10: 28.5,
      co: 1.1,
      no2: 32.0,
      so2: 11.4,
      h2s: 4.8,
      co2: 435,
      voc_index: 108,
      noise_dba: 61.4,
      temperature_c: 21.4,
      humidity_pct: 78.2,
      pressure_hpa: 1013.4,
    },
  },
  {
    node: {
      id: 'demo-trj-02',
      code: 'TRJ-02',
      name: 'Óvalo Papal',
      lat: -8.1062,
      lon: -79.0358,
      status: 'active',
    },
    baseInca: 74,
    baseEpa: 96,
    baseMetrics: {
      pm25: 33.6,
      pm10: 66.4,
      co: 2.6,
      no2: 71.2,
      so2: 20.8,
      h2s: 8.6,
      co2: 528,
      voc_index: 172,
      noise_dba: 74.0,
      temperature_c: 22.3,
      humidity_pct: 73.8,
      pressure_hpa: 1012.8,
    },
  },
  {
    node: {
      id: 'demo-trj-03',
      code: 'TRJ-03',
      name: 'Urb. California / Av. Larco',
      lat: -8.1204,
      lon: -79.0395,
      status: 'active',
    },
    baseInca: 112,
    baseEpa: 138,
    baseMetrics: {
      pm25: 54.0,
      pm10: 108.5,
      co: 4.2,
      no2: 118.4,
      so2: 34.0,
      h2s: 14.5,
      co2: 642,
      voc_index: 238,
      noise_dba: 78.4,
      temperature_c: 23.0,
      humidity_pct: 70.1,
      pressure_hpa: 1012.5,
    },
  },
  {
    node: {
      id: 'demo-hch-01',
      code: 'HCH-01',
      name: 'Muelle Huanchaco',
      lat: -8.0794,
      lon: -79.1206,
      status: 'active',
    },
    baseInca: 31,
    baseEpa: 41,
    baseMetrics: {
      pm25: 7.4,
      pm10: 21.0,
      co: 0.7,
      no2: 18.2,
      so2: 6.1,
      h2s: 3.5,
      co2: 418,
      voc_index: 82,
      noise_dba: 54.0,
      temperature_c: 19.8,
      humidity_pct: 84.6,
      pressure_hpa: 1014.1,
    },
  },
]

const STEP_MINUTES = 5
const SLOTS_24H = 288

function buildSampleTimestampList(): string[] {
  const now = new Date()
  now.setSeconds(0, 0)
  const mins = now.getMinutes()
  now.setMinutes(mins - (mins % STEP_MINUTES))

  const timestamps: string[] = []
  for (let i = SLOTS_24H - 1; i >= 0; i--) {
    const t = new Date(now.getTime() - i * STEP_MINUTES * 60_000)
    timestamps.push(t.toISOString())
  }
  return timestamps
}

const SAMPLE_TIMESTAMPS = buildSampleTimestampList()

function distanceMeters(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const toRad = Math.PI / 180
  const dLat = (lat2 - lat1) * toRad
  const dLon = (lon2 - lon1) * toRad
  const a =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(lat1 * toRad) * Math.cos(lat2 * toRad) * Math.sin(dLon / 2) ** 2
  return 6371000 * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function blendHexColors(hexA: string, hexB: string, t: number): string {
  if (t <= 0) return hexA.toUpperCase()
  if (t >= 1) return hexB.toUpperCase()
  const ar = parseInt(hexA.slice(1, 3), 16)
  const ag = parseInt(hexA.slice(3, 5), 16)
  const ab = parseInt(hexA.slice(5, 7), 16)
  const br = parseInt(hexB.slice(1, 3), 16)
  const bg = parseInt(hexB.slice(3, 5), 16)
  const bb = parseInt(hexB.slice(5, 7), 16)
  const r = Math.round(ar + t * (br - ar))
  const g = Math.round(ag + t * (bg - ag))
  const b = Math.round(ab + t * (bb - ab))
  const toHex = (n: number): string => Math.max(0, Math.min(255, n)).toString(16).padStart(2, '0').toUpperCase()
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

function blendHexToGray(baseHex: string, confidence: number): string {
  return blendHexColors(NO_DATA_COLOR, baseHex, confidence)
}

interface ColorStop {
  value: number
  hex: string
}

function interpolateStops(value: number, stops: ColorStop[]): string {
  if (stops.length === 0) return NO_DATA_COLOR
  if (value <= stops[0]!.value) return stops[0]!.hex
  const last = stops[stops.length - 1]!
  if (value >= last.value) return last.hex
  for (let i = 0; i < stops.length - 1; i++) {
    const a = stops[i]!
    const b = stops[i + 1]!
    if (value <= b.value) {
      const span = b.value - a.value
      return span <= 0 ? b.hex : blendHexColors(a.hex, b.hex, (value - a.value) / span)
    }
  }
  return last.hex
}

function fourTierStops(t1: number, t2: number, t3: number): ColorStop[] {
  return [
    { value: t1 * 0.5, hex: '#22C55E' },
    { value: (t1 + t2) * 0.5, hex: '#EAB308' },
    { value: (t2 + t3) * 0.5, hex: '#F97316' },
    { value: t3 * 1.35, hex: '#EF4444' },
  ]
}

const INCA_STOPS: ColorStop[] = [
  { value: 25, hex: '#22C55E' },
  { value: 75, hex: '#EAB308' },
  { value: 125, hex: '#F97316' },
  { value: 200, hex: '#EF4444' },
]

const EPA_STOPS: ColorStop[] = [
  { value: 25, hex: '#00E400' },
  { value: 75, hex: '#FFFF00' },
  { value: 125, hex: '#FF7E00' },
  { value: 175, hex: '#FF0000' },
  { value: 250, hex: '#8F3F97' },
  { value: 400, hex: '#7E0023' },
]

const METRIC_STOPS: Record<MetricKey, ColorStop[]> = {
  pm25: fourTierStops(25, 50, 100),
  pm10: fourTierStops(50, 100, 200),
  co: fourTierStops(10, 30, 45),
  no2: fourTierStops(100, 200, 400),
  so2: fourTierStops(125, 250, 500),
  h2s: fourTierStops(75, 150, 300),
  voc_index: fourTierStops(100, 200, 400),
  co2: [
    { value: 600, hex: '#22C55E' },
    { value: 1500, hex: '#EAB308' },
    { value: 2500, hex: '#EF4444' },
  ],
  noise_dba: [
    { value: 45, hex: '#22C55E' },
    { value: 70, hex: '#EAB308' },
    { value: 90, hex: '#EF4444' },
  ],
  temperature_c: [
    { value: 12, hex: '#3B82F6' },
    { value: 21, hex: '#22C55E' },
    { value: 29, hex: '#F97316' },
    { value: 36, hex: '#EF4444' },
  ],
  humidity_pct: [
    { value: 20, hex: '#EAB308' },
    { value: 52.5, hex: '#22C55E' },
    { value: 85, hex: '#3B82F6' },
  ],
  pressure_hpa: [
    { value: 998, hex: '#3B82F6' },
    { value: 1012.5, hex: '#22C55E' },
    { value: 1026, hex: '#F97316' },
  ],
}

function interpolateSampleColor(layer: LayerKey, standard: AirQualityStandard, value: number): string {
  if (layer === LAYER_AIR_QUALITY) {
    return interpolateStops(value, standard === 'epa' ? EPA_STOPS : INCA_STOPS)
  }
  return interpolateStops(value, METRIC_STOPS[layer])
}

function pickSampleStyle(layer: LayerKey, standard: AirQualityStandard, value: number): { category: string; color: string } {
  if (layer === LAYER_AIR_QUALITY) {
    if (standard === 'epa') {
      if (value <= 50) return { category: 'Good', color: '#00E400' }
      if (value <= 100) return { category: 'Moderate', color: '#FFFF00' }
      if (value <= 150) return { category: 'Unhealthy for Sensitive Groups', color: '#FF7E00' }
      if (value <= 200) return { category: 'Unhealthy', color: '#FF0000' }
      if (value <= 300) return { category: 'Very Unhealthy', color: '#8F3F97' }
      return { category: 'Hazardous', color: '#7E0023' }
    }
    if (value <= 50) return { category: 'Buena', color: '#22C55E' }
    if (value <= 100) return { category: 'Moderada', color: '#EAB308' }
    if (value <= 150) return { category: 'Mala', color: '#F97316' }
    return { category: 'Peligrosa', color: '#EF4444' }
  }

  switch (layer) {
    case 'co2':
      if (value <= 1000) return { category: 'Normal', color: '#22C55E' }
      if (value <= 2000) return { category: 'Elevado', color: '#EAB308' }
      return { category: 'Alto', color: '#EF4444' }
    case 'noise_dba':
      if (value <= 60) return { category: 'Confortable', color: '#22C55E' }
      if (value <= 80) return { category: 'Moderado', color: '#EAB308' }
      return { category: 'Alto', color: '#EF4444' }
    case 'temperature_c':
      if (value < 16) return { category: 'Frío', color: '#3B82F6' }
      if (value <= 26) return { category: 'Templado', color: '#22C55E' }
      if (value <= 32) return { category: 'Cálido', color: '#F97316' }
      return { category: 'Caluroso', color: '#EF4444' }
    case 'humidity_pct':
      if (value < 30) return { category: 'Seco', color: '#EAB308' }
      if (value <= 75) return { category: 'Confortable', color: '#22C55E' }
      return { category: 'Húmedo', color: '#3B82F6' }
    case 'pressure_hpa':
      if (value < 1005) return { category: 'Baja', color: '#3B82F6' }
      if (value <= 1020) return { category: 'Normal', color: '#22C55E' }
      return { category: 'Alta', color: '#F97316' }
    default: {
      const limits: Record<string, [number, number, number]> = {
        pm25: [25, 50, 100],
        pm10: [50, 100, 200],
        co: [10, 30, 45],
        no2: [100, 200, 400],
        so2: [125, 250, 500],
        h2s: [75, 150, 300],
        voc_index: [100, 200, 400],
      }
      const [t1, t2, t3] = limits[layer] ?? [25, 50, 100]
      if (value <= t1) return { category: 'Bueno', color: '#22C55E' }
      if (value <= t2) return { category: 'Moderado', color: '#EAB308' }
      if (value <= t3) return { category: 'Alto', color: '#F97316' }
      return { category: 'Muy alto', color: '#EF4444' }
    }
  }
}

function rawDiurnalCurve(idx: number): number {
  const phase = (idx / Math.max(1, SLOTS_24H - 1)) * Math.PI * 2
  return 0.78 + 0.34 * Math.sin(phase - 0.8) + 0.1 * Math.cos(phase * 2)
}

const LIVE_CURVE_BASELINE = rawDiurnalCurve(SLOTS_24H - 1)

function sampleFactorForTimestamp(at?: string): number {
  if (!at) return 1
  const idx = SAMPLE_TIMESTAMPS.indexOf(at)
  if (idx === -1) return 1
  return rawDiurnalCurve(idx) / LIVE_CURVE_BASELINE
}

function stationSampleValue(
  station: SampleStation,
  layer: LayerKey,
  standard: AirQualityStandard,
  factor: number,
): number {
  const base =
    layer === LAYER_AIR_QUALITY
      ? standard === 'epa'
        ? station.baseEpa
        : station.baseInca
      : station.baseMetrics[layer]

  if (layer === 'pressure_hpa') {
    return Math.round((base + (factor - 1) * 4) * 100) / 100
  }
  if (layer === 'temperature_c') {
    return Math.round((base + (factor - 1) * 5.2) * 100) / 100
  }
  if (layer === 'humidity_pct') {
    return Math.round(Math.min(98, Math.max(25, base - (factor - 1) * 10)) * 100) / 100
  }
  return Math.round(Math.max(1, base * factor) * 100) / 100
}

function computeSampleConfidence(dMin: number): number {
  if (dMin <= 400) return 1
  if (dMin >= 600) return 0
  return Math.max(0, Math.min(1, 1 - (dMin - 400) / 200))
}

function evaluateFixtureCell(
  lat: number,
  lon: number,
  layer: LayerKey,
  standard: AirQualityStandard,
  factor: number,
): HeatmapCell {
  const cosLat = Math.cos((lat * Math.PI) / 180)
  let dMin = Infinity
  let nearestVal = 0
  let weightSum = 0
  let valueSum = 0
  let confSum = 0
  let dirWeightSum = 0
  let dirX = 0
  let dirY = 0
  let exact: number | null = null

  for (const st of SAMPLE_STATIONS) {
    const d = distanceMeters(lat, lon, st.node.lat, st.node.lon)
    if (d > 600) continue
    const val = stationSampleValue(st, layer, standard, factor)
    if (d < dMin) {
      dMin = d
      nearestVal = val
    }
    confSum += computeSampleConfidence(d)

    const dy = (st.node.lat - lat) * 111320
    const dx = (st.node.lon - lon) * 111320 * cosLat
    const norm = Math.hypot(dx, dy)
    const dirW = 1 - d / 600
    if (dirW > 0 && norm > 1e-6) {
      dirWeightSum += dirW
      dirX += dirW * (dx / norm)
      dirY += dirW * (dy / norm)
    }

    if (d < 1) {
      exact = val
    } else if (d < 600) {
      const taper = (600 - d) / (600 * d)
      const w = taper * taper
      weightSum += w
      valueSum += w * val
    }
  }

  if (!Number.isFinite(dMin)) {
    return {
      lat: Math.round(lat * 1e6) / 1e6,
      lon: Math.round(lon * 1e6) / 1e6,
      value: null,
      category: NO_DATA_LABEL,
      color: NO_DATA_COLOR,
      confidence: 0,
      has_data: false,
    }
  }

  const raw = exact !== null ? exact : weightSum > 0 ? valueSum / weightSum : nearestVal
  const value = Math.round(raw * 100) / 100
  const baseConf = computeSampleConfidence(dMin)
  let rawConf = baseConf
  if (exact === null && dirWeightSum > 0 && baseConf < 1 && confSum > 0) {
    const rx = dirX / dirWeightSum
    const ry = dirY / dirWeightSum
    const rSq = Math.min(1, rx * rx + ry * ry)
    const interiority = 1 - rSq
    const supportEnvelope = Math.min(1, confSum / 0.35)
    rawConf = Math.min(1, baseConf + interiority * supportEnvelope * (1 - baseConf))
  }
  const confidence = Math.round(rawConf * 10000) / 10000
  const styled = pickSampleStyle(layer, standard, value)
  const continuousColor = interpolateSampleColor(layer, standard, value)
  return {
    lat: Math.round(lat * 1e6) / 1e6,
    lon: Math.round(lon * 1e6) / 1e6,
    value,
    category: styled.category,
    color: blendHexToGray(continuousColor, rawConf),
    confidence,
    has_data: true,
  }
}

export function getDemoPublicNodes(): PublicNodesResponse {
  return {
    nodes: ensureDemoNodes()
      .filter((node) => node.status === 'active')
      .map(({ id, code, name, lat, lon, status }) => ({ id, code, name, lat, lon, status })),
  }
}

export function getDemoPlaybackTimestamps(): PlaybackTimestampsResponse {
  return {
    timestamps: [...SAMPLE_TIMESTAMPS],
  }
}

export function getDemoHeatmap(
  layer: LayerKey,
  standard: AirQualityStandard,
  at?: string,
): HeatmapResponse {
  const factor = sampleFactorForTimestamp(at)
  const recordedAt = at ?? SAMPLE_TIMESTAMPS[SAMPLE_TIMESTAMPS.length - 1] ?? new Date().toISOString()

  // Local patches around Trujillo Centro and Huanchaco in coherent 50m lattice cells (includes >600m gray rim)
  const sectors = [
    { minLat: -8.128, maxLat: -8.0985, minLon: -79.048, maxLon: -79.0205 },
    { minLat: -8.087, maxLat: -8.0715, minLon: -79.1285, maxLon: -79.1125 },
  ]

  const originLat = -8.128
  const originLon = -79.1285
  const dLat = 50 / 111320
  const dLon = 50 / (111320 * Math.cos((-8.1 * Math.PI) / 180))
  const cells: HeatmapCell[] = []

  for (const box of sectors) {
    const startR = Math.max(0, Math.floor((box.minLat - originLat) / dLat))
    const endR = Math.ceil((box.maxLat - originLat) / dLat)
    const startC = Math.max(0, Math.floor((box.minLon - originLon) / dLon))
    const endC = Math.ceil((box.maxLon - originLon) / dLon)

    for (let r = startR; r <= endR; r++) {
      const lat = originLat + (r + 0.5) * dLat
      if (lat < box.minLat || lat > box.maxLat) continue
      for (let c = startC; c <= endC; c++) {
        const lon = originLon + (c + 0.5) * dLon
        if (lon < box.minLon || lon > box.maxLon) continue
        let near = Infinity
        for (const st of SAMPLE_STATIONS) {
          const d = distanceMeters(lat, lon, st.node.lat, st.node.lon)
          if (d < near) near = d
        }
        // Include full coverage + fade + an honest outer #9CA3AF band up to 800m
        if (near <= 820) {
          cells.push(evaluateFixtureCell(lat, lon, layer, standard, factor))
        }
      }
    }
  }

  return {
    recorded_at: recordedAt,
    layer,
    standard,
    unit: layer === LAYER_AIR_QUALITY ? (standard === 'inca' ? 'INCA' : 'AQI') : METRIC_SPECS[layer].unit,
    cell_size_m: 50,
    bounds: {
      min_lat: -8.128,
      min_lon: -79.1285,
      max_lat: -8.0715,
      max_lon: -79.0205,
    },
    cells,
  }
}

export function getDemoPoint(
  lat: number,
  lon: number,
  layer: LayerKey,
  standard: AirQualityStandard,
  at?: string,
): PointResponse {
  const unit = layer === LAYER_AIR_QUALITY ? (standard === 'inca' ? 'INCA' : 'AQI') : METRIC_SPECS[layer].unit
  const factor = sampleFactorForTimestamp(at)
  const activeTimestamp = at ?? SAMPLE_TIMESTAMPS[SAMPLE_TIMESTAMPS.length - 1] ?? new Date().toISOString()
  const cell = evaluateFixtureCell(lat, lon, layer, standard, factor)

  if (!cell.has_data || cell.value === null) {
    return {
      lat,
      lon,
      layer,
      standard,
      unit,
      has_data: false,
      label: NO_DATA_LABEL,
      current: null,
      series: [],
    }
  }

  const series: PointSeriesItem[] = SAMPLE_TIMESTAMPS.map((ts, index) => {
    // Demonstrate an honest brief gap near early morning in demo fixtures without substituting 0
    if (index >= 48 && index <= 53) {
      return { recorded_at: ts, value: null, has_data: false }
    }
    const f = sampleFactorForTimestamp(ts)
    const pt = evaluateFixtureCell(lat, lon, layer, standard, f)
    return {
      recorded_at: ts,
      value: pt.value,
      has_data: pt.has_data,
    }
  })

  return {
    lat,
    lon,
    layer,
    standard,
    unit,
    has_data: true,
    label: cell.category,
    current: {
      recorded_at: activeTimestamp,
      value: cell.value,
      category: cell.category,
      color: cell.color,
      confidence: cell.confidence,
    },
    series,
  }
}

let demoNodesStore: Node[] | null = null
const demoReadingsStore = new Map<string, AdminReadingItem[]>()
let demoSequence = 1

function buildInitialStationReadings(station: SampleStation): AdminReadingItem[] {
  return SAMPLE_TIMESTAMPS.map((ts, index) => {
    const factor = sampleFactorForTimestamp(ts)
    const item: AdminReadingItem = { recorded_at: ts }
    for (const metric of METRIC_KEYS) {
      // Vary slight phase per index while staying within realistic station baseline
      const micro = 1 + 0.015 * Math.sin(index * 0.4 + metric.length)
      item[metric] = Math.round(stationSampleValue(station, metric, 'inca', factor) * micro * 100) / 100
    }
    return item
  })
}

function ensureDemoNodes(): Node[] {
  if (demoNodesStore !== null) {
    return demoNodesStore
  }
  const baseCreated = SAMPLE_TIMESTAMPS[0] ?? new Date().toISOString()
  const baseUpdated = SAMPLE_TIMESTAMPS[SAMPLE_TIMESTAMPS.length - 1] ?? baseCreated

  demoNodesStore = SAMPLE_STATIONS.map((station) => {
    demoReadingsStore.set(station.node.id, buildInitialStationReadings(station))
    return {
      ...station.node,
      is_virtual: false,
      created_at: baseCreated,
      updated_at: baseUpdated,
    }
  })

  // Seed one virtual demonstration station (active near Chan Chan / Mansiche) so virtual CRUD is self-evident
  const virtualId = 'demo-trj-v1'
  demoNodesStore.push({
    id: virtualId,
    code: 'TRJ-V-01',
    name: 'Av. Mansiche · Nodo Virtual',
    lat: -8.1018,
    lon: -79.0442,
    status: 'inactive',
    is_virtual: true,
    created_at: baseCreated,
    updated_at: baseUpdated,
  })
  demoReadingsStore.set(virtualId, [])

  return demoNodesStore
}

export function getDemoAdminNodes(): AdminNodesResponse {
  return {
    nodes: ensureDemoNodes().map((node) => ({ ...node })),
  }
}

export function createDemoAdminNode(req: CreateNodeRequest): Node {
  const nodes = ensureDemoNodes()
  const nowIso = new Date().toISOString()
  const suffix = String(demoSequence++).padStart(2, '0')
  const trimmedCode = req.code.trim()
  const code = trimmedCode.length > 0 ? trimmedCode : `TRJ-V-${suffix}`
  const id = `demo-node-${Date.now()}-${demoSequence}`

  const created: Node = {
    id,
    code,
    name: req.name.trim(),
    lat: Math.round(req.lat * 1e6) / 1e6,
    lon: Math.round(req.lon * 1e6) / 1e6,
    status: req.status,
    is_virtual: true,
    created_at: nowIso,
    updated_at: nowIso,
  }
  nodes.push(created)
  demoReadingsStore.set(id, [])
  return { ...created }
}

export function updateDemoAdminNode(nodeId: string, req: UpdateNodeRequest): Node | null {
  const nodes = ensureDemoNodes()
  const target = nodes.find((node) => node.id === nodeId)
  if (!target) {
    return null
  }
  if (req.name !== undefined) {
    target.name = req.name.trim()
  }
  if (req.lat !== undefined) {
    target.lat = Math.round(req.lat * 1e6) / 1e6
  }
  if (req.lon !== undefined) {
    target.lon = Math.round(req.lon * 1e6) / 1e6
  }
  if (req.status !== undefined) {
    target.status = req.status
  }
  target.updated_at = new Date().toISOString()
  return { ...target }
}

export function deleteDemoAdminNode(nodeId: string): boolean {
  const nodes = ensureDemoNodes()
  const idx = nodes.findIndex((node) => node.id === nodeId)
  if (idx === -1) {
    return false
  }
  nodes.splice(idx, 1)
  demoReadingsStore.delete(nodeId)
  return true
}

export function getDemoAdminNodeReadings(nodeId: string, hours = 24): AdminReadingsResponse {
  ensureDemoNodes()
  const all = demoReadingsStore.get(nodeId) ?? []
  const latestMs =
    all.length > 0 ? Math.max(...all.map((item) => Date.parse(item.recorded_at))) : Date.now()
  const cutoffMs = latestMs - Math.max(1, hours) * 3600_000
  const filtered = all
    .filter((item) => Date.parse(item.recorded_at) >= cutoffMs)
    .sort((a, b) => Date.parse(a.recorded_at) - Date.parse(b.recorded_at))
    .map((item) => ({ ...item }))

  return {
    node_id: nodeId,
    readings: filtered,
  }
}

export function upsertDemoNodeReadingsBatch(
  nodeId: string,
  req: BatchReadingsRequest,
): BatchReadingsResponse {
  ensureDemoNodes()
  const startMs = Date.parse(req.start_time)
  const stepMs = 5 * 60 * 1000
  const byTime = new Map<number, AdminReadingItem>()

  for (const existing of demoReadingsStore.get(nodeId) ?? []) {
    const ms = Date.parse(existing.recorded_at)
    if (!Number.isNaN(ms)) {
      byTime.set(ms, { ...existing })
    }
  }

  let touchedSlots = 0
  let minMs = startMs
  let maxMs = startMs

  for (const key of METRIC_KEYS) {
    const values = req.metrics[key]
    if (!values || values.length === 0) continue
    for (let i = 0; i < values.length; i += 1) {
      const slotMs = startMs - i * stepMs
      if (slotMs < minMs) minMs = slotMs
      if (slotMs > maxMs) maxMs = slotMs
      const slotIso = new Date(slotMs).toISOString()
      const current = byTime.get(slotMs) ?? { recorded_at: slotIso }
      current[key] = values[i]
      byTime.set(slotMs, current)
      touchedSlots += 1
    }
  }

  const merged = [...byTime.values()].sort(
    (a, b) => Date.parse(a.recorded_at) - Date.parse(b.recorded_at),
  )
  demoReadingsStore.set(nodeId, merged)

  return {
    inserted_or_updated: touchedSlots,
    from: new Date(minMs).toISOString(),
    to: new Date(maxMs).toISOString(),
  }
}
