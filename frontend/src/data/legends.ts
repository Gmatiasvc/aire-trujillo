import {
  LAYER_AIR_QUALITY,
  METRIC_KEYS,
  METRIC_SPECS,
  NO_DATA_COLOR,
  NO_DATA_LABEL,
  type AirQualityStandard,
  type LayerKey,
  type MetricKey,
} from '../types/telemetry'

export interface LegendBand {
  color: string
  label: string
  range: string
}

export interface LegendSpec {
  title: string
  subtitle: string
  unit: string
  bands: LegendBand[]
  noDataColor: string
  noDataLabel: string
}

const INCA_BANDS: LegendBand[] = [
  { color: '#22C55E', label: 'Buena', range: '0 – 50' },
  { color: '#EAB308', label: 'Moderada', range: '51 – 100' },
  { color: '#F97316', label: 'Mala', range: '101 – 150' },
  { color: '#EF4444', label: 'Peligrosa', range: '> 150' },
]

const EPA_BANDS: LegendBand[] = [
  { color: '#00E400', label: 'Good', range: '0 – 50' },
  { color: '#FFFF00', label: 'Moderate', range: '51 – 100' },
  { color: '#FF7E00', label: 'Unhealthy (Sens.)', range: '101 – 150' },
  { color: '#FF0000', label: 'Unhealthy', range: '151 – 200' },
  { color: '#8F3F97', label: 'Very Unhealthy', range: '201 – 300' },
  { color: '#7E0023', label: 'Hazardous', range: '301 – 500' },
]

function fourTierBands(t1: number, t2: number, t3: number, unit: string): LegendBand[] {
  return [
    { color: '#22C55E', label: 'Bueno', range: `≤ ${t1} ${unit}` },
    { color: '#EAB308', label: 'Moderado', range: `${t1} – ${t2} ${unit}` },
    { color: '#F97316', label: 'Alto', range: `${t2} – ${t3} ${unit}` },
    { color: '#EF4444', label: 'Muy alto', range: `> ${t3} ${unit}` },
  ]
}

const METRIC_BANDS: Record<MetricKey, LegendBand[]> = {
  pm25: fourTierBands(25, 50, 100, 'µg/m³'),
  pm10: fourTierBands(50, 100, 200, 'µg/m³'),
  co: fourTierBands(10, 30, 45, 'mg/m³'),
  no2: fourTierBands(100, 200, 400, 'µg/m³'),
  so2: fourTierBands(125, 250, 500, 'µg/m³'),
  h2s: fourTierBands(75, 150, 300, 'µg/m³'),
  co2: [
    { color: '#22C55E', label: 'Normal', range: '≤ 1000 ppm' },
    { color: '#EAB308', label: 'Elevado', range: '1001 – 2000 ppm' },
    { color: '#EF4444', label: 'Alto', range: '> 2000 ppm' },
  ],
  voc_index: fourTierBands(100, 200, 400, 'idx'),
  noise_dba: [
    { color: '#22C55E', label: 'Confortable', range: '≤ 60 dBA' },
    { color: '#EAB308', label: 'Moderado', range: '61 – 80 dBA' },
    { color: '#EF4444', label: 'Alto', range: '> 80 dBA' },
  ],
  temperature_c: [
    { color: '#3B82F6', label: 'Frío', range: '< 16 °C' },
    { color: '#22C55E', label: 'Templado', range: '16 – 26 °C' },
    { color: '#F97316', label: 'Cálido', range: '26 – 32 °C' },
    { color: '#EF4444', label: 'Caluroso', range: '> 32 °C' },
  ],
  humidity_pct: [
    { color: '#EAB308', label: 'Seco', range: '< 30 %' },
    { color: '#22C55E', label: 'Confortable', range: '30 – 75 %' },
    { color: '#3B82F6', label: 'Húmedo', range: '> 75 %' },
  ],
  pressure_hpa: [
    { color: '#3B82F6', label: 'Baja', range: '< 1005 hPa' },
    { color: '#22C55E', label: 'Normal', range: '1005 – 1020 hPa' },
    { color: '#F97316', label: 'Alta', range: '> 1020 hPa' },
  ],
}

export function getLegendSpec(layer: LayerKey, standard: AirQualityStandard): LegendSpec {
  if (layer === LAYER_AIR_QUALITY) {
    return {
      title: standard === 'inca' ? 'ECA / INCA (Perú)' : 'AQI (US EPA)',
      subtitle: 'Índice compuesto máx. por contaminante criterio',
      unit: standard === 'inca' ? 'INCA' : 'AQI',
      bands: standard === 'inca' ? INCA_BANDS : EPA_BANDS,
      noDataColor: NO_DATA_COLOR,
      noDataLabel: `Sin cobertura / ${NO_DATA_LABEL}`,
    }
  }

  const spec = METRIC_SPECS[layer]
  return {
    title: spec.label,
    subtitle: `Capa individual · ${spec.sensor}`,
    unit: spec.unit,
    bands: METRIC_BANDS[layer],
    noDataColor: NO_DATA_COLOR,
    noDataLabel: `Sin cobertura / ${NO_DATA_LABEL}`,
  }
}

export interface LayerOption {
  key: LayerKey
  label: string
  unit: string
}

export const INDIVIDUAL_LAYER_OPTIONS: LayerOption[] = METRIC_KEYS.map((key) => ({
  key,
  label: METRIC_SPECS[key].label,
  unit: METRIC_SPECS[key].unit,
}))

const limaTimeFormatter = new Intl.DateTimeFormat('es-PE', {
  timeZone: 'America/Lima',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
})

const limaDateTimeFormatter = new Intl.DateTimeFormat('es-PE', {
  timeZone: 'America/Lima',
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
})

export function formatLimaTime(iso: string | null | undefined): string {
  if (!iso) {
    return '--:--'
  }
  const parsed = new Date(iso)
  if (Number.isNaN(parsed.getTime())) {
    return '--:--'
  }
  return limaTimeFormatter.format(parsed)
}

export function formatLimaDateTime(iso: string | null | undefined): string {
  if (!iso) {
    return 'En vivo · America/Lima'
  }
  const parsed = new Date(iso)
  if (Number.isNaN(parsed.getTime())) {
    return iso
  }
  return `${limaDateTimeFormatter.format(parsed)} UTC-5`
}
