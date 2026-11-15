export const LAYER_AIR_QUALITY = 'air_quality' as const

export const METRIC_KEYS = [
  'pm25',
  'pm10',
  'co',
  'no2',
  'so2',
  'h2s',
  'co2',
  'voc_index',
  'noise_dba',
  'temperature_c',
  'humidity_pct',
  'pressure_hpa',
] as const

export type MetricKey = (typeof METRIC_KEYS)[number]
export type LayerKey = typeof LAYER_AIR_QUALITY | MetricKey
export type AirQualityStandard = 'inca' | 'epa'
export type NodeStatus = 'active' | 'inactive'
export type TelemetrySource = 'mqtt' | 'admin_sim'

export const NO_DATA_COLOR = '#9CA3AF' as const
export const NO_DATA_LABEL = 'No hay datos' as const

export interface MetricSpec {
  key: MetricKey
  label: string
  unit: string
  min: number
  max: number
  sensor: string
}

export const METRIC_SPECS: Record<MetricKey, MetricSpec> = {
  pm25: { key: 'pm25', label: 'PM2.5', unit: 'µg/m³', min: 0, max: 1000, sensor: 'Sensirion SPS30' },
  pm10: { key: 'pm10', label: 'PM10', unit: 'µg/m³', min: 0, max: 1000, sensor: 'Sensirion SPS30' },
  co: { key: 'co', label: 'CO', unit: 'mg/m³', min: 0, max: 100, sensor: 'Winsen ZE03-CO' },
  no2: { key: 'no2', label: 'NO2', unit: 'µg/m³', min: 0, max: 2000, sensor: 'Winsen ZE03-NO2' },
  so2: { key: 'so2', label: 'SO2', unit: 'µg/m³', min: 0, max: 2000, sensor: 'Winsen ZE03-SO2' },
  h2s: { key: 'h2s', label: 'H2S', unit: 'µg/m³', min: 0, max: 2000, sensor: 'Winsen ZE03-H2S' },
  co2: { key: 'co2', label: 'CO2', unit: 'ppm', min: 400, max: 10000, sensor: 'Sensirion SCD41' },
  voc_index: { key: 'voc_index', label: 'Índice VOC', unit: 'index', min: 1, max: 500, sensor: 'Sensirion SGP40' },
  noise_dba: { key: 'noise_dba', label: 'Ruido', unit: 'dBA', min: 20, max: 130, sensor: 'InvenSense ICS-43434' },
  temperature_c: { key: 'temperature_c', label: 'Temperatura', unit: '°C', min: -10, max: 60, sensor: 'Bosch BME280' },
  humidity_pct: { key: 'humidity_pct', label: 'Humedad', unit: '%', min: 0, max: 100, sensor: 'Bosch BME280' },
  pressure_hpa: { key: 'pressure_hpa', label: 'Presión', unit: 'hPa', min: 800, max: 1100, sensor: 'Bosch BME280' },
}

export interface ErrorResponse {
  error: string
  message: string
}

export interface HealthResponse {
  status: string
}

export interface Node {
  id: string
  code: string
  name: string
  lat: number
  lon: number
  status: NodeStatus
  is_virtual: boolean
  created_at: string
  updated_at: string
}

export interface PublicNode {
  id: string
  code: string
  name: string
  lat: number
  lon: number
  status: NodeStatus
}

export interface AdminNodesResponse {
  nodes: Node[]
}

export interface PublicNodesResponse {
  nodes: PublicNode[]
}

export interface CreateNodeRequest {
  code: string
  name: string
  lat: number
  lon: number
  status: NodeStatus
}

export interface UpdateNodeRequest {
  name?: string
  lat?: number
  lon?: number
  status?: NodeStatus
}

export interface TelemetryMetrics {
  pm25?: number
  pm10?: number
  co?: number
  no2?: number
  so2?: number
  h2s?: number
  co2?: number
  voc_index?: number
  noise_dba?: number
  temperature_c?: number
  humidity_pct?: number
  pressure_hpa?: number
}

export interface MQTTTelemetryPayload {
  node_code: string
  recorded_at: string
  metrics: TelemetryMetrics
}

export interface AdminReadingItem extends TelemetryMetrics {
  recorded_at: string
}

export interface AdminReadingsResponse {
  node_id: string
  readings: AdminReadingItem[]
}

export interface Bounds {
  min_lat: number
  min_lon: number
  max_lat: number
  max_lon: number
}

export interface HeatmapCell {
  lat: number
  lon: number
  value: number | null
  category: string
  color: string
  confidence: number
  has_data: boolean
}

export interface HeatmapResponse {
  recorded_at: string | null
  layer: LayerKey
  standard: AirQualityStandard
  unit: string
  cell_size_m: number
  bounds: Bounds | null
  cells: HeatmapCell[]
}

export interface PointCurrent {
  recorded_at: string
  value: number
  category: string
  color: string
  confidence: number
}

export interface PointSeriesItem {
  recorded_at: string
  value: number | null
  has_data: boolean
}

export interface PointResponse {
  lat: number
  lon: number
  layer: LayerKey
  standard: AirQualityStandard
  unit: string
  has_data: boolean
  label: string
  current: PointCurrent | null
  series: PointSeriesItem[]
}

export interface PlaybackTimestampsResponse {
  timestamps: string[]
}

export interface AccessVerifyRequest {
  phrase: string
}

export interface AccessResponse {
  ok: boolean
}

export interface BatchReadingsRequest {
  start_time: string
  metrics: Partial<Record<MetricKey, number[]>>
}

export interface BatchReadingsResponse {
  inserted_or_updated: number
  from: string
  to: string
}
