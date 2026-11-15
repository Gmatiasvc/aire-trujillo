import type { AirQualityStandard, MetricKey } from '../types/telemetry'
import { NO_DATA_COLOR, NO_DATA_LABEL } from '../types/telemetry'

export interface DemoStationRow {
  code: string
  name: string
  sector: string
  distanceLabel: string
  confidence: number
  hasData: boolean
  recordedAt: string
  incaValue: number | null
  incaCategory: string
  incaColor: string
  epaValue: number | null
  epaCategory: string
  epaColor: string
  metrics: Partial<Record<MetricKey, number>>
}

export const DEMO_STATION_ROWS: DemoStationRow[] = [
  {
    code: 'TRJ-01',
    name: 'Plaza de Armas',
    sector: 'Centro Histórico',
    distanceLabel: 'd = 120 m',
    confidence: 1.0,
    hasData: true,
    recordedAt: '21:15 UTC-5',
    incaValue: 42,
    incaCategory: 'Buena',
    incaColor: '#22C55E',
    epaValue: 56,
    epaCategory: 'Moderate',
    epaColor: '#EAB308',
    metrics: {
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
    code: 'TRJ-02',
    name: 'Óvalo Papal',
    sector: 'Av. América Norte',
    distanceLabel: 'd = 470 m',
    confidence: 0.65,
    hasData: true,
    recordedAt: '21:15 UTC-5',
    incaValue: 64,
    incaCategory: 'Moderada',
    incaColor: '#EAB308',
    epaValue: 92,
    epaCategory: 'Moderate',
    epaColor: '#EAB308',
    metrics: {
      pm25: 31.8,
      pm10: 62.4,
      co: 2.4,
      no2: 68.5,
      so2: 19.6,
      h2s: 8.2,
      co2: 512,
      voc_index: 164,
      noise_dba: 72.1,
      temperature_c: 22.1,
      humidity_pct: 74.5,
      pressure_hpa: 1012.9,
    },
  },
  {
    code: 'HCH-01',
    name: 'Muelle Huanchaco',
    sector: 'Borde Costero',
    distanceLabel: 'd = 210 m',
    confidence: 1.0,
    hasData: true,
    recordedAt: '21:15 UTC-5',
    incaValue: 31,
    incaCategory: 'Buena',
    incaColor: '#22C55E',
    epaValue: 41,
    epaCategory: 'Good',
    epaColor: '#00E400',
    metrics: {
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
  {
    code: 'SIN-COB',
    name: 'Periferia sin nodo',
    sector: 'Fuera de R_max',
    distanceLabel: 'd = 680 m',
    confidence: 0.0,
    hasData: false,
    recordedAt: '21:15 UTC-5',
    incaValue: null,
    incaCategory: NO_DATA_LABEL,
    incaColor: NO_DATA_COLOR,
    epaValue: null,
    epaCategory: NO_DATA_LABEL,
    epaColor: NO_DATA_COLOR,
    metrics: {},
  },
]

export interface CoverageCheckpoint {
  id: string
  distanceM: number
  confidence: number
  hasData: boolean
  zone: 'full' | 'fade' | 'nodata'
  categoryLabel: string
  valueDisplay: string
  hexColor: string
  note: string
}

export const COVERAGE_CHECKPOINTS: CoverageCheckpoint[] = [
  {
    id: 'd-0',
    distanceM: 0,
    confidence: 1.0,
    hasData: true,
    zone: 'full',
    categoryLabel: 'Buena (lectura directa)',
    valueDisplay: '42 INCA',
    hexColor: '#22C55E',
    note: 'Con d < 1 m sobre el poste activo, el backend adopta el valor exacto de la estación.',
  },
  {
    id: 'd-300',
    distanceM: 300,
    confidence: 1.0,
    hasData: true,
    zone: 'full',
    categoryLabel: 'Buena (cobertura plena)',
    valueDisplay: '42 INCA',
    hexColor: '#22C55E',
    note: 'Dentro del radio R_fade = 400 m la celda de 50 m mantiene confianza máxima (1.00).',
  },
  {
    id: 'd-450',
    distanceM: 450,
    confidence: 0.75,
    hasData: true,
    zone: 'fade',
    categoryLabel: 'Buena (borde 400–600 m)',
    valueDisplay: '42 INCA',
    hexColor: '#40B570',
    note: 'A 450 m el backend inicia la mezcla lineal en RGB hacia #9CA3AF (confianza 0.75).',
  },
  {
    id: 'd-550',
    distanceM: 550,
    confidence: 0.25,
    hasData: true,
    zone: 'fade',
    categoryLabel: 'Buena (límite cercano)',
    valueDisplay: '42 INCA',
    hexColor: '#7D9B9F',
    note: 'Cerca de los 600 m el color se apaga visualmente para advertir menor respaldo espacial.',
  },
  {
    id: 'd-700',
    distanceM: 700,
    confidence: 0.0,
    hasData: false,
    zone: 'nodata',
    categoryLabel: NO_DATA_LABEL,
    valueDisplay: NO_DATA_LABEL,
    hexColor: NO_DATA_COLOR,
    note: 'Superado R_max = 600 m: has_data = false, value = null, color = #9CA3AF. Prohibido devolver 0.',
  },
]

export function formatStandardTitle(std: AirQualityStandard): string {
  return std === 'inca' ? 'ECA / INCA (Perú)' : 'AQI (US EPA)'
}
